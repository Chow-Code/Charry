package priority

// 启动优先级（数值越小越先执行）
const (
	// ConsulConfigLoad 配置加载
	ConsulConfigLoad uint32 = 0

	// RPCServerStart RPC 服务器启动
	RPCServerStart uint32 = 1

	// ConsulServiceRegister Consul 服务注册
	ConsulServiceRegister uint32 = 2
)

// 关闭优先级（数值越小越先执行，与启动相反）
const (
	// ConsulServiceDeregister Consul 服务注销
	ConsulServiceDeregister uint32 = 0

	// RPCServerStop RPC 服务器停止
	RPCServerStop uint32 = 1

	// ConsulClientClose Consul 客户端关闭（停止配置监听）
	ConsulClientClose uint32 = 2
)
