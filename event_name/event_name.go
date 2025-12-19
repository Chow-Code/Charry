package event_name

// 应用生命周期事件
const (
	// AppShutdown 应用关闭事件
	AppShutdown = "app.shutdown"
)

// Consul 相关事件
const (
	// ConsulClientCreated Consul 客户端创建完成事件
	ConsulClientCreated = "consul.client.created"

	// ConsulKVChanged Consul KV 值变化事件
	ConsulKVChanged = "consul.kv.changed"
)

// 配置相关事件
const (
	// ConfigChanged 配置变更事件
	ConfigChanged = "config.changed"
)

