package topic

import (
	"context"

	"github.com/jimiechen/mineplanet/go/modules/social/event"
	"github.com/jimiechen/mineplanet/go/modules/social/member"
)

// Servant TarsGo Servant 注册与转发
// 职责：注册自身到 TarsServer + 提取 minType + 转发给 Handler
type Servant struct {
	handler    *Handler
	jwtMgr     *member.JWTManager  // 用于从 token 提取 jti 做黑名单检查
	tokenStore member.TokenStore  // 黑名单存储（可能为 nil，nil 时跳过检查）
}

// NewServant 创建 Servant 实例
func NewServant(repo Repository, publisher event.Publisher) *Servant {
	return &Servant{
		handler: NewHandler(repo, publisher),
	}
}

// InjectJWTManager 向 Handler 注入 JWT 管理器（延迟注入）
func (s *Servant) InjectJWTManager(m *member.JWTManager) {
	s.jwtMgr = m
	s.handler.InjectJWTManager(m)
}

// InjectTokenStore 向 Servant 注入令牌黑名单存储（延迟注入）
func (s *Servant) InjectTokenStore(ts member.TokenStore) {
	s.tokenStore = ts
	s.handler.InjectTokenStore(ts)
}

// Handle 实现 TarsGo Servant 接口
// 从 extend["minType"] 提取协议号，转发给 Handler.Dispatch
// 将 extend["user_id"] 桥接到 context.Context（AuthMiddleware 注入 → svc 使用）
//
// 鉴权路径黑名单检查（§8.2.1）：
//   在写入 user_id 到 context 之前，通过公共函数 CheckTokenBlacklist 检查 token 的 jti 是否在黑名单中。
//   若在黑名单中，返回空响应 + 错误码，阻断请求进入业务 SVC。
func (s *Servant) Handle(ctx context.Context, req []byte, extend map[string]string) (int, []byte, error) {
	// 鉴权路径黑名单检查：token 的 jti 在黑名单中则拒绝
	if code, respBytes, err := member.CheckTokenBlacklist(ctx, extend, s.jwtMgr, s.tokenStore); code != 0 {
		return code, respBytes, err
	}

	// 将 Gateway AuthMiddleware 注入的 user_id 从 extend 桥接到 Go context
	if userID, ok := extend["user_id"]; ok && userID != "" {
		ctx = WithUserID(ctx, userID)
	}

	minType := extend["minType"]
	respBytes, err := s.handler.Dispatch(ctx, minType, req)
	if err != nil {
		return 500, nil, err
	}
	return 200, respBytes, nil
}
