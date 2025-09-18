package event

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestEvent 测试事件结构
func TestEvent(t *testing.T) {
	event := NewEvent("test.event", "test-service", "test data")

	if event.Id == "" {
		t.Error("Event ID should not be empty")
	}

	if event.Type != "test.event" {
		t.Errorf("Expected event type 'test.event', got '%s'", event.Type)
	}

	if event.Source != "test-service" {
		t.Errorf("Expected event source 'test-service', got '%s'", event.Source)
	}

	if event.Data != "test data" {
		t.Errorf("Expected event data 'test data', got '%v'", event.Data)
	}

	// 测试添加元数据
	event = event.WithMetadata("test_key", "test_value")
	if event.Metadata["test_key"] != "test_value" {
		t.Error("Metadata was not added correctly")
	}
}

// TestEventManager 测试事件管理器
func TestEventManager(t *testing.T) {
	em := NewManager(2)

	// 测试启动和停止
	if err := em.Start(); err != nil {
		t.Fatalf("Failed to start event manager: %v", err)
	}

	if err := em.Stop(); err != nil {
		t.Fatalf("Failed to stop event manager: %v", err)
	}
}

// TestSubscribeAndPublish 测试订阅和发布
func TestSubscribeAndPublish(t *testing.T) {
	em := NewManager(2)
	if err := em.Start(); err != nil {
		t.Fatalf("Failed to start event manager: %v", err)
	}
	defer em.Stop()

	// 创建一个测试处理器
	var receivedEvents []Event
	var mutex sync.Mutex

	testHandler := &TestHandler{
		handleFunc: func(ctx context.Context, event Event) error {
			mutex.Lock()
			receivedEvents = append(receivedEvents, event)
			mutex.Unlock()
			return nil
		},
		canHandleFunc: func(eventType string) bool {
			return eventType == "test.event"
		},
	}

	// 订阅事件
	subscriptionId, err := em.Subscribe("test.event", testHandler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 发布事件
	testEvent := NewEvent("test.event", "test", "test data")
	if err := em.Publish(testEvent); err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// 等待事件处理
	time.Sleep(100 * time.Millisecond)

	// 检查事件是否被接收
	mutex.Lock()
	if len(receivedEvents) != 1 {
		t.Errorf("Expected 1 received event, got %d", len(receivedEvents))
	} else if receivedEvents[0].Id != testEvent.Id {
		t.Errorf("Expected event ID %s, got %s", testEvent.Id, receivedEvents[0].Id)
	}
	mutex.Unlock()

	// 测试取消订阅
	if err := em.Unsubscribe(subscriptionId); err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}

	// 再次发布事件，应该不会被接收
	testEvent2 := NewEvent("test.event", "test", "test data 2")
	if err := em.Publish(testEvent2); err != nil {
		t.Fatalf("Failed to publish second event: %v", err)
	}

	// 等待
	time.Sleep(100 * time.Millisecond)

	// 检查事件数量没有增加
	mutex.Lock()
	if len(receivedEvents) != 1 {
		t.Errorf("Expected 1 received event after unsubscribe, got %d", len(receivedEvents))
	}
	mutex.Unlock()
}

