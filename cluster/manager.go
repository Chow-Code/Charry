package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"charry/event"
	"charry/logger"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// Manager 集群管理器
type Manager struct {
	config       *NacosConfig
	configClient config_client.IConfigClient
	eventManager *event.Manager

	// 节点管理
	nodes    map[string]*Node // nodeId -> Node
	nodesMux sync.RWMutex

	// 状态管理
	isRunning   bool
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	lastVersion string // 配置版本号，用于检测变化
}

// NewManager 创建新的集群管理器
func NewManager(config *NacosConfig, eventManager *event.Manager) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		config:       config,
		eventManager: eventManager,
		nodes:        make(map[string]*Node),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start 启动集群管理器
func (m *Manager) Start() error {
	if m.isRunning {
		return fmt.Errorf("集群管理器已在运行")
	}

	// 初始化 Nacos 客户端
	if err := m.initNacosClient(); err != nil {
		return fmt.Errorf("初始化 Nacos 客户端失败: %v", err)
	}

	// 首次拉取配置
	if err := m.loadInitialConfig(); err != nil {
		logger.Error("首次加载配置失败", "error", err)
		return err
	}

	// 启动配置监听
	if err := m.startConfigWatcher(); err != nil {
		logger.Error("启动配置监听失败", "error", err)
		return err
	}

	m.isRunning = true

	// 发布集群连接事件
	nodes := m.GetAllNodes()
	connectEvent := CreateClusterConnectedEvent(nodes)
	m.eventManager.Publish(connectEvent)

	logger.Info("集群管理器已启动", "nodeCount", len(nodes))
	return nil
}

// Stop 停止集群管理器
func (m *Manager) Stop() error {
	if !m.isRunning {
		return fmt.Errorf("集群管理器尚未运行")
	}

	m.isRunning = false
	m.cancel()
	m.wg.Wait()

	// 发布集群断开连接事件
	disconnectEvent := CreateClusterDisconnectedEvent("管理器手动停止")
	m.eventManager.Publish(disconnectEvent)

	logger.Info("集群管理器已停止")
	return nil
}

// initNacosClient 初始化 Nacos 客户端
func (m *Manager) initNacosClient() error {
	// 构建服务器配置
	var serverConfigs []constant.ServerConfig
	for _, sc := range m.config.ServerConfigs {
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr:      sc.IpAddr,
			Port:        sc.Port,
			ContextPath: sc.ContextPath,
			Scheme:      sc.Scheme,
		})
	}

	// 构建客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         m.config.ClientConfig.NamespaceId,
		TimeoutMs:           m.config.ClientConfig.TimeoutMs,
		NotLoadCacheAtStart: m.config.ClientConfig.NotLoadCacheAtStart,
		LogDir:              m.config.ClientConfig.LogDir,
		CacheDir:            m.config.ClientConfig.CacheDir,
		LogLevel:            m.config.ClientConfig.LogLevel,
		Username:            m.config.ClientConfig.Username,
		Password:            m.config.ClientConfig.Password,
	}

	// 创建配置客户端
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)

	if err != nil {
		return fmt.Errorf("创建 Nacos 配置客户端失败: %v", err)
	}

	m.configClient = configClient
	logger.Info("Nacos 客户端初始化成功")
	return nil
}

// loadInitialConfig 首次加载配置
func (m *Manager) loadInitialConfig() error {
	content, err := m.configClient.GetConfig(vo.ConfigParam{
		DataId: m.config.ClusterConfig.DataId,
		Group:  m.config.ClusterConfig.Group,
	})

	if err != nil {
		return fmt.Errorf("获取 Nacos 配置失败: %v", err)
	}

	if content == "" {
		logger.Info("Nacos 配置为空，初始化为空节点列表")
		return nil
	}

	return m.parseAndUpdateNodes(content, "初始化加载")
}

// startConfigWatcher 启动配置监听器
func (m *Manager) startConfigWatcher() error {
	err := m.configClient.ListenConfig(vo.ConfigParam{
		DataId: m.config.ClusterConfig.DataId,
		Group:  m.config.ClusterConfig.Group,
		OnChange: func(namespace, group, dataId, data string) {
			logger.Info("检测到 Nacos 配置变化",
				"namespace", namespace,
				"group", group,
				"dataId", dataId)

			if err := m.parseAndUpdateNodes(data, "配置变化"); err != nil {
				logger.Error("处理配置变化失败", "error", err)
			}
		},
	})

	if err != nil {
		return fmt.Errorf("启动 Nacos 配置监听失败: %v", err)
	}

	logger.Info("Nacos 配置监听器已启动")
	return nil
}

