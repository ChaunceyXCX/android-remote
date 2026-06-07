package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"nhooyr.io/websocket"
)

type DeviceStatus struct {
	Connected bool   `json:"connected"`
	DeviceID  string `json:"deviceId,omitempty"`
	Model     string `json:"model,omitempty"`
	Android   string `json:"android,omitempty"`
	IP        string `json:"ip,omitempty"`
	Note      string `json:"note,omitempty"`
	Error     string `json:"error,omitempty"`
}

type DeviceInfo struct {
	ID       string `json:"id"`
	Model    string `json:"model,omitempty"`
	Android  string `json:"android,omitempty"`
	IP       string `json:"ip,omitempty"`
	Type     string `json:"type"`
	Note     string `json:"note,omitempty"`
	Selected bool   `json:"selected"`
}

type DeviceListResponse struct {
	Success bool         `json:"success"`
	Devices []DeviceInfo `json:"devices,omitempty"`
	Error   string       `json:"error,omitempty"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ScreenshotResponse struct {
	Success  bool   `json:"success"`
	ImageURL string `json:"imageUrl,omitempty"`
	Error    string `json:"error,omitempty"`
}

type ScreenSizeResponse struct {
	Success bool   `json:"success"`
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
	Error   string `json:"error,omitempty"`
}

type AppInfo struct {
	Package string `json:"package"`
	Name    string `json:"name,omitempty"`
}

type AppListResponse struct {
	Success bool      `json:"success"`
	Apps    []AppInfo `json:"apps,omitempty"`
	Error   string    `json:"error,omitempty"`
}

var (
	screenshotMutex sync.Mutex
	currentDevice   string
	deviceNotes     map[string]string
	notesFile       = "device_notes.json"
	configFile      = "device_config.json"
	scrcpyClient    *ScrcpyClient
	videoHub        *VideoHub
	lastMirrorError error
	portFlag        = flag.Int("port", 8080, "HTTP服务端口")
	adbFlag         = flag.String("adb", "adb", "ADB命令路径")
)

type DeviceConfig struct {
	LastDevice string `json:"lastDevice"`
}

func loadDeviceNotes() {
	deviceNotes = make(map[string]string)
	data, err := os.ReadFile(notesFile)
	if err != nil {
		return
	}
	json.Unmarshal(data, &deviceNotes)
}

func saveDeviceNotes() {
	data, err := json.MarshalIndent(deviceNotes, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(notesFile, data, 0644)
}

func loadDeviceConfig() DeviceConfig {
	var config DeviceConfig
	data, err := os.ReadFile(configFile)
	if err != nil {
		return config
	}
	json.Unmarshal(data, &config)
	return config
}

func saveDeviceConfig() {
	config := DeviceConfig{LastDevice: currentDevice}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(configFile, data, 0644)
}

func tryReconnectLastDevice() {
	config := loadDeviceConfig()
	if config.LastDevice == "" {
		return
	}

	log.Printf("尝试重连上次设备: %s", config.LastDevice)
	
	if strings.Contains(config.LastDevice, ":") {
		_, err := executeADB("connect", config.LastDevice)
		if err != nil {
			log.Printf("重连失败: %v", err)
			return
		}
	}

	output, err := executeADB("devices")
	if err != nil {
		log.Printf("获取设备列表失败: %v", err)
		return
	}

	if strings.Contains(output, config.LastDevice) {
		currentDevice = config.LastDevice
		log.Printf("重连成功: %s", currentDevice)
	}
}

func main() {
	flag.Parse()
	loadDeviceNotes()
	tryReconnectLastDevice()
	
	mux := http.NewServeMux()
	
	mux.HandleFunc("/api/status", handleDeviceStatus)
	mux.HandleFunc("/api/key/", handleKeyPress)
	mux.HandleFunc("/api/input/text", handleInputText)
	mux.HandleFunc("/api/app/start", handleStartApp)
	mux.HandleFunc("/api/app/stop", handleStopApp)
	mux.HandleFunc("/api/app/list", handleListApps)
	mux.HandleFunc("/api/system/screenshot", handleScreenshot)
	mux.HandleFunc("/api/system/screen-size", handleScreenSize)
	mux.HandleFunc("/api/system/swipe", handleSwipe)
	mux.HandleFunc("/api/system/tap", handleTap)
	mux.HandleFunc("/api/connect", handleConnect)
	mux.HandleFunc("/api/devices", handleDeviceList)
	mux.HandleFunc("/api/device/switch", handleDeviceSwitch)
	mux.HandleFunc("/api/device/switch-mode", handleDeviceSwitchMode)
	mux.HandleFunc("/api/device/note", handleDeviceNote)
	mux.HandleFunc("/api/media/status", handleMediaStatus)
	mux.HandleFunc("/api/system/volume", handleVolumeSet)
	mux.HandleFunc("/api/system/volume/status", handleVolumeStatus)
	mux.HandleFunc("/api/screen/stream", handleStream)
	mux.HandleFunc("/api/screen/mirror/start", handleMirrorStart)
	mux.HandleFunc("/api/screen/mirror/stop", handleMirrorStop)
	mux.HandleFunc("/api/screen/mirror/status", handleMirrorStatus)
	
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)
	
	fmt.Println("🚀 Android Remote 控制服务启动")
	fmt.Println("📱 请确保ADB已连接设备")
	fmt.Printf("🌐 访问地址: http://localhost:%d\n", *portFlag)
	
	addr := fmt.Sprintf(":%d", *portFlag)
	server := &http.Server{Addr: addr, Handler: mux}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("收到信号 %v，正在清理...", sig)

		if scrcpyClient != nil && scrcpyClient.State() == StateRunning {
			log.Printf("停止 scrcpy 会话")
			scrcpyClient.Stop()
		}

		log.Printf("清理端口转发")
		executeADB("forward", "--remove-all")
		executeADB("forward", "--remove-all")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)

		log.Println("清理完成")
		os.Exit(0)
	}()

	log.Fatal(server.ListenAndServe())
}

func executeADB(args ...string) (string, error) {
	if currentDevice != "" && len(args) > 0 && args[0] != "devices" && args[0] != "connect" {
		args = append([]string{"-s", currentDevice}, args...)
		log.Printf("执行ADB命令: %s %s", *adbFlag, strings.Join(args, " "))
	} else {
		log.Printf("执行ADB命令(无设备): %s %s", *adbFlag, strings.Join(args, " "))
	}
	cmd := exec.Command(*adbFlag, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ADB命令失败: %s - %v", string(output), err)
		return "", fmt.Errorf("ADB命令执行失败: %s - %v", string(output), err)
	}
	log.Printf("ADB命令成功: %s", string(output))
	return strings.TrimSpace(string(output)), nil
}

func handleDeviceStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持GET请求"})
		return
	}
	
	output, err := executeADB("devices")
	if err != nil {
		sendJSON(w, http.StatusOK, DeviceStatus{Connected: false, Error: "ADB服务未运行"})
		return
	}
	
	lines := strings.Split(output, "\n")
	var deviceID string
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, "device") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				deviceID = parts[0]
				break
			}
		}
	}
	
	if deviceID == "" {
		sendJSON(w, http.StatusOK, DeviceStatus{Connected: false})
		return
	}
	
	// 只在首次连接时设置默认设备，不覆盖用户手动切换的设备
	if currentDevice == "" {
		currentDevice = deviceID
		log.Printf("设置默认设备: %s", currentDevice)
	} else {
		log.Printf("当前设备: %s (保持不变)", currentDevice)
	}
	
	model, _ := executeADB("shell", "getprop", "ro.product.model")
	android, _ := executeADB("shell", "getprop", "ro.build.version.release")
	ip, _ := executeADB("shell", "ip", "addr", "show", "wlan0")
	
	var deviceIP string
	for _, line := range strings.Split(ip, "\n") {
		if strings.Contains(line, "inet ") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "inet" && i+1 < len(parts) {
					deviceIP = strings.Split(parts[i+1], "/")[0]
					break
				}
			}
		}
	}
	
	note := deviceNotes[currentDevice]
	log.Printf("设备状态: 连接=%v, 型号=%s, Android=%s, IP=%s, 备注=%s", true, model, android, deviceIP, note)
	sendJSON(w, http.StatusOK, DeviceStatus{
		Connected: true,
		DeviceID:  currentDevice,
		Model:     model,
		Android:   android,
		IP:        deviceIP,
		Note:      note,
	})
}

func handleKeyPress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}
	
	keyName := strings.TrimPrefix(r.URL.Path, "/api/key/")
	if keyName == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "未指定按键名称"})
		return
	}
	
	keyMap := map[string]string{
		"back":          "KEYCODE_BACK",
		"home":          "KEYCODE_HOME",
		"menu":          "KEYCODE_MENU",
		"power":         "KEYCODE_POWER",
		"volumeup":      "KEYCODE_VOLUME_UP",
		"volumedown":    "KEYCODE_VOLUME_DOWN",
		"mute":          "KEYCODE_VOLUME_MUTE",
		"play":          "KEYCODE_MEDIA_PLAY",
		"pause":         "KEYCODE_MEDIA_PAUSE",
		"playpause":     "KEYCODE_MEDIA_PLAY_PAUSE",
		"next":          "KEYCODE_MEDIA_NEXT",
		"previous":      "KEYCODE_MEDIA_PREVIOUS",
		"stop":          "KEYCODE_MEDIA_STOP",
		"rewind":        "KEYCODE_MEDIA_REWIND",
		"forward":       "KEYCODE_MEDIA_FAST_FORWARD",
		"up":            "KEYCODE_DPAD_UP",
		"down":          "KEYCODE_DPAD_DOWN",
		"left":          "KEYCODE_DPAD_LEFT",
		"right":         "KEYCODE_DPAD_RIGHT",
		"center":        "KEYCODE_DPAD_CENTER",
		"enter":         "KEYCODE_ENTER",
		"del":           "KEYCODE_DEL",
		"tab":           "KEYCODE_TAB",
		"space":         "KEYCODE_SPACE",
		"search":        "KEYCODE_SEARCH",
		"camera":        "KEYCODE_CAMERA",
		"brightnessup":  "KEYCODE_BRIGHTNESS_UP",
		"brightnessdown":"KEYCODE_BRIGHTNESS_DOWN",
	}
	
	keycode, ok := keyMap[keyName]
	if !ok {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "不支持的按键: " + keyName})
		return
	}
	
	_, err := executeADB("shell", "input", "keyevent", keycode)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{Error: "按键发送失败: " + err.Error()})
		return
	}
	
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "按键已发送: " + keyName})
}

func handleInputText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}
	
	var request struct {
		Text string `json:"text"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "无效的请求格式"})
		return
	}
	
	if request.Text == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "文本内容不能为空"})
		return
	}
	
	escapedText := strings.ReplaceAll(request.Text, " ", "%s")
	escapedText = strings.ReplaceAll(escapedText, "'", "\\'")
	escapedText = strings.ReplaceAll(escapedText, "\"", "\\\"")
	
	_, err := executeADB("shell", "input", "text", escapedText)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{Error: "文本输入失败: " + err.Error()})
		return
	}
	
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "文本已输入"})
}

func handleStartApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}
	
	var request struct {
		Package string `json:"package"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "无效的请求格式"})
		return
	}
	
	if request.Package == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "包名不能为空"})
		return
	}
	
	_, err := executeADB("shell", "monkey", "-p", request.Package, "-c", "android.intent.category.LAUNCHER", "1")
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{Error: "应用启动失败: " + err.Error()})
		return
	}
	
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "应用已启动: " + request.Package})
}

func handleStopApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}
	
	var request struct {
		Package string `json:"package"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "无效的请求格式"})
		return
	}
	
	if request.Package == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "包名不能为空"})
		return
	}
	
	_, err := executeADB("shell", "am", "force-stop", request.Package)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{Error: "应用停止失败: " + err.Error()})
		return
	}
	
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "应用已停止: " + request.Package})
}

func handleListApps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持GET请求"})
		return
	}

	output, err := executeADB("shell", "pm", "list", "packages", "-3")
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, AppListResponse{Error: "获取应用列表失败: " + err.Error()})
		return
	}

	var apps []AppInfo
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package:") {
			pkg := strings.TrimPrefix(line, "package:")
			name := getAppName(pkg)
			apps = append(apps, AppInfo{Package: pkg, Name: name})
		}
	}

	sendJSON(w, http.StatusOK, AppListResponse{Success: true, Apps: apps})
}