// TestSyncPublish 测试同步发布
func TestSyncPublish(t *testing.T) {
	em := NewManager(2)
	if err := em.Start(); err != nil {
		t.Fatalf("Failed to start event manager: %v", err)
	}
	defer em.Stop()

	// 创建测试处理器
	var receivedEvents []Event
	var mutex sync.Mutex

	testHandler := &TestHandler{
		handleFunc: func(ctx context.Context, event Event) error {
			mutex.Lock()
			receivedEvents = append(receivedEvents, event)
			mutex.Unlock()
			return nil
		},
		canHandleFunc: func(eventType string) bool {
			return eventType == "test.sync"
		},
	}

	// 订阅事件
	_, err := em.Subscribe("test.sync", testHandler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 同步发布事件
	testEvent := NewEvent("test.sync", "test", "sync test data")
	ctx := context.Background()
	if err := em.PublishSync(ctx, testEvent); err != nil {
		t.Fatalf("Failed to publish sync event: %v", err)
	}

	// 检查事件是否被立即处理
	mutex.Lock()
	if len(receivedEvents) != 1 {
		t.Errorf("Expected 1 received event, got %d", len(receivedEvents))
	}
	mutex.Unlock()
}

// TestEventFilter 测试事件过滤器
func TestEventFilter(t *testing.T) {
	em := NewManager(2)
	if err := em.Start(); err != nil {
		t.Fatalf("Failed to start event manager: %v", err)
	}
	defer em.Stop()

	// 创建测试处理器
	var receivedEvents []Event
	var mutex sync.Mutex

	testHandler := &TestHandler{
		handleFunc: func(ctx context.Context, event Event) error {
			mutex.Lock()
			receivedEvents = append(receivedEvents, event)
			mutex.Unlock()
			return nil
		},
		canHandleFunc: func(eventType string) bool {
			return eventType == "test.filter"
		},
	}

	// 创建过滤器 - 只处理优先级为"high"的事件
	filter := func(event Event) bool {
		if priority, exists := event.Metadata["priority"]; exists {
			return priority == "high"
		}
		return false
	}

	// 订阅事件并应用过滤器
	_, err := em.Subscribe("test.filter", testHandler, filter)
	if err != nil {
		t.Fatalf("Failed to subscribe with filter: %v", err)
	}

	// 发布高优先级事件
	highPriorityEvent := NewEvent("test.filter", "test", "high priority data").
		WithMetadata("priority", "high")
	if err := em.Publish(highPriorityEvent); err != nil {
		t.Fatalf("Failed to publish high priority event: %v", err)
	}

	// 发布低优先级事件
	lowPriorityEvent := NewEvent("test.filter", "test", "low priority data").
		WithMetadata("priority", "low")
	if err := em.Publish(lowPriorityEvent); err != nil {
		t.Fatalf("Failed to publish low priority event: %v", err)
	}

	// 等待处理
	time.Sleep(100 * time.Millisecond)

	// 检查只有高优先级事件被处理
	mutex.Lock()
	if len(receivedEvents) != 1 {
		t.Errorf("Expected 1 received event (high priority only), got %d", len(receivedEvents))
	} else if receivedEvents[0].Metadata["priority"] != "high" {
		t.Errorf("Expected high priority event, got %v", receivedEvents[0].Metadata["priority"])
	}
	mutex.Unlock()
}

// TestFunctionHandler 测试函数处理器
func TestFunctionHandler(t *testing.T) {
	var handledEvent Event
	var handled bool

	// 创建函数处理器
	functionHandler := NewFunctionHandler(
		"测试处理器",
		func(ctx context.Context, event Event) error {
			handled = true
			handledEvent = event
			return nil
		},
		func(eventType string) bool {
			return eventType == "test.function"
		},
	)

	testEvent := NewEvent("test.function", "test", "function test data")
	ctx := context.Background()

	if err := functionHandler.Handle(ctx, testEvent); err != nil {
		t.Errorf("Function handler failed: %v", err)
	}

	if !handled {
		t.Error("Function handler should have handled the event")
	}

	if handledEvent.Id != testEvent.Id {
		t.Errorf("Expected event ID %s, got %s", testEvent.Id, handledEvent.Id)
	}

	if !functionHandler.CanHandle("test.function") {
		t.Error("Function handler should handle test.function events")
	}

	if functionHandler.CanHandle("other.event") {
		t.Error("Function handler should not handle other.event events")
	}

	// 测试处理器名称为空时的错误处理
	emptyHandler := NewFunctionHandler("空处理器", nil, nil)
	if err := emptyHandler.Handle(ctx, testEvent); err == nil {
		t.Error("Empty function handler should return error")
	}
}

// TestChainHandler 测试链式处理器
func TestChainHandler(t *testing.T) {
	var executionOrder []string
	var mutex sync.Mutex

	// 创建多个测试处理器
	handler1 := &TestHandler{
		handleFunc: func(ctx context.Context, event Event) error {
			mutex.Lock()
			executionOrder = append(executionOrder, "handler1")
			mutex.Unlock()
			return nil
		},
		canHandleFunc: func(eventType string) bool { return true },
	}

	handler2 := &TestHandler{
		handleFunc: func(ctx context.Context, event Event) error {
			mutex.Lock()
			executionOrder = append(executionOrder, "handler2")
			mutex.Unlock()
			return nil
		},
		canHandleFunc: func(eventType string) bool { return true },
	}

	// 创建链式处理器
	chainHandler := NewChainHandler(false, handler1, handler2)

	testEvent := NewEvent("test.chain", "test", "chain test data")
	ctx := context.Background()

	if err := chainHandler.Handle(ctx, testEvent); err != nil {
		t.Errorf("Chain handler failed: %v", err)
	}

	// 检查执行顺序
	mutex.Lock()
	if len(executionOrder) != 2 {
		t.Errorf("Expected 2 handlers to execute, got %d", len(executionOrder))
	} else if executionOrder[0] != "handler1" || executionOrder[1] != "handler2" {
		t.Errorf("Expected execution order [handler1, handler2], got %v", executionOrder)
	}
	mutex.Unlock()
}

// TestHandler 测试用的处理器实现
type TestHandler struct {
	handleFunc    func(ctx context.Context, event Event) error
	canHandleFunc func(eventType string) bool
}

func (h *TestHandler) Handle(ctx context.Context, event Event) error {
	if h.handleFunc != nil {
		return h.handleFunc(ctx, event)
	}
	return nil
}

func (h *TestHandler) CanHandle(eventType string) bool {
	if h.canHandleFunc != nil {
		return h.canHandleFunc(eventType)
	}
	return true
}

// TestStats 测试统计信息
func TestStats(t *testing.T) {
	em := NewManager(3)

	// 订阅几个事件
	testHandler := &TestHandler{}
	em.Subscribe("test.event1", testHandler)
	em.Subscribe("test.event2", testHandler)
	em.Subscribe("test.event1", testHandler) // 同一事件类型多个订阅

	stats := em.GetStats()

	if stats["workerPool"] != 3 {
		t.Errorf("Expected workerPool 3, got %v", stats["workerPool"])
	}

	if stats["totalSubscriptions"] != 3 {
		t.Errorf("Expected totalSubscriptions 3, got %v", stats["totalSubscriptions"])
	}

	eventTypes, ok := stats["eventTypes"].([]string)
	if !ok {
		t.Error("eventTypes should be a slice of strings")
	} else if len(eventTypes) != 2 {
		t.Errorf("Expected 2 event types, got %d", len(eventTypes))
	}

	// 测试获取订阅信息
	subscriptions := em.GetSubscriptions()
	if len(subscriptions) != 2 {
		t.Errorf("Expected 2 event types in subscriptions, got %d", len(subscriptions))
	}

	if len(subscriptions["test.event1"]) != 2 {
		t.Errorf("Expected 2 subscriptions for test.event1, got %d", len(subscriptions["test.event1"]))
	}

	if len(subscriptions["test.event2"]) != 1 {
		t.Errorf("Expected 1 subscription for test.event2, got %d", len(subscriptions["test.event2"]))
	}
}
