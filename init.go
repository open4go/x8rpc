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
	instance *ConnectionPool
	once     sync.Once
)

// GetDefaultPool 获取默认连接池的单例
func GetDefaultPool(serverName string) *ConnectionPool {
	address := viper.GetString("grpc." + serverName)
	once.Do(func() {
		instance = &ConnectionPool{
			connections: make(chan *Connection, defaultPool),
			address:     address,
			maxSize:     defaultPool,
			nextID:      0,
		}
	})
	return instance
}

// GetConnectionPool 获取连接池的单例
func GetConnectionPool(address string, maxSize int) *ConnectionPool {
	once.Do(func() {
		instance = &ConnectionPool{
			connections: make(chan *Connection, maxSize),
			address:     address,
			maxSize:     maxSize,
			nextID:      0,
		}
	})
	return instance
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
