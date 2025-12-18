package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/charry/config"
	"github.com/charry/logger"
)

func main() {
	// 第一步：从默认配置文件加载
	cfg, err := config.LoadFromFile("default.config.json")
	if err != nil {
		logger.Fatalf("加载默认配置失败: %v", err)
	}
	logger.Info("✓ 默认配置已加载")

	// 第二步：加载环境变量
	env := config.LoadEnvArgs()
	logger.Info("✓ 环境变量已加载")
	logger.Infof("\n%s", env.ToJSON())

	// 第三步：环境变量覆写配置
	cfg.ApplyEnvArgs(env)

	// 第四步：设置应用特定配置
	cfg.App.Type = "test-service"
	cfg.App.Environment = "dev"
	cfg.App.Metadata = map[string]any{
		"version": "1.0.0",
	}

	// 第五步：启动应用（会先从 Consul KV 拉取配置）
	if err := StartUp(cfg); err != nil {
		logger.Fatalf("应用启动失败: %v", err)
	}

	// 这里可以注册您的业务服务
	// grpcServer := rpc.GlobalServer.GetGRPCServer()
	// pb.RegisterYourServiceServer(grpcServer, &yourServiceImpl{})

	logger.Info("\n服务运行中... 按 Ctrl+C 关闭")

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭（按顺序关闭各个模块）
	logger.Info("\n收到关闭信号，开始优雅关闭...")
	Shutdown()
}
