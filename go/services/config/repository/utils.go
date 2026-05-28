package repository

import (
	"database/sql"
	"strings"
	"time"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// boolToInt 将 Go bool 转换为 SQLite INTEGER (0/1)
// SQLite 不支持原生 bool 类型，存储时需做此转换
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// timePtrToStr 将 *time.Time 转换为 SQLite 兼容的字符串格式
// nil 指针返回 NULL（Go 的 nil 对应 SQL NULL）
func timePtrToStr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02 15:04:05")
	return &s
}

// parseTime 将 SQLite 时间字符串解析为 time.Time
// 兼容多种 SQLite datetime 输出格式
func parseTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Parse("2006-01-02 15:04:05", s)
}

// intToBool 将 SQLite INTEGER (0/1) 转换为 Go bool
func intToBool(i int) bool {
	return i != 0
}

// strPtrToString 安全解引用 *string，nil 时返回空串
func strPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// joinNonEmpty 过滤空串后用 sep 连接，用于构建错误信息
func joinNonEmpty(parts []string, sep string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, sep)
}

// scanFieldSchema 从 sql.Rows 扫描一条 FieldSchema 记录
func scanFieldSchema(rows *sql.Rows) (*domain.FieldSchema, error) {
	fs := &domain.FieldSchema{}
	var fieldTypeStr string
	err := rows.Scan(
		&fs.ID, &fs.ModuleKey, &fs.FieldKey, &fieldTypeStr,
		&fs.DefaultValue, &fs.Validator,
		&fs.IsRequired, &fs.IsSecret, &fs.Description,
		&fs.ClientScope, &fs.MinAppVer,
		&fs.SortOrder, &fs.IsEnabled,
	)
	if err != nil {
		return nil, err
	}
	fs.FieldType = domain.FieldType(fieldTypeStr)
	return fs, nil
}

// scanFieldSchemaRow 从 sql.Row（QueryRow 结果）扫描单条 FieldSchema 记录
func scanFieldSchemaRow(row *sql.Row) (*domain.FieldSchema, error) {
	fs := &domain.FieldSchema{}
	var fieldTypeStr string
	err := row.Scan(
		&fs.ID, &fs.ModuleKey, &fs.FieldKey, &fieldTypeStr,
		&fs.DefaultValue, &fs.Validator,
		&fs.IsRequired, &fs.IsSecret, &fs.Description,
		&fs.ClientScope, &fs.MinAppVer,
		&fs.SortOrder, &fs.IsEnabled,
	)
	if err != nil {
		return nil, err
	}
	fs.FieldType = domain.FieldType(fieldTypeStr)
	return fs, nil
}
