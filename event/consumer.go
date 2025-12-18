package event

// Consumer 事件消费者接口
type Consumer interface {
	// CaseEvent 返回关注的事件名列表
	// 允许关注多个事件
	CaseEvent() []string

	// Triggered 事件触发时调用
	// 传入事件数据
	Triggered(event *Event) error

	// Async 是否异步执行
	// 返回 true：异步执行（默认）
	// 返回 false：同步执行（由生产者线程直接执行）
	Async() bool
}

