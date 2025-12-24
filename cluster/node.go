package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/charry/config"
	"github.com/charry/logger"
	"github.com/charry/tcp"
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

	// TCP 连接池
	connPool *ConnectionPool
	poolMu   sync.RWMutex

	// 消息路由器
	router *Router

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
		router:        NewRouter(),
	}
}

// Connect 建立 TCP 连接池
func (n *Node) Connect(ctx context.Context) error {
	n.poolMu.Lock()
	defer n.poolMu.Unlock()

	if n.connPool != nil {
		return nil // 已连接
	}

	n.setStatus(NodeStatusConnecting)

	target := fmt.Sprintf("%s:%d", n.Config.Addr.Host, n.Config.Addr.Port)
	logger.Infof("连接到节点: %s (%s)", n.ServiceID, target)

	// 获取连接池大小配置
	cfg := config.Get()
	poolSize := cfg.Server.ClusterConnCount
	if poolSize <= 0 {
		poolSize = 4 // 默认 4 个连接
	}

	// 创建连接池
	pool, err := NewConnectionPool(target, poolSize)
	if err != nil {
		n.setStatus(NodeStatusFailed)
		return fmt.Errorf("创建连接池失败: %w", err)
	}

	n.connPool = pool
	n.setStatus(NodeStatusConnected)
	logger.Infof("✓ 已连接到节点: %s (连接数: %d)", n.ServiceID, poolSize)

	// 立即发送第一次心跳（避免对方超时）
	go func() {
		conn, err := pool.Get()
		if err == nil {
			// 发送心跳
			err := tcp.SendHeartbeat(conn)
			if err != nil {
				return
			}
			// 等待响应
			_, err = tcp.DecodeMsg(conn)
			if err != nil {
				return
			}
			pool.Put(conn)
			logger.Infof("✓ 已发送初始心跳: %s", n.ServiceID)
		}
	}()

	// 启动监控协程和心跳
	go n.monitorConnection() // ⭐ 修复：启动监控协程
	go n.sendHeartbeat()

	return nil
}

// Disconnect 断开连接池
func (n *Node) Disconnect() {
	n.poolMu.Lock()
	defer n.poolMu.Unlock()

	close(n.stopChan)

	if n.connPool != nil {
		n.connPool.Close()
		n.connPool = nil
		n.setStatus(NodeStatusDisconnected)
		logger.Infof("已断开节点: %s", n.ServiceID)
	}
}

// GetPool 获取连接池
func (n *Node) GetPool() *ConnectionPool {
	n.poolMu.RLock()
	defer n.poolMu.RUnlock()
	return n.connPool
}

// RegisterHandler 注册消息处理器
func (n *Node) RegisterHandler(module, cmd uint32, handler MessageHandler) {
	n.router.Register(module, cmd, handler)
}

// SendReq 异步发送请求消息（不等待响应）
func (n *Node) SendReq(req *tcp.ReqMsg) error {
	pool := n.GetPool()
	if pool == nil {
		return fmt.Errorf("节点未连接")
	}

	// 从连接池获取连接
	conn, err := pool.Get()
	if err != nil {
		return fmt.Errorf("获取连接失败: %w", err)
	}
	defer pool.Put(conn) // 归还连接

	// 编码并发送
	data := tcp.EncodeReqMsg(req)
	_, err = conn.Write(data)
	if err != nil {
		// 触发重连
		select {
		case n.reconnectChan <- struct{}{}:
		default:
		}
		return fmt.Errorf("发送失败: %w", err)
	}

	return nil
}

