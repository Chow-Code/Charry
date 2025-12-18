package consul

import (
	"github.com/charry/config"
	"github.com/charry/event"
	"github.com/charry/logger"
)

// 事件名称常量
const (
	// ConfigChangedEventName 配置变更事件名
	ConfigChangedEventName = "config.changed"
)

// ConfigChangedConsumer 配置变更事件消费者
// 当 Consul 配置更新时，会收到此事件
type ConfigChangedConsumer struct{}

// CaseEvent 关注配置变更事件
func (c *ConfigChangedConsumer) CaseEvent() []string {
	return []string{ConfigChangedEventName}
}

// Triggered 配置变更时触发
func (c *ConfigChangedConsumer) Triggered(evt *event.Event) error {
	logger.Info("收到配置变更事件")

	// 获取更新后的配置
	if cfg, ok := evt.Data.(*config.Config); ok {
		logger.Info("配置已更新，执行相关逻辑...")
		logger.Infof("当前服务类型: %s", cfg.App.Type)
		logger.Infof("当前环境: %s", cfg.App.Environment)
		logger.Infof("监听地址: %s:%d", cfg.App.Addr.Host, cfg.App.Addr.Port)

		// 在这里添加配置更新后的处理逻辑
		// 例如：
		// - 重新加载缓存
		// - 更新连接池配置
		// - 通知其他服务
	}

	return nil
}

// Async 异步执行
func (c *ConfigChangedConsumer) Async() bool {
	return true
}
