package config

// ServerConfig 统一服务端配置
// 从 configs/server.yaml 加载，支持环境变量覆盖
type ServerConfig struct {
	MySQL MySQLConfig `yaml:"mysql"`
	Redis RedisConfig `yaml:"redis"`
	Cache CacheConfig `yaml:"cache"`
}

// MySQLConfig MySQL 数据库配置
type MySQLConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Charset         string `yaml:"charset"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime string `yaml:"conn_max_idle_time"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	ConfigTTLSeconds int  `yaml:"config_ttl_seconds"`
	I18nTTLSeconds   int  `yaml:"i18n_ttl_seconds"`
	PubSubEnabled    bool `yaml:"pubsub_enabled"`
}
