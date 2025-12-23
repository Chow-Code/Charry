package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/charry/config"
	"github.com/charry/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// Node 节点信息
type Node struct {
	// 服务标识
	ServiceID   string // Consul 服务 ID
	Id          uint16
	Type        string
	Environment string

	// 服务配置
	Config *config.AppConfig

	// gRPC 连接
	conn   *grpc.ClientConn
	connMu sync.RWMutex

	// 状态
	status     NodeStatus
	statusMu   sync.RWMutex
	lastUpdate time.Time

	// 重连控制
	reconnectChan chan struct{}
	stopChan      chan struct{}
}

// NodeStatus 节点状态
type NodeStatus int

const (
	NodeStatusDisconnected NodeStatus = 0 // 未连接
	NodeStatusConnecting   NodeStatus = 1 // 连接中
	NodeStatusConnected    NodeStatus = 2 // 已连接
	NodeStatusFailed       NodeStatus = 3 // 连接失败
)

// NewNode 创建新节点
func NewNode(serviceID string, appConfig *config.AppConfig) *Node {
	return &Node{
		ServiceID:     serviceID,
		Id:            appConfig.Id,
		Type:          appConfig.Type,
		Environment:   appConfig.Environment,
		Config:        appConfig,
		status:        NodeStatusDisconnected,
		lastUpdate:    time.Now(),
		reconnectChan: make(chan struct{}, 1),
		stopChan:      make(chan struct{}),
	}
}

// Connect 建立 gRPC 连接
// 注意：超时由传入的 ctx 控制
func (n *Node) Connect(ctx context.Context) error {
	n.connMu.Lock()
	defer n.connMu.Unlock()

	if n.conn != nil {
		return nil // 已连接
	}

	n.setStatus(NodeStatusConnecting)

	target := fmt.Sprintf("%s:%d", n.Config.Addr.Host, n.Config.Addr.Port)
	logger.Infof("连接到节点: %s (%s)", n.ServiceID, target)

	// 创建 gRPC 连接
	// 注意：grpc.DialContext 在 1.x 中仍然支持，2.x 才会移除
	conn, err := grpc.DialContext(ctx, target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // 阻塞直到连接建立
	)
	if err != nil {
		n.setStatus(NodeStatusFailed)
		return fmt.Errorf("连接节点失败: %w", err)
	}

	n.conn = conn
	n.setStatus(NodeStatusConnected)
	logger.Infof("✓ 已连接到节点: %s", n.ServiceID)

	// 启动连接监控
	go n.monitorConnection()

	return nil
}

// Disconnect 断开连接
func (n *Node) Disconnect() {
	n.connMu.Lock()
	defer n.connMu.Unlock()

	close(n.stopChan)

	if n.conn != nil {
		n.conn.Close()
		n.conn = nil
		n.setStatus(NodeStatusDisconnected)
		logger.Infof("已断开节点: %s", n.ServiceID)
	}
}

// GetConn 获取 gRPC 连接
func (n *Node) GetConn() *grpc.ClientConn {
	n.connMu.RLock()
	defer n.connMu.RUnlock()
	return n.conn
}

// UpdateConfig 更新节点配置
func (n *Node) UpdateConfig(appConfig *config.AppConfig) {
	n.Config = appConfig
	n.lastUpdate = time.Now()
	logger.Infof("节点配置已更新: %s", n.ServiceID)
}

// ToJSON 转换节点信息为 JSON
func (n *Node) ToJSON() string {
	data := map[string]interface{}{
		"service_id":  n.ServiceID,
		"id":          n.Id,
		"type":        n.Type,
		"environment": n.Environment,
		"status":      n.GetStatus(),
		"last_update": n.lastUpdate.Format(time.RFC3339),
		"config":      n.Config,
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"error\": \"%v\"}", err)
	}
	return string(jsonBytes)
}

// GetStatus 获取节点状态
func (n *Node) GetStatus() NodeStatus {
	n.statusMu.RLock()
	defer n.statusMu.RUnlock()
	return n.status
}

// setStatus 设置节点状态
func (n *Node) setStatus(status NodeStatus) {
	n.statusMu.Lock()
	defer n.statusMu.Unlock()
	n.status = status
}

// monitorConnection 监控连接状态
func (n *Node) monitorConnection() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.stopChan:
			return
		case <-ticker.C:
			n.checkConnectionState()
		case <-n.reconnectChan:
			n.tryReconnect()
		}
	}
}

// checkConnectionState 检查连接状态
func (n *Node) checkConnectionState() {
	n.connMu.RLock()
	conn := n.conn
	n.connMu.RUnlock()

	if conn == nil {
		return
	}

	state := conn.GetState()
	if state == connectivity.TransientFailure || state == connectivity.Shutdown {
		logger.Warnf("节点连接异常: %s, 状态: %v", n.ServiceID, state)
		// 触发重连
		select {
		case n.reconnectChan <- struct{}{}:
		default:
		}
	}
}

// tryReconnect 尝试重连
func (n *Node) tryReconnect() {
	n.connMu.Lock()
	oldConn := n.conn
	if oldConn != nil {
		oldConn.Close()
		n.conn = nil
	}
	n.connMu.Unlock()

	logger.Infof("尝试重连节点: %s", n.ServiceID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := n.Connect(ctx); err != nil {
		logger.Errorf("重连节点失败: %s, %v", n.ServiceID, err)
		// 5 秒后再次尝试
		time.AfterFunc(5*time.Second, func() {
			select {
			case n.reconnectChan <- struct{}{}:
			default:
			}
		})
	}
}
