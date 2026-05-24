class AndroidRemote {
    constructor() {
        this.isConnected = false;
        this.deviceIP = '';
        this.currentPanel = 0;
        this.totalPanels = 3;
        this.touchStartX = 0;
        this.touchStartY = 0;
        this.touchEndX = 0;
        this.isSwiping = false;
        this.init();
    }

    init() {
        this.bindElements();
        this.bindEvents();
        this.checkStatus();
        this.loadDeviceList();
        this.initDeviceListToggle();
        this.initConnectionToggle();
        this.initSwipe();
        this.initI18n();
        this.initMediaStatusPolling();
    }

    bindElements() {
        this.elements = {
            deviceIP: document.getElementById('deviceIP'),
            connectBtn: document.getElementById('connectBtn'),
            refreshDevicesBtn: document.getElementById('refreshDevicesBtn'),
            deviceStatus: document.getElementById('deviceStatus'),
            connectionSection: document.getElementById('connectionSection'),
            connectionHeader: document.getElementById('connectionHeader'),
            connectionStatusText: document.getElementById('connectionStatusText'),
            connectionStatusDetail: document.getElementById('connectionStatusDetail'),
            deviceDisplayName: document.getElementById('deviceDisplayName'),
            deviceModel: document.getElementById('deviceModel'),
            androidVersion: document.getElementById('androidVersion'),
            deviceIPDisplay: document.getElementById('deviceIPDisplay'),
            deviceList: document.getElementById('deviceList'),
            deviceListContent: document.getElementById('deviceListContent'),
            deviceListToggle: document.getElementById('deviceListToggle'),
            textInput: document.getElementById('textInput'),
            sendTextBtn: document.getElementById('sendTextBtn'),
            packageName: document.getElementById('packageName'),
            startAppBtn: document.getElementById('startAppBtn'),
            stopAppBtn: document.getElementById('stopAppBtn'),
            listAppsBtn: document.getElementById('listAppsBtn'),
            appList: document.getElementById('appList'),
            screenshotBtn: document.getElementById('screenshotBtn'),
            screenshotImage: document.getElementById('screenshotImage'),
            screenshotPlaceholder: document.getElementById('screenshotPlaceholder'),
            screenshotPreview: document.getElementById('screenshotPreview'),
            clickIndicator: document.getElementById('clickIndicator'),
            clickToTap: document.getElementById('clickToTap'),
            tapX: document.getElementById('tapX'),
            tapY: document.getElementById('tapY'),
            tapBtn: document.getElementById('tapBtn'),
            swipeX1: document.getElementById('swipeX1'),
            swipeY1: document.getElementById('swipeY1'),
            swipeX2: document.getElementById('swipeX2'),
            swipeY2: document.getElementById('swipeY2'),
            swipeBtn: document.getElementById('swipeBtn'),
            toast: document.getElementById('toast'),
            swipePanels: document.getElementById('swipePanels'),
            swipeDots: document.getElementById('swipeDots'),
            playPauseBtn: document.querySelector('[data-key="playpause"]')
        };
    }

    bindEvents() {
        this.elements.connectBtn.addEventListener('click', () => this.connectDevice());
        this.elements.refreshDevicesBtn.addEventListener('click', () => this.loadDeviceList());
        this.elements.sendTextBtn.addEventListener('click', () => this.sendText());
        this.elements.startAppBtn.addEventListener('click', () => this.startApp());
        this.elements.stopAppBtn.addEventListener('click', () => this.stopApp());
        this.elements.listAppsBtn.addEventListener('click', () => this.listApps());
        this.elements.screenshotBtn.addEventListener('click', () => this.takeScreenshot());
        this.elements.tapBtn.addEventListener('click', () => this.simulateTap());
        this.elements.swipeBtn.addEventListener('click', () => this.simulateSwipe());
        this.elements.screenshotPreview.addEventListener('click', (e) => this.handleScreenshotClick(e));

        this.elements.textInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.sendText();
        });

        this.elements.deviceIP.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.connectDevice();
        });

        document.querySelectorAll('[data-key]').forEach(btn => {
            btn.addEventListener('click', () => {
                const key = btn.getAttribute('data-key');
                this.sendKey(key);
            });
        });
    }

    initDeviceListToggle() {
        const toggle = this.elements.deviceListToggle;
        const deviceList = this.elements.deviceList;
        const header = deviceList.querySelector('.device-list-header');

        const savedState = localStorage.getItem('deviceListCollapsed');
        if (savedState === 'true') {
            deviceList.classList.add('collapsed');
        }

        const toggleHandler = () => {
            deviceList.classList.toggle('collapsed');
            const isCollapsed = deviceList.classList.contains('collapsed');
            localStorage.setItem('deviceListCollapsed', isCollapsed);
        };

        toggle.addEventListener('click', (e) => {
            e.stopPropagation();
            toggleHandler();
        });

        header.addEventListener('click', toggleHandler);
    }

    initConnectionToggle() {
        const section = this.elements.connectionSection;
        const header = this.elements.connectionHeader;
        if (!section || !header) return;

        const savedState = localStorage.getItem('connectionCollapsed');
        if (savedState === 'true') {
            section.classList.add('collapsed');
        }

        header.addEventListener('click', () => {
            section.classList.toggle('collapsed');
            const isCollapsed = section.classList.contains('collapsed');
            localStorage.setItem('connectionCollapsed', isCollapsed);
        });
    }

    initI18n() {
        const langBtn = document.getElementById('langToggle');
        if (langBtn) {
            langBtn.addEventListener('click', () => {
                const newLang = i18nInstance.toggleLanguage();
                langBtn.textContent = newLang === 'zh' ? 'EN' : '中';
            });
        }
        i18nInstance.updatePageTexts();
    }

    initMediaStatusPolling() {
        this.checkMediaStatus();
        setInterval(() => this.checkMediaStatus(), 3000);
    }

    async checkMediaStatus() {
        try {
            const response = await fetch('/api/media/status');
            const data = await response.json();
            this.updatePlayPauseIcon(data.playing);
        } catch (error) {
        }
    }

    updatePlayPauseIcon(isPlaying) {
        const btn = this.elements.playPauseBtn;
        if (!btn) return;
        
        const icon = btn.querySelector('i');
        if (!icon) return;

        if (isPlaying) {
            icon.className = 'fas fa-pause';
        } else {
            icon.className = 'fas fa-play';
        }
    }

    initSwipe() {
        if (!this.elements.swipePanels) return;

        const isMobile = () => window.innerWidth < 768;

        this.elements.swipePanels.addEventListener('touchstart', (e) => {
            if (!isMobile()) return;
            this.touchStartX = e.changedTouches[0].screenX;
            this.touchStartY = e.changedTouches[0].screenY;
            this.isSwiping = true;
        }, { passive: true });

        this.elements.swipePanels.addEventListener('touchmove', (e) => {
            if (!isMobile() || !this.isSwiping) return;
            const diffX = Math.abs(e.changedTouches[0].screenX - this.touchStartX);
            const diffY = Math.abs(e.changedTouches[0].screenY - this.touchStartY);
            if (diffX > diffY) {
                e.preventDefault();
            }
        }, { passive: false });

        this.elements.swipePanels.addEventListener('touchend', (e) => {
            if (!isMobile() || !this.isSwiping) return;
            this.touchEndX = e.changedTouches[0].screenX;
            this.isSwiping = false;
            this.handleSwipe();
        }, { passive: true });

        if (this.elements.swipeDots) {
            this.elements.swipeDots.addEventListener('click', (e) => {
                const dot = e.target.closest('.swipe-dot');
                if (dot) {
                    const panelIndex = parseInt(dot.dataset.dot);
                    this.goToPanel(panelIndex);
                }
            });
        }

        this.elements.swipePanels.addEventListener('scroll', () => {
            if (!isMobile()) return;
            const scrollLeft = this.elements.swipePanels.scrollLeft;
            const panelWidth = this.elements.swipePanels.clientWidth;
            if (panelWidth > 0) {
                const newPanel = Math.round(scrollLeft / panelWidth);
                if (newPanel !== this.currentPanel) {
                    this.currentPanel = newPanel;
                    this.updateDots();
                }
            }
        }, { passive: true });
    }

    handleSwipe() {
        const diff = this.touchStartX - this.touchEndX;
        const threshold = 50;

        if (Math.abs(diff) < threshold) return;

        if (diff > 0 && this.currentPanel < this.totalPanels - 1) {
            this.goToPanel(this.currentPanel + 1);
        } else if (diff < 0 && this.currentPanel > 0) {
            this.goToPanel(this.currentPanel - 1);
        }
    }

    goToPanel(index) {
        if (index < 0 || index >= this.totalPanels) return;
        this.currentPanel = index;
        const panelWidth = this.elements.swipePanels.clientWidth;
        this.elements.swipePanels.scrollTo({
            left: index * panelWidth,
            behavior: 'smooth'
        });
        this.updateDots();
    }

    updateDots() {
        if (!this.elements.swipeDots) return;
        const dots = this.elements.swipeDots.querySelectorAll('.swipe-dot');
        dots.forEach((dot, i) => {
            dot.classList.toggle('active', i === this.currentPanel);
        });
    }

    async checkStatus() {
        try {
            const response = await fetch('/api/status');
            const data = await response.json();
            this.updateDeviceStatus(data);
        } catch (error) {
            console.error('状态检查失败:', error);
        }
    }

    updateDeviceStatus(data) {
        this.isConnected = data.connected;
        const statusDot = this.elements.deviceStatus.querySelector('.status-dot');
        const statusText = this.elements.deviceStatus.querySelector('.status-text');

        if (data.connected) {
            statusDot.classList.add('connected');
            statusText.textContent = i18nInstance.t('connected');
            if (this.elements.connectionStatusText) {
                const mainText = data.note || data.model || '-';
                const detailText = data.ip || '';
                this.elements.connectionStatusText.textContent = mainText;
                this.elements.connectionStatusText.style.color = 'var(--success)';
                if (this.elements.connectionStatusDetail) {
                    this.elements.connectionStatusDetail.textContent = detailText;
                }
            }
            const displayName = data.note || data.model || data.deviceId || '-';
            this.elements.deviceDisplayName.textContent = displayName;
            this.elements.deviceModel.textContent = data.model || '-';
            this.elements.androidVersion.textContent = data.android || '-';
            this.elements.deviceIPDisplay.textContent = data.ip || '-';
        } else {
            statusDot.classList.remove('connected');
            statusText.textContent = i18nInstance.t('disconnected');
            if (this.elements.connectionStatusText) {
                this.elements.connectionStatusText.textContent = i18nInstance.t('disconnected');
                this.elements.connectionStatusText.style.color = 'var(--text-secondary)';
                if (this.elements.connectionStatusDetail) {
                    this.elements.connectionStatusDetail.textContent = '';
                }
            }
            this.elements.deviceDisplayName.textContent = '-';
            this.elements.deviceModel.textContent = '-';
            this.elements.androidVersion.textContent = '-';
            this.elements.deviceIPDisplay.textContent = '-';
        }
    }

    async connectDevice() {
        const ip = this.elements.deviceIP.value.trim();
        if (!ip) {
            this.showToast(i18nInstance.t('enterIP'), 'error');
            return;
        }

        try {
            const response = await fetch('/api/connect', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ip })
            });
            const data = await response.json();

            if (data.success) {
                this.showToast(data.message, 'success');
                this.elements.deviceIP.value = '';
                setTimeout(() => {
                    this.checkStatus();
                    this.loadDeviceList();
                }, 1000);
            } else {
                this.showToast(data.error || i18nInstance.t('connectionFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('connectionRequestFailed') + ': ' + error.message, 'error');
        }
    }

    async sendKey(keyName) {
        try {
            const response = await fetch(`/api/key/${keyName}`, {
                method: 'POST'
            });
            const data = await response.json();

            if (data.success) {
                this.showToast(data.message, 'success');
            } else {
                this.showToast(data.error || i18nInstance.t('keySendFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('keyRequestFailed') + ': ' + error.message, 'error');
        }
    }

    async sendText() {
        const text = this.elements.textInput.value.trim();
        if (!text) {
            this.showToast(i18nInstance.t('enterText'), 'error');
            return;
        }

        try {
            const response = await fetch('/api/input/text', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ text })
            });
            const data = await response.json();

            if (data.success) {
                this.showToast(data.message, 'success');
                this.elements.textInput.value = '';
            } else {
                this.showToast(data.error || i18nInstance.t('textSendFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('textRequestFailed') + ': ' + error.message, 'error');
        }
    }

    async startApp() {
        const packageName = this.elements.packageName.value.trim();
        if (!packageName) {
            this.showToast(i18nInstance.t('enterPackage'), 'error');
            return;
        }

        try {
            const response = await fetch('/api/app/start', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ package: packageName })
            });
            const data = await response.json();

            if (data.success) {
                this.showToast(data.message, 'success');
            } else {
                this.showToast(data.error || i18nInstance.t('appStartFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('appStartRequestFailed') + ': ' + error.message, 'error');
        }
    }

    async stopApp() {
        const packageName = this.elements.packageName.value.trim();
        if (!packageName) {
            this.showToast(i18nInstance.t('enterPackage'), 'error');
            return;
        }

        try {
            const response = await fetch('/api/app/stop', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ package: packageName })
            });
            const data = await response.json();

            if (data.success) {
                this.showToast(data.message, 'success');
            } else {
                this.showToast(data.error || i18nInstance.t('appStopFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('appStopRequestFailed') + ': ' + error.message, 'error');
        }
    }

    async listApps() {
        try {
            const response = await fetch('/api/app/list');
            const data = await response.json();

            if (data.success && data.apps) {
                this.renderAppList(data.apps);
            } else {
                this.showToast(data.error || i18nInstance.t('getAppListFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('getAppListRequestFailed') + ': ' + error.message, 'error');
        }
    }

    renderAppList(apps) {
        this.elements.appList.innerHTML = '';
        
        if (apps.length === 0) {
            this.elements.appList.innerHTML = `<div class="app-item">${i18nInstance.t('noAppFound')}</div>`;
            return;
        }

        apps.forEach(app => {
            const item = document.createElement('div');
            item.className = 'app-item';
            const displayName = app.name || app.package;
            const showPackage = app.name && app.name !== app.package;
            item.innerHTML = `
                <div class="app-info">
                    <span class="app-name">${displayName}</span>
                    ${showPackage ? `<span class="app-package">${app.package}</span>` : ''}
                </div>
            `;
            item.addEventListener('click', () => {
                this.elements.packageName.value = app.package;
            });
            this.elements.appList.appendChild(item);
        });
    }

    async takeScreenshot() {
        this.elements.screenshotBtn.disabled = true;
        this.elements.screenshotBtn.innerHTML = `<i class="fas fa-spinner fa-spin"></i> ${i18nInstance.t('screenshoting')}`;

        try {
            const response = await fetch('/api/system/screenshot');
            const data = await response.json();

            if (data.success && data.imageUrl) {
                this.elements.screenshotImage.src = data.imageUrl + '?t=' + Date.now();
                this.elements.screenshotImage.style.display = 'block';
                this.elements.screenshotPlaceholder.style.display = 'none';
                this.showToast(i18nInstance.t('screenshotSuccess'), 'success');
            } else {
                this.showToast(data.error || i18nInstance.t('screenshotFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('screenshotRequestFailed') + ': ' + error.message, 'error');
        } finally {
            this.elements.screenshotBtn.disabled = false;
            this.elements.screenshotBtn.innerHTML = `<i class="fas fa-camera"></i> ${i18nInstance.t('takeScreenshot')}`;
        }
    }

    handleScreenshotClick(e) {
        if (!this.elements.screenshotImage.src || this.elements.screenshotImage.style.display === 'none') {
            return;
        }

        const img = this.elements.screenshotImage;
        const rect = img.getBoundingClientRect();
        
        const imgNaturalWidth = img.naturalWidth;
        const imgNaturalHeight = img.naturalHeight;
        const displayWidth = rect.width;
        const displayHeight = rect.height;
        
        const clickX = e.clientX - rect.left;
        const clickY = e.clientY - rect.top;
        
        const deviceX = Math.round(clickX * (imgNaturalWidth / displayWidth));
        const deviceY = Math.round(clickY * (imgNaturalHeight / displayHeight));

        this.showClickIndicator(clickX, clickY);

        if (this.elements.clickToTap.checked) {
            this.executeTap(deviceX, deviceY);
        } else {
            this.elements.tapX.value = deviceX;
            this.elements.tapY.value = deviceY;
            this.showToast(`${i18nInstance.t('coordFilled')}: (${deviceX}, ${deviceY})`, 'info');
        }
    }

    showClickIndicator(x, y) {
        const indicator = this.elements.clickIndicator;
        indicator.style.left = x + 'px';
        indicator.style.top = y + 'px';
        indicator.classList.add('active');
        
        setTimeout(() => {
            indicator.classList.remove('active');
        }, 500);
    }

    async executeTap(x, y) {
        try {
            const response = await fetch('/api/system/tap', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ x, y })
            });
            const data = await response.json();

            if (data.success) {
                this.showToast(`${i18nInstance.t('tapAt')} (${x}, ${y})`, 'success');
                setTimeout(() => this.takeScreenshot(), 300);
            } else {
                this.showToast(data.error || i18nInstance.t('tapFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('tapRequestFailed') + ': ' + error.message, 'error');
        }
    }

    async simulateTap() {
        const x = parseInt(this.elements.tapX.value);
        const y = parseInt(this.elements.tapY.value);

        if (isNaN(x) || isNaN(y)) {
            this.showToast(i18nInstance.t('enterValidCoords'), 'error');
            return;
        }

        try {
            const response = await fetch('/api/system/tap', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ x, y })
            });
            const data = await response.json();

            if (data.success) {
                this.showToast(data.message, 'success');
            } else {
                this.showToast(data.error || i18nInstance.t('tapFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('tapRequestFailed') + ': ' + error.message, 'error');
        }
    }

    async simulateSwipe() {
        const x1 = parseInt(this.elements.swipeX1.value);
        const y1 = parseInt(this.elements.swipeY1.value);
        const x2 = parseInt(this.elements.swipeX2.value);
        const y2 = parseInt(this.elements.swipeY2.value);

        if (isNaN(x1) || isNaN(y1) || isNaN(x2) || isNaN(y2)) {
            this.showToast(i18nInstance.t('enterValidCoords'), 'error');
            return;
        }

        try {
            const response = await fetch('/api/system/swipe', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ x1, y1, x2, y2 })
            });
            const data = await response.json();

            if (data.success) {
                this.showToast(data.message, 'success');
            } else {
                this.showToast(data.error || i18nInstance.t('swipeFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('swipeRequestFailed') + ': ' + error.message, 'error');
        }
    }

    async loadDeviceList() {
        try {
            const response = await fetch('/api/devices');
            const data = await response.json();

            if (data.success && data.devices) {
                this.renderDeviceList(data.devices);
            } else {
                this.elements.deviceListContent.innerHTML = `<div class="device-item-empty">${i18nInstance.t('noDevices')}</div>`;
            }
        } catch (error) {
            console.error('加载设备列表失败:', error);
            this.elements.deviceListContent.innerHTML = `<div class="device-item-empty">${i18nInstance.t('noDevices')}</div>`;
        }
    }

    renderDeviceList(devices) {
        if (devices.length === 0) {
            this.elements.deviceListContent.innerHTML = `<div class="device-item-empty">${i18nInstance.t('noDevices')}</div>`;
            return;
        }

        this.elements.deviceListContent.innerHTML = '';
        
        devices.forEach(device => {
            const item = document.createElement('div');
            item.className = `device-item ${device.selected ? 'selected' : ''}`;
            
            let switchModeBtn = '';
            if (device.selected) {
                if (device.type === 'usb' && device.ip) {
                    switchModeBtn = `<button class="switch-mode-btn to-wireless" 
                                             data-device-id="${device.id}" data-mode="wireless"
                                             title="${i18nInstance.t('toWireless')}">
                                        ${i18nInstance.t('toWireless')}
                                     </button>`;
                } else if (device.type === 'wireless') {
                    switchModeBtn = `<button class="switch-mode-btn to-usb" 
                                             data-device-id="${device.id}" data-mode="usb"
                                             title="${i18nInstance.t('toUSB')}">
                                        ${i18nInstance.t('toUSB')}
                                     </button>`;
                }
            }

            const displayName = device.note || device.model || device.id;
            const showModel = device.note && device.model && device.model !== device.id;

            item.innerHTML = `
                <div class="device-item-left">
                    <div class="device-item-info">
                        <div class="device-item-name">${displayName}</div>
                        ${showModel ? `<div class="device-item-model">${device.model}</div>` : ''}
                        <div class="device-item-id">${device.id}</div>
                    </div>
                </div>
                <div class="device-item-right">
                    <span class="device-type-badge ${device.type}">${device.type === 'usb' ? 'USB' : i18nInstance.t('toWireless').replace('→ ', '')}</span>
                    ${switchModeBtn}
                    <input type="text" class="device-note-input" placeholder="${i18nInstance.t('addNote')}" 
                           value="${device.note || ''}" data-device-id="${device.id}">
                </div>
            `;
            
            const noteInput = item.querySelector('.device-note-input');
            noteInput.addEventListener('click', (e) => e.stopPropagation());
            noteInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.saveDeviceNote(device.id, noteInput.value);
                }
            });
            noteInput.addEventListener('blur', () => {
                this.saveDeviceNote(device.id, noteInput.value);
            });

            const switchModeBtnEl = item.querySelector('.switch-mode-btn');
            if (switchModeBtnEl) {
                switchModeBtnEl.addEventListener('click', (e) => {
                    e.stopPropagation();
                    this.switchDeviceMode(switchModeBtnEl.dataset.deviceId, switchModeBtnEl.dataset.mode);
                });
            }
            
            item.addEventListener('click', () => this.switchDevice(device.id));
            this.elements.deviceListContent.appendChild(item);
        });
    }

    async switchDevice(deviceId) {
        try {
            const response = await fetch('/api/device/switch', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ deviceId })
            });
            const data = await response.json();

            if (data.success) {
                const deviceItem = document.querySelector(`[data-device-id="${deviceId}"]`);
                const deviceCard = deviceItem?.closest('.device-item');
                const noteEl = deviceCard?.querySelector('.device-item-name');
                const note = noteEl?.textContent || deviceId;
                this.showToast(`${i18nInstance.t('switchedToDevice')}: ${note}`, 'success');
                this.loadDeviceList();
                setTimeout(() => this.checkStatus(), 500);
            } else {
                this.showToast(data.error || i18nInstance.t('switchFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('switchRequestFailed') + ': ' + error.message, 'error');
        }
    }

    async switchDeviceMode(deviceId, mode) {
        const modeText = mode === 'wireless' ? i18nInstance.t('switchingToWireless') : i18nInstance.t('switchingToUSB');
        this.showToast(modeText, 'info');
        
        try {
            const response = await fetch('/api/device/switch-mode', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ deviceId, mode })
            });
            const data = await response.json();

            if (data.success) {
                this.showToast(data.message, 'success');
                setTimeout(() => this.loadDeviceList(), 1000);
            } else {
                this.showToast(data.error || i18nInstance.t('switchModeFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('switchRequestFailed') + ': ' + error.message, 'error');
        }
    }

    async saveDeviceNote(deviceId, note) {
        try {
            const response = await fetch('/api/device/note', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ deviceId, note })
            });
            const data = await response.json();

            if (data.success) {
                this.showToast(i18nInstance.t('noteSaved'), 'success');
            } else {
                this.showToast(data.error || i18nInstance.t('noteSaveFailed'), 'error');
            }
        } catch (error) {
            this.showToast(i18nInstance.t('noteSaveRequestFailed') + ': ' + error.message, 'error');
        }
    }

    showToast(message, type = 'info') {
        const toast = this.elements.toast;
        toast.textContent = message;
        toast.className = 'toast show ' + type;
        
        setTimeout(() => {
            toast.className = 'toast';
        }, 3000);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new AndroidRemote();
});
