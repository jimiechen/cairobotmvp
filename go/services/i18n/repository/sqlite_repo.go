package repository

import (
	"database/sql"
)

// SQLiteRepo I18nRepository 的 SQLite 实现
// 用于开发和测试环境，生产环境应使用 MySQL
type SQLiteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) *SQLiteRepo {
	return &SQLiteRepo{db: db}
}

func (r *SQLiteRepo) DB() *sql.DB {
	return r.db
}
