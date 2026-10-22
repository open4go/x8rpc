# x8rpc 优化方案

| # | 优先级 | 问题 | 位置 | 建议 | 工作量 |
|---|---|---|---|---|---|
| 1 | 高 | 连接池上限不起作用：池空时 `len(p.connections)` 是 0，永远小于 maxSize，所以每个并发请求都会新建连接（拨号还是 `WithBlock`），高并发时连接数没有上限，延迟也会被拨号拖慢 | `init.go:107` | `grpc.ClientConn` 本身支持 HTTP/2 多路复用，**每个地址只需要一个长连接**：改成 `map[address]*grpc.ClientConn` 加懒加载，去掉 channel 池 | M |
| 2 | 高 | `UpdateOrderStatus` 发出的是空请求 `&order.UpdateOrderStatusRequest{}`，参数 openId/inviter/ip 都没用上，调用必定达不到预期 | `call/order.go` | 把入参填进请求，或者删除这个未完成的函数（member.go、finance.go 都正常填了参数） | S |
| 3 | 中 | handler 为 nil 时 `return nil, err`，而此时 err 是 nil，调用方拿到 `(nil, nil)` 后做类型断言会 panic | `call.go:28` | 返回 `errors.New("nil grpc connection")` | S |
| 4 | 中 | 用了已废弃的 `grpc.DialContext` + `WithInsecure` + `WithBlock` | `init.go:118-125` | 改用 `grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))`；生产环境支持 TLS 或 mTLS | S |
| 5 | 中 | 连接池保存了第一个调用方的 ctx（可能是请求 ctx），后续日志都会带上这个早已结束的请求的 trace | `init.go` ConnectionPool.ctx | 不保存 ctx，日志用 `context.Background()` 或调用方传入的 ctx | S |
| 6 | 中 | `CallGrpcService` 返回 `any`，调用方要强制断言 | `call.go` | 改为泛型返回 `func CallGrpc[C, R any](...) (R, error)` | S |
| 7 | 中 | 没有超时、重试和拦截器（trace 透传） | `call.go` | 默认 `context.WithTimeout`；加 unary interceptor 透传 `X-Request-ID` 的 metadata | M |
| 8 | 中 | govulncheck：grpc v1.70.0（2 个）、x/net、docker（来自旧版 log） | `go.mod` | 升级 grpc、log、x9proto | S |
| 9 | 低 | README 只有一行 “test”，没有测试也没有 CI | — | 用 bufconn 写集成测试，加 CI | M |