var commonApps = map[string]string{
	"com.tencent.qqmusicpad":              "QQ音乐HD",
	"com.netease.cloudmusic.tv":           "网易云音乐TV",
	"com.dangbeimarket":                   "当贝市场",
	"com.google.android.inputmethod.pinyin": "谷歌拼音输入法",
	"com.ktcp.video":                      "云视听极光",
	"com.gitvvideo.tencent":               "云视听极光",
	"com.galaxy.tv":                       "银河奇异果",
	"com.qiyi.video":                      "爱奇艺",
	"com.cibn.tv":                         "CIBN酷喵",
	"com.youku.phone":                     "优酷",
	"tv.danmaku.bili":                     "云视听小电视",
	"com.bilibili.app.in":                 "哔哩哔哩",
	"com.android.settings":                "设置",
	"com.android.browser":                 "浏览器",
	"com.android.gallery3d":               "图库",
	"com.android.music":                   "音乐",
	"com.android.vending":                 "Google Play Store",
	"com.dangbei.dblauncher":              "当贝桌面",
	"com.dangbei.tvmaster":                "当贝助手",
	"com.shafa.market":                    "沙发管家",
	"com.dianshijia.live":                 "电视家",
	"com.dianshijia.newlive":              "电视家",
	"org.xbmc.kodi":                       "Kodi",
	"com.mxtech.videoplayer.ad":           "MX Player",
	"com.mxtech.videoplayer.pro":          "MX Player Pro",
	"com.jellyfin.androidtv":              "Jellyfin",
	"com.plexapp.android":                 "Plex",
	"com.embymedia.embyatv":               "Emby",
}

