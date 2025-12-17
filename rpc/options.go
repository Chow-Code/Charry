package rpc

import (
	"google.golang.org/grpc"
)

// RpcConfig RPC 服务器配置
type RpcConfig struct {
	// gRPC 服务器选项（拦截器、认证、压缩等）
	GrpcOptions []grpc.ServerOption
}

// NewDefaultRpcConfig 创建默认 RPC 配置
func NewDefaultRpcConfig() *RpcConfig {
	return &RpcConfig{
		GrpcOptions: []grpc.ServerOption{},
	}
}
