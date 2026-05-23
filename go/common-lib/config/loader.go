package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load 从指定路径加载服务端配置
// 支持从 YAML 文件加载基础配置，并通过环境变量覆盖
// 环境变量前缀为 CAIROBOT，例如 CAIROBOT_MYSQL_HOST
func Load(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 替换环境变量占位符，例如 ${MYSQL_PASSWORD}
	content := replaceEnvVars(string(data))

	var cfg ServerConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 从环境变量覆盖配置
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

// LoadWithDefaults 加载配置并使用默认值填充缺失项
func LoadWithDefaults(path string) (*ServerConfig, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}

	applyDefaults(cfg)
	return cfg, nil
}

// applyDefaults 为配置项设置默认值
func applyDefaults(cfg *ServerConfig) {
	if cfg.MySQL.Port == 0 {
		cfg.MySQL.Port = 3306
	}
	if cfg.MySQL.Charset == "" {
		cfg.MySQL.Charset = "utf8mb4"
	}
	if cfg.MySQL.MaxOpenConns == 0 {
		cfg.MySQL.MaxOpenConns = 100
	}
	if cfg.MySQL.MaxIdleConns == 0 {
		cfg.MySQL.MaxIdleConns = 20
	}
	if cfg.MySQL.ConnMaxLifetime == "" {
		cfg.MySQL.ConnMaxLifetime = "1h"
	}
	if cfg.MySQL.ConnMaxIdleTime == "" {
		cfg.MySQL.ConnMaxIdleTime = "10m"
	}

	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}

	if cfg.Cache.ConfigTTLSeconds == 0 {
		cfg.Cache.ConfigTTLSeconds = 30
	}
	if cfg.Cache.I18nTTLSeconds == 0 {
		cfg.Cache.I18nTTLSeconds = 60
	}
}

// replaceEnvVars 替换字符串中的环境变量占位符
// 支持 ${VAR_NAME} 格式
func replaceEnvVars(content string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		varName := match[2 : len(match)-1]
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match
	})
}

// applyEnvOverrides 从环境变量覆盖配置值
// 环境变量格式: CAIROBOT_MYSQL_HOST, CAIROBOT_REDIS_PORT 等
func applyEnvOverrides(cfg *ServerConfig) {
	prefix := "CAIROBOT_"

	if v := os.Getenv(prefix + "MYSQL_HOST"); v != "" {
		cfg.MySQL.Host = v
	}
	if v := os.Getenv(prefix + "MYSQL_PORT"); v != "" {
		var port int
		fmt.Sscanf(v, "%d", &port)
		cfg.MySQL.Port = port
	}
	if v := os.Getenv(prefix + "MYSQL_USERNAME"); v != "" {
		cfg.MySQL.Username = v
	}
	if v := os.Getenv(prefix + "MYSQL_PASSWORD"); v != "" {
		cfg.MySQL.Password = v
	}
	if v := os.Getenv(prefix + "MYSQL_DATABASE"); v != "" {
		cfg.MySQL.Database = v
	}

	if v := os.Getenv(prefix + "REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := os.Getenv(prefix + "REDIS_PORT"); v != "" {
		var port int
		fmt.Sscanf(v, "%d", &port)
		cfg.Redis.Port = port
	}
	if v := os.Getenv(prefix + "REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv(prefix + "REDIS_DB"); v != "" {
		var db int
		fmt.Sscanf(v, "%d", &db)
		cfg.Redis.DB = db
	}
	if v := os.Getenv(prefix + "REDIS_ENABLED"); v != "" {
		cfg.Redis.Enabled = strings.ToLower(v) == "true" || v == "1"
	}
}
