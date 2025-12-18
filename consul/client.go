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

// GetKV 从 Consul 获取 Key/Value
func (c *Client) GetKV(key string) (string, error) {
	pair, _, err := c.client.KV().Get(key, nil)
	if err != nil {
		return "", fmt.Errorf("获取 KV 失败: %w", err)
	}

	if pair == nil {
		return "", fmt.Errorf("配置键不存在: %s", key)
	}

	return string(pair.Value), nil
}

// PutKV 设置 Key/Value 到 Consul
func (c *Client) PutKV(key, value string) error {
	p := &consulapi.KVPair{Key: key, Value: []byte(value)}
	_, err := c.client.KV().Put(p, nil)
	if err != nil {
		return fmt.Errorf("设置 KV 失败: %w", err)
	}
	return nil
}

// DeleteKV 删除 Consul 中的 Key/Value
func (c *Client) DeleteKV(key string) error {
	_, err := c.client.KV().Delete(key, nil)
	if err != nil {
		return fmt.Errorf("删除 KV 失败: %w", err)
	}
	return nil
}
