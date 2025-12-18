package event

// Event 事件类型
type Event struct {
	// 事件名称
	Name string

	// 事件对象（任意类型）
	Data interface{}
}

// NewEvent 创建新事件
func NewEvent(name string, data interface{}) *Event {
	return &Event{
		Name: name,
		Data: data,
	}
}

