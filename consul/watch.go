package consul

import (
	"time"

	"github.com/charry/event"
	"github.com/charry/logger"
	consulapi "github.com/hashicorp/consul/api"
)

var (
	// kvWatchStopChans KV 监听停止通道映射 key -> stopChan
	kvWatchStopChans map[string]chan struct{}
)

// StopWatch 停止所有 KV 监听
func StopWatch() {
	// 停止所有 KV 监听
	for key, stopChan := range kvWatchStopChans {
		close(stopChan)
		logger.Infof("停止监听 KV: %s", key)
	}
	kvWatchStopChans = nil
}

// RegisterWatch 注册监听指定的 KV
// 当 KV 值发生变化时，发布 KVChangedEvent 事件
func RegisterWatch(key string) {
	if key == "" {
		return
	}

	if GlobalClient == nil {
		logger.Warn("Consul 客户端未初始化，无法注册 KV 监听")
		return
	}

	// 初始化 kvWatchStopChans
	if kvWatchStopChans == nil {
		kvWatchStopChans = make(map[string]chan struct{})
	}

	// 检查是否已经在监听
	if _, exists := kvWatchStopChans[key]; exists {
		logger.Warnf("KV %s 已在监听中", key)
		return
	}

	stopChan := make(chan struct{})
	kvWatchStopChans[key] = stopChan

	logger.Infof("开始监听 KV: %s", key)

	go func() {
		var lastIndex uint64
		isFirstCheck := true

		for {
			select {
			case <-stopChan:
				logger.Infof("停止监听 KV: %s", key)
				return
			default:
				// 使用阻塞查询监听 KV 变化
				pair, meta, err := GlobalClient.GetClient().KV().Get(key, &consulapi.QueryOptions{
					WaitIndex: lastIndex,
					WaitTime:  30 * time.Second,
				})

				if err != nil {
					logger.Errorf("监听 KV %s 失败: %v", key, err)
					time.Sleep(5 * time.Second)
					continue
				}

				// 第一次查询，只初始化 lastIndex
				if isFirstCheck {
					lastIndex = meta.LastIndex
					isFirstCheck = false
					logger.Infof("✓ KV 监听已就绪: %s", key)
					continue
				}

				// 检查是否有变化
				if meta.LastIndex > lastIndex {
					lastIndex = meta.LastIndex

					var value string
					if pair != nil {
						value = string(pair.Value)
					}

					logger.Infof("检测到 KV 变化: %s", key)

					// 发布 KV 变化事件
					kvEvent := &KVChangedEvent{
						Key:   key,
						Value: value,
					}
					event.PublishEvent(KVChangedEventName, kvEvent)
				}
			}
		}
	}()
}
