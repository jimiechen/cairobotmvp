package sqlitex

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Mode SQLite 运行模式
type Mode string

const (
	ModeMemory = "memory" // 纯内存模式，适合测试
	ModeFile   = "file"   // 文件持久化模式，适合开发
)

// Config SQLite 连接配置
type Config struct {
	// Mode 运行模式：memory 或 file
	Mode Mode
	// DSN 文件模式下为数据库文件路径，内存模式可留空
	DSN string
	// WAL 是否启用 WAL 日志模式（提升并发性能）
	WAL bool
	// BusyTimeout 忙等待超时（毫秒）
	BusyTimeout int
}

// DefaultMemoryConfig 返回内存模式默认配置（用于测试）
func DefaultMemoryConfig() *Config {
	return &Config{
		Mode:         ModeMemory,
		DSN:          ":memory:",
		WAL:          true,
		BusyTimeout: 5000,
	}
}

// DefaultFileConfig 返回文件模式默认配置
func DefaultFileConfig(dbPath string) *Config {
	return &Config{
		Mode:         ModeFile,
		DSN:          dbPath,
		WAL:          true,
		BusyTimeout: 5000,
	}
}

// DB 数据库操作抽象接口（与 mysqlx.DB 接口一致）
// 业务代码可通过此接口在 MySQL/SQLite 间切换
type DB interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Ping(ctx context.Context) error
	Close() error
}

// sqliteDB DB 接口的 SQLite 实现
type sqliteDB struct {
	db *sql.DB
}

// Open 创建 SQLite 数据库连接
func Open(cfg *Config) (DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("sqlitex: config is nil")
	}

	dsn := cfg.DSN
	if cfg.Mode == ModeMemory && dsn == "" {
		dsn = ":memory:"
	}

	params := ""
	if cfg.WAL {
		params += "_journal_mode=WAL&"
	}
	if cfg.BusyTimeout > 0 {
		params += fmt.Sprintf("_busy_timeout=%d&", cfg.BusyTimeout)
	}

	if len(params) > 0 {
		params = "?" + params[:len(params)-1]
		dsn = dsn + params
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlitex: open failed: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlitex: ping failed: %w", err)
	}

	return &sqliteDB{db: db}, nil
}

// Exec 执行无返回行数的 SQL
func (s *sqliteDB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

// Query 执行查询并返回多行结果
func (s *sqliteDB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

// QueryRow 执行查询并返回单行结果
func (s *sqliteDB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}

// BeginTx 开启事务
func (s *sqliteDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, opts)
}

// Ping 检查连接健康状态
func (s *sqliteDB) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close 关闭连接
func (s *sqliteDB) Close() error {
	return s.db.Close()
}
