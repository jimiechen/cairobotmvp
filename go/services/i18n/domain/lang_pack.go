package domain

import "time"

// LangPack 语言包实体
// 对应 sys_lang_pack 表，表示一个语言包的元信息
//
// 职责：
// - 封装语言包的基本属性
// - 提供发布状态判断方法
//
// 不负责：
// - 数据库操作（由 Repository 负责）
// - 字符串内容管理（由 LangString 负责）
type LangPack struct {
	ID          int64
	PackName    string
	Env         string
	Version     int
	LangCode    string
	Description string
	IsPublished bool
	PublishedAt *time.Time
	PublishedBy int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsPublished 判断语言包是否已发布
func (p *LangPack) IsPublishedStatus() bool {
	return p.IsPublished
}
