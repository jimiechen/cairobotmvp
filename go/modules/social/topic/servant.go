package topic

import (
	"context"
)

// Servant TarsGo Servant 注册与转发
type Servant struct {
	handler *Handler
}

// NewServant 创建 Servant 实例
func NewServant(repo Repository) *Servant {
	return &Servant{
		handler: NewHandler(repo),
	}
}

// Handle 实现 TarsGo Servant 接口
func (s *Servant) Handle(ctx context.Context, req []byte, extend map[string]string) (int, []byte, error) {
	minType := extend["minType"]
	respBytes, err := s.handler.Dispatch(ctx, minType, req)
	if err != nil {
		return 500, nil, err
	}
	return 200, respBytes, nil
}
