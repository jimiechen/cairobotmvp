package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	commonlib "github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/common-lib/health"
	"github.com/jimiechen/mineplanet/go/common-lib/i18n"
	"github.com/jimiechen/mineplanet/go/common-lib/module"
)

// Usecase Health 模块核心业务逻辑
type Usecase struct {
	cfg      module.ConfigReader
	i18n     module.I18nRenderer
	checkersMu sync.RWMutex
	checkers  []health.Checker
}

// NewUsecase 创建 Usecase 实例
func NewUsecase(cfg module.ConfigReader, i18n module.I18nRenderer, checkers []health.Checker) *Usecase {
	return &Usecase{
		cfg:      cfg,
		i18n:     i18n,
		checkers: checkers,
	}
}

// RegisterChecker 动态注册额外 Checker（线程安全）
func (u *Usecase) RegisterChecker(checker health.Checker) {
	u.checkersMu.Lock()
	defer u.checkersMu.Unlock()
	u.checkers = append(u.checkers, checker)
}

// getCheckers 线程安全地获取 Checker 列表
func (u *Usecase) getCheckers() []health.Checker {
	u.checkersMu.RLock()
	defer u.checkersMu.RUnlock()
	return u.checkers
}

// DoCheck 执行健康检查业务逻辑
func (u *Usecase) DoCheck(ctx context.Context, req *pb.ServiceHealthCheckRequest) (*pb.ServiceHealthCheckResponse, error) {
	version, _ := u.cfg.GetString(ctx, "system_cfg", "build_version")
	if version == "" {
		version = "0.0.0-dev"
	}

	var components []health.ComponentStatus
	checkers := u.getCheckers()
	if len(checkers) > 0 {
		components = u.runCheckersConcurrently(ctx, checkers)
	}

	healthyCount := countHealthyComponents(components)
	totalCount := len(components)

	status := "OK"
	resultCode := int32(commonlib.CodeSuccess)
	if healthyCount < totalCount {
		status = "Unhealthy"
		resultCode = int32(commonlib.CodeInternalError)
	}

	lang := i18n.ResolveLang(ctx, "", "", u.cfg)

	_ = u.renderStatusSummary(ctx, lang, healthyCount, totalCount)

	rsp := &pb.ServiceHealthCheckResponse{
		Result: &pb.Result{
			Code:    resultCode,
			Message: status,
		},
		Status:    status,
		Timestamp: time.Now().Unix(),
	}

	return rsp, nil
}

// GetComponentStatuses 返回各 Checker 的组件状态
// 注：proto 重生成（make proto）后，此数据将移入 rsp.Components 字段
func (u *Usecase) GetComponentStatuses(ctx context.Context) []health.ComponentStatus {
	checkers := u.getCheckers()
	return u.runCheckersConcurrently(ctx, checkers)
}

// GetVersion 返回版本号（供调用方获取详情）
// 注：proto 重生成后，此数据将移入 rsp.Version 字段
func (u *Usecase) GetVersion(ctx context.Context) string {
	v, _ := u.cfg.GetString(ctx, "system_cfg", "build_version")
	if v == "" {
		return "0.0.0-dev"
	}
	return v
}

// GetMessage 返回状态摘要（供调用方获取详情）
// 注：proto 重生成后，此数据将移入 rsp.Message 字段
func (u *Usecase) GetMessage(ctx context.Context) string {
	checkers := u.getCheckers()
	components := u.runCheckersConcurrently(ctx, checkers)
	healthyCount := countHealthyComponents(components)
	totalCount := len(components)
	lang := i18n.ResolveLang(ctx, "", "", u.cfg)
	return u.renderStatusSummary(ctx, lang, healthyCount, totalCount)
}

// runCheckersConcurrently 并发执行所有 Checker
func (u *Usecase) runCheckersConcurrently(ctx context.Context, checkers []health.Checker) []health.ComponentStatus {
	var wg sync.WaitGroup
	results := make([]health.ComponentStatus, len(checkers))

	for i, checker := range checkers {
		wg.Add(1)
		go func(idx int, c health.Checker) {
			defer wg.Done()
			status := health.CheckWithTimeout(ctx, c)
			results[idx] = status
		}(i, checker)
	}

	wg.Wait()
	return results
}

// countHealthyComponents 统计健康组件数量
func countHealthyComponents(components []health.ComponentStatus) int {
	count := 0
	for _, c := range components {
		if c.Healthy {
			count++
		}
	}
	return count
}

// renderStatusSummary 渲染状态摘要（i18n ICU plural 模板）
func (u *Usecase) renderStatusSummary(ctx context.Context, lang string, healthy int, total int) string {
	if total == 0 {
		return "No dependencies to check."
	}

	if u.i18n == nil {
		return fmt.Sprintf("%d of %d dependencies healthy", healthy, total)
	}

	message, err := u.i18n.T(ctx, lang, "svc_health_status_summary", map[string]any{
		"healthy": healthy,
		"total":   total,
	})
	if err != nil {
		return fmt.Sprintf("%d of %d dependencies healthy", healthy, total)
	}

	return message
}
