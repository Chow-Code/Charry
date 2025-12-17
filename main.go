package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/charry/config"
	"github.com/charry/rpc"
)

func main() {
	// 设置环境变量
	os.Setenv("CONSUL_ADDRESS", "192.168.30.230:8500")

	// 创建应用配置
	appConfig := &config.AppConfig{
		Id:          1,
		Type:        "test-service",
		Environment: "dev",
		Addr: config.Addr{
			Host: "localhost",
			Port: 50051,
		},
		Metadata: map[string]any{
			"version": "1.0.0",
		},
	}

	// 创建 gRPC 服务器并注册到 Consul
	server, err := rpc.NewServerWithConsul(
		appConfig,
		rpc.WithPort(appConfig.Addr.Port),
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	log.Println("✓ gRPC server created and registered to Consul")
	log.Printf("✓ Server listening on :%d", appConfig.Addr.Port)
	log.Println("✓ Using TCP health check (port reachability)")

	// 这里可以注册您的业务服务
	// grpcServer := server.GetGRPCServer()
	// pb.RegisterYourServiceServer(grpcServer, &yourServiceImpl{})

	// 启动服务器
	server.StartAsync()
	log.Println("✓ gRPC server started")
	log.Println("\nService is running... Press Ctrl+C to shutdown")

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	log.Println("\n收到关闭信号，开始优雅关闭...")
	server.Shutdown()
	log.Println("✓ Service shutdown complete")
}

