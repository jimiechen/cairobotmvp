package redisx

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
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

	err := client.Set(m.Ctx, "test:key", "hello", time.Minute)
	require.NoError(t, err)

	result, err := client.Get(m.Ctx, "test:key")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)

	err = client.Delete(m.Ctx, "test:key")
	require.NoError(t, err)

	_, err = client.Get(m.Ctx, "test:key")
	assert.Error(t, err)
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

	err := client.Ping(context.Background())
	require.NoError(t, err)
}

func TestInvalidate_EmptyMatch(t *testing.T) {
	m, client := setupTestRedis(t)

	m.Set("other:key", "value")

	err := client.Invalidate(m.Ctx, "nonexistent:*")
	require.NoError(t, err)

	keys, _ := client.Scan(m.Ctx, "*")
	assert.Equal(t, 1, len(keys))
}

func TestInvalidate_SingleKey(t *testing.T) {
	m, client := setupTestRedis(t)

	m.Set("config:schema:default:field1", `{"type":"string"}`)

	err := client.Invalidate(m.Ctx, "config:schema:default:*")
	require.NoError(t, err)

	keys, _ := client.Scan(m.Ctx, "*")
	assert.Empty(t, keys)
}

func TestInvalidate_MultipleKeysSingleBatch(t *testing.T) {
	m, client := setupTestRedis(t)

	for i := 0; i < 10; i++ {
		m.Set(fmt.Sprintf("cache:value:prod:module_%d", i), fmt.Sprintf(`"val%d"`, i))
	}

	err := client.Invalidate(m.Ctx, "cache:value:prod:*")
	require.NoError(t, err)

	keys, _ := client.Scan(m.Ctx, "cache:value:prod:*")
	assert.Empty(t, keys)
}

func TestInvalidate_MultipleKeysMultiBatch(t *testing.T) {
	m, client := setupTestRedis(t)

	totalKeys := 1200
	for i := 0; i < totalKeys; i++ {
		m.Set(fmt.Sprintf("bulk:key:%04d", i), fmt.Sprintf("v%d", i))
	}

	err := client.Invalidate(m.Ctx, "bulk:key:*")
	require.NoError(t, err)

	keys, _ := client.Scan(m.Ctx, "bulk:key:*")
	assert.Empty(t, keys)
}

func TestInvalidate_ScanError(t *testing.T) {
	m, client := setupTestRedis(t)

	m.Close()

	err := client.Invalidate(m.Ctx, "test:*")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalidate scan failed")
}

func TestInvalidate_CtxCancel(t *testing.T) {
	m, client := setupTestRedis(t)

	for i := 0; i < 100; i++ {
		m.Set(fmt.Sprintf("slow:key:%d", i), "val")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.Invalidate(ctx, "slow:key:*")
	assert.Error(t, err)
}
