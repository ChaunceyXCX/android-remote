class AdaptivePlayer {
    constructor(containerElement) {
        this.container = containerElement;
        this.player = null;
        this.playerType = null;
        this.canvas = null;
        this.video = null;
        this.onStateChange = null;
        this.onError = null;
        this.width = 0;
        this.height = 0;

        this.initElements();
    }

    initElements() {
        this.canvas = document.createElement('canvas');
        this.canvas.style.display = 'none';
        this.canvas.style.width = '100%';
        this.canvas.style.height = '100%';
        this.container.appendChild(this.canvas);

        this.video = document.createElement('video');
        this.video.style.display = 'none';
        this.video.style.width = '100%';
        this.video.style.height = '100%';
        this.video.autoplay = true;
        this.video.muted = true;
        this.video.playsInline = true;
        this.container.appendChild(this.video);
    }

    async connect() {
        const bestPlayer = this.detectBestPlayer();
        console.log('选择播放器类型:', bestPlayer);

        try {
            if (bestPlayer === 'webcodecs') {
                await this.connectWithWebCodecs();
            } else if (bestPlayer === 'mse') {
                await this.connectWithMSE();
            } else {
                throw new Error('没有可用的播放器');
            }
        } catch (error) {
            console.error(`${bestPlayer} 连接失败，尝试降级:`, error);
            await this.tryFallback(bestPlayer);
        }
    }

    detectBestPlayer() {
        if (typeof VideoDecoder !== 'undefined' && typeof EncodedVideoChunk !== 'undefined') {
            return 'webcodecs';
        }

        if (typeof MediaSource !== 'undefined' && MediaSource.isTypeSupported('video/mp4; codecs="avc1.42001f"')) {
            return 'mse';
        }

        return null;
    }

    async connectWithWebCodecs() {
        const { ScrcpyPlayer } = await this.loadModule('scrcpy-player.js');
        this.player = new ScrcpyPlayer(this.canvas);
        this.playerType = 'webcodecs';

        this.canvas.style.display = 'block';
        this.video.style.display = 'none';

        this.player.onStateChange = (state) => {
            this.updateState(state);
        };
        this.player.onError = (error) => {
            if (this.onError) this.onError(error);
        };

        await this.player.connect();
    }

    async connectWithMSE() {
        const { MSEPlayer } = await this.loadModule('mse-player.js');
        this.player = new MSEPlayer(this.video);
        this.playerType = 'mse';

        this.canvas.style.display = 'none';
        this.video.style.display = 'block';

        this.player.onStateChange = (state) => {
            this.updateState(state);
        };
        this.player.onError = (error) => {
            if (this.onError) this.onError(error);
        };

        await this.player.connect();
    }

    async tryFallback(failedPlayerType) {
        if (failedPlayerType === 'webcodecs') {
            console.log('WebCodecs 失败，降级到 MSE');
            try {
                await this.connectWithMSE();
                return;
            } catch (error) {
                console.error('MSE 也失败:', error);
            }
        }

        console.log('所有播放器都失败');
        this.updateState('error');
        if (this.onError) this.onError(new Error('没有可用的播放器'));
    }

    loadModule(src) {
        return new Promise((resolve, reject) => {
            const script = document.createElement('script');
            script.src = src;
            script.onload = () => {
                resolve(window);
            };
            script.onerror = () => {
                reject(new Error(`加载 ${src} 失败`));
            };
            document.head.appendChild(script);
        });
    }

    disconnect() {
        if (this.player) {
            this.player.destroy();
            this.player = null;
        }
        this.playerType = null;
        this.updateState('disconnected');
    }

    updateState(state) {
        if (this.onStateChange) {
            this.onStateChange(state);
        }
    }

    getState() {
        if (!this.player) return 'disconnected';
        return this.player.getState();
    }

    getPlayerType() {
        return this.playerType;
    }

    destroy() {
        this.disconnect();
        if (this.canvas && this.canvas.parentNode) {
            this.canvas.parentNode.removeChild(this.canvas);
        }
        if (this.video && this.video.parentNode) {
            this.video.parentNode.removeChild(this.video);
        }
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = AdaptivePlayer;
}
