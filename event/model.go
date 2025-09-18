package event

import (
	"context"
	"time"
)

// Event 事件
type Event struct {
	Id        string                 `json:"id"`        // 事件唯一标识
	Type      string                 `json:"type"`      // 事件类型
	Data      interface{}            `json:"data"`      // 事件数据
	Timestamp time.Time              `json:"timestamp"` // 事件时间戳
	Source    string                 `json:"source"`    // 事件源
	Metadata  map[string]interface{} `json:"metadata"`  // 元数据
}

// Handler 事件处理器接口
type Handler interface {
	Handle(ctx context.Context, event Event) error
	CanHandle(eventType string) bool
}

// AsyncEventHandler 异步事件处理器接口
type AsyncEventHandler interface {
	Handler
	HandleAsync(ctx context.Context, event Event) <-chan error
}

// Filter 事件过滤器
type Filter func(event Event) bool

// Subscription 订阅信息
type Subscription struct {
	Id        string    `json:"id"`         // 订阅ID
	EventType string    `json:"event_type"` // 订阅的事件类型
	Handler   Handler   `json:"-"`          // 事件处理器（不序列化）
	Filter    Filter    `json:"-"`          // 事件过滤器（不序列化）
	CreatedAt time.Time `json:"created_at"` // 订阅创建时间
	IsActive  bool      `json:"is_active"`  // 是否激活
}

// NewEvent 创建新事件
func NewEvent(eventType, source string, data interface{}) Event {
	return Event{
		Id:        generateEventId(),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
		Source:    source,
		Metadata:  make(map[string]interface{}),
	}
}

// WithMetadata 添加元数据
func (e Event) WithMetadata(key string, value interface{}) Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// generateEventId 生成事件ID
func generateEventId() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}
