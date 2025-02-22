package call

import (
	"context"
	"github.com/open4go/x8rpc"
	"github.com/open4go/x9proto/pb/order"
	"google.golang.org/grpc"
)

// UpdateOrderStatus 订单状态
func UpdateOrderStatus(ctx context.Context, openId string, inviter string, ip string) (*order.UpdateOrderStatusResponse, error) {
	// 定义客户端创建函数
	clientCreator := func(conn *grpc.ClientConn) order.OrderServiceClient {
		return order.NewOrderServiceClient(conn)
	}

	// 定义服务调用方法
	serviceMethod := func(client order.OrderServiceClient, ctx context.Context) (any, error) {
		reqPayload := &order.UpdateOrderStatusRequest{}
		return client.UpdateOrderStatus(ctx, reqPayload)
	}

	// 调用通用函数
	resp, err := x8rpc.CallGrpcService(ctx, "order", clientCreator, serviceMethod)
	if err != nil {
		return nil, err
	}

	// 强制类型断言并返回响应
	return resp.(*order.UpdateOrderStatusResponse), nil
}
