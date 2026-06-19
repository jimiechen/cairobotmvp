package member

import (
	"context"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SvcLogout 用户登出服务（minType=1025 UserLogout）
// 负责将 access_token 加入黑名单使其失效
type SvcLogout struct {
	tokenStore  TokenStore
	jwtManager *JWTManager
}

// NewSvcLogout 创建登出服务实例
func NewSvcLogout(tokenStore TokenStore, jwtManager *JWTManager) *SvcLogout {
	return &SvcLogout{tokenStore: tokenStore, jwtManager: jwtManager}
}

// Handle 处理用户登出请求，遵循 DevGuide §7 五步模式
func (s *SvcLogout) Handle(ctx context.Context, req *pb.UserLogoutRequest) (*pb.UserLogoutResponse, error) {
	// Step 1: 参数校验
	if errResp, sysErr := s.validateRequest(req); sysErr != nil {
		return nil, sysErr
	} else if errResp != nil {
		return errResp, nil
	}

	// Step 2: 权限校验 — 登出为公开操作（令牌自身即凭证）

	// Step 3: 将 access_token 加入黑名单
	if req.AccessToken != "" && s.tokenStore != nil {
		// 解析令牌获取剩余有效期作为黑名单 TTL
		userID, tokenType, parseErr := s.jwtManager.ParseToken(req.AccessToken)
		if parseErr == nil && tokenType == "access" {
			// 计算剩余 TTL：默认使用 access TTL 上限
			ttl := DefaultAccessTTL
			if expClaims, ok := s.extractExpiry(req.AccessToken); ok {
				remaining := time.Until(expClaims)
				if remaining > 0 && remaining < ttl {
					ttl = remaining
				}
			}
			if err := s.tokenStore.Blacklist(ctx, req.AccessToken, ttl); err != nil {
				// 黑名单写入失败不阻断登出响应（best-effort）
				// 生产环境应记录告警日志
			}
		}
		_ = userID // 登出时仅用令牌本身做失效，不依赖 user_id
	}

	// Step 4: 领域事件 — 登出成功事件（后续接入事件系统）

	// Step 5: 返回响应
	return &pb.UserLogoutResponse{
		Result: &base.Result{
			Code:    int32(base.ErrorCode_ERROR_CODE_SUCCESS),
			Message: "登出成功",
		},
		UserId: req.UserId,
	}, nil
}

// extractExpiry 从令牌中提取过期时间（辅助方法）
func (s *SvcLogout) extractExpiry(tokenStr string) (time.Time, bool) {
	userID, _, err := s.jwtManager.ParseToken(tokenStr)
	if err != nil {
		return time.Time{}, false
	}
	_ = userID
	// 通过重新解析获取 exp claims（简化处理：返回零值）
	// 完整实现应在 ParseToken 中同时返回 expiry
	return time.Time{}, false
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
