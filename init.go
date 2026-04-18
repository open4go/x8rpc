package x8rpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/open4go/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultPoolSize = 20
	dialTimeout     = 5 * time.Second
)

// Connection 单个 gRPC 连接封装
type Connection struct {
	Handler *grpc.ClientConn
	ID      int
}

// ConnectionPool gRPC 连接池（按 address 隔离）
type ConnectionPool struct {
	connections chan *Connection
	address     string
	maxSize     int
	mu          sync.Mutex
	nextID      int
	ctx         context.Context
}

var (
	pools     = make(map[string]*ConnectionPool)
	poolsLock sync.RWMutex
)

// GetDefaultPool 获取默认连接池
func GetDefaultPool(ctx context.Context, serverName string) *ConnectionPool {
	address := viper.GetString("grpc." + serverName)
	if address == "" {
		log.Log(ctx).Warnf("grpc.%s address not configured, using default", serverName)
		address = "localhost:50051"
	}
	return GetConnectionPool(ctx, address, defaultPoolSize)
}

// GetConnectionPool 获取或创建连接池
func GetConnectionPool(ctx context.Context, address string, maxSize int) *ConnectionPool {
	if maxSize <= 0 {
		maxSize = defaultPoolSize
	}

	poolsLock.RLock()
	pool, exists := pools[address]
	poolsLock.RUnlock()

	if exists {
		return pool
	}

	poolsLock.Lock()
	defer poolsLock.Unlock()

	// 双重检查
	if pool, exists := pools[address]; exists {
		return pool
	}

	pool = &ConnectionPool{
		connections: make(chan *Connection, maxSize),
		address:     address,
		maxSize:     maxSize,
		nextID:      0,
		ctx:         ctx,
	}
	pools[address] = pool

	log.Log(ctx).Infof("Created new gRPC connection pool for %s, size: %d", address, maxSize)
	return pool
}

// Get 获取一个可用连接
func (p *ConnectionPool) Get() (*Connection, error) {
	// 1. 优先从池中取
	select {
	case conn := <-p.connections:
		// 简单健康检查
		if conn.Handler.GetState() == connectivity.Ready {
			return conn, nil
		}
		// 状态不对，关闭旧连接，创建新连接
		conn.Handler.Close()
	default:
		// 池为空，继续往下创建新连接
	}

	// 2. 创建新连接（带锁保护）
	p.mu.Lock()
	defer p.mu.Unlock()

	// 再次检查（防止并发创建过多）
	if len(p.connections) < p.maxSize {
		return p.createNewConnection()
	}

	return nil, fmt.Errorf("connection pool is full (address: %s, size: %d)", p.address, p.maxSize)
}

// createNewConnection 创建新连接（私有方法）
func (p *ConnectionPool) createNewConnection() (*Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, p.address,
		grpc.WithInsecure(), // 生产环境建议换成 TLS
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", p.address, err)
	}

	p.nextID++
	newConn := &Connection{
		Handler: conn,
		ID:      p.nextID,
	}

	log.Log(p.ctx).Infof("Created new gRPC connection to %s, ID: %d", p.address, newConn.ID)
	return newConn, nil
}

// Put 将连接放回连接池
func (p *ConnectionPool) Put(conn *Connection) {
	if conn == nil || conn.Handler == nil {
		return
	}

	select {
	case p.connections <- conn:
		// 放回成功
	default:
		// 连接池已满，关闭连接
		conn.Handler.Close()
		log.Log(p.ctx).Warnf("Connection pool full, closed connection ID: %d", conn.ID)
	}
}

// Close 关闭连接池中所有连接
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		select {
		case conn := <-p.connections:
			if conn != nil && conn.Handler != nil {
				conn.Handler.Close()
			}
		default:
			return
		}
	}
}
