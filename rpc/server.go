package rpc

import (
	"fmt"
	"net"

	"github.com/charry/config"
	"github.com/charry/logger"
	"google.golang.org/grpc"
)

// Server gRPC 服务器封装
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	addr       config.Addr
}

// NewServer 创建一个新的 gRPC 服务器
// rpcConfig: RPC 配置（可选，传 nil 则使用默认配置）
// appConfig: 应用配置（使用其中的 Addr）
func NewServer(rpcConfig *RpcConfig, appConfig *config.AppConfig) (*Server, error) {
	if appConfig == nil {
		return nil, fmt.Errorf("appConfig is nil")
	}

	// 使用默认配置（如果未提供）
	if rpcConfig == nil {
		rpcConfig = NewDefaultRpcConfig()
	}

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer(rpcConfig.GrpcOptions...)

	// 创建监听器
	addr := fmt.Sprintf("%s:%d", appConfig.Addr.Host, appConfig.Addr.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("监听地址 %s 失败: %w", addr, err)
	}

	return &Server{
		grpcServer: grpcServer,
		listener:   lis,
		addr:       appConfig.Addr,
	}, nil
}

// GetGRPCServer 获取原生 gRPC 服务器
// 用于注册您的业务服务
func (s *Server) GetGRPCServer() *grpc.Server {
	return s.grpcServer
}

// Start 启动 gRPC 服务器
func (s *Server) Start() error {
	if err := s.grpcServer.Serve(s.listener); err != nil {
		return fmt.Errorf("启动服务失败: %w", err)
	}

	return nil
}

// StartAsync 异步启动 gRPC 服务器
func (s *Server) StartAsync() {
	go func() {
		if err := s.Start(); err != nil {
			logger.Errorf("gRPC 服务器错误: %v", err)
		}
	}()
}

// Stop 停止 gRPC 服务器（优雅关闭）
func (s *Server) Stop() {
	logger.Info("正在停止 gRPC 服务器...")
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
	return s.addr.Port
}

// GetAddr 获取服务器监听地址配置
func (s *Server) GetAddr() config.Addr {
	return s.addr
}
