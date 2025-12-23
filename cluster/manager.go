package cluster

import (
	"context"
	"sync"
	"time"

	"github.com/charry/config"
	"github.com/charry/logger"
	consulapi "github.com/hashicorp/consul/api"
)

// Manager 集群管理器
type Manager struct {
	// 节点列表：serviceID -> Node
	nodes   map[string]*Node
	nodesMu sync.RWMutex

	// Consul 客户端（用于监听服务变化）
	consulClient *consulapi.Client

	// 停止通道
	stopChan chan struct{}
}

// NewManager 创建集群管理器
func NewManager(consulClient *consulapi.Client) *Manager {
	return &Manager{
		nodes:        make(map[string]*Node),
		consulClient: consulClient,
		stopChan:     make(chan struct{}),
	}
}

// AddNode 添加节点
func (m *Manager) AddNode(serviceID string, appConfig *config.AppConfig) error {
	m.nodesMu.Lock()
	defer m.nodesMu.Unlock()

	// 检查是否已存在
	if _, exists := m.nodes[serviceID]; exists {
		logger.Infof("节点已存在: %s", serviceID)
		return nil
	}

	// 创建节点
	node := NewNode(serviceID, appConfig)
	m.nodes[serviceID] = node

	logger.Infof("✓ 节点已添加: %s", serviceID)

	// 异步建立连接
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := node.Connect(ctx); err != nil {
			logger.Errorf("连接节点失败: %s, %v", serviceID, err)
		}
	}()

	return nil
}

// RemoveNode 移除节点
func (m *Manager) RemoveNode(serviceID string) {
	m.nodesMu.Lock()
	node, exists := m.nodes[serviceID]
	if exists {
		delete(m.nodes, serviceID)
	}
	m.nodesMu.Unlock()

	if node != nil {
		node.Disconnect()
		logger.Infof("✓ 节点已移除: %s", serviceID)
	}
}

// UpdateNode 更新节点配置
func (m *Manager) UpdateNode(serviceID string, appConfig *config.AppConfig) {
	m.nodesMu.RLock()
	node, exists := m.nodes[serviceID]
	m.nodesMu.RUnlock()

	if exists {
		node.UpdateConfig(appConfig)
	}
}

// GetNode 获取节点
func (m *Manager) GetNode(serviceID string) *Node {
	m.nodesMu.RLock()
	defer m.nodesMu.RUnlock()
	return m.nodes[serviceID]
}

// GetAllNodes 获取所有节点
func (m *Manager) GetAllNodes() []*Node {
	m.nodesMu.RLock()
	defer m.nodesMu.RUnlock()

	nodes := make([]*Node, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNodesByType 按类型获取节点
func (m *Manager) GetNodesByType(typ string) []*Node {
	m.nodesMu.RLock()
	defer m.nodesMu.RUnlock()

	nodes := make([]*Node, 0)
	for _, node := range m.nodes {
		if node.Type == typ {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// Close 关闭管理器
func (m *Manager) Close() {
	close(m.stopChan)

	m.nodesMu.Lock()
	defer m.nodesMu.Unlock()

	// 断开所有节点连接
	for _, node := range m.nodes {
		node.Disconnect()
	}
	m.nodes = make(map[string]*Node)

	logger.Info("✓ 集群管理器已关闭")
}

