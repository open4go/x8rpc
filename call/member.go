package call

import (
	"context"
	"github.com/open4go/x8rpc"
	"github.com/open4go/x9proto/pb/member"
	"google.golang.org/grpc"
)

// RegisterMember 注册会员
func RegisterMember(ctx context.Context, openId string, inviter string, ip string) (*member.RegisterMemberResponse, error) {
	// 定义客户端创建函数
	clientCreator := func(conn *grpc.ClientConn) member.MembershipServiceClient {
		return member.NewMembershipServiceClient(conn)
	}

	// 定义服务调用方法
	serviceMethod := func(client member.MembershipServiceClient, ctx context.Context) (any, error) {
		reqPayload := &member.RegisterMemberRequest{
			ThirdParty: &member.ThirdPartyInfo{
				Openid: openId,
			},
			AdditionalInfo: inviter,
		}
		return client.RegisterMember(ctx, reqPayload)
	}

	// 调用通用函数
	resp, err := x8rpc.CallGrpcService(ctx, "member", clientCreator, serviceMethod)
	if err != nil {
		return nil, err
	}

	// 强制类型断言并返回响应
	return resp.(*member.RegisterMemberResponse), nil
}

// FetchMemberByID 获取会员信息
func FetchMemberByID(ctx context.Context, id string) (*member.MemberDetail, error) {
	// 定义客户端创建函数
	clientCreator := func(conn *grpc.ClientConn) member.MembershipServiceClient {
		return member.NewMembershipServiceClient(conn)
	}

	// 定义服务调用方法
	serviceMethod := func(client member.MembershipServiceClient, ctx context.Context) (any, error) {
		reqPayload := &member.FetchByIDRequest{
			Id: id,
		}
		return client.FetchMemberByID(ctx, reqPayload)
	}

	// 调用通用函数
	resp, err := x8rpc.CallGrpcService(ctx, "member", clientCreator, serviceMethod)
	if err != nil {
		return nil, err
	}

	// 强制类型断言并返回响应
	return resp.(*member.MemberDetail), nil
}

// FetchMemberByOpenID 获取会员信息
func FetchMemberByOpenID(ctx context.Context, id string) (*member.MemberDetail, error) {
	// 定义客户端创建函数
	clientCreator := func(conn *grpc.ClientConn) member.MembershipServiceClient {
		return member.NewMembershipServiceClient(conn)
	}

	// 定义服务调用方法
	serviceMethod := func(client member.MembershipServiceClient, ctx context.Context) (any, error) {
		reqPayload := &member.FetchByThirdPartyRequest{
			Openid: id,
		}
		return client.FetchMemberByOpenID(ctx, reqPayload)
	}

	// 调用通用函数
	resp, err := x8rpc.CallGrpcService(ctx, "member", clientCreator, serviceMethod)
	if err != nil {
		return nil, err
	}

	// 强制类型断言并返回响应
	return resp.(*member.MemberDetail), nil
}

// FetchMemberIdentityInfo 获取会员信息
func FetchMemberIdentityInfo(ctx context.Context, id string) (*member.IdentityInfo, error) {
	// 定义客户端创建函数
	clientCreator := func(conn *grpc.ClientConn) member.MembershipServiceClient {
		return member.NewMembershipServiceClient(conn)
	}

	// 定义服务调用方法
	serviceMethod := func(client member.MembershipServiceClient, ctx context.Context) (any, error) {
		reqPayload := &member.FetchByThirdPartyRequest{
			Openid: id,
		}
		return client.FetchMemberByOpenID(ctx, reqPayload)
	}

	// 调用通用函数
	resp, err := x8rpc.CallGrpcService(ctx, "member", clientCreator, serviceMethod)
	if err != nil {
		return nil, err
	}

	// 强制类型断言并返回响应
	return resp.(*member.IdentityInfo), nil
}
