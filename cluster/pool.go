package cluster

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/charry/logger"
)

// ConnectionPool TCP 连接池
type ConnectionPool struct {
	// 连接列表
	conns []net.Conn
	mu    sync.RWMutex

	// 空闲连接队列（索引）
	freeConns chan int

	// 连接配置
	target   string
	poolSize int

	// 状态
	closed bool
}

// NewConnectionPool 创建连接池
func NewConnectionPool(target string, poolSize int) (*ConnectionPool, error) {
	if poolSize <= 0 {
		poolSize = 4 // 默认 4 个连接
	}

	pool := &ConnectionPool{
		conns:     make([]net.Conn, poolSize),
		freeConns: make(chan int, poolSize),
		target:    target,
		poolSize:  poolSize,
	}

	// 初始化连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var d net.Dialer
	for i := 0; i < poolSize; i++ {
		conn, err := d.DialContext(ctx, "tcp", target)
		if err != nil {
			// 清理已创建的连接
			pool.Close()
			return nil, fmt.Errorf("创建连接 %d 失败: %w", i, err)
		}
		pool.conns[i] = conn
		pool.freeConns <- i // 标记为空闲
	}

	logger.Infof("连接池创建成功: %s, 连接数: %d", target, poolSize)
	return pool, nil
}

// Get 获取一个连接（阻塞直到有可用连接）
func (p *ConnectionPool) Get() (net.Conn, error) {
	if p.closed {
		return nil, fmt.Errorf("连接池已关闭")
	}

	// 从空闲队列获取索引
	idx := <-p.freeConns

	p.mu.RLock()
	conn := p.conns[idx]
	p.mu.RUnlock()

	return conn, nil
}

// Put 归还连接
func (p *ConnectionPool) Put(conn net.Conn) {
	if p.closed {
		return
	}

	// 找到连接的索引
	p.mu.RLock()
	var idx int
	found := false
	for i, c := range p.conns {
		if c == conn {
			idx = i
			found = true
			break
		}
	}
	p.mu.RUnlock()

	if found {
		// 归还到空闲队列
		select {
		case p.freeConns <- idx:
		default:
			// 队列满了，不应该发生
			logger.Warn("连接池空闲队列已满")
		}
	}
}

// Close 关闭连接池
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}
	p.closed = true

	// 关闭所有连接
	for _, conn := range p.conns {
		if conn != nil {
			err := conn.Close()
			if err != nil {
				return
			}
		}
	}

	close(p.freeConns)
	logger.Infof("连接池已关闭: %s", p.target)
}

// GetPoolSize 获取连接池大小
func (p *ConnectionPool) GetPoolSize() int {
	return p.poolSize
}

// GetFreeCount 获取空闲连接数
func (p *ConnectionPool) GetFreeCount() int {
	return len(p.freeConns)
}
