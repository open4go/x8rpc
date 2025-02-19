package call

import (
	"context"
	"github.com/open4go/x8rpc"
	"github.com/open4go/x9proto/pb/finance"
	"google.golang.org/grpc"
)

// FetchFinanceKeys 获取财务密钥
func FetchFinanceKeys(ctx context.Context, storeId string) (*finance.FianceKeyRsp, error) {
	// 定义客户端创建函数
	clientCreator := func(conn *grpc.ClientConn) finance.FinanceServiceClient {
		return finance.NewFinanceServiceClient(conn)
	}

	// 定义服务调用方法
	serviceMethod := func(client finance.FinanceServiceClient, ctx context.Context) (any, error) {
		reqPayload := &finance.FianceRequest{
			FinanceKeyId: storeId,
		}
		return client.FetchFinanceKey(ctx, reqPayload)
	}

	// 调用通用函数
	resp, err := x8rpc.CallGrpcService(ctx, "merchant", clientCreator, serviceMethod)
	if err != nil {
		return nil, err
	}

	// 强制类型断言并返回响应
	return resp.(*finance.FianceKeyRsp), nil
}
