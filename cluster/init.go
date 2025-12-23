package cluster

import (
	"fmt"

	"github.com/charry/config"
	"github.com/charry/consul"
	"github.com/charry/logger"
)

var (
	// GlobalManager 全局集群管理器
	GlobalManager *Manager
)

// Init 初始化集群模块
func Init() error {
	logger.Info("初始化集群模块...")

	if consul.GlobalClient == nil {
		return fmt.Errorf("Consul 客户端未初始化")
	}

	// 创建集群管理器
	GlobalManager = NewManager(consul.GlobalClient.GetClient())

	// 获取配置
	cfg := config.Get()

	// 监听同类型服务
	serviceName := fmt.Sprintf("%s-%s", cfg.App.Type, cfg.App.Environment)
	GlobalManager.WatchServices(serviceName)

	logger.Info("✓ 集群模块初始化完成")
	return nil
}

// Close 关闭集群模块
func Close() {
	if GlobalManager != nil {
		logger.Info("关闭集群模块...")
		GlobalManager.Close()
		logger.Info("✓ 集群模块已关闭")
	}
}

