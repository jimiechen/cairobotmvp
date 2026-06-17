package member

import (
	"context"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// ========== 以下为 MVP-P0 白名单协议的 svc 存根 ==========
// 每个 svc 存根提供编译所需的类型定义和 Handle 方法签名
// 后续按 TDD 流程逐个替换为完整实现
// 规则：一协议一 svc 文件，实现后删除此处的对应存根

// SvcRefresh 令牌刷新服务（minType=1027 RefreshToken）
// TODO(ai, batch1): TDD 实现
type SvcRefresh struct{}

func NewSvcRefresh() *SvcRefresh {
	return &SvcRefresh{}
}

func (s *SvcRefresh) Handle(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	return &pb.RefreshTokenResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_UNAUTHORIZED),
			Message: "TODO: Refresh 未实现",
		},
	}, nil
}
