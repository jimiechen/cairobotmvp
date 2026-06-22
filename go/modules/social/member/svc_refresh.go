package member

import (
	"context"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcRefresh 令牌刷新服务（minType=1027 RefreshToken）
// 负责用 refresh_token 换取新的 access_token + refresh_token 对
// 刷新流程：验证 refresh_token → 检查黑名单 → 签发新令牌对 → 返回
type SvcRefresh struct {
	tokenStore  TokenStore
	jwtManager *JWTManager
	repo       Repository
}

// NewSvcRefresh 创建刷新服务实例
func NewSvcRefresh(tokenStore TokenStore, jwtManager *JWTManager, repo Repository) *SvcRefresh {
	return &SvcRefresh{
		tokenStore:  tokenStore,
		jwtManager: jwtManager,
		repo:       repo,
	}
}

// Handle 处理刷新令牌请求，遵循 DevGuide §7 五步模式
func (s *SvcRefresh) Handle(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	// Step 1: 参数校验
	if req.RefreshToken == "" {
		return &pb.RefreshTokenResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "refresh_token 不能为空",
			},
		}, nil
	}

	// Step 2: 解析并验证 refresh_token
	userID, tokenType, err := s.jwtManager.ParseToken(req.RefreshToken)
	if err != nil {
			code := base.ErrorCode_ERROR_CODE_UNAUTHORIZED
			if err == ErrTokenExpired {
				// proto 无 TOKEN_EXPIRED 码，使用 UNAUTHORIZED + 特殊 message 区分
				code = base.ErrorCode_ERROR_CODE_UNAUTHORIZED
			}
			return &pb.RefreshTokenResponse{
				Result: &base.Result{
					Code:    int32(code),
					Message: "refresh_token 无效或已过期",
				},
			}, nil
		}

	// 必须是 refresh 类型的令牌
	if tokenType != "refresh" {
		return &pb.RefreshTokenResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST),
				Message: "令牌类型错误：期望 refresh",
			},
		}, nil
	}

	// Step 3: 检查 refresh_token 的 jti 是否在黑名单中
	if s.tokenStore != nil && s.jwtManager != nil {
		jti := s.jwtManager.ParseJTI(req.RefreshToken)
		if jti != "" {
			blacklisted, checkErr := s.tokenStore.Exists(ctx, jti)
			if checkErr != nil {
				return nil, checkErr
			}
			if blacklisted {
				return &pb.RefreshTokenResponse{
					Result: &base.Result{
						Code:    int32(base.ErrorCode_ERROR_CODE_UNAUTHORIZED),
						Message: "refresh_token 已被撤销",
					},
				}, nil
			}
		}
	}

	// 验证用户仍然存在且状态正常
	user, findErr := s.repo.GetUserByID(ctx, userID)
	if findErr != nil || user == nil {
		return &pb.RefreshTokenResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_NOT_FOUND),
				Message: "用户不存在",
			},
		}, nil
	}
	if user.Status != UserStatusActive {
		return &pb.RefreshTokenResponse{
			Result: &base.Result{
				Code:    int32(base.ErrorCode_ERROR_CODE_FORBIDDEN),
				Message: "用户状态异常",
			},
		}, nil
	}

	// Step 4: 将旧 refresh_token 的 jti 加入黑名单（一次性使用，Rotation 策略）
	if s.tokenStore != nil && s.jwtManager != nil {
		oldJTI := s.jwtManager.ParseJTI(req.RefreshToken)
		if oldJTI != "" {
			_ = s.tokenStore.Store(ctx, oldJTI, int64(DefaultRefreshTTL.Seconds()))
		}
	}

	// 签发新的令牌对
	newAccessToken, accessExp, genErr := s.jwtManager.GenerateAccessToken(userID)
	if genErr != nil {
		return nil, genErr
	}
	newRefreshToken, _, genErr := s.jwtManager.GenerateRefreshToken(userID)
	if genErr != nil {
		return nil, genErr
	}

	// Step 5: 返回新令牌对
	return &pb.RefreshTokenResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "刷新成功",
		},
		UserId:       userID,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    accessExp,
	}, nil
}