// Send 发送原始字节流（兼容旧接口）
func (n *Node) Send(data []byte) ([]byte, error) {
	pool := n.GetPool()
	if pool == nil {
		return nil, fmt.Errorf("节点未连接")
	}

	conn, err := pool.Get()
	if err != nil {
		return nil, err
	}
	defer pool.Put(conn)

	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	response := make([]byte, 4096)
	bytesRead, err := conn.Read(response)
	if err != nil {
		return nil, err
	}

	return response[:bytesRead], nil
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

// checkConnectionState 检查连接池状态
// 通过心跳机制检测，这里可以简化
func (n *Node) checkConnectionState() {
	pool := n.GetPool()
	if pool == nil {
		return
	}

	// 连接池状态由心跳机制维护
	// 如果心跳失败会自动触发重连
}

// tryReconnect 尝试重连
func (n *Node) tryReconnect() {
	n.poolMu.Lock()
	oldPool := n.connPool
	if oldPool != nil {
		oldPool.Close()
		n.connPool = nil
	}
	n.poolMu.Unlock()

	logger.Infof("尝试重连节点: %s", n.ServiceID)

	// 创建新连接池（不启动新协程）
	target := fmt.Sprintf("%s:%d", n.Config.Addr.Host, n.Config.Addr.Port)
	cfg := config.Get()
	poolSize := cfg.Server.ClusterConnCount
	if poolSize <= 0 {
		poolSize = 4
	}

	pool, err := NewConnectionPool(target, poolSize)
	if err != nil {
		logger.Errorf("重连节点失败: %s, %v", n.ServiceID, err)
		// 5 秒后再次尝试
		time.AfterFunc(5*time.Second, func() {
			select {
			case n.reconnectChan <- struct{}{}:
			default:
			}
		})
		return
	}

	n.poolMu.Lock()
	n.connPool = pool
	n.poolMu.Unlock()
	n.setStatus(NodeStatusConnected)

	logger.Infof("✓ 节点重连成功: %s", n.ServiceID)
}

// receiveLoop 接收协程（每个连接一个）
func (n *Node) receiveLoop(conn net.Conn, connIndex int) {
	logger.Infof("接收协程启动: %s, 连接%d", n.ServiceID, connIndex)

	for {
		select {
		case <-n.stopChan:
			return
		default:
			// 解码消息
			msg, err := tcp.DecodeMsg(conn)
			if err != nil {
				logger.Warnf("连接%d 接收消息失败: %s, %v", connIndex, n.ServiceID, err)
				// 触发重连
				select {
				case n.reconnectChan <- struct{}{}:
				default:
				}
				return
			}

			// 分发消息
			switch v := msg.(type) {
			case *tcp.ReqMsg:
				// 收到请求消息（不应该发生，节点是客户端）
				logger.Warnf("节点收到请求消息: module=%d, cmd=%d", v.Module, v.Cmd)
			case *tcp.RespMsg:
				// 收到响应消息
				if tcp.IsHeartbeatMsg(v.Module, v.Cmd) {
					// 心跳响应，忽略
					continue
				}
				// 处理业务响应
				if err := n.router.HandleResp(v); err != nil {
					logger.Warnf("处理响应失败: %v", err)
				}
			}
		}
	}
}

// sendHeartbeat 定时发送心跳（所有连接都发送，异步不等待响应）
func (n *Node) sendHeartbeat() {
	ticker := time.NewTicker(tcp.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.stopChan:
			return
		case <-ticker.C:
			pool := n.GetPool()
			status := n.GetStatus()

			if pool != nil && status == NodeStatusConnected {
				// 对所有连接发送心跳
				poolSize := pool.GetPoolSize()
				var lastErr error

				for i := 0; i < poolSize; i++ {
					conn, err := pool.Get()
					if err != nil {
						lastErr = err
						continue
					}

					// 只发送心跳，不等待响应（接收协程会处理）
					err = tcp.SendHeartbeat(conn)
					pool.Put(conn) // 立即归还

					if err != nil {
						lastErr = err
					}
				}

				// 如果所有连接都失败，触发重连
				if lastErr != nil {
					logger.Warnf("发送心跳失败: %s, %v", n.ServiceID, lastErr)
					// 触发重连
					select {
					case n.reconnectChan <- struct{}{}:
					default:
					}
				}
			}
		}
	}
}
