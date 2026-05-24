package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
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
	mux.HandleFunc("/api/system/swipe", handleSwipe)
	mux.HandleFunc("/api/system/tap", handleTap)
	mux.HandleFunc("/api/connect", handleConnect)
	mux.HandleFunc("/api/devices", handleDeviceList)
	mux.HandleFunc("/api/device/switch", handleDeviceSwitch)
	mux.HandleFunc("/api/device/switch-mode", handleDeviceSwitchMode)
	mux.HandleFunc("/api/device/note", handleDeviceNote)
	
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)
	
	fmt.Println("🚀 Android Remote 控制服务启动")
	fmt.Println("📱 请确保ADB已连接设备")
	fmt.Println("🌐 访问地址: http://localhost:8080")
	
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func executeADB(args ...string) (string, error) {
	if currentDevice != "" && len(args) > 0 && args[0] != "devices" && args[0] != "connect" {
		args = append([]string{"-s", currentDevice}, args...)
		log.Printf("执行ADB命令: adb %s", strings.Join(args, " "))
	} else {
		log.Printf("执行ADB命令(无设备): adb %s", strings.Join(args, " "))
	}
	cmd := exec.Command("adb", args...)
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

func getAppName(pkg string) string {
	output, err := executeADB("shell", "dumpsys", "package", pkg)
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Application Label:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
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

		_, err = executeADB("connect", deviceIP+":5555")
		if err != nil {
			sendJSON(w, http.StatusOK, APIResponse{Error: "无线连接失败: " + err.Error()})
			return
		}

		log.Printf("设备 %s 已切换到无线模式: %s:5555", request.DeviceID, deviceIP)
		sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "已切换到无线模式: " + deviceIP + ":5555"})

	} else if request.Mode == "usb" && isWireless {
		_, err := executeADB("-s", request.DeviceID, "usb")
		if err != nil {
			sendJSON(w, http.StatusOK, APIResponse{Error: "切换USB模式失败"})
			return
		}

		log.Printf("设备 %s 已切换到USB模式", request.DeviceID)
		sendJSON(w, http.StatusOK, APIResponse{Success: true, Message: "已切换到USB模式"})

	} else {
		sendJSON(w, http.StatusBadRequest, APIResponse{Error: "无效的切换模式"})
	}
}

func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}