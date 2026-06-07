package main

import "sync"

// VideoHub 广播视频包到多个 WebSocket 客户端
type VideoHub struct {
	clients map[chan *VideoPacket]struct{}
	mu      sync.RWMutex
}

func NewVideoHub() *VideoHub {
	return &VideoHub{
		clients: make(map[chan *VideoPacket]struct{}),
	}
}

func (h *VideoHub) Broadcast(packet *VideoPacket) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- packet:
		default:
		}
	}
}

func (h *VideoHub) Subscribe() chan *VideoPacket {
	ch := make(chan *VideoPacket, 100)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[ch] = struct{}{}
	return ch
}

func (h *VideoHub) Unsubscribe(ch chan *VideoPacket) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, ch)
	close(ch)
}
