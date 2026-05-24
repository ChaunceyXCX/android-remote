<div align="center">

# 🎮 Android Remote

**网页版安卓遥控器 | Web-based Android Remote Control**

[English](#english) | [中文](#中文)

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![HTML5](https://img.shields.io/badge/HTML5-E34F26?style=flat&logo=html5&logoColor=white)
![CSS3](https://img.shields.io/badge/CSS3-1572B6?style=flat&logo=css3&logoColor=white)
![JavaScript](https://img.shields.io/badge/JavaScript-F7DF1E?style=flat&logo=javascript&logoColor=black)
![License](https://img.shields.io/badge/License-MIT-blue)

</div>

---

<a name="english"></a>

## 🇬🇧 English

### Features

- 📱 **Device Management** - Auto-detect USB/Wireless devices, switch between them
- 🎮 **Remote Control** - Navigation, media, power, volume controls
- 📸 **Screenshot** - Real-time screenshot with click-to-tap functionality
- 📝 **Text Input** - Send text to Android device directly
- 📦 **App Management** - List, start, stop applications
- 🌐 **Bilingual UI** - Chinese/English language switching
- 📱 **Mobile Optimized** - Responsive design with swipe navigation

### Quick Start

#### Prerequisites

- Go 1.21 or higher
- ADB (Android Debug Bridge) installed
- Android device with USB debugging enabled

#### Installation

```bash
# Clone the repository
git clone https://github.com/YOUR_USERNAME/android-remote.git
cd android-remote

# Build and run
go build -o android-remote .
./android-remote
```

#### Usage

1. Connect your Android device via USB or wireless ADB
2. Open `http://localhost:8080` in your browser
3. For mobile access, use your computer's IP address

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/status` | Get device status |
| GET | `/api/devices` | List connected devices |
| POST | `/api/device/switch` | Switch active device |
| POST | `/api/connect` | Connect to wireless device |
| POST | `/api/key/:keyname` | Send key press |
| POST | `/api/input/text` | Input text |
| GET | `/api/system/screenshot` | Take screenshot |
| POST | `/api/system/tap` | Simulate tap |
| POST | `/api/system/swipe` | Simulate swipe |

### Tech Stack

- **Backend**: Go with net/http
- **Frontend**: Vanilla HTML/CSS/JavaScript
- **Communication**: ADB (Android Debug Bridge)

---

<a name="中文"></a>

## 🇨🇳 中文

### 功能特性

- 📱 **设备管理** - 自动检测 USB/无线设备，支持设备切换
- 🎮 **远程控制** - 导航键、媒体控制、电源、音量调节
- 📸 **屏幕截图** - 实时截图，支持点击截图直接触摸
- 📝 **文本输入** - 直接向安卓设备发送文本
- 📦 **应用管理** - 查看、启动、停止应用
- 🌐 **中英双语** - 支持中英文界面切换
- 📱 **移动端优化** - 响应式设计，支持滑动翻页

### 快速开始

#### 环境要求

- Go 1.21 或更高版本
- 已安装 ADB (Android Debug Bridge)
- 安卓设备已开启 USB 调试

#### 安装运行

```bash
# 克隆仓库
git clone https://github.com/YOUR_USERNAME/android-remote.git
cd android-remote

# 编译运行
go build -o android-remote .
./android-remote
```

#### 使用方法

1. 通过 USB 或无线 ADB 连接安卓设备
2. 在浏览器中打开 `http://localhost:8080`
3. 移动端访问请使用电脑的 IP 地址

### API 接口

| 方法 | 接口 | 说明 |
|------|------|------|
| GET | `/api/status` | 获取设备状态 |
| GET | `/api/devices` | 获取设备列表 |
| POST | `/api/device/switch` | 切换当前设备 |
| POST | `/api/connect` | 连接无线设备 |
| POST | `/api/key/:keyname` | 发送按键 |
| POST | `/api/input/text` | 输入文本 |
| GET | `/api/system/screenshot` | 截图 |
| POST | `/api/system/tap` | 模拟点击 |
| POST | `/api/system/swipe` | 模拟滑动 |

### 技术栈

- **后端**: Go + net/http
- **前端**: 原生 HTML/CSS/JavaScript
- **通信**: ADB (Android Debug Bridge)

---

## 📄 License

MIT License © 2024
