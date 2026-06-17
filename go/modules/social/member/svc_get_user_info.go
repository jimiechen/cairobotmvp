package member

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// ctxKeyUserID 上下文键类型，用于从认证中间件获取当前用户 ID
type ctxKeyUserIDType struct{}

// ctxKeyUserID 当前用户 ID 的上下文键（由认证中间件注入）
var ctxKeyUserID = ctxKeyUserIDType{}

// SvcGetUserInfo 查询用户信息服务（minType=1029 GetUserInfo）
// 负责根据认证上下文中的 userId 查询并返回完整用户信息
// 不负责用户认证（由认证中间件负责），不负责权限判断
type SvcGetUserInfo struct {
	repo Repository
}

// NewSvcGetUserInfo 创建查询用户信息服务实例
func NewSvcGetUserInfo(repo Repository) *SvcGetUserInfo {
	return &SvcGetUserInfo{repo: repo}
}

// Handle 处理查询用户信息请求，遵循 DevGuide §7 五步模式
func (s *SvcGetUserInfo) Handle(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.GetUserInfoResponse, error) {
	// Step 1: 参数校验 — 从上下文提取 userId
	userID := ctx.Value(ctxKeyUserID)
	if userID == nil || userID.(string) == "" {
		return &pb.GetUserInfoResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "缺少用户身份信息",
			},
		}, nil
	}
	uidStr := userID.(string)

	// Step 2: 权限校验 — 查询自身信息为公开操作，无需额外权限

	// Step 3: 1级数据读写 — 根据 ID 查询用户
	user, err := s.repo.GetUserByID(ctx, uidStr)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return &pb.GetUserInfoResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_FORBIDDEN),
				Message: "用户不存在",
			},
		}, nil
	}

	// Step 4: 领域事件 — 无

	// Step 5: 返回响应
	return &pb.GetUserInfoResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "查询成功",
		},
		UserInfo: &base.UserInfo{
			UserId:          user.ID,
			Uid:             user.UID,
			Username:        user.Username,
			Email:           user.Email,
			Phone:           user.Phone,
			Nickname:        user.Nickname,
			Avatar:          user.Avatar,
			Status:          base.UserStatus(user.Status),
			CreatedAt:       user.CreatedAt,
			UpdatedAt:       user.UpdatedAt,
			LastLoginAt:     user.LastLoginAt,
			MembershipLevel: base.MembershipLevel(base.MembershipLevel_value[user.MembershipLevel]),
		},
		UserId: user.ID,
	}, nil
}
