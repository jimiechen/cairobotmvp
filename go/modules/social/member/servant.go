package member

import (
	"context"

	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// Servant TarsGo Servant 注册与转发
// 职责：注册自身到 TarsServer + 提取 minType + 转发给 Handler
// 不解析 proto bytes，不包含业务逻辑
type Servant struct {
	handler *Handler
}

// NewServant 创建 Servant 实例
func NewServant(repo Repository, publisher event.Publisher) *Servant {
	return &Servant{
		handler: NewHandler(repo, publisher),
	}
}

// InjectJWTManager 向 Handler 注入 JWT 管理器（延迟注入）
// 用于解决 Module 创建时 JWT 依赖尚未就绪的循环依赖问题
// 会重建内部 Handler 实例（保留原 repo 和 publisher）
func (s *Servant) InjectJWTManager(m *JWTManager) {
	s.handler.InjectJWTManager(m)
}

// InjectTokenStore 向 Handler 注入令牌黑名单存储（延迟注入）
// 用于解决 Module 创建时 Redis 依赖尚未就绪的循环依赖问题
func (s *Servant) InjectTokenStore(ts TokenStore) {
	s.handler.InjectTokenStore(ts)
}

// Handle 实现 TarsGo Servant 接口
// 从 extend["minType"] 提取协议号，转发给 Handler.Dispatch
// 将 extend["user_id"] 桥接到 context.Context（AuthMiddleware 注入 → svc 使用）
func (s *Servant) Handle(ctx context.Context, req []byte, extend map[string]string) (int, []byte, error) {
	// 将 Gateway AuthMiddleware 注入的 user_id 从 extend 桥接到 Go context
	if userID, ok := extend["user_id"]; ok && userID != "" {
		ctx = context.WithValue(ctx, CtxKeyUserID, userID)
	}

	minType := extend["minType"]
	respBytes, err := s.handler.Dispatch(ctx, minType, req)
	if err != nil {
		return 500, nil, err
	}
	return 200, respBytes, nil
}