func getAppName(pkg string) string {
	output, err := executeADB("shell", "dumpsys", "package", pkg)
	if err == nil {
		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Application Label:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}
	
	if name, found := commonApps[pkg]; found {
		return name
	}
	return ""
}

func handleScreenshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持GET请求"})
		return
	}
	
	screenshotMutex.Lock()
	defer screenshotMutex.Unlock()
	
	_, err := executeADB("shell", "screencap", "-p", "/sdcard/screenshot.png")
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, ScreenshotResponse{Error: "截图失败: " + err.Error()})
		return
	}
	
	_, err = executeADB("pull", "/sdcard/screenshot.png", "./static/screenshot.png")
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, ScreenshotResponse{Error: "截图拉取失败: " + err.Error()})
		return
	}
	
	executeADB("shell", "rm", "/sdcard/screenshot.png")
	
	sendJSON(w, http.StatusOK, ScreenshotResponse{
		Success:  true,
		ImageURL: "/screenshot.png",
	})
}

func handleSwipe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}
	
	var request struct {
		X1 int `json:"x1"`
		Y1 int `json:"y1"`
		X2 int `json:"x2"`
		Y2 int `json:"y2"`
		Duration int `json:"duration,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "无效的请求格式"})
		return
	}
	
	duration := 300
	if request.Duration > 0 {
		duration = request.Duration
	}
	
	_, err := executeADB("shell", "input", "swipe", 
		fmt.Sprintf("%d", request.X1), 
		fmt.Sprintf("%d", request.Y1),
		fmt.Sprintf("%d", request.X2),
		fmt.Sprintf("%d", request.Y2),
		fmt.Sprintf("%d", duration))
	
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{Error: "滑动失败: " + err.Error()})
		return
	}
	
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "滑动已执行"})
}

func handleScreenSize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持GET请求"})
		return
	}
	
	output, err := executeADB("shell", "wm", "size")
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, ScreenSizeResponse{Error: "获取屏幕尺寸失败: " + err.Error()})
		return
	}
	
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var sizeStr string
	for _, line := range lines {
		if strings.Contains(line, "Physical size:") {
			sizeStr = strings.TrimSpace(strings.Split(line, ":")[1])
			break
		} else if strings.Contains(line, "Override size:") {
			sizeStr = strings.TrimSpace(strings.Split(line, ":")[1])
			break
		}
	}
	
	if sizeStr == "" {
		sendJSON(w, http.StatusInternalServerError, ScreenSizeResponse{Error: "无法解析屏幕尺寸"})
		return
	}
	
	var width, height int
	_, err = fmt.Sscanf(sizeStr, "%dx%d", &width, &height)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, ScreenSizeResponse{Error: "解析屏幕尺寸失败: " + err.Error()})
		return
	}
	
	sendJSON(w, http.StatusOK, ScreenSizeResponse{
		Success: true,
		Width:   width,
		Height:  height,
	})
}

func handleTap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}
	
	var request struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "无效的请求格式"})
		return
	}
	
	_, err := executeADB("shell", "input", "tap", 
		fmt.Sprintf("%d", request.X), 
		fmt.Sprintf("%d", request.Y))
	
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{Error: "点击失败: " + err.Error()})
		return
	}
	
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "点击已执行"})
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}
	
	var request struct {
		IP string `json:"ip"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "无效的请求格式"})
		return
	}
	
	if request.IP == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "IP地址不能为空"})
		return
	}
	
	address := request.IP
	if !strings.Contains(address, ":") {
		address = address + ":5555"
	}
	
	output, err := executeADB("connect", address)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{Error: "连接失败: " + err.Error()})
		return
	}
	
	if strings.Contains(output, "connected") {
		currentDevice = address
		saveDeviceConfig()
		sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "已连接到设备: " + address})
	} else {
		sendJSON(w, http.StatusOK, APIResponse{Success: false, Error: output})
	}
}

