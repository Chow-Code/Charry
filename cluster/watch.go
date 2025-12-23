package cluster

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charry/config"
	"github.com/charry/logger"
	consulapi "github.com/hashicorp/consul/api"
)

// WatchServices 监听 Consul 服务变化
func (m *Manager) WatchServices(serviceName string) {
	logger.Infof("开始监听服务变化: %s", serviceName)

	go func() {
		var lastIndex uint64
		isFirstCheck := true

		for {
			select {
			case <-m.stopChan:
				logger.Info("停止监听服务变化")
				return
			default:
				// 使用阻塞查询监听服务变化
				services, meta, err := m.consulClient.Health().Service(
					serviceName,
					"",
					true, // 只获取健康的服务
					&consulapi.QueryOptions{
						WaitIndex: lastIndex,
						WaitTime:  30 * time.Second,
					},
				)

				if err != nil {
					logger.Errorf("查询服务失败: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// 第一次查询，只初始化索引
				if isFirstCheck {
					lastIndex = meta.LastIndex
					isFirstCheck = false

					// 初始化时加载现有服务
					m.loadExistingServices(services)
					logger.Info("✓ 服务监听已就绪")
					continue
				}

				// 检查是否有变化
				if meta.LastIndex > lastIndex {
					lastIndex = meta.LastIndex
					logger.Info("检测到服务变化")

					// 处理服务变化
					m.handleServiceChange(services)

					// 打印当前所有节点
					m.printAllNodes()
				}
			}
		}
	}()
}

// loadExistingServices 加载现有服务
func (m *Manager) loadExistingServices(services []*consulapi.ServiceEntry) {
	logger.Infof("加载现有服务，共 %d 个", len(services))

	for _, service := range services {
		// 跳过自己
		cfg := config.Get()
		selfServiceID := fmt.Sprintf("%s-%s-%d", cfg.App.Type, cfg.App.Environment, cfg.App.Id)
		if service.Service.ID == selfServiceID {
			continue
		}

		// 解析配置
		appConfig, err := parseServiceConfig(service)
		if err != nil {
			logger.Errorf("解析服务配置失败: %s, %v", service.Service.ID, err)
			continue
		}

		// 添加节点
		m.AddNode(service.Service.ID, appConfig)
	}
}

// handleServiceChange 处理服务变化
func (m *Manager) handleServiceChange(services []*consulapi.ServiceEntry) {
	// 当前服务列表
	currentServices := make(map[string]*consulapi.ServiceEntry)
	for _, service := range services {
		currentServices[service.Service.ID] = service
	}

	// 获取现有节点列表
	existingNodes := m.GetAllNodes()
	existingNodeMap := make(map[string]*Node)
	for _, node := range existingNodes {
		existingNodeMap[node.ServiceID] = node
	}

	// 跳过自己
	cfg := config.Get()
	selfServiceID := fmt.Sprintf("%s-%s-%d", cfg.App.Type, cfg.App.Environment, cfg.App.Id)

	// 1. 检查新增的服务
	for serviceID, service := range currentServices {
		if serviceID == selfServiceID {
			continue
		}

		if _, exists := existingNodeMap[serviceID]; !exists {
			// 新增服务
			logger.Infof("发现新服务: %s", serviceID)
			appConfig, err := parseServiceConfig(service)
			if err != nil {
				logger.Errorf("解析服务配置失败: %v", err)
				continue
			}
			m.AddNode(serviceID, appConfig)
		} else {
			// 检查服务是否真的更新
			newConfig, err := parseServiceConfig(service)
			if err != nil {
				continue
			}
			
			// 比较配置是否变化
			existingNode := existingNodeMap[serviceID]
			if isConfigChanged(existingNode.Config, newConfig) {
				m.UpdateNode(serviceID, newConfig)
			}
		}
	}

	// 2. 检查下线的服务
	for serviceID := range existingNodeMap {
		if _, exists := currentServices[serviceID]; !exists {
			// 服务下线
			logger.Infof("服务下线: %s", serviceID)
			m.RemoveNode(serviceID)
		}
	}
}

// parseServiceConfig 从 Consul 服务解析 AppConfig
// Meta 中已经包含了 AppConfig 的所有字段（展开后）
func parseServiceConfig(service *consulapi.ServiceEntry) (*config.AppConfig, error) {
	meta := service.Service.Meta

	appConfig := &config.AppConfig{
		Type:        meta["type"],
		Environment: meta["environment"],
		Addr: config.Addr{
			Host: meta["host"],
		},
		Data: make(map[string]any),
	}

	// 解析 id
	if idStr, ok := meta["id"]; ok {
		var id uint16
		fmt.Sscanf(idStr, "%d", &id)
		appConfig.Id = id
	}

	// 解析 port
	if portStr, ok := meta["port"]; ok {
		var port int
		fmt.Sscanf(portStr, "%d", &port)
		appConfig.Addr.Port = port
	}

	// 解析 data（JSON 字符串）
	if dataJSON, ok := meta["data"]; ok && dataJSON != "" {
		var data map[string]any
		if err := json.Unmarshal([]byte(dataJSON), &data); err == nil {
			appConfig.Data = data
		}
	}

	return appConfig, nil
}

// printAllNodes 打印所有节点信息
func (m *Manager) printAllNodes() {
	nodes := m.GetAllNodes()
	logger.Infof("当前集群节点数: %d", len(nodes))

	for _, node := range nodes {
		logger.Infof("\n%s", node.ToJSON())
	}
}

// isConfigChanged 比较两个 AppConfig 是否发生变化
func isConfigChanged(old, new *config.AppConfig) bool {
	if old == nil || new == nil {
		return true
	}

	// 比较基本字段
	if old.Id != new.Id ||
		old.Type != new.Type ||
		old.Environment != new.Environment ||
		old.Addr.Host != new.Addr.Host ||
		old.Addr.Port != new.Addr.Port {
		return true
	}

	// 比较 data（转换为 JSON 字符串比较）
	oldDataJSON, _ := json.Marshal(old.Data)
	newDataJSON, _ := json.Marshal(new.Data)
	return string(oldDataJSON) != string(newDataJSON)
}

