package service

import (
	"github.com/jimiechen/mineplanet/go/services/config/cache"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
)

// ConfigService 应用配置核心服务接口
// 对外暴露配置查询能力，协调 Repository / Cache / Parse / Compose 各子流程
type ConfigService interface {
	GetAppConfigs(req *AppConfigRequest) (*AppConfigResponse, error)
	GetVersionInfo(env string, knownVersions map[string]int64) (*VersionInfoResponse, error)
}

// AppConfigService ConfigService 的默认实现
// 组合依赖：ConfigRepository（持久化）+ Cache（缓存）+ SchemaRepository（元数据）
type AppConfigService struct {
	configRepo repository.ConfigRepository
	schemaRepo repository.SchemaRepository
	cache      cache.Cache
}

// AppConfigRequest 客户端请求参数，映射自 proto AppConfigsReq
type AppConfigRequest struct {
	Env            string
	ClientScope    string
	ClientVersion  string
	RequestedModules []string
}

// AppConfigService 配置响应，包含强类型字段 + 动态模块列表
// 对应 proto AppConfigsRsp 的业务视图（不含 Result 码，由传输层填充）
type AppConfigResponse struct {
	StaticModules   map[string]map[string]*domain.TypedValue
	DynamicModules []*DynamicModuleView
}

// DynamicModuleView 动态模块的业务视图
// 从 ConfigVersion + ModuleSchema 组装而来
type DynamicModuleView struct {
	ModuleKey   string
	Version     int64
	Fields      map[string]*domain.TypedValue
	Descriptors []*FieldDescriptorView
}

// FieldDescriptorView 字段描述符视图，对应 proto FieldDescriptor
type FieldDescriptorView struct {
	FieldKey   string
	FieldType  string
	IsRequired bool
	DefaultVal string
}

// VersionInfoResponse 版本轮询响应
// 对应 proto AppConfigVersionRsp 的业务视图
type VersionInfoResponse struct {
	ConfigVersions map[string]int64
	HasChanges     bool
}

// NewAppConfigService 创建配置服务实例
func NewAppConfigService(
	configRepo repository.ConfigRepository,
	schemaRepo repository.SchemaRepository,
	cache cache.Cache,
) *AppConfigService {
	return &AppConfigService{
		configRepo: configRepo,
		schemaRepo: schemaRepo,
		cache:      cache,
	}
}
