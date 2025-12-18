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
// 创建并启动 gRPC 服务器
func Init(cfg *config.Config) error {
	logger.Info("初始化 RPC 模块...")

	// 创建 gRPC 服务器
	server, err := NewServer(nil, &cfg.App)
	if err != nil {
		return err
	}

	// 保存全局服务器
	GlobalServer = server

	// 启动服务器
	GlobalServer.StartAsync()

	logger.Infof("✓ RPC 模块初始化完成，服务器已启动: %s:%d",
		cfg.App.Addr.Host, cfg.App.Addr.Port)
	return nil
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
