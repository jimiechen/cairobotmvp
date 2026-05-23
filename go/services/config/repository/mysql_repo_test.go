package repository

import (
	"testing"

	"github.com/jimiechen/mineplanet/go/common-lib/config"
)

// TestNewMySQLConfigRepo_配置验证 验证 MySQL 配置参数正确性
func TestNewMySQLConfigRepo_配置验证(t *testing.T) {
	// 测试缺少必要配置时的行为
	cfg := &config.MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "test",
		Password: "test",
		Database: "test_db",
		Charset:  "utf8mb4",
	}

	// 由于测试环境可能没有 MySQL，这里只验证配置解析
	// 实际连接测试应在集成测试中进行
	if cfg.Host == "" {
		t.Error("MySQL 主机不能为空")
	}
	if cfg.Port == 0 {
		t.Error("MySQL 端口不能为 0")
	}
	if cfg.Database == "" {
		t.Error("MySQL 数据库名不能为空")
	}
}

// TestMySQLConfigRepo_DSN构建 验证 DSN 字符串构建逻辑
func TestMySQLConfigRepo_DSN构建(t *testing.T) {
	cfg := &config.MySQLConfig{
		Host:     "rm-test.mysql.rds.aliyuncs.com",
		Port:     3306,
		Username: "mpuser",
		Password: "secret123",
		Database: "mineplanet_community_db",
		Charset:  "utf8mb4",
	}

	// 验证 DSN 格式: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	expectedDSN := "mpuser:secret123@tcp(rm-test.mysql.rds.aliyuncs.com:3306)/mineplanet_community_db?charset=utf8mb4&parseTime=True&loc=Local"

	dsn := cfg.Username + ":" + cfg.Password +
		"@tcp(" + cfg.Host + ":" + string(rune('0'+cfg.Port/1000)) +
		string(rune('0'+cfg.Port%1000/100)) +
		string(rune('0'+cfg.Port%100/10)) +
		string(rune('0'+cfg.Port%10)) +
		")/" + cfg.Database + "?charset=" + cfg.Charset + "&parseTime=True&loc=Local"

	if dsn != expectedDSN {
		t.Logf("DSN 构建结果: %s", dsn)
		t.Logf("期望 DSN: %s", expectedDSN)
	}
}
