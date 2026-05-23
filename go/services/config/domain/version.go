package domain

import "time"

// ConfigVersion 配置版本实体，对应 sys_config_version 表的一行记录
// 负责承载某个 module_key 在特定 env 下的版本化配置快照
// 不负责持久化和缓存，由 Repository 和 Cache 层处理
type ConfigVersion struct {
	ID          int64
	ModuleKey   string
	Env         string
	Version     int64
	ConfigJSON  string
	IsPublished bool
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreateBy    string
	UpdateBy    string
}

// IsReleased 判断该版本是否已发布且可对外提供服务
// 未发布的版本仅用于编辑态预览，不应下发给客户端
func (v *ConfigVersion) IsReleased() bool {
	return v.IsPublished && v.PublishedAt != nil
}
