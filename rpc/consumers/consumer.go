package consumers

import (
	"github.com/charry/config"
	"github.com/charry/consul"
	"github.com/charry/event"
	"github.com/charry/logger"
	"github.com/charry/rpc"
)

// RPCStartConsumer RPC 服务器启动消费者
// 监听配置变更事件，在配置加载完成后启动 RPC 服务器
type RPCStartConsumer struct{}

func (c *RPCStartConsumer) CaseEvent() []string {
	return []string{consul.ClientCreatedEventName}
}

func (c *RPCStartConsumer) Triggered(evt *event.Event) error {
	logger.Info("初始化 RPC 服务器...")

	// 获取最新配置
	cfg := config.Get()

	// 初始化 RPC 模块
	if err := rpc.Init(cfg); err != nil {
		logger.Errorf("初始化 RPC 模块失败: %v", err)
		return err
	}

	return nil
}

func (c *RPCStartConsumer) Async() bool {
	return false // 同步执行
}

func (c *RPCStartConsumer) Priority() uint32 {
	return uint32(event.RPCServerStart)
}

// init 自动注册 RPC 相关的事件消费者
func init() {
	event.RegisterConsumer(&RPCStartConsumer{})
}
