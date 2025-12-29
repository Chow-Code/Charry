package cluster

import (
	"fmt"
	"sync"

	"github.com/charry/logger"
	"github.com/charry/tcp"
)

// MessageHandler 消息处理器
type MessageHandler func(payload []byte) error

// Router 消息路由器
type Router struct {
	// 路由表：(module << 32 | cmd) -> handler
	handlers map[uint64]MessageHandler
	mu       sync.RWMutex
}

// NewRouter 创建路由器
func NewRouter() *Router {
	return &Router{
		handlers: make(map[uint64]MessageHandler),
	}
}

// Register 注册消息处理器
func (r *Router) Register(module, cmd uint32, handler MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := makeRouteKey(module, cmd)
	r.handlers[key] = handler
	logger.Infof("注册消息处理器: module=%d, cmd=%d", module, cmd)
}

// Unregister 注销消息处理器
func (r *Router) Unregister(module, cmd uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := makeRouteKey(module, cmd)
	delete(r.handlers, key)
}

// Handle 处理消息
func (r *Router) Handle(module, cmd uint32, payload []byte) error {
	r.mu.RLock()
	key := makeRouteKey(module, cmd)
	handler, exists := r.handlers[key]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("未注册的消息: module=%d, cmd=%d", module, cmd)
	}

	return handler(payload)
}

// HandleReq 处理请求消息
func (r *Router) HandleReq(req *tcp.ClusterReqMsg) error {
	return r.Handle(req.Module, req.Cmd, req.Payload)
}

// HandleResp 处理响应消息
func (r *Router) HandleResp(resp *tcp.ClusterRespMsg) error {
	return r.Handle(resp.Module, resp.Cmd, resp.Payload)
}

// makeRouteKey 生成路由键
func makeRouteKey(module, cmd uint32) uint64 {
	return (uint64(module) << 32) | uint64(cmd)
}

