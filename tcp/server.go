package tcp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charry/config"
	"github.com/charry/logger"
)

// Server TCP 服务器
type Server struct {
	addr     string
	listener net.Listener

	// 连接管理
	conns   map[net.Conn]struct{}
	connsMu sync.RWMutex

	// 状态
	running atomic.Bool

	// 停止信号
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 处理器
	handler ConnectionHandler
}

// ConnectionHandler 连接处理器接口
type ConnectionHandler interface {
	// HandleConnection 处理新连接
	HandleConnection(conn net.Conn)
}

// DefaultHandler 默认处理器（支持协议解析和心跳）
type DefaultHandler struct{}

func (h *DefaultHandler) HandleConnection(conn net.Conn) {
	defer conn.Close()

	// 设置初始读超时（心跳3秒一次，给予足够余量）
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	for {
		// 解码消息
		msg, err := DecodeMsg(conn)
		if err != nil {
			// 读取失败，结束连接
			return
		}

		// 收到消息后，重置超时（30秒，心跳10秒一次足够）
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		// 处理消息
		switch v := msg.(type) {
		case *ClusterReqMsg:
			// 处理请求消息
			if IsHeartbeatMsg(v.Module, v.Cmd) {
				// 处理心跳请求
				HandleHeartbeatReq(conn, v)
			} else {
				// 处理业务请求（回显）
				resp := &ClusterRespMsg{
					Module:    v.Module,
					Cmd:       v.Cmd,
					SessionId: v.SessionId,
					Code:      0,
					Payload:   v.Payload,
				}
				data := EncodeClusterRespMsg(resp)
				conn.Write(data)
			}

		case *ClusterRespMsg:
			// 收到响应消息（客户端模式）
			logger.Infof("收到响应: module=%d, cmd=%d, sessionId=%s, code=%d",
				v.Module, v.Cmd, v.SessionId, v.Code)
		}
	}
}

// NewServer 创建 TCP 服务器
func NewServer(appConfig *config.AppConfig) (*Server, error) {
	addr := fmt.Sprintf("%s:%d", appConfig.Addr.Host, appConfig.Addr.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("创建 TCP 监听失败: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	server := &Server{
		addr:     addr,
		listener: listener,
		conns:    make(map[net.Conn]struct{}),
		ctx:      ctx,
		cancel:   cancel,
		handler:  &DefaultHandler{}, // 默认处理器
	}

	logger.Infof("TCP 服务器创建成功: %s", addr)
	return server, nil
}

// SetHandler 设置连接处理器
func (s *Server) SetHandler(handler ConnectionHandler) {
	s.handler = handler
}

// Start 启动服务器
func (s *Server) Start() error {
	if !s.running.CompareAndSwap(false, true) {
		return fmt.Errorf("服务器已在运行")
	}

	logger.Infof("TCP 服务器启动: %s", s.addr)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil // 正常关闭
			default:
				logger.Errorf("接受连接失败: %v", err)
				continue
			}
		}

		// 记录连接
		s.addConn(conn)

		// 处理连接
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer s.removeConn(conn)

			s.handler.HandleConnection(conn)
		}()
	}
}

// StartAsync 异步启动服务器
func (s *Server) StartAsync() {
	go func() {
		if err := s.Start(); err != nil {
			logger.Errorf("TCP 服务器运行错误: %v", err)
		}
	}()
}

// Stop 停止服务器
func (s *Server) Stop() {
	if !s.running.CompareAndSwap(true, false) {
		return // 已停止
	}

	logger.Info("停止 TCP 服务器...")

	// 取消 context
	s.cancel()

	// 关闭监听器
	if s.listener != nil {
		s.listener.Close()
	}

	// 关闭所有连接
	s.closeAllConns()

	// 等待所有处理协程结束
	s.wg.Wait()

	logger.Info("✓ TCP 服务器已停止")
}

// addConn 添加连接
func (s *Server) addConn(conn net.Conn) {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()
	s.conns[conn] = struct{}{}

	// 识别健康检查连接（来自 Consul），不打印日志
	if !isHealthCheckConn(conn) {
		logger.Infof("新连接: %s", conn.RemoteAddr())
	}
}

// removeConn 移除连接
func (s *Server) removeConn(conn net.Conn) {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()
	delete(s.conns, conn)

	// 健康检查连接不打印日志
	if !isHealthCheckConn(conn) {
		logger.Infof("连接断开: %s", conn.RemoteAddr())
	}
}

// isHealthCheckConn 判断是否为健康检查连接
// 通过来源 IP 判断（Consul 地址）
func isHealthCheckConn(conn net.Conn) bool {
	remoteAddr := conn.RemoteAddr().String()

	// 获取配置中的 Consul 地址
	cfg := config.Get()
	consulAddr := cfg.Consul.Address

	// 如果配置了 Consul 地址，检查是否来自该地址
	if consulAddr != "" {
		// 提取 IP（去除端口）
		host, _, _ := net.SplitHostPort(consulAddr)
		if host != "" && len(remoteAddr) > len(host) {
			return remoteAddr[:len(host)] == host
		}
	}

	return false
}

// closeAllConns 关闭所有连接
func (s *Server) closeAllConns() {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()

	for conn := range s.conns {
		conn.Close()
	}
	s.conns = make(map[net.Conn]struct{})
}

// GetAddr 获取监听地址
func (s *Server) GetAddr() string {
	return s.addr
}

// GetConnCount 获取当前连接数
func (s *Server) GetConnCount() int {
	s.connsMu.RLock()
	defer s.connsMu.RUnlock()
	return len(s.conns)
}
