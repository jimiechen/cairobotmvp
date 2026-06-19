package member

import (
	"context"
	"fmt"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcUpdateUserInfo 更新用户信息服务（minType=1031 UpdateUserInfo）
// 负责更新当前登录用户的昵称、头像、邮箱、手机号等资料
// 不负责用户认证，不负责权限判断（只能修改自己的信息）
type SvcUpdateUserInfo struct {
	repo Repository
}

// NewSvcUpdateUserInfo 创建更新用户信息服务实例
func NewSvcUpdateUserInfo(repo Repository) *SvcUpdateUserInfo {
	return &SvcUpdateUserInfo{repo: repo}
}

// Handle 处理更新用户信息请求，遵循 DevGuide §7 五步模式
func (s *SvcUpdateUserInfo) Handle(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UpdateUserInfoResponse, error) {
	// Step 1: 参数校验 — 从上下文提取 userId
	userID := ctx.Value(CtxKeyUserID)
	if userID == nil || userID.(string) == "" {
		return &pb.UpdateUserInfoResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "缺少用户身份信息",
			},
		}, nil
	}
	uidStr := userID.(string)

	// Step 2: 权限校验 — 修改自身信息为已认证操作

	// Step 3: 1级数据读写 — 查询用户 → 应用变更 → 保存
	user, err := s.repo.GetUserByID(ctx, uidStr)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return &pb.UpdateUserInfoResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_FORBIDDEN),
				Message: "用户不存在",
			},
		}, nil
	}

	// 仅更新非空字段（部分更新语义）
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	user.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	// Step 4: 领域事件 — 无

	// Step 5: 返回响应
	return &pb.UpdateUserInfoResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "更新成功",
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
			MembershipLevel: base.MembershipLevel(base.MembershipLevel_value[user.MembershipLevel]),
		},
		UserId: user.ID,
	}, nil
}
