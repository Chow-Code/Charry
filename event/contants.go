package event

type AfterStartupRunOrder uint32

const (
	// 配置加载
	ConsulConfigLoad AfterStartupRunOrder = 0
	// RPC 服务器启动
	RPCServerStart AfterStartupRunOrder = 1
	// Consul 服务注册
	ConsulServiceRegister AfterStartupRunOrder = 2
)
