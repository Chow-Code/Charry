package main

import (
	"github.com/charry/config"
	_ "github.com/charry/config/consumers" // 自动注册配置消费者
	"github.com/charry/consul"
	_ "github.com/charry/consul/consumers" // 自动注册 consul 消费者
	"github.com/charry/event"
	"github.com/charry/logger"
	"github.com/charry/rpc"
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
	if err := event.Init(10); err != nil {
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
