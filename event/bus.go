package event

import (
	"sync"

	"github.com/charry/logger"
)

// Bus 事件总线
type Bus struct {
	// 事件消费者映射: eventName -> []Consumer
	consumers map[string][]Consumer

	// 事件队列（用于异步消费者）
	eventChan chan *Event

	// 停止通道
	stopChan chan struct{}

	// 互斥锁
	mu sync.RWMutex

	// 工作协程数量
	workerCount int
}

// NewBus 创建新的事件总线
func NewBus(workerCount int) *Bus {
	if workerCount <= 0 {
		workerCount = 10 // 默认 10 个工作协程
	}

	return &Bus{
		consumers:   make(map[string][]Consumer),
		eventChan:   make(chan *Event, 1000), // 缓冲 1000 个事件
		stopChan:    make(chan struct{}),
		workerCount: workerCount,
	}
}

// Register 注册事件消费者
func (b *Bus) Register(consumer Consumer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	events := consumer.CaseEvent()
	for _, eventName := range events {
		b.consumers[eventName] = append(b.consumers[eventName], consumer)
		logger.Infof("注册消费者到事件: %s", eventName)
	}
}

// Publish 发布事件
// 注意：消费者只在启动时注册，运行时只读，因此不需要加锁
func (b *Bus) Publish(event *Event) {
	consumers := b.consumers[event.Name]

	if len(consumers) == 0 {
		// 没有消费者关注此事件
		return
	}

	for _, consumer := range consumers {
		if consumer.Async() {
			// 异步执行：放入队列
			select {
			case b.eventChan <- event:
				// 成功放入队列
			default:
				logger.Warnf("事件队列已满，丢弃事件: %s", event.Name)
			}
		} else {
			// 同步执行：由当前线程直接执行
			b.handleEvent(consumer, event)
		}
	}
}

// Start 启动事件总线（启动工作协程处理异步事件）
func (b *Bus) Start() {
	logger.Infof("启动事件总线，工作协程数: %d", b.workerCount)

	for i := 0; i < b.workerCount; i++ {
		go b.worker(i)
	}
}

// Stop 停止事件总线
func (b *Bus) Stop() {
	logger.Info("停止事件总线...")
	close(b.stopChan)
	close(b.eventChan)
}

// worker 工作协程，处理异步事件
func (b *Bus) worker(id int) {
	for {
		select {
		case <-b.stopChan:
			logger.Infof("事件总线工作协程 %d 已停止", id)
			return
		case event, ok := <-b.eventChan:
			if !ok {
				return
			}

			// 查找消费者并执行（运行时只读，不需要加锁）
			consumers := b.consumers[event.Name]

			for _, consumer := range consumers {
				if consumer.Async() {
					b.handleEvent(consumer, event)
				}
			}
		}
	}
}

// handleEvent 处理事件
func (b *Bus) handleEvent(consumer Consumer, event *Event) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("事件处理发生 panic: %v, 事件: %s", r, event.Name)
		}
	}()

	if err := consumer.Triggered(event); err != nil {
		logger.Errorf("事件处理失败: %v, 事件: %s", err, event.Name)
	}
}

// GetConsumerCount 获取指定事件的消费者数量
func (b *Bus) GetConsumerCount(eventName string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.consumers[eventName])
}
