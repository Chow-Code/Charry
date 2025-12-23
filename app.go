package main

import (
	_ "github.com/charry/cluster/consumers" // 自动注册集群消费者
	"github.com/charry/config"
	_ "github.com/charry/config/consumers" // 自动注册配置消费者
	"github.com/charry/constants/event_name"
	"github.com/charry/consul"
	_ "github.com/charry/consul/consumers" // 自动注册 consul 消费者
	"github.com/charry/event"
	"github.com/charry/logger"
	_ "github.com/charry/rpc/consumers" // 自动注册 rpc 消费者
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

	// 3. 初始化事件模块
	if err := event.Init(); err != nil {
		logger.Errorf("初始化事件模块失败: %v", err)
		return err
	}

	// 4. 初始化日志模块（已在 logger.init() 中完成）
	logger.Info("✓ 日志模块已初始化")

	// 5. 初始化 Consul 客户端（创建全局 client）
	// 注意：所有消费者已通过 init() 自动注册
	// 创建后会触发 ClientCreatedEvent，按优先级自动执行：
	//   [0] ClientCreatedConsumer - 加载 Consul 配置
	//   [1] RPCStartConsumer - 启动 RPC 服务器
	//   [2] ServiceRegisterConsumer - 注册服务到 Consul
	cfg := config.Get()
	if err := consul.Init(cfg); err != nil {
		logger.Errorf("初始化 Consul 客户端失败: %v", err)
		return err
	}

	logger.Info("========================================")
	logger.Info("✓ 应用启动完成")
	logger.Info("========================================")

	return nil
}

// Shutdown 关闭应用
// 通过发布关闭事件，让各模块按优先级自动关闭
func Shutdown() {
	logger.Info("========================================")
	logger.Info("开始关闭应用...")
	logger.Info("========================================")

	// 发布关闭事件，各模块按优先级自动关闭：
	//   [0] ServiceDeregisterConsumer - 注销服务
	//   [1] RPCStopConsumer - 停止 RPC 服务器
	//   [2] ShutdownConsumer - 停止配置监听
	event.PublishEvent(event_name.AppShutdown, nil)

	// 等待所有同步消费者执行完成（已经在 PublishEvent 中同步执行）

	// 关闭事件模块
	event.Close()

	// 刷新日志
	logger.Info("✓ 应用已关闭")
	logger.Sync()
}
