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

// Handle 实现 TarsGo Servant 接口
// 从 extend["minType"] 提取协议号，转发给 Handler.Dispatch
func (s *Servant) Handle(ctx context.Context, req []byte, extend map[string]string) (int, []byte, error) {
	minType := extend["minType"]
	respBytes, err := s.handler.Dispatch(ctx, minType, req)
	if err != nil {
		return 500, nil, err
	}
	return 200, respBytes, nil
}
