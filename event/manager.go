package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"charry/logger"
)

// EventManager 事件管理器
type EventManager struct {
	subscriptions map[string]map[string]*Subscription // eventType -> subscriptionId -> subscription
	handlers      map[string][]EventHandler           // eventType -> handlers
	mutex         sync.RWMutex
	eventChan     chan Event
	workerPool    int
	running       bool
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewEventManager 创建新的事件管理器
func NewEventManager(workerPoolSize int) *EventManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &EventManager{
		subscriptions: make(map[string]map[string]*Subscription),
		handlers:      make(map[string][]EventHandler),
		eventChan:     make(chan Event, 1000), // 缓冲区大小为1000
		workerPool:    workerPoolSize,
		ctx:           ctx,
		cancel:        cancel,
		running:       false,
	}
}

// Start 启动事件管理器
func (em *EventManager) Start() error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if em.running {
		return fmt.Errorf("事件管理器已经在运行")
	}

	em.running = true

	// 启动worker池
	for i := 0; i < em.workerPool; i++ {
		em.wg.Add(1)
		go em.worker(i)
	}

	logger.Info("事件管理器已启动", "workerPool", em.workerPool)
	return nil
}

// Stop 停止事件管理器
func (em *EventManager) Stop() error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if !em.running {
		return fmt.Errorf("事件管理器尚未运行")
	}

	em.running = false
	em.cancel()
	close(em.eventChan)
	em.wg.Wait()

	logger.Info("事件管理器已停止")
	return nil
}

// Subscribe 订阅事件
func (em *EventManager) Subscribe(eventType string, handler EventHandler, filters ...EventFilter) (string, error) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	// 创建订阅
	subscription := &Subscription{
		Id:        generateSubscriptionId(),
		EventType: eventType,
		Handler:   handler,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	// 如果有过滤器，使用第一个
	if len(filters) > 0 {
		subscription.Filter = filters[0]
	}

	// 存储订阅
	if em.subscriptions[eventType] == nil {
		em.subscriptions[eventType] = make(map[string]*Subscription)
	}
	em.subscriptions[eventType][subscription.Id] = subscription

	// 添加处理器到快速查找列表
	em.handlers[eventType] = append(em.handlers[eventType], handler)

	logger.Info("新事件订阅已创建",
		"subscriptionId", subscription.Id,
		"eventType", eventType)

	return subscription.Id, nil
}

// Unsubscribe 取消订阅
func (em *EventManager) Unsubscribe(subscriptionId string) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	// 查找并删除订阅
	for eventType, subs := range em.subscriptions {
		if sub, exists := subs[subscriptionId]; exists {
			delete(subs, subscriptionId)

			// 从处理器列表中移除
			for i, handler := range em.handlers[eventType] {
				if handler == sub.Handler {
					em.handlers[eventType] = append(em.handlers[eventType][:i], em.handlers[eventType][i+1:]...)
					break
				}
			}

			// 如果该事件类型没有订阅了，清理空map
			if len(subs) == 0 {
				delete(em.subscriptions, eventType)
				delete(em.handlers, eventType)
			}

			logger.Info("事件订阅已取消",
				"subscriptionId", subscriptionId,
				"eventType", eventType)

			return nil
		}
	}

	return fmt.Errorf("未找到订阅ID: %s", subscriptionId)
}

// Publish 发布事件（异步）
func (em *EventManager) Publish(event Event) error {
	if !em.running {
		return fmt.Errorf("事件管理器尚未启动")
	}

	select {
	case em.eventChan <- event:
		logger.Debug("事件已发布到队列",
			"eventId", event.Id,
			"eventType", event.Type)
		return nil
	case <-em.ctx.Done():
		return fmt.Errorf("事件管理器已停止")
	default:
		return fmt.Errorf("事件队列已满，无法发布事件")
	}
}

// PublishSync 同步发布事件
func (em *EventManager) PublishSync(ctx context.Context, event Event) error {
	em.mutex.RLock()
	subscriptions := em.subscriptions[event.Type]
	em.mutex.RUnlock()

	if len(subscriptions) == 0 {
		logger.Debug("没有找到事件订阅者", "eventType", event.Type)
		return nil
	}

	var errs []error
	for _, sub := range subscriptions {
		if !sub.IsActive {
			continue
		}

		// 应用过滤器
		if sub.Filter != nil && !sub.Filter(event) {
			continue
		}

		// 处理事件
		if err := sub.Handler.Handle(ctx, event); err != nil {
			logger.Error("事件处理失败",
				"eventId", event.Id,
				"eventType", event.Type,
				"subscriptionId", sub.Id,
				"error", err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分事件处理失败: %v", errs)
	}

	return nil
}

// worker 工作协程
func (em *EventManager) worker(workerId int) {
	defer em.wg.Done()

	logger.Debug("事件处理器worker已启动", "workerId", workerId)

	for {
		select {
		case event, ok := <-em.eventChan:
			if !ok {
				logger.Debug("事件处理器worker已停止", "workerId", workerId)
				return
			}

			em.handleEvent(event)

		case <-em.ctx.Done():
			logger.Debug("事件处理器worker已停止", "workerId", workerId)
			return
		}
	}
}

// handleEvent 处理事件
func (em *EventManager) handleEvent(event Event) {
	em.mutex.RLock()
	subscriptions := em.subscriptions[event.Type]
	em.mutex.RUnlock()

	if len(subscriptions) == 0 {
		logger.Debug("没有找到事件订阅者", "eventType", event.Type)
		return
	}

	logger.Debug("开始处理事件",
		"eventId", event.Id,
		"eventType", event.Type,
		"subscriberCount", len(subscriptions))

	for _, sub := range subscriptions {
		if !sub.IsActive {
			continue
		}

		// 应用过滤器
		if sub.Filter != nil && !sub.Filter(event) {
			continue
		}

		// 异步处理事件
		go func(subscription *Subscription) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := subscription.Handler.Handle(ctx, event); err != nil {
				logger.Error("事件处理失败",
					"eventId", event.Id,
					"eventType", event.Type,
					"subscriptionId", subscription.Id,
					"error", err)
			} else {
				logger.Debug("事件处理成功",
					"eventId", event.Id,
					"eventType", event.Type,
					"subscriptionId", subscription.Id)
			}
		}(sub)
	}
}

// GetSubscriptions 获取所有订阅信息
func (em *EventManager) GetSubscriptions() map[string][]*Subscription {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	result := make(map[string][]*Subscription)
	for eventType, subs := range em.subscriptions {
		for _, sub := range subs {
			result[eventType] = append(result[eventType], sub)
		}
	}

	return result
}

// GetStats 获取统计信息
func (em *EventManager) GetStats() map[string]interface{} {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	stats := map[string]interface{}{
		"running":          em.running,
		"workerPool":       em.workerPool,
		"eventQueueLength": len(em.eventChan),
		"totalSubscriptions": func() int {
			count := 0
			for _, subs := range em.subscriptions {
				count += len(subs)
			}
			return count
		}(),
		"eventTypes": func() []string {
			types := make([]string, 0, len(em.subscriptions))
			for eventType := range em.subscriptions {
				types = append(types, eventType)
			}
			return types
		}(),
	}

	return stats
}

// generateSubscriptionId 生成订阅ID
func generateSubscriptionId() string {
	return "sub_" + time.Now().Format("20060102150405") + randomString(8)
}
