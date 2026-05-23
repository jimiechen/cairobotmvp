package domain

// PackVersion 语言包版本
// 简单包装语言代码和版本号，用于增量查询
//
// 职责：
// - 封装版本号信息
// - 提供版本比较方法
//
// 不负责：
// - 版本号生成（由业务层负责）
type PackVersion struct {
	LangCode string
	Version  int
}

// IsNewerThan 判断当前版本是否比目标版本更新
func (v *PackVersion) IsNewerThan(other int) bool {
	return v.Version > other
}
