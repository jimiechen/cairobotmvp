package mysqlx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, 3306, cfg.Port)
	assert.Equal(t, "root", cfg.User)
	assert.Equal(t, "cairobot", cfg.Database)
}

func TestConfig_DSN(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     3307,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
	}
	dsn := cfg.DSN()
	assert.Contains(t, dsn, "testuser:testpass@tcp(localhost:3307)/testdb")
	assert.Contains(t, dsn, "charset=utf8mb4")
	assert.Contains(t, dsn, "parseTime=True")
}

func TestNewDB_NilConfig(t *testing.T) {
	_, err := NewDB(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config is nil")
}

func TestNewDB_InvalidDSN(t *testing.T) {
	cfg := &Config{
		Host:     "invalid-host-that-does-not-exist",
		Port:     3306,
		User:     "root",
		Database: "nonexistent",
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	_, err := NewDB(cfg)
	assert.Error(t, err)
}

func TestConfig_DefaultValues(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 10, cfg.MaxOpenConns)
	assert.Equal(t, 5, cfg.MaxIdleConns)
	assert.Equal(t, time.Hour, cfg.ConnMaxLifetime)
	require.NotNil(t, cfg)
}
