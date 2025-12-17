package main

import (
	"github.com/charry/config"
	"github.com/charry/consul"
	"github.com/charry/logger"
	"github.com/charry/rpc"
)

// StartUp 启动应用
// 按顺序初始化各个模块
func StartUp(cfg *config.Config) error {
	logger.Info("========================================")
	logger.Info("开始启动应用...")
	logger.Info("========================================")

	// 1. 初始化日志模块（已在 logger.init() 中完成）
	logger.Info("✓ 日志模块已初始化")

	// 2. 初始化 RPC 模块
	if err := rpc.Init(cfg); err != nil {
		logger.Errorf("初始化 RPC 模块失败: %v", err)
		return err
	}

	// 3. 初始化 Consul 模块
	if err := consul.Init(cfg); err != nil {
		logger.Errorf("初始化 Consul 模块失败: %v", err)
		return err
	}

	// 4. 启动 RPC 服务器
	rpc.Start()

	logger.Info("========================================")
	logger.Info("✓ 应用启动完成")
	logger.Info("========================================")

	return nil
}

// Shutdown 关闭应用
// 按顺序关闭各个模块
func Shutdown() {
	logger.Info("========================================")
	logger.Info("开始关闭应用...")
	logger.Info("========================================")

	// 按照相反的顺序关闭模块

	// 1. 关闭 Consul 模块（先注销服务）
	consul.Close()

	// 2. 关闭 RPC 模块（再停止服务器）
	rpc.Close()

	// 3. 刷新日志
	logger.Info("✓ 应用已关闭")
	logger.Sync()
}
