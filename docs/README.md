# x8rpc — gRPC 客户端连接池与调用封装

> 仓库：https://github.com/open4go/x8rpc ｜ 模块：`github.com/open4go/x8rpc` ｜ Go 1.22.3 ｜ 最近提交：2026-04-19（最新 tag v0.1.3）

## 项目简介
按 `grpc.<serverName>` 配置的地址，为每个地址维护一个 `*grpc.ClientConn` 池，并提供泛型调用函数 `CallGrpcService`；`call/` 包封装了订单、会员、财务服务的具体调用。

## 目录结构
```
x8rpc/
├── init.go   ConnectionPool：GetDefaultPool / GetConnectionPool / Get / Put / Close
├── call.go   CallGrpcService[T any](ctx, serverName, clientCreator, serviceMethod) (any, error)
└── call/     order.go（UpdateOrderStatus）、member.go、finance.go
```

## 使用示例（call/order.go）
```go
resp, err := x8rpc.CallGrpcService(ctx, "order",
    func(conn *grpc.ClientConn) order.OrderServiceClient { return order.NewOrderServiceClient(conn) },
    func(client order.OrderServiceClient, ctx context.Context) (any, error) {
        return client.UpdateOrderStatus(ctx, &order.UpdateOrderStatusRequest{})
    })
return resp.(*order.UpdateOrderStatusResponse), nil
```
配置：
```yaml
grpc:
  order: order-api:50051   # 不配置时默认 localhost:50051，并打 Warning
```

## 依赖
`google.golang.org/grpc v1.70.0`、`viper`、`open4go/log v0.0.16`、`open4go/x9proto v0.0.27`

## 构建
`go build`、`go vet` 通过；没有测试，没有 CI。

## 与其他仓库的关系
依赖 `x9proto` 的生成代码（落后最新版 3 个 tag）；被 r2day 业务服务用于服务间调用。
