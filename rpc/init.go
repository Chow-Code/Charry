package rpc

import (
	"github.com/charry/config"
	"github.com/charry/logger"
)

var (
	// GlobalServer 全局 gRPC 服务器
	GlobalServer *Server
)

// Init 初始化 RPC 模块
// 创建 gRPC 服务器
func Init(rpcConfig *RpcConfig, appConfig *config.AppConfig) error {
	logger.Info("初始化 RPC 模块...")

	// 创建 gRPC 服务器
	server, err := NewServer(rpcConfig, appConfig)
	if err != nil {
		return err
	}

	// 保存全局服务器
	GlobalServer = server

	logger.Infof("✓ RPC 模块初始化完成，监听地址: %s:%d", 
		appConfig.Addr.Host, appConfig.Addr.Port)
	return nil
}

// Start 启动 RPC 服务器
func Start() {
	if GlobalServer != nil {
		GlobalServer.StartAsync()
		logger.Info("✓ RPC 服务器已启动")
	}
}

// Close 关闭 RPC 模块
// 停止 gRPC 服务器
func Close() {
	if GlobalServer != nil {
		logger.Info("关闭 RPC 模块...")
		GlobalServer.Stop()
		logger.Info("✓ RPC 模块已关闭")
	}
}

