const i18n = {
    zh: {
        appTitle: 'Android Remote',
        connected: '已连接',
        disconnected: '未连接',

        deviceManagement: '设备管理',
        ipPlaceholder: '设备IP地址或IP:端口 (例如: 192.168.1.100:5555)',
        connect: '连接',
        refreshDevices: '刷新设备',
        connectionTip: '无线ADB连接前，请确保在安卓设备的开发者选项中开启"无线调试"，并勾选"永久允许"该电脑进行调试',
        connectedDevices: '已连接设备',
        noDevices: '暂无设备连接',
        device: '设备',
        model: '型号',
        androidVersion: 'Android',
        ip: 'IP',
        addNote: '备注...',
        toWireless: '→ 无线',
        toUSB: '→ USB',

        navigation: '导航控制',
        back: '返回',
        home: '主页',
        recent: '最近',
        menu: '菜单',
        search: '搜索',
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
        noAppFound: '未找到应用',

        screenshot: '获取屏幕',
        takeScreenshot: '获取屏幕',
        screenshoting: '正在获取...',
        clickToTap: '点击屏幕直接触摸',
        syncActionRefresh: '同步操作刷新屏幕',
        screenshotPreview: '屏幕预览',
        screenshotPlaceholder: '屏幕画面将显示在这里',

        footerTitle: '安卓远程控制',
        footerDesc: 'ADB命令转发',

        enterIP: '请输入设备IP地址',
        connectionFailed: '连接失败',
        connectionRequestFailed: '连接请求失败',
        keySendFailed: '按键发送失败',
        keyRequestFailed: '按键请求失败',
        enterText: '请输入要发送的文本',
        textSendFailed: '文本发送失败',
        textRequestFailed: '文本请求失败',
        enterPackage: '请输入应用包名',
        appStartFailed: '应用启动失败',
        appStopFailed: '应用停止失败',
        appStartRequestFailed: '启动请求失败',
        appStopRequestFailed: '停止请求失败',
        getAppListFailed: '获取应用列表失败',
        getAppListRequestFailed: '列表请求失败',
        screenshotSuccess: '获取屏幕成功（点击画面可直接触摸）',
        screenshotFailed: '获取屏幕失败',
        screenshotRequestFailed: '获取屏幕请求失败',
        coordFilled: '坐标已填入',
        noteSaved: '备注已保存',
        noteSaveFailed: '保存失败',
        noteSaveRequestFailed: '保存备注失败',
        switchedToDevice: '已切换到设备',
        switchFailed: '切换失败',
        switchRequestFailed: '切换请求失败',
        switchingToWireless: '正在切换到无线模式...',
        switchingToUSB: '正在切换到USB模式...',
        switchedToWireless: '已切换到无线模式',
        switchedToUSB: '已切换到USB模式',
        switchModeFailed: '切换失败',

        manageDevices: '管理设备',
        backToHome: '返回主页',
        selectDevice: '选择设备',
        deviceManagementPage: '设备管理',
        noDeviceSelected: '未选择设备',

        navMedia: '导航/媒体',
        inputApps: '输入/应用',
        touchControl: '触摸'
    },

    en: {
        appTitle: 'Android Remote',
        connected: 'Connected',
        disconnected: 'Disconnected',

        deviceManagement: 'Device Management',
        ipPlaceholder: 'Device IP or IP:port (e.g., 192.168.1.100:5555)',
        connect: 'Connect',
        refreshDevices: 'Refresh',
        connectionTip: 'Before wireless ADB connection, please enable "Wireless debugging" in Developer options and check "Always allow" this computer',
        connectedDevices: 'Connected Devices',
        noDevices: 'No devices connected',
        device: 'Device',
        model: 'Model',
        androidVersion: 'Android',
        ip: 'IP',
        addNote: 'Note...',
        toWireless: '→ Wireless',
        toUSB: '→ USB',

        navigation: 'Navigation',
        back: 'Back',
        home: 'Home',
        recent: 'Recent',
        menu: 'Menu',
        search: 'Search',
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
        noAppFound: 'No app found',

        screenshot: 'Get Screen',
        takeScreenshot: 'Get Screen',
        screenshoting: 'Fetching...',
        clickToTap: 'Click screen to tap',
        syncActionRefresh: 'Sync actions to refresh screen',
        screenshotPreview: 'Screen Preview',
        screenshotPlaceholder: 'Screen image will appear here',

        footerTitle: 'Android Remote Control',
        footerDesc: 'ADB Command Forwarding',

        enterIP: 'Please enter device IP',
        connectionFailed: 'Connection failed',
        connectionRequestFailed: 'Connection request failed',
        keySendFailed: 'Key send failed',
        keyRequestFailed: 'Key request failed',
        enterText: 'Please enter text to send',
        textSendFailed: 'Text send failed',
        textRequestFailed: 'Text request failed',
        enterPackage: 'Please enter package name',
        appStartFailed: 'App start failed',
        appStopFailed: 'App stop failed',
        appStartRequestFailed: 'Start request failed',
        appStopRequestFailed: 'Stop request failed',
        getAppListFailed: 'Get app list failed',
        getAppListRequestFailed: 'List request failed',
        screenshotSuccess: 'Screen fetched (click screen to tap)',
        screenshotFailed: 'Failed to fetch screen',
        screenshotRequestFailed: 'Screen fetch request failed',
        coordFilled: 'Coord filled',
        noteSaved: 'Note saved',
        noteSaveFailed: 'Save failed',
        noteSaveRequestFailed: 'Note save failed',
        switchedToDevice: 'Switched to device',
        switchFailed: 'Switch failed',
        switchRequestFailed: 'Switch request failed',
        switchingToWireless: 'Switching to wireless...',
        switchingToUSB: 'Switching to USB...',
        switchedToWireless: 'Switched to wireless',
        switchedToUSB: 'Switched to USB',
        switchModeFailed: 'Switch failed',

        manageDevices: 'Manage Devices',
        backToHome: 'Back to Home',
        selectDevice: 'Select Device',
        deviceManagementPage: 'Device Management',
        noDeviceSelected: 'No device selected',

        navMedia: 'Nav/Media',
        inputApps: 'Input/Apps',
        touchControl: 'Touch'
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
        const statusText = document.querySelector('.status-text');
        if (statusText) {
            const isConnected = statusText.textContent === this.translations.zh.connected || 
                               statusText.textContent === this.translations.en.connected;
            statusText.textContent = isConnected ? this.t('connected') : this.t('disconnected');
        }

        this.setTextBySelector('.section-header-left span', 'deviceManagement');
        this.setPlaceholder('deviceIP', 'ipPlaceholder');
        this.setTextById('connectBtn', 'connect');
        this.setTextById('refreshDevicesBtn', 'refreshDevices');
        this.setTextBySelector('.connection-tip span', 'connectionTip');
        this.setTextBySelector('.device-list-header h3', 'connectedDevices');
        this.setTextBySelector('.device-item-empty', 'noDevices');

        this.setTextById('screenshotBtn', 'takeScreenshot');
        this.setTextById('clickToTapLabel', 'clickToTap');
        this.setTextById('syncActionRefreshLabel', 'syncActionRefresh');
        this.setTextBySelector('.screenshot-preview .placeholder span', 'screenshotPlaceholder');

        const sections = {
            '.navigation .section-title': 'navigation',
            '.media .section-title': 'mediaControl',
            '.power .section-title': 'powerControl',
            '.input .section-title': 'textInput',
            '.apps .section-title': 'appManagement',
            '.touch .section-title': 'touchSimulator'
        };
        for (const [selector, key] of Object.entries(sections)) {
            this.setTextBySelector(selector, key);
        }

        const keyButtons = {
            'back': 'back', 'home': 'home', 'recent': 'recent',
            'up': 'up', 'down': 'down', 'left': 'left', 'right': 'right',
            'center': 'confirm', 'menu': 'menu', 'search': 'search',
            'previous': 'previous', 'playpause': 'playPause', 'next': 'next',
            'volumedown': 'volumeDown', 'mute': 'mute', 'volumeup': 'volumeUp',
            'power': 'power', 'brightnessup': 'brightnessUp', 'brightnessdown': 'brightnessDown'
        };
        document.querySelectorAll('[data-key]').forEach(btn => {
            const key = btn.dataset.key;
            if (keyButtons[key]) {
                const hasIcon = btn.querySelector('i');
                const hasText = btn.textContent.trim().length > (hasIcon ? 1 : 0);
                
                if (hasIcon && hasText) {
                    const icon = btn.querySelector('i');
                    btn.innerHTML = '';
                    btn.appendChild(icon);
                    btn.appendChild(document.createTextNode(' ' + this.t(keyButtons[key])));
                } else if (hasIcon && !hasText) {
                    btn.title = this.t(keyButtons[key]);
                }
            }
        });

        this.setPlaceholder('textInput', 'textPlaceholder');
        this.setTextById('sendTextBtn', 'send');
        this.setPlaceholder('packageName', 'packagePlaceholder');
        this.setTextById('startAppBtn', 'start');
        this.setTextById('stopAppBtn', 'stop');
        this.setTextById('listAppsBtn', 'getAppList');

        const footerSpans = document.querySelectorAll('.footer-content span');
        if (footerSpans[0]) footerSpans[0].textContent = this.t('footerTitle');
        if (footerSpans[2]) footerSpans[2].textContent = this.t('footerDesc');

        const langSelect = document.getElementById('langSelect');
        if (langSelect) {
            langSelect.value = this.currentLang;
        }

        const deviceDropdown = document.getElementById('deviceDropdown');
        if (deviceDropdown) {
            const firstOption = deviceDropdown.querySelector('option[value=""]');
            if (firstOption) firstOption.textContent = this.t('selectDevice');
        }

        this.setTextById('manageDevicesLink', 'manageDevices');
        this.setTextById('backToHomeLink', 'backToHome');

        this.setTextById('devicePageTitle', 'deviceManagementPage');
    }

    setTextById(id, key) {
        const el = document.getElementById(id);
        if (el) el.textContent = this.t(key);
    }

    setTextBySelector(selector, key) {
        const el = document.querySelector(selector);
        if (el) el.textContent = this.t(key);
    }

    setPlaceholder(id, key) {
        const el = document.getElementById(id);
        if (el) el.placeholder = this.t(key);
    }
}

const i18nInstance = new I18n();
