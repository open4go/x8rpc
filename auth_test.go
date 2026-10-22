package x8rpc

import (
	"context"
	"net"
	"testing"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func setCfg(t *testing.T, key string, v any) {
	t.Helper()
	prev := viper.Get(key)
	viper.Set(key, v)
	t.Cleanup(func() { viper.Set(key, prev) })
}

func startServer(t *testing.T) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(AuthUnaryInterceptor()))
	healthpb.RegisterHealthServer(s, health.NewServer())
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)
	return lis.Addr().String()
}

func call(t *testing.T, addr string, opts ...grpc.DialOption) codes.Code {
	t.Helper()
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	cc, err := grpc.NewClient(addr, opts...)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cc.Close() }()
	_, err = healthpb.NewHealthClient(cc).Check(context.Background(), &healthpb.HealthCheckRequest{})
	return status.Code(err)
}

func TestAuthUnaryInterceptor(t *testing.T) {
	setCfg(t, "rpc.auth_token", "s3cret")
	setCfg(t, "middle.rpc.allowUnauthenticated", false)
	addr := startServer(t)

	if got := call(t, addr); got != codes.Unauthenticated {
		t.Fatalf("no token: %v", got)
	}
	if got := call(t, addr, grpc.WithUnaryInterceptor(ClientTokenInterceptor())); got != codes.OK {
		t.Fatalf("with token: %v", got)
	}

	// 客户端令牌错误
	wrong := grpc.WithUnaryInterceptor(func(ctx context.Context, m string, req, reply any, c *grpc.ClientConn, inv grpc.UnaryInvoker, o ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, InternalTokenMetadata, "wrong")
		return inv(ctx, m, req, reply, c, o...)
	})
	if got := call(t, addr, wrong); got != codes.Unauthenticated {
		t.Fatalf("wrong token: %v", got)
	}
}

func TestAuthUnaryInterceptorNoConfigRejects(t *testing.T) {
	setCfg(t, "rpc.auth_token", "")
	setCfg(t, "rpc.token", "")
	setCfg(t, "middle.rpc.allowUnauthenticated", false)
	addr := startServer(t)
	if got := call(t, addr, grpc.WithUnaryInterceptor(ClientTokenInterceptor())); got != codes.Unauthenticated {
		t.Fatalf("no config: %v", got)
	}
}

func TestAuthUnaryInterceptorOptOut(t *testing.T) {
	setCfg(t, "rpc.auth_token", "s3cret")
	setCfg(t, "middle.rpc.allowUnauthenticated", true)
	addr := startServer(t)
	if got := call(t, addr); got != codes.OK {
		t.Fatalf("opt-out: %v", got)
	}
}
