package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/charry/config"
	"github.com/charry/logger"
)

func main() {
	// 创建应用配置
	// 环境变量在 .vscode/launch.json 或启动脚本中设置
	appConfig := &config.AppConfig{
		Id:          config.LoadIdFromEnv(),   // 从环境变量加载 APP_ID
		Type:        "test-service",           // 代码中设置
		Environment: "dev",                    // 代码中设置
		Addr:        config.LoadAddrFromEnv(), // 从环境变量加载 APP_HOST, APP_PORT
		Metadata: map[string]any{
			"version": "1.0.0",
		},
	}

	logger.Infof("加载配置: ID=%d, 类型=%s, 环境=%s, 地址=%s:%d",
		appConfig.Id, appConfig.Type, appConfig.Environment,
		appConfig.Addr.Host, appConfig.Addr.Port)

	// 启动应用（按顺序初始化各个模块）
	if err := StartUp(appConfig); err != nil {
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
