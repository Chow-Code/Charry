package consul

import (
	"github.com/charry/config"
	"github.com/charry/logger"
)

var (
	// GlobalClient 全局 Consul 客户端
	GlobalClient *Client
	// GlobalAppConfig 全局应用配置
	GlobalAppConfig *config.AppConfig
)

// Init 初始化 Consul 模块
// 创建客户端并注册服务到 Consul
func Init(appConfig *config.AppConfig) error {
	logger.Info("初始化 Consul 模块...")

	// 保存全局配置
	GlobalAppConfig = appConfig

	// 从环境变量创建客户端并注册服务
	client, err := RegisterFromEnv(appConfig)
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
	if GlobalClient != nil && GlobalAppConfig != nil {
		logger.Info("关闭 Consul 模块...")
		GlobalClient.GracefulShutdown(GlobalAppConfig)
		logger.Info("✓ Consul 模块已关闭")
	}
}

