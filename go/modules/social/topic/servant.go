package topic

import (
	"context"

	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// Servant TarsGo Servant 注册与转发
type Servant struct {
	handler *Handler
}

// NewServant 创建 Servant 实例
func NewServant(repo Repository, publisher event.Publisher) *Servant {
	return &Servant{
		handler: NewHandler(repo, publisher),
	}
}

// Handle 实现 TarsGo Servant 接口
// 从 extend["minType"] 提取协议号，转发给 Handler.Dispatch
// 将 extend["user_id"] 桥接到 context.Context（AuthMiddleware 注入 → svc 使用）
func (s *Servant) Handle(ctx context.Context, req []byte, extend map[string]string) (int, []byte, error) {
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
