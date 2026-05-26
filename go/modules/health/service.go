package health

import (
	"context"

	"github.com/jimiechen/mineplanet/go/common-lib/health"
	"github.com/jimiechen/mineplanet/go/common-lib/module"
	"github.com/jimiechen/mineplanet/go/third_party/mysqlx"
	"github.com/jimiechen/mineplanet/go/third_party/redisx"
)

// Service Health 模块服务
type Service struct {
	handler *Handler
}

// New 创建 Health Service 实例（统一 Deps 装配入口）
func New(deps module.Deps, checkers []health.Checker) *Service {
	allCheckers := buildDefaultCheckers(deps, checkers)

	usecase := NewUsecase(deps.Config, deps.I18n, allCheckers)
	handler := NewHandler(usecase, deps.Logger)

	return &Service{
		handler: handler,
	}
}

// Register 动态注册额外 Checker
func (s *Service) Register(checker health.Checker) {
	s.handler.Register(checker)
}

// buildDefaultCheckers 构建默认 Checker 列表
func buildDefaultCheckers(deps module.Deps, extra []health.Checker) []health.Checker {
	checkers := make([]health.Checker, 0, len(extra)+4)

	checkers = append(checkers,
		NewConfigChecker(deps.Config),
		NewI18nChecker(deps.I18n),
	)

	var db mysqlx.DB
	if deps.DB != nil {
		if v, ok := deps.DB.(mysqlx.DB); ok {
			db = v
		}
	}
	checkers = append(checkers, NewMySQLChecker(db))

	var cache redisx.Client
	if deps.Cache != nil {
		if v, ok := deps.Cache.(redisx.Client); ok {
			cache = v
		}
	}
	checkers = append(checkers, NewRedisChecker(cache))

	checkers = append(checkers, extra...)

	return checkers
}

// Check 执行健康检查
func (s *Service) Check(ctx context.Context, request []byte) ([]byte, error) {
	return s.handler.HandleCheck(ctx, request)
}
