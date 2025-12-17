package consul

import (
	"fmt"

	"github.com/charry/config"
	"github.com/charry/logger"
)

// RegisterFromEnv 从环境变量创建 Consul 客户端并注册服务
// 这是一个便捷方法，简化了客户端创建和服务注册的过程
func RegisterFromEnv(appConfig *config.AppConfig) (*Client, error) {
	// 从环境变量加载 Consul 配置
	consulConfig := NewConfigFromEnv()

	// 创建 Consul 客户端
	client, err := NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 Consul 客户端失败: %w", err)
	}

	// 测试连接
	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("连接 Consul 失败: %w", err)
	}

	// 注册服务
	if err := client.RegisterService(appConfig); err != nil {
		return nil, fmt.Errorf("注册服务失败: %w", err)
	}

	logger.Infof("服务注册成功: %s-%s-%d", 
		appConfig.Type, appConfig.Environment, appConfig.Id)

	return client, nil
}

// MustRegisterFromEnv 同 RegisterFromEnv，但失败时会 panic
func MustRegisterFromEnv(appConfig *config.AppConfig) *Client {
	client, err := RegisterFromEnv(appConfig)
	if err != nil {
		panic(fmt.Sprintf("注册服务到 Consul 失败: %v", err))
	}
	return client
}

// GracefulShutdown 优雅关闭时注销服务
func (c *Client) GracefulShutdown(appConfig *config.AppConfig) {
	if err := c.DeregisterService(appConfig); err != nil {
		logger.Errorf("注销服务失败: %v", err)
	} else {
		logger.Infof("服务注销成功: %s-%s-%d",
			appConfig.Type, appConfig.Environment, appConfig.Id)
	}
}

