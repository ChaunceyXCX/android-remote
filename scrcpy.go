package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type ScrcpyState int

const (
	StateIdle ScrcpyState = iota
	StateStarting
	StateRunning
	StateStopping
	StateError
)

func (s ScrcpyState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

type ScrcpyClient struct {
	deviceID string
	adbPath  string
	port     int

	serverCmd   *exec.Cmd
	videoConn   net.Conn
	controlConn net.Conn

	streamHeader VideoStreamHeader
	videoHub     *VideoHub

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu      sync.Mutex
	state   ScrcpyState
	stateMu sync.RWMutex
}

func NewScrcpyClient(deviceID, adbPath string, hub *VideoHub) *ScrcpyClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &ScrcpyClient{
		deviceID: deviceID,
		adbPath:  adbPath,
		videoHub: hub,
		ctx:      ctx,
		cancel:   cancel,
		state:    StateIdle,
	}
}

func (c *ScrcpyClient) State() ScrcpyState {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state
}

func (c *ScrcpyClient) setState(s ScrcpyState) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	c.state = s
}

func (c *ScrcpyClient) StreamHeader() VideoStreamHeader {
	return c.streamHeader
}

func (c *ScrcpyClient) Start() error {
	if c.State() != StateIdle && c.State() != StateError {
		return fmt.Errorf("无法启动：当前状态 %s", c.State())
	}

	c.setState(StateStarting)
	log.Printf("[ScrcpyClient] 启动 scrcpy 连接，设备: %s", c.deviceID)

	if err := c.pushScrcpyServer(); err != nil {
		c.setState(StateError)
		return fmt.Errorf("推送 scrcpy-server 失败: %w", err)
	}

	// 关键：先启动服务器（tunnel_forward=true 时服务器在设备上创建 abstract socket 并监听）
	if err := c.launchServer(); err != nil {
		c.setState(StateError)
		return fmt.Errorf("启动 scrcpy-server 失败: %w", err)
	}

	// 设置 adb forward 桥接：host TCP → device localabstract:scrcpy
	if err := c.setupPortForward(); err != nil {
		c.killServer()
		c.setState(StateError)
		return fmt.Errorf("设置端口转发失败: %w", err)
	}

	// 作为客户端连接到服务器（通过 adb forward）
	if err := c.connectVideo(); err != nil {
		c.killServer()
		c.cleanupPortForward()
		c.setState(StateError)
		return fmt.Errorf("建立视频连接失败: %w", err)
	}

	c.wg.Add(1)
	go c.readLoop()

	c.setState(StateRunning)
	log.Printf("[ScrcpyClient] scrcpy 连接已建立，分辨率: %dx%d", c.streamHeader.Width, c.streamHeader.Height)
	return nil
}

// Stop 停止 scrcpy 连接并清理资源
func (c *ScrcpyClient) Stop() {
	if c.State() == StateIdle || c.State() == StateStopping {
		return
	}

	c.setState(StateStopping)
	log.Printf("[ScrcpyClient] 停止 scrcpy 连接，设备: %s", c.deviceID)

	c.cancel()

	if c.videoConn != nil {
		c.videoConn.Close()
	}

	// 3. 关闭控制连接
	if c.controlConn != nil {
		c.controlConn.Close()
	}

	c.killServer()

	c.cleanupPortForward()

	c.wg.Wait()

	c.setState(StateIdle)
	log.Printf("[ScrcpyClient] scrcpy 连接已停止，设备: %s", c.deviceID)
}

func (c *ScrcpyClient) ReadPacket() (*VideoPacket, error) {
	return ReadFullPacket(c.videoConn)
}

func (c *ScrcpyClient) pushScrcpyServer() error {
	tmpFile := filepath.Join(os.TempDir(), "scrcpy-server.jar")
	if err := os.WriteFile(tmpFile, scrcpyServerJar, 0644); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile)

	cmd := exec.Command(c.adbPath, "-s", c.deviceID, "push", tmpFile, "/data/local/tmp/scrcpy-server.jar")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb push 失败: %s, %w", string(output), err)
	}

	log.Printf("[ScrcpyClient] scrcpy-server.jar 已推送到设备 %s", c.deviceID)
	return nil
}

func (c *ScrcpyClient) setupPortForward() error {
	c.port = 27183

	exec.Command(c.adbPath, "-s", c.deviceID, "forward", "--remove", fmt.Sprintf("tcp:%d", c.port)).Run()

	cmd := exec.Command(c.adbPath, "-s", c.deviceID, "forward", fmt.Sprintf("tcp:%d", c.port), "localabstract:scrcpy")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb forward 失败: %s, %w", string(output), err)
	}

	log.Printf("[ScrcpyClient] adb forward 已设置: tcp:%d → localabstract:scrcpy", c.port)
	return nil
}

