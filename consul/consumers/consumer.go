package consumers

import (
	"github.com/charry/constants/event_name"
	"github.com/charry/constants/priority"
	"github.com/charry/consul"
	"github.com/charry/event"
	"github.com/charry/logger"
)

// ServiceRegisterConsumer Consul 服务注册消费者
// 在 RPC 服务器启动后注册服务到 Consul
type ServiceRegisterConsumer struct{}

func (c *ServiceRegisterConsumer) CaseEvent() []string {
	return []string{event_name.ConsulClientCreated}
}

func (c *ServiceRegisterConsumer) Triggered(evt *event.Event) error {
	logger.Info("注册服务到 Consul...")

	// 注册服务
	if err := consul.Register(); err != nil {
		logger.Errorf("注册服务失败: %v", err)
		return err
	}

	return nil
}

func (c *ServiceRegisterConsumer) Async() bool {
	return false // 同步执行
}

func (c *ServiceRegisterConsumer) Priority() uint32 {
	return priority.ConsulServiceRegister
}

// ServiceDeregisterConsumer Consul 服务注销消费者
type ServiceDeregisterConsumer struct{}

func (c *ServiceDeregisterConsumer) CaseEvent() []string {
	return []string{event_name.AppShutdown}
}

func (c *ServiceDeregisterConsumer) Triggered(evt *event.Event) error {
	logger.Info("关闭 Consul 模块...")
	consul.Close()
	return nil
}

func (c *ServiceDeregisterConsumer) Async() bool {
	return false // 同步执行
}

func (c *ServiceDeregisterConsumer) Priority() uint32 {
	return priority.ConsulServiceDeregister
}

// init 自动注册 Consul 相关的事件消费者
func init() {
	event.RegisterConsumer(&ServiceRegisterConsumer{})
	event.RegisterConsumer(&ServiceDeregisterConsumer{})
}
