package consul

import (
	"time"

	"github.com/charry/config"
	"github.com/charry/logger"
	consulapi "github.com/hashicorp/consul/api"
)

var (
	// watchStopChan 停止监听的通道
	watchStopChan chan struct{}
)

// WatchConfig 监听 Consul KV 配置变化
// 当配置发生变化时，自动重新加载并合并配置
func WatchConfig(cfg *config.Config, configKey string) {
	if configKey == "" {
		logger.Info("未配置 APP_CONFIG_KEY，跳过配置监听")
		return
	}

	watchStopChan = make(chan struct{})

	logger.Infof("开始监听配置变化: %s", configKey)

	go func() {
		var lastIndex uint64
		isFirstCheck := true
		
		for {
			select {
			case <-watchStopChan:
				logger.Info("配置监听已停止")
				return
			default:
				// 使用阻塞查询监听配置变化
				pair, meta, err := GlobalClient.GetClient().KV().Get(configKey, &consulapi.QueryOptions{
					WaitIndex: lastIndex,
					WaitTime:  30 * time.Second,
				})

				if err != nil {
					logger.Errorf("监听配置失败: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// 第一次查询，只初始化 lastIndex，不触发更新
				if isFirstCheck {
					lastIndex = meta.LastIndex
					isFirstCheck = false
					logger.Info("✓ 配置监听已就绪")
					continue
				}

				// 检查是否有变化
				if meta.LastIndex > lastIndex {
					lastIndex = meta.LastIndex

					if pair != nil {
						logger.Info("检测到配置变化，重新加载...")
						
						// 解析并合并配置
						if err := cfg.ReadFromJSON(string(pair.Value)); err != nil {
							logger.Errorf("解析配置失败: %v", err)
							continue
						}

						logger.Info("✓ 配置已更新")
						if jsonStr, err := cfg.ToJSON(); err == nil {
							logger.Infof("\n%s", jsonStr)
						}

						// 通知各模块配置已更新
						onConfigChanged(cfg)
					}
				}
			}
		}
	}()
}

// StopWatch 停止配置监听
func StopWatch() {
	if watchStopChan != nil {
		close(watchStopChan)
		watchStopChan = nil
	}
}

// onConfigChanged 配置变更回调
// 当配置更新时，通知各模块
func onConfigChanged(cfg *config.Config) {
	logger.Info("配置已更新，通知各模块...")
	
	// 可以在这里添加通知逻辑
	// 例如：重新初始化某些模块、更新缓存等
	
	// 示例：打印重要配置项
	logger.Infof("当前服务类型: %s", cfg.App.Type)
	logger.Infof("当前环境: %s", cfg.App.Environment)
	logger.Infof("监听地址: %s:%d", cfg.App.Addr.Host, cfg.App.Addr.Port)
}

