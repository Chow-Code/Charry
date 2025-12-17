package rpc

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

// Server gRPC 服务器封装
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	options    *ServerOptions
}

// NewServer 创建一个新的 gRPC 服务器
func NewServer(opts ...ServerOption) (*Server, error) {
	options := defaultServerOptions()
	for _, opt := range opts {
		opt(options)
	}

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer(options.grpcOptions...)

	// 创建监听器
	addr := fmt.Sprintf("%s:%d", options.host, options.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	return &Server{
		grpcServer: grpcServer,
		listener:   lis,
		options:    options,
	}, nil
}

// GetGRPCServer 获取原生 gRPC 服务器
// 用于注册您的业务服务
func (s *Server) GetGRPCServer() *grpc.Server {
	return s.grpcServer
}

// Start 启动 gRPC 服务器
func (s *Server) Start() error {
	addr := s.listener.Addr().String()
	log.Printf("Starting gRPC server on %s", addr)

	if err := s.grpcServer.Serve(s.listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// StartAsync 异步启动 gRPC 服务器
func (s *Server) StartAsync() {
	go func() {
		if err := s.Start(); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()
}

// Stop 停止 gRPC 服务器（优雅关闭）
func (s *Server) Stop() {
	log.Println("Stopping gRPC server...")
	s.grpcServer.GracefulStop()
}

// Shutdown Stop 的别名，优雅关闭服务器
func (s *Server) Shutdown() {
	s.Stop()
}

// GetAddress 获取服务器监听地址
func (s *Server) GetAddress() string {
	return s.listener.Addr().String()
}

// GetPort 获取服务器监听端口
func (s *Server) GetPort() int {
	return s.options.port
}

