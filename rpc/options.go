package rpc

import (
	"google.golang.org/grpc"
)

// ServerOptions gRPC 服务器配置选项
type ServerOptions struct {
	host        string
	port        int
	grpcOptions []grpc.ServerOption
}

// ServerOption 服务器选项函数
type ServerOption func(*ServerOptions)

// defaultServerOptions 默认服务器选项
func defaultServerOptions() *ServerOptions {
	return &ServerOptions{
		host:        "0.0.0.0",
		port:        50051,
		grpcOptions: []grpc.ServerOption{},
	}
}

// WithHost 设置监听主机
func WithHost(host string) ServerOption {
	return func(o *ServerOptions) {
		o.host = host
	}
}

// WithPort 设置监听端口
func WithPort(port int) ServerOption {
	return func(o *ServerOptions) {
		o.port = port
	}
}

// WithGRPCOptions 设置 gRPC 服务器选项
// 例如：拦截器、认证、压缩等
func WithGRPCOptions(opts ...grpc.ServerOption) ServerOption {
	return func(o *ServerOptions) {
		o.grpcOptions = append(o.grpcOptions, opts...)
	}
}
