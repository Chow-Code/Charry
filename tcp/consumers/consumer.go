package consumers

import (
	"github.com/charry/config"
	"github.com/charry/constants/event_name"
	"github.com/charry/constants/priority"
	"github.com/charry/event"
	"github.com/charry/logger"
	"github.com/charry/tcp"
)

// TCPServerStartConsumer TCP 服务器启动消费者
// 监听配置变更事件，在配置加载完成后启动 TCP 服务器
type TCPServerStartConsumer struct{}

func (c *TCPServerStartConsumer) CaseEvent() []string {
	return []string{event_name.ConsulClientCreated}
}

func (c *TCPServerStartConsumer) Triggered(evt *event.Event) error {
	logger.Info("初始化 TCP 服务器...")

	// 获取最新配置
	cfg := config.Get()

	// 初始化 TCP 模块
	if err := tcp.Init(cfg); err != nil {
		logger.Errorf("初始化 TCP 模块失败: %v", err)
		return err
	}

	return nil
}

func (c *TCPServerStartConsumer) Async() bool {
	return false // 同步执行
}

func (c *TCPServerStartConsumer) Priority() uint32 {
	return priority.RPCServerStart
}

// TCPServerStopConsumer TCP 服务器停止消费者
type TCPServerStopConsumer struct{}

func (c *TCPServerStopConsumer) CaseEvent() []string {
	return []string{event_name.AppShutdown}
}

func (c *TCPServerStopConsumer) Triggered(evt *event.Event) error {
	logger.Info("关闭 TCP 模块...")
	tcp.Close()
	return nil
}

func (c *TCPServerStopConsumer) Async() bool {
	return false // 同步执行
}

func (c *TCPServerStopConsumer) Priority() uint32 {
	return priority.RPCServerStop
}

// init 自动注册 TCP 相关的事件消费者
func init() {
	event.RegisterConsumer(&TCPServerStartConsumer{})
	event.RegisterConsumer(&TCPServerStopConsumer{})
}
