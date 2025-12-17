package consul

import (
	"fmt"

	consulapi "github.com/hashicorp/consul/api"
)

// Client Consul 客户端封装
type Client struct {
	client *consulapi.Client
	config *Config
}

// NewClient 创建 Consul 客户端
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("consul config is nil")
	}

	if cfg.Address == "" {
		return nil, fmt.Errorf("consul address is required")
	}

	// 创建 Consul API 配置
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = cfg.Address
	if cfg.Datacenter != "" {
		consulConfig.Datacenter = cfg.Datacenter
	}

	// 创建 Consul 客户端
	client, err := consulapi.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// GetClient 获取原生 Consul API 客户端
func (c *Client) GetClient() *consulapi.Client {
	return c.client
}

// GetConfig 获取配置
func (c *Client) GetConfig() *Config {
	return c.config
}

// Ping 测试 Consul 连接
func (c *Client) Ping() error {
	_, err := c.client.Agent().Self()
	if err != nil {
		return fmt.Errorf("failed to ping consul: %w", err)
	}
	return nil
}
