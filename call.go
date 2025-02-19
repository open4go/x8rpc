package x8rpc

import (
	"context"
	"github.com/open4go/log"
	"google.golang.org/grpc"
)

// CallGrpcService 通用 gRPC 服务调用函数
func CallGrpcService[T any](
	ctx context.Context,
	serverName string,
	clientCreator func(conn *grpc.ClientConn) T,
	serviceMethod func(client T, ctx context.Context) (any, error),
) (any, error) {
	// 获取连接池
	pool := GetDefaultPool(serverName)

	// 获取连接
	conn, err := pool.Get()
	if err != nil {
		log.Log(ctx).Error(err)
		return nil, err
	}

	if conn == nil || conn.Handler == nil {
		log.Log(ctx).Error("connection handler is nil")
		return nil, err
	}
	defer pool.Put(conn)

	// 创建客户端
	client := clientCreator(conn.Handler)

	// 调用服务方法
	resp, err := serviceMethod(client, ctx)
	if err != nil {
		log.Log(ctx).WithField("reason", "调用服务失败").Error(err)
		return nil, err
	}

	return resp, nil
}
