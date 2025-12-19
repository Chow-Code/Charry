package consul

// 事件名称常量
const (
	// ConfigChangedEventName 配置变更事件名
	ConfigChangedEventName = "config.changed"

	// ClientCreatedEventName Consul 客户端创建完成事件名
	ClientCreatedEventName = "consul.client.created"

	// KVChangedEventName KV 值变化事件名
	KVChangedEventName = "consul.kv.changed"
)

// KVChangedEvent KV 变化事件数据
type KVChangedEvent struct {
	Key   string // KV 的 key
	Value string // KV 的新值
}
