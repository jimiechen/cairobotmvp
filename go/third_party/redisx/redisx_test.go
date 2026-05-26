package redisx

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, Client) {
	t.Helper()
	
	m, err := miniredis.Run()
	require.NoError(t, err)
	
	t.Cleanup(m.Close)
	
	cfg := &Config{
		Addr: m.Addr(),
		DB:   0,
	}
	
	client, err := NewClient(cfg)
	require.NoError(t, err)
	
	t.Cleanup(func() { _ = client.Close() })
	
	return m, client
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "127.0.0.1:6379", cfg.Addr)
	assert.Equal(t, 0, cfg.DB)
	assert.Equal(t, 10, cfg.PoolSize)
}

func TestNewClient_NilConfig(t *testing.T) {
	_, err := NewClient(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config is nil")
}

func TestGetSetDelete(t *testing.T) {
	m, client := setupTestRedis(t)

	val, err := client.Set(m.Ctx, "test:key", "hello", time.Minute)
	require.NoError(t, err)
	assert.NotNil(t, val)

	result, err := client.Get(m.Ctx, "test:key")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)

	err = client.Delete(m.Ctx, "test:key")
	require.NoError(t, err)

	result, err = client.Get(m.Ctx, "test:key")
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestScan(t *testing.T) {
	m, client := setupTestRedis(t)

	m.Set("cache:user:1", `{"name":"test"}`)
	m.Set("cache:user:2", `{"name":"test2"}`)
	m.Set("other:key", "value")

	keys, err := client.Scan(m.Ctx, "cache:user:*")
	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestPing(t *testing.T) {
	_, client := setupTestRedis(t)
	
	err := client.Ping(m.Ctx())
	require.NoError(t, err)
}
