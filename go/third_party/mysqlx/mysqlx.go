package mysqlx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config MySQL 连接配置
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig 返回开发环境默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:            "127.0.0.1",
		Port:            3306,
		User:            "root",
		Password:        "",
		Database:        "cairobot",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 10,
	}
}

// DSN 从配置生成 MySQL DSN 字符串
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

// DB 数据库操作抽象接口
// 统一 MySQL/SQLite 的访问方式，便于测试时替换为 mock 或 SQLite
type DB interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Ping(ctx context.Context) error
	Close() error
}

// mysqlDB DB 接口的 MySQL 实现
type mysqlDB struct {
	db *sql.DB
}

// NewDB 创建 MySQL 数据库连接
// 前置条件：cfg 不能为 nil，且包含有效的连接信息
func NewDB(cfg *Config) (DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("mysqlx: config is nil")
	}

	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("mysqlx: open failed: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("mysqlx: ping failed: %w", err)
	}

	return &mysqlDB{db: db}, nil
}

// Exec 执行无返回行数的 SQL（INSERT/UPDATE/DELETE）
func (m *mysqlDB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return m.db.ExecContext(ctx, query, args...)
}

// Query 执行查询并返回多行结果
func (m *mysqlDB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return m.db.QueryContext(ctx, query, args...)
}

// QueryRow 执行查询并返回单行结果
func (m *mysqlDB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return m.db.QueryRowContext(ctx, query, args...)
}

// BeginTx 开启事务
func (m *mysqlDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.db.BeginTx(ctx, opts)
}

// Ping 检查数据库连接健康状态
func (m *mysqlDB) Ping(ctx context.Context) error {
	return m.db.PingContext(ctx)
}

// Close 关闭数据库连接
func (m *mysqlDB) Close() error {
	return m.db.Close()
}
