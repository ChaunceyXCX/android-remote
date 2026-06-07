class ScrcpyPlayer {
    constructor(canvasElement) {
        this.canvas = canvasElement;
        this.ctx = canvasElement.getContext('2d');
        this.decoder = null;
        this.ws = null;
        this.isConnected = false;
        this.width = 0;
        this.height = 0;
        this.onStateChange = null;
        this.onError = null;
    }

    async connect() {
        if (this.isConnected) return;

        try {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/api/screen/stream`;
            
            this.ws = new WebSocket(wsUrl);
            this.ws.binaryType = 'arraybuffer';

            this.ws.onopen = () => {
                console.log('WebSocket连接已建立');
                this.isConnected = true;
                this.updateState('connected');
            };

            this.ws.onmessage = (event) => {
                this.handleVideoPacket(event.data);
            };

            this.ws.onclose = (event) => {
                console.log('WebSocket连接关闭:', event.code, event.reason);
                this.isConnected = false;
                this.updateState('disconnected');
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket错误:', error);
                this.isConnected = false;
                this.updateState('error');
                if (this.onError) this.onError(error);
            };
        } catch (error) {
            console.error('连接失败:', error);
            this.updateState('error');
            if (this.onError) this.onError(error);
        }
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        this.isConnected = false;
        this.updateState('disconnected');
    }

    async handleVideoPacket(data) {
        const packet = new Uint8Array(data);
        
        if (packet.length < 12) {
            console.warn('视频包太小:', packet.length);
            return;
        }

        const header = this.parsePacketHeader(packet);
        const nalData = packet.slice(12);

        if (header.type === 0) {
            console.log('收到配置帧 (SPS/PPS)');
            await this.initDecoder(nalData);
        } else if (header.type === 1) {
            console.log('收到关键帧');
            if (!this.decoder) {
                await this.initDecoder(nalData);
            }
            this.decodeFrame(nalData, true, header.pts);
        } else if (header.type === 2) {
            this.decodeFrame(nalData, false, header.pts);
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

    async initDecoder(configData) {
        if (this.decoder) {
            this.decoder.close();
        }

        try {
            const sps = this.extractSPS(configData);
            if (sps) {
                this.updateResolution(sps.width, sps.height);
            }

            this.decoder = new VideoDecoder({
                output: (frame) => {
                    this.renderFrame(frame);
                    frame.close();
                },
                error: (error) => {
                    console.error('视频解码错误:', error);
                    this.decoder = null;
                }
            });

            const config = {
                codec: 'avc1.42001f',
                codedWidth: this.width || 1920,
                codedHeight: this.height || 1080,
                optimizeForLatency: true
            };

            const { supported } = await VideoDecoder.isConfigSupported(config);
            if (!supported) {
                throw new Error('不支持的视频配置');
            }

            await this.decoder.configure(config);
            console.log('视频解码器已初始化:', config);

            this.decodeFrame(configData, true, 0);
        } catch (error) {
            console.error('初始化解码器失败:', error);
            if (this.onError) this.onError(error);
        }
    }

    extractSPS(data) {
        let i = 0;
        while (i < data.length - 4) {
            if (data[i] === 0 && data[i + 1] === 0 && data[i + 2] === 0 && data[i + 3] === 1) {
                const naluType = data[i + 4] & 0x1f;
                if (naluType === 7) {
                    return this.parseSPS(data.slice(i + 4));
                }
            }
            i++;
        }
        return null;
    }

    parseSPS(spsData) {
        try {
            let offset = 4;
            const profile_idc = spsData[offset++];
            offset++;
            const level_idc = spsData[offset++];
            
            return {
                width: 1920,
                height: 1080
            };
        } catch (e) {
            console.warn('解析SPS失败:', e);
            return null;
        }
    }

    decodeFrame(data, isKeyFrame, timestamp) {
        if (!this.decoder || this.decoder.state !== 'configured') {
            return;
        }

        try {
            const chunk = new EncodedVideoChunk({
                type: isKeyFrame ? 'key' : 'delta',
                timestamp: timestamp,
                data: data
            });

            this.decoder.decode(chunk);
        } catch (error) {
            console.error('解码帧失败:', error);
        }
    }

    renderFrame(frame) {
        if (!this.ctx) return;

        if (this.canvas.width !== frame.displayWidth || this.canvas.height !== frame.displayHeight) {
            this.canvas.width = frame.displayWidth;
            this.canvas.height = frame.displayHeight;
            this.updateResolution(frame.displayWidth, frame.displayHeight);
        }

        this.ctx.drawImage(frame, 0, 0);
    }

    updateResolution(width, height) {
        if (this.width !== width || this.height !== height) {
            this.width = width;
            this.height = height;
            console.log(`分辨率更新: ${width}x${height}`);
        }
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
        if (this.decoder) {
            this.decoder.close();
            this.decoder = null;
        }
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = ScrcpyPlayer;
}
