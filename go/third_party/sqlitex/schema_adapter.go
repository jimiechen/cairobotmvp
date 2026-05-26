package sqlitex

import (
	"strings"
)

// SchemaAdapter MySQL DDL → SQLite DDL 适配器
// 解决两种数据库语法差异，使同一套 schema 可同时在 MySQL 和 SQLite 上运行
type SchemaAdapter struct{}

// NewSchemaAdapter 创建适配器实例
func NewSchemaAdapter() *SchemaAdapter {
	return &SchemaAdapter{}
}

// ConvertDDL 将 MySQL DDL 转换为兼容 SQLite 的版本
// 处理以下差异：
// - BIGINT AUTO_INCREMENT → INTEGER PRIMARY KEY AUTOINCREMENT
// - TINYINT(1) → INTEGER（SQLite 无布尔类型）
// - DATETIME → TEXT（SQLite 无原生日期类型）
// - ENGINE=InnoDB → 删除（SQLite 不支持）
// - DEFAULT CURRENT_TIMESTAMP → DEFAULT (datetime('now'))
// - COMMENT 'xxx' → -- xxx
func (a *SchemaAdapter) ConvertDDL(mysqlDDL string) string {
	result := mysqlDDL

	result = a.convertAutoIncrement(result)
	result = a.convertBigInt(result)
	result = a.convertTinyInt(result)
	result = a.convertDatetime(result)
	result = a.removeEngineClause(result)
	result = a.removeComments(result)
	result = a.convertDefaultTimestamp(result)
	result = a.convertJSON(result)

	return result
}

func (a *SchemaAdapter) convertAutoIncrement(sql string) string {
	return strings.ReplaceAll(sql, "BIGINT AUTO_INCREMENT", "INTEGER PRIMARY KEY AUTOINCREMENT")
}

func (a *SchemaAdapter) convertBigInt(sql string) string {
	return strings.ReplaceAll(sql, "BIGINT", "INTEGER")
}

func (a *SchemaAdapter) convertTinyInt(sql string) string {
	return strings.ReplaceAll(sql, "TINYINT(1)", "INTEGER")
}

func (a *SchemaAdapter) convertDatetime(sql string) string {
	sql = strings.ReplaceAll(sql, "DATETIME", "TEXT")
	return strings.ReplaceAll(sql, "TIMESTAMP", "TEXT")
}

func (a *SchemaAdapter) removeEngineClause(sql string) string {
	for _, engine := range []string{"ENGINE=InnoDB", "ENGINE=MyISAM"} {
		sql = strings.ReplaceAll(sql, engine, "")
	}
	return sql
}

func (a *SchemaAdapter) removeComments(sql string) strings.Builder {
	var result strings.Builder
	inComment := false
	
	for i := 0; i < len(sql); i++ {
		if i+1 < len(sql) && sql[i] == '-' && sql[i+1] == '-' {
			inComment = true
			i++
			continue
		}
		
		if inComment && sql[i] == '\n' {
			inComment = false
			continue
		}
		
		if !inComment {
			result.WriteByte(sql[i])
		}
	}
	
	return result
}

func (a *SchemaAdapter) convertDefaultTimestamp(sql string) string {
	return strings.ReplaceAll(
		sql,
		"DEFAULT CURRENT_TIMESTAMP",
		"DEFAULT (datetime('now'))",
	)
}

func (a *SchemaAdapter) convertJSON(sql string) string {
	return strings.ReplaceAll(sql, "JSON", "TEXT")
}
