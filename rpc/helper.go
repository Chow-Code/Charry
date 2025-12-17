package rpc

import (
	"fmt"

	"github.com/charry/config"
	"github.com/charry/consul"
)

// ServerWithConsul 创建带 Consul 注册的 gRPC 服务器
type ServerWithConsul struct {
	*Server
	consulClient *consul.Client
	appConfig    *config.AppConfig
}

// NewServerWithConsul 创建 gRPC 服务器并注册到 Consul
// 这是一个便捷方法，整合了 gRPC 服务器创建和 Consul 注册
func NewServerWithConsul(appConfig *config.AppConfig, opts ...ServerOption) (*ServerWithConsul, error) {
	// 创建 gRPC 服务器
	server, err := NewServer(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC server: %w", err)
	}

	// 注册到 Consul
	consulClient, err := consul.RegisterFromEnv(appConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to register to Consul: %w", err)
	}

	return &ServerWithConsul{
		Server:       server,
		consulClient: consulClient,
		appConfig:    appConfig,
	}, nil
}

// GetConsulClient 获取 Consul 客户端
func (s *ServerWithConsul) GetConsulClient() *consul.Client {
	return s.consulClient
}

// Shutdown 优雅关闭服务器并从 Consul 注销
func (s *ServerWithConsul) Shutdown() {
	// 从 Consul 注销服务
	s.consulClient.GracefulShutdown(s.appConfig)

	// 关闭 gRPC 服务器
	s.Server.Stop()
}
