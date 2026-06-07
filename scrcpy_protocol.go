package main

import _ "embed"
import "encoding/binary"
import "fmt"
import "io"
import "net"

//go:embed assets/scrcpy-server.jar
var scrcpyServerJar []byte

const (
	ScrcpyVersion   = "2.7"
	VideoSocket     = "scrcpy"
	DeviceNameLen   = 64
	PacketHeaderLen = 12
)

// PacketType 定义视频包类型
type PacketType byte

const (
	PacketConfig   PacketType = 0 // SPS/PPS 配置帧
	PacketKeyframe PacketType = 1 // 关键帧
	PacketDelta    PacketType = 2 // 普通帧
	PacketEOS      PacketType = 3 // 流结束
)

// VideoPacket 表示一个视频包
type VideoPacket struct {
	PTS  int64      // 纳秒时间戳
	Type PacketType
	Data []byte // 原始 H264 NAL 数据
}

// VideoStreamHeader 视频流头部信息
type VideoStreamHeader struct {
	DeviceName string
	Width      uint16
	Height     uint16
}

// ParsePacketHeader 解析 12 字节视频包头部
// 格式: [8B PTS+flags (big-endian uint64)] [4B size (big-endian uint32)]
// PTS 高位标志: bit63=config, bit62=keyframe
func ParsePacketHeader(header []byte) (pts int64, isConfig bool, isKeyframe bool, size uint32, err error) {
	if len(header) < PacketHeaderLen {
		return 0, false, false, 0, fmt.Errorf("header too short: %d < %d", len(header), PacketHeaderLen)
	}
	ptsFlags := binary.BigEndian.Uint64(header[0:8])
	isConfig = (ptsFlags & (1 << 63)) != 0
	isKeyframe = (ptsFlags & (1 << 62)) != 0
	pts = int64(ptsFlags & 0x3FFFFFFFFFFFFFFF) // 低62位为PTS
	size = binary.BigEndian.Uint32(header[8:12])
	return pts, isConfig, isKeyframe, size, nil
}

// ReadFullPacket 从连接读取一个完整的视频包
func ReadFullPacket(conn net.Conn) (*VideoPacket, error) {
	header := make([]byte, PacketHeaderLen)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	pts, isConfig, isKeyframe, size, err := ParsePacketHeader(header)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}

	var pktType PacketType
	if isConfig {
		pktType = PacketConfig
	} else if isKeyframe {
		pktType = PacketKeyframe
	} else {
		pktType = PacketDelta
	}

	data := make([]byte, size)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, fmt.Errorf("read data: %w", err)
	}

	return &VideoPacket{PTS: pts, Type: pktType, Data: data}, nil
}
