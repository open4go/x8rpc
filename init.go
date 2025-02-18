package x8rpc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"sync"
)

const defaultPool = 20

type Connection struct {
	Handler *grpc.ClientConn
	ID      int
}

type ConnectionPool struct {
	connections chan *Connection
	address     string
	maxSize     int
	mu          sync.Mutex
	nextID      int
}

var (
	pools     = make(map[string]*ConnectionPool) // 使用 map 来存储多个连接池
	poolsLock sync.RWMutex                       // 为了安全地读写连接池，使用读写锁
)

// GetDefaultPool 获取默认连接池的单例
func GetDefaultPool(serverName string) *ConnectionPool {
	address := viper.GetString("grpc." + serverName)
	return GetConnectionPool(address, defaultPool)
}

// GetConnectionPool 获取连接池，根据不同的 address 创建单独的池
func GetConnectionPool(address string, maxSize int) *ConnectionPool {
	poolsLock.RLock() // 使用读锁防止在获取时的并发问题
	pool, exists := pools[address]
	poolsLock.RUnlock()

	// 如果连接池已存在，直接返回
	if exists {
		return pool
	}

	// 如果池不存在，创建一个新的连接池
	poolsLock.Lock() // 获取写锁来创建新池
	defer poolsLock.Unlock()

	// 防止并发情况下多次创建相同的连接池
	if pool, exists := pools[address]; exists {
		return pool
	}

	// 创建并初始化新的连接池
	pool = &ConnectionPool{
		connections: make(chan *Connection, maxSize),
		address:     address,
		maxSize:     maxSize,
		nextID:      0,
	}
	pools[address] = pool
	return pool
}

// Get 获取连接
func (p *ConnectionPool) Get() (*Connection, error) {
	select {
	case conn := <-p.connections:
		return conn, nil
	default:
		p.mu.Lock()
		defer p.mu.Unlock()

		if len(p.connections) < p.maxSize {
			conn, err := grpc.Dial(p.address, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				return nil, err
			}
			p.nextID++
			return &Connection{Handler: conn, ID: p.nextID}, nil
		}

		return nil, nil // 连接池已满
	}
}

// Put 将连接放回连接池
func (p *ConnectionPool) Put(conn *Connection) {
	select {
	case p.connections <- conn:
	default:
		conn.Handler.Close() // 连接池已满，关闭连接
	}
}

// Cleanup 清理超过最大限制的连接
func (p *ConnectionPool) Cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.connections) > p.maxSize {
		conn := <-p.connections
		conn.Handler.Close() // 关闭多余的连接
	}
}

// Close 关闭连接池中的所有连接
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for {
		select {
		case conn := <-p.connections:
			conn.Handler.Close()
		default:
			return
		}
	}
}
