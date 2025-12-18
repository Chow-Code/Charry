package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/charry/logger"
)

func main() {
	// 启动应用（完整的启动流程）
	if err := StartUp(); err != nil {
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