func handleDeviceList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, DeviceListResponse{Error: "只支持GET请求"})
		return
	}
	
	output, err := executeADB("devices", "-l")
	if err != nil {
		sendJSON(w, http.StatusOK, DeviceListResponse{Error: "ADB服务未运行"})
		return
	}
	
	lines := strings.Split(output, "\n")
	devices := make([]DeviceInfo, 0)
	
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		
		if parts[1] != "device" {
			continue
		}
		
		deviceID := parts[0]
		deviceType := "usb"
		if strings.Contains(deviceID, ":") {
			deviceType = "wireless"
		}
		
		model, _ := executeADB("-s", deviceID, "shell", "getprop", "ro.product.model")
		android, _ := executeADB("-s", deviceID, "shell", "getprop", "ro.build.version.release")
		
		var deviceIP string
		if deviceType == "wireless" {
			deviceIP = strings.Split(deviceID, ":")[0]
		} else {
			ipOutput, _ := executeADB("-s", deviceID, "shell", "ip", "addr", "show", "wlan0")
			for _, ipLine := range strings.Split(ipOutput, "\n") {
				if strings.Contains(ipLine, "inet ") {
					ipParts := strings.Fields(ipLine)
					for i, part := range ipParts {
						if part == "inet" && i+1 < len(ipParts) {
							deviceIP = strings.Split(ipParts[i+1], "/")[0]
							break
						}
					}
				}
			}
		}
		
		note := deviceNotes[deviceID]
		
		devices = append(devices, DeviceInfo{
			ID:       deviceID,
			Model:    model,
			Android:  android,
			IP:       deviceIP,
			Type:     deviceType,
			Note:     note,
			Selected: deviceID == currentDevice,
		})
	}
	
	sendJSON(w, http.StatusOK, DeviceListResponse{Success: true, Devices: devices})
}

func handleDeviceSwitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}
	
	var request struct {
		DeviceID string `json:"deviceId"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "无效的请求格式"})
		return
	}
	
	if request.DeviceID == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "设备ID不能为空"})
		return
	}
	
	if scrcpyClient != nil && scrcpyClient.State() == StateRunning {
		log.Printf("切换设备前停止 scrcpy 会话")
		scrcpyClient.Stop()
	}
	
	currentDevice = request.DeviceID
	saveDeviceConfig()
	log.Printf("切换到设备: %s", currentDevice)
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "已切换到设备: " + request.DeviceID})
}

func handleDeviceNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}
	
	var request struct {
		DeviceID string `json:"deviceId"`
		Note     string `json:"note"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "无效的请求格式"})
		return
	}
	
	if request.DeviceID == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "设备ID不能为空"})
		return
	}
	
	deviceNotes[request.DeviceID] = request.Note
	saveDeviceNotes()
	
	log.Printf("保存设备备注: %s -> %s", request.DeviceID, request.Note)
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "备注已保存"})
}

func handleDeviceSwitchMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}

	var request struct {
		DeviceID string `json:"deviceId"`
		Mode     string `json:"mode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "无效的请求格式"})
		return
	}

	if request.DeviceID == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "设备ID不能为空"})
		return
	}

	isWireless := strings.Contains(request.DeviceID, ":")

	if request.Mode == "wireless" && !isWireless {
		ipOutput, err := executeADB("-s", request.DeviceID, "shell", "ip", "addr", "show", "wlan0")
		if err != nil {
			sendJSON(w, http.StatusOK, APIResponse{Error: "获取设备IP失败"})
			return
		}

		var deviceIP string
		for _, line := range strings.Split(ipOutput, "\n") {
			if strings.Contains(line, "inet ") {
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "inet" && i+1 < len(parts) {
						deviceIP = strings.Split(parts[i+1], "/")[0]
						break
					}
				}
			}
		}

		if deviceIP == "" {
			sendJSON(w, http.StatusOK, APIResponse{Error: "无法获取设备IP，请确保设备已连接WiFi"})
			return
		}

		_, err = executeADB("-s", request.DeviceID, "tcpip", "5555")
		if err != nil {
			sendJSON(w, http.StatusOK, APIResponse{Error: "开启无线调试失败"})
			return
		}

		time.Sleep(2 * time.Second)

		newDeviceID := deviceIP + ":5555"
		_, err = executeADB("connect", newDeviceID)
		if err != nil {
			sendJSON(w, http.StatusOK, APIResponse{Error: "无线连接失败: " + err.Error()})
			return
		}

		if note, exists := deviceNotes[request.DeviceID]; exists {
			deviceNotes[newDeviceID] = note
			saveDeviceNotes()
		}

		currentDevice = newDeviceID
		saveDeviceConfig()

		log.Printf("设备 %s 已切换到无线模式: %s", request.DeviceID, newDeviceID)
		sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "已切换到无线模式: " + newDeviceID})

	} else if request.Mode == "usb" && isWireless {
		_, err := executeADB("-s", request.DeviceID, "usb")
		if err != nil {
			sendJSON(w, http.StatusOK, APIResponse{Error: "切换USB模式失败"})
			return
		}

		time.Sleep(1 * time.Second)

		output, _ := executeADB("devices")
		var usbDeviceID string
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "device") && !strings.Contains(line, ":") {
				parts := strings.Fields(line)
				if len(parts) >= 2 && parts[1] == "device" {
					usbDeviceID = parts[0]
					break
				}
			}
		}

		if usbDeviceID != "" {
			if note, exists := deviceNotes[request.DeviceID]; exists {
				deviceNotes[usbDeviceID] = note
				saveDeviceNotes()
			}
			currentDevice = usbDeviceID
			saveDeviceConfig()
		}

		log.Printf("设备 %s 已切换到USB模式, 新设备ID: %s", request.DeviceID, usbDeviceID)
		sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "已切换到USB模式"})

	} else {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "无效的切换模式"})
	}
}

type MediaStatus struct {
	Playing bool   `json:"playing"`
	Package string `json:"package,omitempty"`
}

type VolumeStatus struct {
	Current int `json:"current"`
	Max     int `json:"max"`
}

func handleMediaStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, MediaStatus{})
		return
	}

	output, err := executeADB("shell", "dumpsys", "media_session")
	if err != nil {
		sendJSON(w, http.StatusOK, MediaStatus{Playing: false})
		return
	}

	playing := false
	if strings.Contains(output, "state=3") {
		playing = true
	}

	sendJSON(w, http.StatusOK, MediaStatus{Playing: playing})
}

func handleVolumeSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}

	var request struct {
		Level int `json:"level"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "无效的请求格式"})
		return
	}

	if request.Level < 0 || request.Level > 100 {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "音量级别必须在0-100之间"})
		return
	}

	output, err := executeADB("shell", "cmd", "media_session", "volume", "--get", "--stream", "3")
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{Error: "获取音量失败: " + err.Error()})
		return
	}

	currentVolume := 0
	maxVolume := 30
	if strings.Contains(output, "volume is") {
		parts := strings.Split(output, "volume is")
		if len(parts) > 1 {
			volumePart := strings.TrimSpace(parts[1])
			fmt.Sscanf(volumePart, "%d", &currentVolume)
		}
	}

	if strings.Contains(output, "in range") {
		rangeParts := strings.Split(output, "in range")
		if len(rangeParts) > 1 {
			rangeStr := rangeParts[1]
			if idx := strings.Index(rangeStr, ".."); idx != -1 {
				endPart := rangeStr[idx+2:]
				if endIdx := strings.Index(endPart, "]"); endIdx != -1 {
					fmt.Sscanf(endPart[:endIdx], "%d", &maxVolume)
				}
			}
		}
	}

	targetVolume := request.Level * maxVolume / 100
	diff := targetVolume - currentVolume

	if diff > 0 {
		for i := 0; i < diff; i++ {
			executeADB("shell", "input", "keyevent", "24")
		}
	} else if diff < 0 {
		for i := 0; i < -diff; i++ {
			executeADB("shell", "input", "keyevent", "25")
		}
	}

	sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: fmt.Sprintf("音量已设置为 %d%%", request.Level)})
}

func handleVolumeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, VolumeStatus{})
		return
	}

	output, err := executeADB("shell", "cmd", "media_session", "volume", "--get", "--stream", "3")
	if err != nil {
		sendJSON(w, http.StatusOK, VolumeStatus{Current: 0, Max: 100})
		return
	}

	current := 0
	maxVolume := 30 // 默认最大值

	// 解析 "volume is 7 in range [0..30]"
	if strings.Contains(output, "volume is") {
		parts := strings.Split(output, "volume is")
		if len(parts) > 1 {
			volumePart := strings.TrimSpace(parts[1])
			// 提取当前音量值
			fmt.Sscanf(volumePart, "%d", &current)
		}
	}

	if strings.Contains(output, "in range") {
		rangeParts := strings.Split(output, "in range")
		if len(rangeParts) > 1 {
			rangeStr := rangeParts[1]
			// 提取 [0..30] 中的最大值
			if idx := strings.Index(rangeStr, ".."); idx != -1 {
				endPart := rangeStr[idx+2:]
				if endIdx := strings.Index(endPart, "]"); endIdx != -1 {
					fmt.Sscanf(endPart[:endIdx], "%d", &maxVolume)
				}
			}
		}
	}

	// 将实际音量映射到0-100百分比
	percentCurrent := current * 100 / maxVolume

	sendJSON(w, http.StatusOK, VolumeStatus{Current: percentCurrent, Max: 100})
}

