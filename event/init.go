package event

import (
	"github.com/charry/logger"
)

var (
	// GlobalBus 全局事件总线
	GlobalBus *Bus
)

// Init 初始化事件模块
func Init(workerCount int) error {
	logger.Info("初始化事件模块...")

	// 创建事件总线
	GlobalBus = NewBus(workerCount)

	// 启动事件总线
	GlobalBus.Start()

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

