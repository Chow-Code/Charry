package consumers

import (
	"github.com/charry/cluster"
	"github.com/charry/constants/event_name"
	"github.com/charry/constants/priority"
	"github.com/charry/event"
	"github.com/charry/logger"
)

// ClusterInitConsumer 集群初始化消费者
type ClusterInitConsumer struct{}

func (c *ClusterInitConsumer) CaseEvent() []string {
	return []string{event_name.ConsulClientCreated}
}

func (c *ClusterInitConsumer) Triggered(evt *event.Event) error {
	logger.Info("初始化集群模块...")
	return cluster.Init()
}

func (c *ClusterInitConsumer) Async() bool {
	return false // 同步执行
}

func (c *ClusterInitConsumer) Priority() uint32 {
	return priority.ConsulServiceRegister + 1 // 在服务注册之后
}

// ClusterStopConsumer 集群停止消费者
type ClusterStopConsumer struct{}

func (c *ClusterStopConsumer) CaseEvent() []string {
	return []string{event_name.AppShutdown}
}

func (c *ClusterStopConsumer) Triggered(evt *event.Event) error {
	cluster.Close()
	return nil
}

func (c *ClusterStopConsumer) Async() bool {
	return false // 同步执行
}

func (c *ClusterStopConsumer) Priority() uint32 {
	return priority.ConsulServiceDeregister + 1 // 在服务注销之后
}

// init 自动注册集群相关的事件消费者
func init() {
	event.RegisterConsumer(&ClusterInitConsumer{})
	event.RegisterConsumer(&ClusterStopConsumer{})
}