func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func handleStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持GET请求"})
		return
	}

	if scrcpyClient == nil || scrcpyClient.State() != StateRunning {
		sendJSON(w, http.StatusServiceUnavailable, APIResponse{Error: "屏幕镜像未启动"})
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ch := videoHub.Subscribe()
	defer videoHub.Unsubscribe(ch)

	ctx := conn.CloseRead(r.Context())

	for {
		select {
		case <-ctx.Done():
			return
		case packet, ok := <-ch:
			if !ok {
				return
			}
			// 构建 12 字节头部 + NAL 数据
			// 格式: [8B PTS+flags] [4B size]
			ptsFlags := uint64(packet.PTS) & 0x3FFFFFFFFFFFFFFF
			if packet.Type == PacketConfig {
				ptsFlags |= (1 << 63)
			} else if packet.Type == PacketKeyframe {
				ptsFlags |= (1 << 62)
			}
			fullPacket := make([]byte, 12+len(packet.Data))
			binary.BigEndian.PutUint64(fullPacket[0:8], ptsFlags)
			binary.BigEndian.PutUint32(fullPacket[8:12], uint32(len(packet.Data)))
			copy(fullPacket[12:], packet.Data)

			err := conn.Write(ctx, websocket.MessageBinary, fullPacket)
			if err != nil {
				log.Printf("WebSocket写入失败: %v", err)
				return
			}
		}
	}
}

type MirrorStatusResponse struct {
	Success bool   `json:"success"`
	State   string `json:"state,omitempty"`
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
	Error   string `json:"error,omitempty"`
}

func handleMirrorStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}

	if currentDevice == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "未连接设备"})
		return
	}

	if scrcpyClient != nil && scrcpyClient.State() == StateRunning {
		sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "镜像已在运行"})
		return
	}

	videoHub = NewVideoHub()
	scrcpyClient = NewScrcpyClient(currentDevice, *adbFlag, videoHub)
	lastMirrorError = nil

	// 异步启动，避免阻塞 HTTP handler
	go func() {
		if err := scrcpyClient.Start(); err != nil {
			log.Printf("启动scrcpy失败: %v", err)
			lastMirrorError = err
		}
	}()

	sendJSON(w, http.StatusOK, MirrorStatusResponse{
		Success: true,
		State:   "starting",
	})
}

func handleMirrorStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持POST请求"})
		return
	}

	if scrcpyClient == nil {
		sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "镜像未运行"})
		return
	}

	scrcpyClient.Stop()
	scrcpyClient = nil
	sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "镜像已停止"})
}

func handleMirrorStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "只支持GET请求"})
		return
	}

	if scrcpyClient == nil {
		state := "idle"
		errMsg := ""
		if lastMirrorError != nil {
			state = "error"
			errMsg = lastMirrorError.Error()
		}
		sendJSON(w, http.StatusOK, MirrorStatusResponse{
			Success: true,
			State:   state,
			Error:   errMsg,
		})
		return
	}

	header := scrcpyClient.StreamHeader()
	resp := MirrorStatusResponse{
		Success: true,
		State:   scrcpyClient.State().String(),
		Width:   int(header.Width),
		Height:  int(header.Height),
	}
	if scrcpyClient.State() == StateError && lastMirrorError != nil {
		resp.Error = lastMirrorError.Error()
	}
	sendJSON(w, http.StatusOK, resp)
}