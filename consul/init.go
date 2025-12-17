package consul

import (
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
		GlobalClient.GracefulShutdown(&GlobalConfig.App)
		logger.Info("✓ Consul 模块已关闭")
	}
}