// parseAndUpdateNodes 解析配置并更新节点
func (m *Manager) parseAndUpdateNodes(configData, reason string) error {
	// 解析节点配置
	var newNodes []*Node
	if err := json.Unmarshal([]byte(configData), &newNodes); err != nil {
		return fmt.Errorf("解析节点配置失败: %v", err)
	}

	m.nodesMux.Lock()
	defer m.nodesMux.Unlock()

	// 记录变化
	var addedNodes, updatedNodes, removedNodes []*Node

	// 构建新节点映射
	newNodeMap := make(map[string]*Node)
	for _, node := range newNodes {
		newNodeMap[node.Id] = node
	}

	// 检查新增和更新的节点
	for nodeId, newNode := range newNodeMap {
		if oldNode, exists := m.nodes[nodeId]; exists {
			// 检查是否有变化
			if !reflect.DeepEqual(oldNode, newNode) {
				updatedNodes = append(updatedNodes, newNode)

				// 发布节点更新事件
				updateEvent := CreateNodeUpdatedEvent(newNode, oldNode, reason)
				m.eventManager.Publish(updateEvent)
			}
		} else {
			// 新增节点
			addedNodes = append(addedNodes, newNode)

			// 发布节点添加事件
			addEvent := CreateNodeAddedEvent(newNode, reason)
			m.eventManager.Publish(addEvent)
		}
	}

	// 检查删除的节点
	for nodeId, oldNode := range m.nodes {
		if _, exists := newNodeMap[nodeId]; !exists {
			removedNodes = append(removedNodes, oldNode)

			// 发布节点删除事件
			removeEvent := CreateNodeRemovedEvent(oldNode, reason)
			m.eventManager.Publish(removeEvent)
		}
	}

	// 更新内部节点映射
	m.nodes = newNodeMap

	// 如果有变化，发布集群整体变化事件
	if len(addedNodes) > 0 || len(updatedNodes) > 0 || len(removedNodes) > 0 {
		allNodes := make([]*Node, 0, len(newNodes))
		for _, node := range newNodes {
			allNodes = append(allNodes, node)
		}

		clusterEvent := CreateClusterChangedEvent(allNodes, addedNodes, updatedNodes, removedNodes)
		m.eventManager.Publish(clusterEvent)

		logger.Info("集群节点已更新",
			"totalNodes", len(allNodes),
			"added", len(addedNodes),
			"updated", len(updatedNodes),
			"removed", len(removedNodes),
			"reason", reason)
	}

	return nil
}

// GetAllNodes 获取所有节点
func (m *Manager) GetAllNodes() []*Node {
	m.nodesMux.RLock()
	defer m.nodesMux.RUnlock()

	nodes := make([]*Node, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

// GetNode 根据 ID 获取节点
func (m *Manager) GetNode(nodeId string) (*Node, bool) {
	m.nodesMux.RLock()
	defer m.nodesMux.RUnlock()

	node, exists := m.nodes[nodeId]
	return node, exists
}

// GetNodesByType 根据类型获取节点
func (m *Manager) GetNodesByType(nodeType int) []*Node {
	m.nodesMux.RLock()
	defer m.nodesMux.RUnlock()

	var nodes []*Node
	for _, node := range m.nodes {
		if node.Type == nodeType {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// PublishNodeConfig 发布节点配置到 Nacos
func (m *Manager) PublishNodeConfig(nodes []*Node) error {
	data, err := json.Marshal(nodes)
	if err != nil {
		return fmt.Errorf("序列化节点配置失败: %v", err)
	}

	success, err := m.configClient.PublishConfig(vo.ConfigParam{
		DataId:  m.config.ClusterConfig.DataId,
		Group:   m.config.ClusterConfig.Group,
		Content: string(data),
	})

	if err != nil {
		return fmt.Errorf("发布配置到 Nacos 失败: %v", err)
	}

	if !success {
		return fmt.Errorf("发布配置到 Nacos 失败：操作未成功")
	}

	logger.Info("节点配置已发布到 Nacos", "nodeCount", len(nodes))
	return nil
}

// AddNode 添加节点
func (m *Manager) AddNode(node *Node) error {
	nodes := m.GetAllNodes()

	// 检查节点是否已存在
	for _, existingNode := range nodes {
		if existingNode.Id == node.Id {
			return fmt.Errorf("节点 %s 已存在", node.Id)
		}
	}

	nodes = append(nodes, node)
	return m.PublishNodeConfig(nodes)
}

// UpdateNode 更新节点
func (m *Manager) UpdateNode(node *Node) error {
	nodes := m.GetAllNodes()

	// 查找并更新节点
	found := false
	for i, existingNode := range nodes {
		if existingNode.Id == node.Id {
			nodes[i] = node
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("节点 %s 不存在", node.Id)
	}

	return m.PublishNodeConfig(nodes)
}

// RemoveNode 删除节点
func (m *Manager) RemoveNode(nodeId string) error {
	nodes := m.GetAllNodes()

	// 查找并删除节点
	newNodes := make([]*Node, 0, len(nodes))
	found := false
	for _, node := range nodes {
		if node.Id != nodeId {
			newNodes = append(newNodes, node)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("节点 %s 不存在", nodeId)
	}

	return m.PublishNodeConfig(newNodes)
}

// GetStats 获取集群统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.nodesMux.RLock()
	defer m.nodesMux.RUnlock()

	stats := map[string]interface{}{
		"isRunning":   m.isRunning,
		"totalNodes":  len(m.nodes),
		"lastVersion": m.lastVersion,
	}

	// 按类型统计节点
	typeStats := make(map[int]int)
	for _, node := range m.nodes {
		typeStats[node.Type]++
	}
	stats["nodesByType"] = typeStats

	return stats
}
