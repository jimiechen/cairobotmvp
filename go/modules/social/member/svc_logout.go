package member

import (
	"context"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcLogout 用户登出服务（minType=1025 UserLogout）
// 负责使用户访问令牌失效
// MVP-P0 阶段：仅做参数校验和响应返回（令牌管理在 MVP1 引入 Redis 后实现）
type SvcLogout struct{}

// NewSvcLogout 创建登出服务实例
func NewSvcLogout() *SvcLogout {
	return &SvcLogout{}
}

// Handle 处理用户登出请求，遵循 DevGuide §7 五步模式（简化版）
func (s *SvcLogout) Handle(ctx context.Context, req *pb.UserLogoutRequest) (*pb.UserLogoutResponse, error) {
	// Step 1: 参数校验
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 登出为公开操作（令牌自身即凭证）

	// Step 3: 1级数据读写 — 使令牌失效
	// TODO(security, MVP1): 将 access_token 加入黑名单（Redis Set）
	// MVP-P0: 无状态令牌，客户端丢弃即可

	// Step 4: 领域事件 — 登出成功事件

	// Step 5: 返回响应
	return &pb.UserLogoutResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "登出成功",
		},
		UserId: req.UserId,
	}, nil
}

// validateRequest 校验登出请求必填字段
func (s *SvcLogout) validateRequest(req *pb.UserLogoutRequest) (*pb.UserLogoutResponse, error) {
	if req.UserId == "" {
		return &pb.UserLogoutResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "用户ID不能为空",
			},
		}, nil
	}
	return nil, nil
}
