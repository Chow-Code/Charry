package consul

import (
	"fmt"

	"github.com/charry/config"
	"github.com/charry/logger"
)

var (
	// GlobalClient 全局 Consul 客户端
	GlobalClient *Client
	// GlobalConfig 全局配置
	GlobalConfig *config.Config
)

// Init 初始化 Consul 模块
// 创建客户端并注册服务到 Consul
func Init(cfg *config.Config) error {
	logger.Info("初始化 Consul 模块...")

	// 保存全局配置
	GlobalConfig = cfg

	// 从配置创建客户端并注册服务
	client, err := RegisterFromConfig(cfg)
	if err != nil {
		return err
	}

	// 保存全局客户端
	GlobalClient = client

	logger.Info("✓ Consul 模块初始化完成")
	return nil
}

// Close 关闭 Consul 模块
// 从 Consul 注销服务
func Close() {
	if GlobalClient != nil && GlobalConfig != nil {
		logger.Info("关闭 Consul 模块...")

		// 停止配置监听
		StopWatch()

		// 注销服务
		GlobalClient.GracefulShutdown(&GlobalConfig.App)
		logger.Info("✓ Consul 模块已关闭")
	}
}

// LoadConfigFromConsul 从 Consul 加载配置
// 如果 cfg.AppConfigKey 为空，则跳过加载
func LoadConfigFromConsul(cfg *config.Config) error {
	configKey := cfg.AppConfigKey
	
	if configKey == "" {
		logger.Info("未配置 APP_CONFIG_KEY，跳过从 Consul 加载配置")
		return nil
	}

	logger.Infof("从 Consul 加载配置: %s", configKey)

	// 创建临时 Consul 客户端（用于拉取配置）
	consulConfig := &Config{
		Address:    cfg.Consul.Address,
		Datacenter: cfg.Consul.Datacenter,
	}

	client, err := NewClient(consulConfig)
	if err != nil {
		return fmt.Errorf("创建 Consul 客户端失败: %w", err)
	}

	// 测试连接
	if err := client.Ping(); err != nil {
		return fmt.Errorf("连接 Consul 失败: %w", err)
	}

	// 获取配置
	jsonStr, err := client.GetKV(configKey)
	if err != nil {
		logger.Warnf("从 Consul 加载配置失败: %v，使用本地配置", err)
		return nil // 不阻止启动，使用本地配置
	}

	// 解析并合并配置
	if err := cfg.MergeFromJSON(jsonStr); err != nil {
		return fmt.Errorf("解析 Consul 配置失败: %w", err)
	}

	logger.Info("✓ 配置已从 Consul 加载并合并")
	if jsonStr, err := cfg.ToJSON(); err == nil {
		logger.Infof("\n%s", jsonStr)
	}

	return nil
}
