package x8rpc

import (
	"context"
	"crypto/subtle"
	"strings"

	"github.com/open4go/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// 服务间 gRPC 鉴权，与 open4go/middle.RPCAuthMiddleware 使用同一套配置：
//
//	rpc.auth_token                    服务间共享令牌（兼容旧配置 rpc.token）
//	middle.rpc.allowUnauthenticated   true 时缺少或错误令牌的调用仍放行（只记告警），仅用于灰度窗口，默认 false
//
// 客户端：本包的连接池会自动在每次调用的 metadata 中加入 x-internal-token。
// 服务端：grpc.NewServer(grpc.ChainUnaryInterceptor(x8rpc.AuthUnaryInterceptor()))。
const InternalTokenMetadata = "x-internal-token"

func internalToken() string {
	if t := strings.TrimSpace(viper.GetString("rpc.auth_token")); t != "" {
		return t
	}
	return strings.TrimSpace(viper.GetString("rpc.token"))
}

func validToken(got string) bool {
	want := internalToken()
	got = strings.TrimSpace(got)
	if want == "" || got == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(want), []byte(got)) == 1
}

// WithInternalToken 返回附带服务间令牌的 outgoing context（未配置令牌时原样返回）。
func WithInternalToken(ctx context.Context) context.Context {
	t := internalToken()
	if t == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, InternalTokenMetadata, t)
}

// clientTokenInterceptor 为每次一元调用加上服务间令牌。
func clientTokenInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return invoker(WithInternalToken(ctx), method, req, reply, cc, opts...)
}

// ClientTokenInterceptor 供自行建连的调用方使用：grpc.WithUnaryInterceptor(x8rpc.ClientTokenInterceptor())。
func ClientTokenInterceptor() grpc.UnaryClientInterceptor {
	return clientTokenInterceptor
}

// AuthUnaryInterceptor 服务端校验 x-internal-token，失败返回 Unauthenticated。默认安全：未配置令牌时拒绝所有调用。
func AuthUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		token := ""
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get(InternalTokenMetadata); len(vals) > 0 {
				token = vals[0]
			}
		}
		if validToken(token) {
			return handler(ctx, req)
		}
		if viper.GetBool("middle.rpc.allowUnauthenticated") {
			method := ""
			if info != nil {
				method = info.FullMethod
			}
			log.Log(ctx).WithField("method", method).
				Warning("grpc call without valid internal token allowed by middle.rpc.allowUnauthenticated")
			return handler(ctx, req)
		}
		return nil, status.Error(codes.Unauthenticated, "missing or invalid internal token")
	}
}
