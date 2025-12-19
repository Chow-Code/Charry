package consul

// KVChangedEvent KV 变化事件数据
type KVChangedEvent struct {
	Key   string // KV 的 key
	Value string // KV 的新值
}
