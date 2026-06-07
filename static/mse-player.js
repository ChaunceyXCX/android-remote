class MSEPlayer {
    constructor(videoElement) {
        this.video = videoElement;
        this.mediaSource = null;
        this.sourceBuffer = null;
        this.ws = null;
        this.isConnected = false;
        this.width = 0;
        this.height = 0;
        this.onStateChange = null;
        this.onError = null;
        this.codec = 'video/mp4; codecs="avc1.42001f"';
        this.bufferQueue = [];
        this.isAppending = false;
    }

    async connect() {
        if (this.isConnected) return;

        try {
            this.mediaSource = new MediaSource();
            this.video.src = URL.createObjectURL(this.mediaSource);

            this.mediaSource.addEventListener('sourceopen', () => {
                console.log('MediaSource已打开');
                this.initSourceBuffer();
            });

            this.mediaSource.addEventListener('error', (error) => {
                console.error('MediaSource错误:', error);
                if (this.onError) this.onError(error);
            });

            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/api/screen/stream`;

            this.ws = new WebSocket(wsUrl);
            this.ws.binaryType = 'arraybuffer';

            this.ws.onopen = () => {
                console.log('MSE WebSocket连接已建立');
                this.isConnected = true;
                this.updateState('connected');
            };

            this.ws.onmessage = (event) => {
                this.handleVideoPacket(event.data);
            };

            this.ws.onclose = (event) => {
                console.log('MSE WebSocket连接关闭:', event.code, event.reason);
                this.isConnected = false;
                this.updateState('disconnected');
            };

            this.ws.onerror = (error) => {
                console.error('MSE WebSocket错误:', error);
                this.isConnected = false;
                this.updateState('error');
                if (this.onError) this.onError(error);
            };
        } catch (error) {
            console.error('MSE连接失败:', error);
            this.updateState('error');
            if (this.onError) this.onError(error);
        }
    }

    initSourceBuffer() {
        try {
            this.sourceBuffer = this.mediaSource.addSourceBuffer(this.codec);
            this.sourceBuffer.addEventListener('updateend', () => {
                this.isAppending = false;
                this.appendNextChunk();
            });

            this.sourceBuffer.addEventListener('error', (error) => {
                console.error('SourceBuffer错误:', error);
            });

            console.log('SourceBuffer已初始化');
        } catch (error) {
            console.error('初始化SourceBuffer失败:', error);
            if (this.onError) this.onError(error);
        }
    }

    handleVideoPacket(data) {
        const packet = new Uint8Array(data);

        if (packet.length < 12) {
            console.warn('视频包太小:', packet.length);
            return;
        }

        const header = this.parsePacketHeader(packet);
        const nalData = packet.slice(12);

        const fmp4Data = this.convertToFragmentedMP4(nalData, header.type === 0 || header.type === 1);

        if (fmp4Data) {
            this.appendToBuffer(fmp4Data);
        }
    }

    parsePacketHeader(packet) {
        const pts = (packet[0] << 56) | (packet[1] << 48) | (packet[2] << 40) | (packet[3] << 32) |
                    (packet[4] << 24) | (packet[5] << 16) | (packet[6] << 8) | packet[7];

        const type = (pts >> 62) & 0x03;
        const cleanPts = pts & 0x3FFFFFFFFFFFFFFF;

        const size = (packet[8] << 24) | (packet[9] << 16) | (packet[10] << 8) | packet[11];

        return {
            pts: cleanPts,
            type: type,
            size: size
        };
    }

    convertToFragmentedMP4(nalData, isKeyFrame) {
        if (!this.sourceBuffer) return null;

        try {
            const startCode = new Uint8Array([0x00, 0x00, 0x00, 0x01]);
            const fullNal = new Uint8Array(startCode.length + nalData.length);
            fullNal.set(startCode, 0);
            fullNal.set(nalData, startCode.length);

            return fullNal;
        } catch (error) {
            console.error('转换fMP4失败:', error);
            return null;
        }
    }

    appendToBuffer(data) {
        if (!this.sourceBuffer || this.sourceBuffer.updating) {
            this.bufferQueue.push(data);
            return;
        }

        try {
            this.isAppending = true;
            this.sourceBuffer.appendBuffer(data);
        } catch (error) {
            console.error('追加缓冲区失败:', error);
            this.isAppending = false;

            if (error.name === 'QuotaExceededError') {
                this.removeOldBuffers();
            }
        }
    }

    appendNextChunk() {
        if (this.bufferQueue.length > 0 && !this.isAppending) {
            const nextChunk = this.bufferQueue.shift();
            this.appendToBuffer(nextChunk);
        }
    }

    removeOldBuffers() {
        if (!this.sourceBuffer || this.sourceBuffer.updating) return;

        try {
            const buffered = this.sourceBuffer.buffered;
            if (buffered.length > 0) {
                const removeEnd = buffered.end(buffered.length - 1) - 10;
                if (removeEnd > 0) {
                    this.sourceBuffer.remove(0, removeEnd);
                }
            }
        } catch (error) {
            console.error('清理缓冲区失败:', error);
        }
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }

        if (this.video) {
            this.video.pause();
            this.video.src = '';
        }

        if (this.mediaSource && this.mediaSource.readyState === 'open') {
            try {
                this.mediaSource.endOfStream();
            } catch (error) {
                console.error('结束MediaSource失败:', error);
            }
        }

        this.isConnected = false;
        this.updateState('disconnected');
    }

    updateState(state) {
        if (this.onStateChange) {
            this.onStateChange(state);
        }
    }

    getState() {
        if (!this.ws) return 'disconnected';
        if (this.ws.readyState === WebSocket.CONNECTING) return 'connecting';
        if (this.ws.readyState === WebSocket.OPEN) return 'connected';
        return 'disconnected';
    }

    destroy() {
        this.disconnect();
        this.bufferQueue = [];
        this.sourceBuffer = null;
        this.mediaSource = null;
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = MSEPlayer;
}
