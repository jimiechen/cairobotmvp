package tarsclient

import (
	"fmt"

	"github.com/jimiechen/mineplanet/go/common-lib/config"
	configcache "github.com/jimiechen/mineplanet/go/services/config/cache"
	configrepo "github.com/jimiechen/mineplanet/go/services/config/repository"
	configsvc "github.com/jimiechen/mineplanet/go/services/config/service"
	i18ncache "github.com/jimiechen/mineplanet/go/services/i18n/cache"
	i18nrepo "github.com/jimiechen/mineplanet/go/services/i18n/repository"
	i18nsvc "github.com/jimiechen/mineplanet/go/services/i18n/service"
)

// RealServices 真实服务组装结果
// 包含从 MySQL 仓库构建的 ConfigService 和 I18nService
type RealServices struct {
	ConfigSvc configsvc.ConfigService
	I18nSvc   i18nsvc.I18nService
}

// BuildRealServices 基于 MySQL 连接配置构建真实 Config/I18n 服务
// 用于 S1 阶段替换 noop stub，使 Gateway 能返回真实数据
//
// 组装链路：
//   MySQL (go_biz) → MySQLConfigRepo + MySQLSchemaRepo + MySQLRepo → AppConfigService + I18nServiceImpl
//   Cache 层使用 MockCache（S1 阶段不依赖 Redis）
//
// 前置条件：MySQL 服务可达，sys_config_* / sys_lang_* 表已初始化（002 + 003 迁移）
func BuildRealServices(cfg *config.MySQLConfig, env string) (*RealServices, error) {
	// ---- 1. 创建 Repository 层 ----
	configRepo, err := configrepo.NewMySQLConfigRepo(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 ConfigRepo 失败: %w", err)
	}

	schemaRepo, err := configrepo.NewMySQLSchemaRepo(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 SchemaRepo 失败: %w", err)
	}

	i18nRepo, err := i18nrepo.NewMySQLRepo(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 I18nRepo 失败: %w", err)
	}

	// ---- 2. 创建 Cache 层（MockCache，S1 不依赖 Redis）----
	configCache := configcache.NewMockCache()
	i18nCache := i18ncache.NewMockCache()

	// ---- 3. 组装 Service 层 ----
	configSvc := configsvc.NewAppConfigService(configRepo, schemaRepo, configCache)
	i18nSvc := i18nsvc.NewI18nService(i18nRepo, i18nCache, env)

	return &RealServices{
		ConfigSvc: configSvc,
		I18nSvc:   i18nSvc,
	}, nil
}

// RegisterRealHandlers 使用真实服务注册所有 handler（System + Config + I18n）
// 与 RegisterAllLocalHandlers 的区别：使用 MySQL 真实服务替代 noop stub
func RegisterRealHandlers(invoker *LocalInvoker, svc *RealServices) {
	RegisterModuleHandlers(invoker)
	RegisterConfigI18nHandlers(invoker, svc.ConfigSvc, svc.I18nSvc)
}
