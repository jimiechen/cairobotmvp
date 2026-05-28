package admin

import (
	"context"

	"github.com/jimiechen/mineplanet/go/services/config/repository"
	"github.com/jimiechen/mineplanet/go/services/config/service"
	"github.com/jimiechen/mineplanet/go/third_party/redisx"
)

// Publisher 变更广播接口
type Publisher interface {
	Publish(ctx context.Context, channel, message string) error
}

// ConfigAdminService 配置管理后台服务统一接口
// 插件层（go-admin）依赖此接口，不依赖具体 AdminConfigService 实现
// 合并了 Schema CRUD + Value 发布两类操作
type ConfigAdminService interface {
	ConfigSchemaService
	ConfigValueService
}

// ConfigSchemaService Schema 管理操作接口
type ConfigSchemaService interface {
	ListSchemas(ctx context.Context, moduleKey string) ([]*SchemaItem, error)
	CreateSchema(ctx context.Context, req CreateSchemaRequest) (*SchemaItem, error)
	UpdateSchema(ctx context.Context, req UpdateSchemaRequest) (*SchemaItem, error)
	DeleteSchema(ctx context.Context, id int64, operator string) error
}

// ConfigValueService 配置值管理操作接口
type ConfigValueService interface {
	PublishValue(ctx context.Context, req PublishValueRequest) (*ValueVersion, error)
}

// AdminConfigService 同时实现 ConfigAdminService（嵌入两个子接口）
// 组合依赖：inner SchemaService（校验+CRUD） + SchemaRepository（直接落库） + redisx（失效+广播）
type AdminConfigService struct {
	innerSchema *service.SchemaService
	schemaRepo  repository.SchemaRepository
	configRepo  repository.ConfigRepository
	cache       redisx.Client
	pubsub      Publisher
	auditWriter AuditWriter
}

// AdminOption AdminConfigService 构造选项
type AdminOption func(*AdminConfigService)

// WithCache 注入 Redis 缓存客户端（用于 Invalidate 失效）
func WithCache(cache redisx.Client) AdminOption {
	return func(s *AdminConfigService) {
		s.cache = cache
	}
}

// WithPubSub 注入 Pub/Sub 客户端（用于变更广播）
func WithPubSub(ps Publisher) AdminOption {
	return func(s *AdminConfigService) {
		s.pubsub = ps
	}
}

// WithAuditWriter 注入审计日志写入器
func WithAuditWriter(w AuditWriter) AdminOption {
	return func(s *AdminConfigService) {
		s.auditWriter = w
	}
}

// NewAdminConfigService 创建配置管理后台服务实例
// innerSchema 必须非 nil，用于委托校验和基础 CRUD
func NewAdminConfigService(
	innerSchema *service.SchemaService,
	schemaRepo repository.SchemaRepository,
	configRepo repository.ConfigRepository,
	opts ...AdminOption,
) *AdminConfigService {
	s := &AdminConfigService{
		innerSchema: innerSchema,
		schemaRepo:  schemaRepo,
		configRepo:  configRepo,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// invalidateAndBroadcast 执行缓存失效 + 变更广播
// 统一使用 InvalidateEvent JSON payload 格式
func (s *AdminConfigService) invalidateAndBroadcast(ctx context.Context, moduleKeys []string) error {
	if s.cache == nil {
		return nil
	}
	for _, mk := range moduleKeys {
		if err := s.cache.Invalidate(ctx, "sdk:"+mk); err != nil {
			return err
		}
	}
	if s.pubsub == nil {
		return nil
	}
	evt := sdkInvalidateEvent(moduleKeys)
	payload, err := marshalEvent(evt)
	if err != nil {
		return err
	}
	return s.pubsub.Publish(ctx, "cairobot.config.invalidate", payload)
}
