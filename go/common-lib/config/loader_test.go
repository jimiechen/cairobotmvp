package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_成功加载配置(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_server.yaml")
	configContent := `
mysql:
  host: "127.0.0.1"
  port: 3306
  username: "testuser"
  password: "testpass"
  database: "testdb"
  charset: "utf8mb4"
redis:
  enabled: true
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
cache:
  config_ttl_seconds: 30
  i18n_ttl_seconds: 60
  pubsub_enabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.MySQL.Host != "127.0.0.1" {
		t.Errorf("MySQL Host 期望 127.0.0.1，实际 %s", cfg.MySQL.Host)
	}
	if cfg.MySQL.Port != 3306 {
		t.Errorf("MySQL Port 期望 3306，实际 %d", cfg.MySQL.Port)
	}
	if cfg.Redis.Host != "127.0.0.1" {
		t.Errorf("Redis Host 期望 127.0.0.1，实际 %s", cfg.Redis.Host)
	}
	if cfg.Cache.ConfigTTLSeconds != 30 {
		t.Errorf("ConfigTTLSeconds 期望 30，实际 %d", cfg.Cache.ConfigTTLSeconds)
	}
}

func TestLoad_环境变量覆盖(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_server.yaml")
	configContent := `
mysql:
  host: "127.0.0.1"
  port: 3306
  username: "testuser"
  password: "testpass"
  database: "testdb"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}

	// 设置环境变量
	os.Setenv("CAIROBOT_MYSQL_HOST", "192.168.1.100")
	os.Setenv("CAIROBOT_MYSQL_PORT", "3307")
	defer func() {
		os.Unsetenv("CAIROBOT_MYSQL_HOST")
		os.Unsetenv("CAIROBOT_MYSQL_PORT")
	}()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.MySQL.Host != "192.168.1.100" {
		t.Errorf("MySQL Host 期望被环境变量覆盖为 192.168.1.100，实际 %s", cfg.MySQL.Host)
	}
	if cfg.MySQL.Port != 3307 {
		t.Errorf("MySQL Port 期望被环境变量覆盖为 3307，实际 %d", cfg.MySQL.Port)
	}
}

func TestLoad_环境变量占位符替换(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_server.yaml")
	configContent := `
mysql:
  host: "127.0.0.1"
  port: 3306
  username: "testuser"
  password: "${TEST_MYSQL_PASSWORD}"
  database: "testdb"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}

	// 设置环境变量
	os.Setenv("TEST_MYSQL_PASSWORD", "secret123")
	defer func() {
		os.Unsetenv("TEST_MYSQL_PASSWORD")
	}()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.MySQL.Password != "secret123" {
		t.Errorf("MySQL Password 期望被替换为 secret123，实际 %s", cfg.MySQL.Password)
	}
}

func TestLoadWithDefaults_默认值填充(t *testing.T) {
	// 创建临时配置文件（部分字段缺失）
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_server.yaml")
	configContent := `
mysql:
  host: "127.0.0.1"
  username: "testuser"
  password: "testpass"
  database: "testdb"
redis:
  enabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}

	cfg, err := LoadWithDefaults(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.MySQL.Port != 3306 {
		t.Errorf("MySQL Port 期望默认值为 3306，实际 %d", cfg.MySQL.Port)
	}
	if cfg.MySQL.Charset != "utf8mb4" {
		t.Errorf("MySQL Charset 期望默认值为 utf8mb4，实际 %s", cfg.MySQL.Charset)
	}
	if cfg.MySQL.MaxOpenConns != 100 {
		t.Errorf("MySQL MaxOpenConns 期望默认值为 100，实际 %d", cfg.MySQL.MaxOpenConns)
	}
	if cfg.Redis.Port != 6379 {
		t.Errorf("Redis Port 期望默认值为 6379，实际 %d", cfg.Redis.Port)
	}
	if cfg.Cache.ConfigTTLSeconds != 30 {
		t.Errorf("Cache ConfigTTLSeconds 期望默认值为 30，实际 %d", cfg.Cache.ConfigTTLSeconds)
	}
}

func TestLoad_配置文件不存在(t *testing.T) {
	_, err := Load("/nonexistent/path/server.yaml")
	if err == nil {
		t.Error("期望返回错误，实际返回 nil")
	}
}
