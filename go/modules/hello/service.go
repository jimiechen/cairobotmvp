package hello

import (
	"context"

	"github.com/jimiechen/mineplanet/go/common-lib/module"
)

// Service Hello 模块服务
// 负责依赖装配和对外暴露接口
type Service struct {
	handler *Handler
}

// New 创建 Hello Service 实例（统一 Deps 装配入口）
// deps: 模块依赖，必须包含 Config 和 Logger
func New(deps module.Deps) *Service {
	usecase := NewUsecase(deps.Config, deps.I18n)
	handler := NewHandler(usecase, deps.Logger)

	return &Service{
		handler: handler,
	}
}

// SayHello 执行问候操作（保持接口兼容）
func (s *Service) SayHello(ctx context.Context, request []byte) ([]byte, error) {
	return s.handler.HandleSayHello(ctx, request)
}
