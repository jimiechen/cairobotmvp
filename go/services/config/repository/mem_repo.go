package repository

import (
	"fmt"
	"sync"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// MemConfigRepo 基于内存 Map 的 ConfigRepository 实现
// 零外部依赖，用于单测和开发阶段
// 不持久化，进程重启后数据丢失；生产环境应替换为 MySQL 实现
type MemConfigRepo struct {
	mu       sync.RWMutex
	versions []*domain.ConfigVersion
}

// NewMemConfigRepo 创建空内存配置仓库实例
func NewMemConfigRepo() *MemConfigRepo {
	return &MemConfigRepo{versions: make([]*domain.ConfigVersion, 0)}
}

// GetLatestVersion 查询指定模块在指定环境下最新已发布的版本
func (r *MemConfigRepo) GetLatestVersion(moduleKey, env string) (*domain.ConfigVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *domain.ConfigVersion
	for _, v := range r.versions {
		if v.ModuleKey == moduleKey && v.Env == env && v.IsPublished {
			if latest == nil || v.Version > latest.Version {
				latest = v
			}
		}
	}
	return latest, nil
}

// GetByModuleAndVersion 精确查询某模块在某环境下的特定版本
func (r *MemConfigRepo) GetByModuleAndVersion(moduleKey, env string, version int64) (*domain.ConfigVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, v := range r.versions {
		if v.ModuleKey == moduleKey && v.Env == env && v.Version == version {
			return v, nil
		}
	}
	return nil, nil
}

// ListPublishedVersions 列出指定环境下所有已发布的配置版本
func (r *MemConfigRepo) ListPublishedVersions(env string) ([]*domain.ConfigVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.ConfigVersion
	for _, v := range r.versions {
		if v.Env == env && v.IsPublished {
			result = append(result, v)
		}
	}
	return result, nil
}

// Save 新增配置版本记录到内存
func (r *MemConfigRepo) Save(version *domain.ConfigVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	version.ID = int64(len(r.versions)) + 1
	r.versions = append(r.versions, version)
	return nil
}

// Clear 清空所有数据（仅测试用）
func (r *MemConfigRepo) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.versions = make([]*domain.ConfigVersion, 0)
}

// MemSchemaRepo 基于内存 Map 的 SchemaRepository 实现
type MemSchemaRepo struct {
	mu      sync.RWMutex
	schemas []*domain.FieldSchema
}

// NewMemSchemaRepo 创建空内存 Schema 仓库实例
func NewMemSchemaRepo() *MemSchemaRepo {
	return &MemSchemaRepo{schemas: make([]*domain.FieldSchema, 0)}
}

// ListByModule 查询指定模块下所有字段 Schema
func (r *MemSchemaRepo) ListByModule(moduleKey string) ([]*domain.FieldSchema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.FieldSchema
	for _, s := range r.schemas {
		if s.ModuleKey == moduleKey {
			result = append(result, s)
		}
	}
	return result, nil
}

// Create 新增字段 Schema 记录
func (r *MemSchemaRepo) Create(schema *domain.FieldSchema) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, s := range r.schemas {
		if s.ModuleKey == schema.ModuleKey && s.FieldKey == schema.FieldKey {
			return fmt.Errorf("唯一约束冲突: (%s, %s)", schema.ModuleKey, schema.FieldKey)
		}
	}
	schema.ID = int64(len(r.schemas)) + 1
	r.schemas = append(r.schemas, schema)
	return nil
}

// Update 更新字段 Schema（按 ID）
func (r *MemSchemaRepo) Update(schema *domain.FieldSchema) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, s := range r.schemas {
		if s.ID == schema.ID {
			r.schemas[i] = schema
			return nil
		}
	}
	return fmt.Errorf("未找到 ID=%d 的记录", schema.ID)
}

// DeleteSoft 软删除（标记禁用）
func (r *MemSchemaRepo) DeleteSoft(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, s := range r.schemas {
		if s.ID == id {
			s.IsEnabled = false
			return nil
		}
	}
	return fmt.Errorf("未找到 ID=%d 的记录", id)
}

// SeedWithDefaults 写入 DDL 中定义的种子 Schema 数据
// 用于初始化测试环境的默认字段定义
func (r *MemSchemaRepo) SeedWithDefaults() {
	defaultSchemas := []struct {
		mod, key, fType, defVal, desc string
		sortOrder                      int
	}{
		{"base_cfg", "domain_root", "string", "", "API 根域名", 1},
		{"base_cfg", "domain_wap", "string", "", "WAP 页面域名", 2},
		{"base_cfg", "sign_rand", "string", "", "签名随机盐值", 3},
		{"base_cfg", "construct_email", "string", "", "反馈联系邮箱", 4},
		{"wap_cfg", "user_agreement_url", "string", "", "用户协议 URL", 1},
		{"wap_cfg", "privacy_policy_url", "string", "", "隐私政策 URL", 2},
		{"regex_cfg", "regex_email", "string", "", "邮箱正则表达式", 1},
		{"regex_cfg", "regex_password", "string", "", "密码正则表达式", 2},
		{"regex_cfg", "regex_phone", "string", "", "手机号正则表达式", 3},
		{"regex_cfg", "regex_nick", "string", "", "昵称正则表达式", 4},
		{"regex_cfg", "regex_circle_name", "string", "", "圈子名称正则表达式", 5},
		{"oss_cfg", "oss_host", "string", "", "OSS 主机地址", 1},
		{"oss_cfg", "oss_domain", "string", "", "OSS 域名", 2},
		{"oss_cfg", "cdn_domain", "string", "", "CDN 域名", 3},
		{"lang_cfg", "lang_code", "string", "zh-CN", "默认语言代码", 1},
		{"mute_cfg", "durations", "json", "[]", "静音时长选项列表", 1},
		{"group_cfg", "group_config_pay_notice", "string", "", "群组支付公告文案", 1},
	}
	for _, sd := range defaultSchemas {
		fs := &domain.FieldSchema{
			ModuleKey: sd.mod, FieldKey: sd.key,
			FieldType: domain.FieldType(sd.fType),
			DefaultValue: sd.defVal,
			Description: sd.desc,
			SortOrder:   sd.sortOrder,
			IsEnabled:   true,
		}
		r.Create(fs)
	}
}

// Close 空实现，满足接口约定（内存实现无需关闭资源）
func (r *MemConfigRepo) Close() error { return nil }

// FindSchema 按主键查找单条 Schema（辅助方法）
func (r *MemSchemaRepo) FindSchema(id int64) *domain.FieldSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.schemas {
		if s.ID == id {
			return s
		}
	}
	return nil
}

// Count 返回当前记录数（测试用）
func (r *MemSchemaRepo) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.schemas)
}