// launchServer 在设备上启动 scrcpy-server
func (c *ScrcpyClient) launchServer() error {
	// scrcpy-server 启动参数
	args := []string{
		"-s", c.deviceID,
		"shell",
		"CLASSPATH=/data/local/tmp/scrcpy-server.jar",
		"app_process", "/", "com.genymobile.scrcpy.Server",
		ScrcpyVersion,
		"video_codec=h264",
		"video_bit_rate=8000000",
		"max_fps=30",
		"video_source=display",
		"send_frame_meta=true",
		"tunnel_forward=true",
		"control=false",
		"audio=false",
	}

	c.serverCmd = exec.Command(c.adbPath, args...)
	c.serverCmd.Stdout = os.Stdout
	c.serverCmd.Stderr = os.Stderr

	if err := c.serverCmd.Start(); err != nil {
		return fmt.Errorf("启动 scrcpy-server 失败: %w", err)
	}

	log.Printf("[ScrcpyClient] scrcpy-server 已启动，PID: %d", c.serverCmd.Process.Pid)

	time.Sleep(1 * time.Second)

	return nil
}

// connectVideo 作为客户端连接到 scrcpy-server（通过 adb forward）
// tunnel_forward=true 时服务器在设备上创建 abstract socket 并监听
// 我们通过 adb forward tcp:PORT localabstract:scrcpy 桥接后连接
func (c *ScrcpyClient) connectVideo() error {
	maxRetries := 20
	delay := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			log.Printf("[ScrcpyClient] 重试连接 scrcpy-server，第 %d/%d 次", i+1, maxRetries)
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * 1.3)
			if delay > 3*time.Second {
				delay = 3 * time.Second
			}
		}

		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", c.port), 5*time.Second)
		if err != nil {
			log.Printf("[ScrcpyClient] 连接失败: %v", err)
			continue
		}

		// 读取 dummy byte（服务器发送的连接确认）
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		dummy := make([]byte, 1)
		if _, err := io.ReadFull(conn, dummy); err != nil {
			conn.Close()
			log.Printf("[ScrcpyClient] 读取 dummy byte 失败: %v", err)
			continue
		}

		// 读取设备信息（64B name，宽高在 H264 SPS/PPS 中）
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		header, err := c.readStreamHeaderFrom(conn)
		conn.SetReadDeadline(time.Time{})
		if err != nil {
			conn.Close()
			log.Printf("[ScrcpyClient] 读取设备信息失败: %v", err)
			continue
		}

		c.videoConn = conn
		c.streamHeader = header
		log.Printf("[ScrcpyClient] 已连接到 scrcpy-server: %s %dx%d", header.DeviceName, header.Width, header.Height)
		return nil
	}

	return fmt.Errorf("连接 scrcpy-server 失败，已重试 %d 次", maxRetries)
}

func (c *ScrcpyClient) readStreamHeader() (VideoStreamHeader, error) {
	return c.readStreamHeaderFrom(c.videoConn)
}

func (c *ScrcpyClient) readStreamHeaderFrom(conn net.Conn) (VideoStreamHeader, error) {
	nameBuf := make([]byte, DeviceNameLen)
	if _, err := io.ReadFull(conn, nameBuf); err != nil {
		return VideoStreamHeader{}, fmt.Errorf("读取设备名失败: %w", err)
	}

	name := string(nameBuf)
	for i, b := range nameBuf {
		if b == 0 {
			name = string(nameBuf[:i])
			break
		}
	}

	return VideoStreamHeader{
		DeviceName: name,
		Width:      0,
		Height:     0,
	}, nil
}

func (c *ScrcpyClient) readLoop() {
	defer c.wg.Done()

	log.Printf("[ScrcpyClient] 开始读取视频流...")
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		packet, err := c.ReadPacket()
		if err != nil {
			if c.ctx.Err() != nil {
				return
			}
			log.Printf("[ScrcpyClient] 读取视频包失败: %v", err)
			c.setState(StateError)
			return
		}

		if c.videoHub != nil {
			c.videoHub.Broadcast(packet)
		}

		if packet.Type == PacketEOS {
			log.Printf("[ScrcpyClient] 收到 EOS 包，结束读取")
			return
		}
	}
}

func (c *ScrcpyClient) killServer() {
	if c.serverCmd == nil || c.serverCmd.Process == nil {
		return
	}

	c.serverCmd.Process.Signal(os.Interrupt)

	done := make(chan error, 1)
	go func() {
		done <- c.serverCmd.Wait()
	}()

	select {
	case <-done:
		log.Printf("[ScrcpyClient] scrcpy-server 已正常退出")
	case <-time.After(2 * time.Second):
		c.serverCmd.Process.Kill()
		log.Printf("[ScrcpyClient] scrcpy-server 已被强制杀死")
	}

	cleanupCmd := exec.Command(c.adbPath, "-s", c.deviceID, "shell", "pkill", "-f", "scrcpy.Server")
	cleanupCmd.Run()
}

func (c *ScrcpyClient) cleanupPortForward() {
	if c.port == 0 {
		return
	}

	cmd := exec.Command(c.adbPath, "-s", c.deviceID, "forward", "--remove", fmt.Sprintf("tcp:%d", c.port))
	if err := cmd.Run(); err != nil {
		log.Printf("[ScrcpyClient] 清理端口转发失败: %v", err)
	} else {
		log.Printf("[ScrcpyClient] 端口转发已清理: tcp:%d", c.port)
	}
}
