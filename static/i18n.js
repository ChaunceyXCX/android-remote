const i18n = {
    zh: {
        // Header
        appTitle: 'Android Remote',
        connected: '已连接',
        disconnected: '未连接',

        // Connection Section
        deviceManagement: '设备管理',
        ipPlaceholder: '设备IP地址或IP:端口 (例如: 192.168.1.100:5555)',
        connect: '连接',
        refreshDevices: '刷新设备',
        connectionTip: '无线ADB连接前，请确保在安卓设备的开发者选项中开启"无线调试"，并勾选"永久允许"该电脑进行调试',
        connectedDevices: '已连接设备',
        noDevices: '暂无设备连接',
        device: '设备',
        model: '型号',
        ip: 'IP',
        addNote: '备注...',
        toWireless: '→ 无线',
        toUSB: '→ USB',

        // Control Sections
        navigation: '导航控制',
        back: '返回',
        home: '主页',
        recent: '最近',
        up: '上',
        down: '下',
        left: '左',
        right: '右',
        confirm: '确认',

        mediaControl: '媒体控制',
        previous: '上一曲',
        playPause: '播放/暂停',
        next: '下一曲',
        volumeDown: '音量减',
        mute: '静音',
        volumeUp: '音量加',

        powerControl: '电源控制',
        power: '电源',
        brightnessUp: '亮度+',
        brightnessDown: '亮度-',

        textInput: '文本输入',
        textPlaceholder: '输入要发送的文本...',
        send: '发送',

        appManagement: '应用管理',
        packagePlaceholder: '应用包名',
        start: '启动',
        stop: '停止',
        getAppList: '获取应用列表',

        screenshot: '屏幕截图',
        takeScreenshot: '截图',
        clickToTap: '点击截图直接触摸',
        screenshotPlaceholder: '截图将显示在这里',

        touchSimulator: '触摸模拟',
        xCoord: 'X坐标',
        yCoord: 'Y坐标',
        tap: '点击',
        swipe: '滑动',
        swipeStart: '起点',
        swipeEnd: '终点',

        // Toast messages
        enterIP: '请输入设备IP地址',
        connecting: '连接中...',
        connected_success: '已连接',
        connectionFailed: '连接失败',
        enterText: '请输入文本',
        textSent: '文本已发送',
        enterPackage: '请输入包名',
        appStarted: '应用已启动',
        appStopped: '应用已停止',
        screenshotSuccess: '截图成功（点击图片可直接触摸）',
        screenshotFailed: '截图失败',
        enterCoords: '请输入有效的坐标',
        tapSuccess: '点击成功',
        swipeSuccess: '滑动成功',
        noteSaved: '备注已保存',
        switchedTo: '已切换到设备',
        switchingToWireless: '正在切换到无线模式...',
        switchingToUSB: '正在切换到USB模式...',
        switchedToWireless: '已切换到无线模式',
        switchedToUSB: '已切换到USB模式',

        // Swipe indicators
        navMedia: '导航/媒体',
        inputApps: '输入/应用',
        touch: '触摸'
    },

    en: {
        // Header
        appTitle: 'Android Remote',
        connected: 'Connected',
        disconnected: 'Disconnected',

        // Connection Section
        deviceManagement: 'Device Management',
        ipPlaceholder: 'Device IP or IP:port (e.g., 192.168.1.100:5555)',
        connect: 'Connect',
        refreshDevices: 'Refresh',
        connectionTip: 'Before wireless ADB connection, please enable "Wireless debugging" in Developer options and check "Always allow" this computer',
        connectedDevices: 'Connected Devices',
        noDevices: 'No devices connected',
        device: 'Device',
        model: 'Model',
        ip: 'IP',
        addNote: 'Note...',
        toWireless: '→ Wireless',
        toUSB: '→ USB',

        // Control Sections
        navigation: 'Navigation',
        back: 'Back',
        home: 'Home',
        recent: 'Recent',
        up: 'Up',
        down: 'Down',
        left: 'Left',
        right: 'Right',
        confirm: 'OK',

        mediaControl: 'Media Control',
        previous: 'Previous',
        playPause: 'Play/Pause',
        next: 'Next',
        volumeDown: 'Vol-',
        mute: 'Mute',
        volumeUp: 'Vol+',

        powerControl: 'Power Control',
        power: 'Power',
        brightnessUp: 'Bright+',
        brightnessDown: 'Bright-',

        textInput: 'Text Input',
        textPlaceholder: 'Enter text to send...',
        send: 'Send',

        appManagement: 'App Management',
        packagePlaceholder: 'Package name',
        start: 'Start',
        stop: 'Stop',
        getAppList: 'Get App List',

        screenshot: 'Screenshot',
        takeScreenshot: 'Screenshot',
        clickToTap: 'Click to tap',
        screenshotPlaceholder: 'Screenshot will appear here',

        touchSimulator: 'Touch Simulator',
        xCoord: 'X',
        yCoord: 'Y',
        tap: 'Tap',
        swipe: 'Swipe',
        swipeStart: 'Start',
        swipeEnd: 'End',

        // Toast messages
        enterIP: 'Please enter device IP',
        connecting: 'Connecting...',
        connected_success: 'Connected',
        connectionFailed: 'Connection failed',
        enterText: 'Please enter text',
        textSent: 'Text sent',
        enterPackage: 'Please enter package name',
        appStarted: 'App started',
        appStopped: 'App stopped',
        screenshotSuccess: 'Screenshot taken (click image to tap)',
        screenshotFailed: 'Screenshot failed',
        enterCoords: 'Please enter valid coordinates',
        tapSuccess: 'Tap executed',
        swipeSuccess: 'Swipe executed',
        noteSaved: 'Note saved',
        switchedTo: 'Switched to device',
        switchingToWireless: 'Switching to wireless...',
        switchingToUSB: 'Switching to USB...',
        switchedToWireless: 'Switched to wireless mode',
        switchedToUSB: 'Switched to USB mode',

        // Swipe indicators
        navMedia: 'Nav/Media',
        inputApps: 'Input/Apps',
        touch: 'Touch'
    }
};

