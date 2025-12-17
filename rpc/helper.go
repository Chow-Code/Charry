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
	config       *config.Config
}

// NewServerWithConsul 创建 gRPC 服务器并注册到 Consul
// 注意：推荐使用 app.StartUp() 统一启动流程，而不是直接调用此方法
// cfg: 完整配置（包含 App 和 Consul）
// rpcConfig: RPC 配置（可选，传 nil 则使用默认配置）
func NewServerWithConsul(cfg *config.Config, rpcConfig *RpcConfig) (*ServerWithConsul, error) {
	// 创建 gRPC 服务器
	server, err := NewServer(rpcConfig, &cfg.App)
	if err != nil {
		return nil, fmt.Errorf("创建 gRPC 服务器失败: %w", err)
	}

	// 注册到 Consul
	consulClient, err := consul.RegisterFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("注册到 Consul 失败: %w", err)
	}

	return &ServerWithConsul{
		Server:       server,
		consulClient: consulClient,
		config:       cfg,
	}, nil
}

// GetConsulClient 获取 Consul 客户端
func (s *ServerWithConsul) GetConsulClient() *consul.Client {
	return s.consulClient
}

// Shutdown 优雅关闭服务器并从 Consul 注销
func (s *ServerWithConsul) Shutdown() {
	// 从 Consul 注销服务
	s.consulClient.GracefulShutdown(&s.config.App)

	// 关闭 gRPC 服务器
	s.Server.Stop()
}
