package member

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcLogin 用户登录服务（minType=1023 UserLogin）
// 负责用户身份验证：参数校验 → 查找用户 → 密码比对 → 生成 JWT 令牌
// 不负责注册（由 SvcRegister 负责）
type SvcLogin struct {
	repo       Repository
	jwtManager *JWTManager
}

// NewSvcLogin 创建登录服务实例
func NewSvcLogin(repo Repository, jwtManager *JWTManager) *SvcLogin {
	return &SvcLogin{repo: repo, jwtManager: jwtManager}
}

// Handle 处理用户登录请求，遵循 DevGuide §7 五步模式
func (s *SvcLogin) Handle(ctx context.Context, req *pb.UserLoginRequest) (*pb.UserLoginResponse, error) {
	// Step 1: 参数校验
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 登录为公开操作，无需权限

	// Step 3: 1级数据读写 — 根据用户名查找用户 + 密码验证
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return &pb.UserLoginResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_UNAUTHORIZED),
				Message: "用户名或密码错误",
			},
		}, nil
	}

	// 验证密码（使用 bcrypt 安全比对）
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return &pb.UserLoginResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_UNAUTHORIZED),
				Message: "用户名或密码错误",
			},
		}, nil
	}

	// 检查用户状态：仅 active 用户允许登录
	if user.Status != UserStatusActive {
		return &pb.UserLoginResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_FORBIDDEN),
				Message: "用户状态异常，无法登录",
			},
		}, nil
	}

	// 更新最后登录时间
	now := time.Now().UnixMilli()
	user.LastLoginAt = now
	user.LoginCount++
	_ = s.repo.UpdateUser(ctx, user) // MVP-P0 忽略更新错误

	// Step 4: 领域事件 — 登录成功事件（MVP-P0 可延迟）

	// Step 5: 生成 JWT 令牌对
	accessToken, accessExp, err := s.jwtManager.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("生成 access_token 失败: %w", err)
	}
	refreshToken, _, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("生成 refresh_token 失败: %w", err)
	}

	return &pb.UserLoginResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "登录成功",
		},
		UserId:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserInfo: &base.UserInfo{
			UserId:   user.ID,
			Username: user.Username,
			Email:    user.Email,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		},
		ExpiresAt: accessExp,
	}, nil
}

// validateRequest 校验登录请求必填字段
func (s *SvcLogin) validateRequest(req *pb.UserLoginRequest) (*pb.UserLoginResponse, error) {
	if req.Username == "" {
		return &pb.UserLoginResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "用户名不能为空",
			},
		}, nil
	}
	if req.Password == "" {
		return &pb.UserLoginResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "密码不能为空",
			},
		}, nil
	}
	return nil, nil
}