class I18n {
    constructor() {
        this.currentLang = localStorage.getItem('language') || 'zh';
        this.translations = i18n;
    }

    t(key) {
        return this.translations[this.currentLang]?.[key] || key;
    }

    setLanguage(lang) {
        this.currentLang = lang;
        localStorage.setItem('language', lang);
        this.updatePageTexts();
    }

    toggleLanguage() {
        const newLang = this.currentLang === 'zh' ? 'en' : 'zh';
        this.setLanguage(newLang);
        return newLang;
    }

    updatePageTexts() {
        // Header
        this.setText('appTitle', 'appTitle');
        this.setText('.status-text', 'disconnected', true);

        // Connection section
        this.setPlaceholder('deviceIP', 'ipPlaceholder');
        this.setText('#connectBtn', 'connect');
        this.setText('#refreshDevicesBtn', 'refreshDevices');

        // Section titles
        const sections = {
            '.navigation .section-title': 'navigation',
            '.media .section-title': 'mediaControl',
            '.power .section-title': 'powerControl',
            '.input .section-title': 'textInput',
            '.apps .section-title': 'appManagement',
            '.touch .section-title': 'touchSimulator'
        };

        for (const [selector, key] of Object.entries(sections)) {
            const el = document.querySelector(selector);
            if (el) el.textContent = this.t(key);
        }

        // Buttons with data-key attributes
        const keyButtons = {
            'back': 'back',
            'home': 'home',
            'recent': 'recent',
            'up': 'up',
            'down': 'down',
            'left': 'left',
            'right': 'right',
            'center': 'confirm',
            'previous': 'previous',
            'playpause': 'playPause',
            'next': 'next',
            'volumedown': 'volumeDown',
            'mute': 'mute',
            'volumeup': 'volumeUp',
            'power': 'power',
            'brightnessup': 'brightnessUp',
            'brightnessdown': 'brightnessDown'
        };

        document.querySelectorAll('[data-key]').forEach(btn => {
            const key = btn.dataset.key;
            if (keyButtons[key]) {
                const span = btn.querySelector('span');
                if (span) span.textContent = this.t(keyButtons[key]);
            }
        });

        // Placeholders
        this.setPlaceholder('textInput', 'textPlaceholder');
        this.setPlaceholder('packageName', 'packagePlaceholder');
        this.setPlaceholder('tapX', 'xCoord');
        this.setPlaceholder('tapY', 'yCoord');

        const langBtn = document.getElementById('langToggle');
        if (langBtn) {
            langBtn.textContent = this.currentLang === 'zh' ? 'EN' : '中';
        }
    }

    setText(selector, key, isClass = false) {
        const el = isClass ? document.querySelector(selector) : document.getElementById(selector);
        if (el) el.textContent = this.t(key);
    }

    setPlaceholder(id, key) {
        const el = document.getElementById(id);
        if (el) el.placeholder = this.t(key);
    }
}

// Initialize i18n
const i18nInstance = new I18n();
