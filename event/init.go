package event

import (
	"github.com/charry/config"
	"github.com/charry/logger"
)

var (
	// GlobalBus 全局事件总线
	GlobalBus *Bus

	// pendingConsumers 待注册的消费者列表
	pendingConsumers []Consumer
)

// RegisterConsumer 注册消费者到待注册列表
// 在各 consumers 包的 init() 中调用
func RegisterConsumer(consumer Consumer) {
	pendingConsumers = append(pendingConsumers, consumer)
}

// Init 初始化事件模块
func Init() error {
	logger.Info("初始化事件模块...")

	// 从配置获取工作协程数
	cfg := config.Get()
	workerCount := cfg.Server.EventWorkerCount
	if workerCount <= 0 {
		workerCount = 10 // 默认值
	}

	// 创建事件总线
	GlobalBus = NewBus(workerCount)

	// 启动事件总线
	GlobalBus.Start()

	// 注册所有待注册的消费者
	for _, consumer := range pendingConsumers {
		GlobalBus.Register(consumer)
	}
	logger.Infof("✓ 已自动注册 %d 个事件消费者", len(pendingConsumers))
	pendingConsumers = nil // 清空列表

	logger.Info("✓ 事件模块初始化完成")
	return nil
}

// Close 关闭事件模块
func Close() {
	if GlobalBus != nil {
		logger.Info("关闭事件模块...")
		GlobalBus.Stop()
		logger.Info("✓ 事件模块已关闭")
	}
}

// Register 注册事件消费者到全局事件总线
func Register(consumer Consumer) {
	if GlobalBus != nil {
		GlobalBus.Register(consumer)
	} else {
		logger.Warn("事件总线未初始化，无法注册消费者")
	}
}

// Publish 发布事件到全局事件总线
func Publish(event *Event) {
	if GlobalBus != nil {
		GlobalBus.Publish(event)
	} else {
		logger.Warn("事件总线未初始化，无法发布事件")
	}
}

// PublishEvent 便捷方法：创建并发布事件
func PublishEvent(name string, data interface{}) {
	Publish(NewEvent(name, data))
}
