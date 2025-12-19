package consumers

import (
	"github.com/charry/config"
	"github.com/charry/consul"
	"github.com/charry/event"
	"github.com/charry/logger"
)

// ClientCreatedConsumer Consul 客户端创建完成事件消费者
type ClientCreatedConsumer struct{}

func (c *ClientCreatedConsumer) CaseEvent() []string {
	return []string{consul.ClientCreatedEventName}
}

func (c *ClientCreatedConsumer) Triggered(evt *event.Event) error {
	logger.Info("Consul 客户端已创建，加载配置并注册监听...")

	// 获取配置
	cfg := config.Get()

	// 1. 从 Consul 加载配置并合并
	if cfg.AppConfigKey != "" {
		logger.Infof("从 Consul 加载配置: %s", cfg.AppConfigKey)

		if jsonStr, err := consul.GetKV(cfg.AppConfigKey); err != nil {
			logger.Warnf("从 Consul 加载配置失败: %v，使用本地配置", err)
		} else if jsonStr != "" {
			logger.Info("✓ 配置已从 Consul 加载")

			if err := config.MergeFromJSON(jsonStr); err != nil {
				logger.Errorf("合并配置失败: %v", err)
				return err
			}

			logger.Info("✓ 配置已合并")
			updatedCfg := config.Get()
			if mergedJSON, err := updatedCfg.ToJSON(); err == nil {
				logger.Infof("\n%s", mergedJSON)
			}
		}

		// 2. 注册监听 AppConfigKey
		consul.RegisterWatch(cfg.AppConfigKey)
	} else {
		logger.Info("未配置 APP_CONFIG_KEY，跳过从 Consul 加载配置")
	}

	return nil
}

func (c *ClientCreatedConsumer) Async() bool {
	return false // 同步执行，确保监听立即注册
}

func (c *ClientCreatedConsumer) Priority() uint32 {
	return uint32(event.ConsulConfigLoad) // 最高优先级
}

// KVChangedConsumer KV 变化事件消费者
type KVChangedConsumer struct{}

func (c *KVChangedConsumer) CaseEvent() []string {
	return []string{consul.KVChangedEventName}
}

func (c *KVChangedConsumer) Triggered(evt *event.Event) error {
	kvEvt, ok := evt.Data.(*consul.KVChangedEvent)
	if !ok {
		return nil
	}

	// 获取当前配置
	cfg := config.Get()

	// 判断是否为 AppConfigKey
	if kvEvt.Key == cfg.AppConfigKey {
		logger.Infof("检测到配置变化: %s", kvEvt.Key)

		// 合并配置
		if err := config.MergeFromJSON(kvEvt.Value); err != nil {
			logger.Errorf("合并配置失败: %v", err)
			return err
		}

		logger.Info("✓ 配置已更新")
		updatedCfg := config.Get()
		if jsonStr, err := updatedCfg.ToJSON(); err == nil {
			logger.Infof("\n%s", jsonStr)
		}

		// 发布配置变更事件
		event.PublishEvent(consul.ConfigChangedEventName, &updatedCfg)
	}

	return nil
}

func (c *KVChangedConsumer) Async() bool {
	return true // 异步执行
}

func (c *KVChangedConsumer) Priority() uint32 {
	return 0 // 最高优先级
}

// init 自动注册配置相关的事件消费者
func init() {
	event.RegisterConsumer(&ClientCreatedConsumer{})
	event.RegisterConsumer(&KVChangedConsumer{})
}
