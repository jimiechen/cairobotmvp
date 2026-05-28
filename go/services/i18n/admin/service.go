package admin

import (
	"context"

	"github.com/jimiechen/mineplanet/go/services/i18n/repository"
	"github.com/jimiechen/mineplanet/go/services/i18n/service"
	"github.com/jimiechen/mineplanet/go/third_party/redisx"
)

// I18nAdminService 国际化管理后台服务统一接口
// 插件层（go-admin）依赖此接口，不依赖具体 AdminI18nService 实现
type I18nAdminService interface {
	I18nStringService
	I18nPackService
}

// I18nStringService 语言字符串管理操作接口
type I18nStringService interface {
	CreateString(ctx context.Context, req CreateStringRequest) (*StringItem, error)
	UpdateString(ctx context.Context, req UpdateStringRequest) (*StringItem, error)
	DeleteString(ctx context.Context, id int64, operator string) error
	ListStrings(packID int64) ([]*StringItem, error)
}

// I18nPackService 语言包管理操作接口
type I18nPackService interface {
	PublishPack(ctx context.Context, req PublishPackRequest) (*PackVersion, error)
	RollbackPack(ctx context.Context, packID int64, targetVersion int, operator string) error
	ImportStringsFromCSV(ctx context.Context, reader interface{}, packID int64, operator string) (*ImportResult, error)
	ExportStringsToCSV(ctx context.Context, req ExportCSVRequest) ([]byte, error)
}

// AdminI18nService 同时实现 I18nAdminService（嵌入两个子接口）
// 组合依赖：inner I18nService（校验） + repository（落库） + redisx（失效+广播）
type AdminI18nService struct {
	innerSvc    service.I18nService
	repo        repository.I18nRepository
	cache       redisx.Client
	pubsub      Publisher
	auditWriter AuditWriter
}

// I18nAdminOption 构造选项
type I18nAdminOption func(*AdminI18nService)

// WithCache 注入 Redis 缓存客户端
func WithCache(cache redisx.Client) I18nAdminOption {
	return func(s *AdminI18nService) { s.cache = cache }
}

// WithPubSub 注入 Pub/Sub 客户端
func WithPubSub(ps Publisher) I18nAdminOption {
	return func(s *AdminI18nService) { s.pubsub = ps }
}

// WithAuditWriter 注入审计写入器
func WithAuditWriter(w AuditWriter) I18nAdminOption {
	return func(s *AdminI18nService) { s.auditWriter = w }
}

// NewAdminI18nService 创建国际化管理后台服务实例
func NewAdminI18nService(
	innerSvc service.I18nService,
	repo repository.I18nRepository,
	opts ...I18nAdminOption,
) *AdminI18nService {
	s := &AdminI18nService{
		innerSvc: innerSvc,
		repo:     repo,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// invalidateLangCode 失效指定语言码的缓存 + 广播通知
func (s *AdminI18nService) invalidateLangCode(ctx context.Context, langCodes []string) error {
	if s.cache == nil {
		return nil
	}
	for _, lc := range langCodes {
		if err := s.cache.Invalidate(ctx, "i18n:"+lc); err != nil {
			return err
		}
	}
	if s.pubsub == nil {
		return nil
	}
	evt := i18nInvalidateEvent(langCodes)
	payload, err := marshalEvent(evt)
	if err != nil {
		return err
	}
	return s.pubsub.Publish(ctx, "cairobot.i18n.invalidate", payload)
}

// writeAudit 写入审计日志
func (s *AdminI18nService) writeAudit(ctx context.Context, action, targetType, targetID, operator string) {
	if s.auditWriter == nil {
		return
	}
	_ = s.auditWriter.Write(ctx, AuditEntry{
		TenantID:  "default",
		Action:    action,
		TargetType: targetType,
		TargetID:  targetID,
		Operator:  operator,
	})
}
