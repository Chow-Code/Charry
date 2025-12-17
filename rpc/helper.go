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
// rpcConfig: RPC 配置（可选，传 nil 则使用默认配置）
// appConfig: 应用配置（使用其中的 Addr，并注册到 Consul）
func NewServerWithConsul(rpcConfig *RpcConfig, appConfig *config.AppConfig) (*ServerWithConsul, error) {
	// 创建 gRPC 服务器
	server, err := NewServer(rpcConfig, appConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 gRPC 服务器失败: %w", err)
	}

	// 注册到 Consul
	consulClient, err := consul.RegisterFromEnv(appConfig)
	if err != nil {
		return nil, fmt.Errorf("注册到 Consul 失败: %w", err)
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
