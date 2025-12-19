package consul

import (
	"fmt"

	"github.com/charry/config"
	"github.com/charry/constants/event_name"
	"github.com/charry/event"
	"github.com/charry/logger"
)

var (
	// GlobalClient 全局 Consul 客户端
	GlobalClient *Client
)

// Init 初始化 Consul 模块
// 只负责创建并初始化全局客户端
func Init(cfg config.Config) error {
	logger.Info("初始化 Consul 模块...")

	// 创建 Consul 客户端（直接使用配置）
	client, err := NewClient(&cfg.Consul)
	if err != nil {
		return fmt.Errorf("创建 Consul 客户端失败: %w", err)
	}

	// 测试连接
	if err := client.Ping(); err != nil {
		return fmt.Errorf("连接 Consul 失败: %w", err)
	}

	// 保存全局客户端
	GlobalClient = client

	logger.Info("✓ Consul 模块初始化完成")
	// 发布 Consul 客户端创建完成事件
	event.PublishEvent(event_name.ConsulClientCreated, nil)
	return nil
}

// Register 注册服务到 Consul
// 使用全局客户端
func Register() error {
	if GlobalClient == nil {
		return fmt.Errorf("Consul 客户端未初始化")
	}

	logger.Info("注册服务到 Consul...")

	cfg := config.Get()
	if err := GlobalClient.RegisterService(&cfg.App); err != nil {
		return fmt.Errorf("注册服务失败: %w", err)
	}

	logger.Infof("服务注册成功: %s-%s-%d",
		cfg.App.Type, cfg.App.Environment, cfg.App.Id)

	return nil
}

// GetKV 从 Consul KV 获取值
// 通用方法，可以读取任意 key
func GetKV(key string) (string, error) {
	if key == "" {
		return "", nil
	}

	if GlobalClient == nil {
		return "", fmt.Errorf("Consul 客户端未初始化")
	}

	value, err := GlobalClient.GetKV(key)
	if err != nil {
		return "", err
	}

	return value, nil
}

// PutKV 设置 Consul KV 值
// 通用方法，可以设置任意 key/value
// 注意：不允许直接修改 AppConfigKey，防止配置被意外覆盖
func PutKV(key, value string) error {
	if GlobalClient == nil {
		return fmt.Errorf("Consul 客户端未初始化")
	}

	// 安全检查：禁止直接修改配置 key
	cfg := config.Get()
	if key == cfg.AppConfigKey {
		return fmt.Errorf("禁止直接修改配置 key: %s，请使用配置管理功能", key)
	}

	return GlobalClient.PutKV(key, value)
}

// DeleteKV 删除 Consul KV
// 通用方法，可以删除任意 key
// 注意：不允许删除 AppConfigKey，防止配置被意外删除
func DeleteKV(key string) error {
	if GlobalClient == nil {
		return fmt.Errorf("Consul 客户端未初始化")
	}

	// 安全检查：禁止删除配置 key
	cfg := config.Get()
	if key == cfg.AppConfigKey {
		return fmt.Errorf("禁止删除配置 key: %s，这会导致配置丢失", key)
	}

	return GlobalClient.DeleteKV(key)
}

// Close 关闭 Consul 模块
// 从 Consul 注销服务
func Close() {
	if GlobalClient != nil {
		logger.Info("关闭 Consul 模块...")

		// 停止配置监听
		StopWatch()

		// 注销服务
		cfg := config.Get()
		GlobalClient.GracefulShutdown(&cfg.App)
		logger.Info("✓ Consul 模块已关闭")
	}
}
