package consul

import (
	"fmt"
	"log"

	"github.com/charry/config"
)

// RegisterFromEnv 从环境变量创建 Consul 客户端并注册服务
// 这是一个便捷方法，简化了客户端创建和服务注册的过程
func RegisterFromEnv(appConfig *config.AppConfig) (*Client, error) {
	// 从环境变量加载 Consul 配置
	consulConfig := NewConfigFromEnv()

	// 创建 Consul 客户端
	client, err := NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	// 测试连接
	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to consul: %w", err)
	}

	// 注册服务
	if err := client.RegisterService(appConfig); err != nil {
		return nil, fmt.Errorf("failed to register service: %w", err)
	}

	log.Printf("Service registered successfully: %s-%s-%d", 
		appConfig.Type, appConfig.Environment, appConfig.Id)

	return client, nil
}

// MustRegisterFromEnv 同 RegisterFromEnv，但失败时会 panic
func MustRegisterFromEnv(appConfig *config.AppConfig) *Client {
	client, err := RegisterFromEnv(appConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to register service to consul: %v", err))
	}
	return client
}

// GracefulShutdown 优雅关闭时注销服务
func (c *Client) GracefulShutdown(appConfig *config.AppConfig) {
	if err := c.DeregisterService(appConfig); err != nil {
		log.Printf("Failed to deregister service: %v", err)
	} else {
		log.Printf("Service deregistered successfully: %s-%s-%d",
			appConfig.Type, appConfig.Environment, appConfig.Id)
	}
}

