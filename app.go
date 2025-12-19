package main

import (
	"fmt"

	"github.com/charry/config"
	"github.com/charry/config/consumers"
	"github.com/charry/consul"
	"github.com/charry/event"
	"github.com/charry/logger"
	"github.com/charry/rpc"
)

// StartUp 启动应用
// 完整的启动流程，无需外部参数
func StartUp() error {
	logger.Info("========================================")
	logger.Info("开始启动应用...")
	logger.Info("========================================")

	// 1. 加载环境变量
	env := config.LoadEnvArgs()
	logger.Info("✓ 环境变量已加载")
	logger.Infof("\n%s", env.ToJSON())

	// 2. 初始化配置（从默认配置文件 + 环境变量）
	if err := config.Init(env); err != nil {
		logger.Errorf("初始化配置失败: %v", err)
		return err
	}
	logger.Info("✓ 配置已初始化")

	// 3. 初始化 Consul 客户端（创建全局 client）
	cfg := config.Get()
	if err := consul.Init(cfg); err != nil {
		logger.Errorf("初始化 Consul 客户端失败: %v", err)
		return err
	}

	// 4. 从 Consul 加载配置并合并
	if cfg.AppConfigKey != "" {
		logger.Infof("从 Consul 加载配置: %s", cfg.AppConfigKey)

		if jsonStr, err := consul.GetKV(cfg.AppConfigKey); err != nil {
			logger.Warnf("从 Consul 加载配置失败: %v，使用本地配置", err)
		} else if jsonStr != "" {
			logger.Info("✓ 配置已从 Consul 加载")

			if err := config.MergeFromJSON(jsonStr); err != nil {
				return fmt.Errorf("合并 Consul 配置失败: %w", err)
			}

			logger.Info("✓ 配置已合并")
			cfg = config.Get() // 重新获取合并后的配置
			if mergedJSON, err := cfg.ToJSON(); err == nil {
				logger.Infof("\n%s", mergedJSON)
			}
		}
	} else {
		logger.Info("未配置 APP_CONFIG_KEY，跳过从 Consul 加载配置")
	}

	// 5. 初始化日志模块（已在 logger.init() 中完成）
	logger.Info("✓ 日志模块已初始化")

	// 6. 初始化事件模块
	if err := event.Init(10); err != nil {
		logger.Errorf("初始化事件模块失败: %v", err)
		return err
	}

	// 7. 注册配置相关的事件消费者
	consumers.Register()

	// 8. 初始化 RPC 模块（创建并启动服务器）
	cfg = config.Get() // 获取最新配置
	if err := rpc.Init(cfg); err != nil {
		logger.Errorf("初始化 RPC 模块失败: %v", err)
		return err
	}

	// 9. 注册服务到 Consul
	if err := consul.Register(); err != nil {
		logger.Errorf("注册服务到 Consul 失败: %v", err)
		return err
	}

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

	// 3. 关闭事件模块
	event.Close()

	// 4. 刷新日志
	logger.Info("✓ 应用已关闭")
	logger.Sync()
}
