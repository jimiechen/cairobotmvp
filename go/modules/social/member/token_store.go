package member

import (
	"context"
	"sync"
	"time"
)

// TokenStore 定义令牌黑名单存储抽象
// 生产环境使用 Redis 实现（RedisTokenStore）；单测/本地开发使用内存实现（MemoryTokenStore）
// 接口隔离原则：认证模块不直接依赖 Redis 运行时
type TokenStore interface {
	// Blacklist 将令牌加入黑名单，ttl 为令牌剩余有效期
	Blacklist(ctx context.Context, token string, ttl time.Duration) error
	// IsBlacklisted 检查令牌是否在黑名单中
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}

// ========== 内存实现（用于单测和本地开发）==========

// memoryTokenEntry 黑名单条目
type memoryTokenEntry struct {
	expiry time.Time
}

// MemoryTokenStore 基于内存的令牌黑名单存储
// 使用 map + sync.RWMutex 实现线程安全
// 适用场景：单测试、本地开发、无 Redis 环境
type MemoryTokenStore struct {
	mu       sync.RWMutex
	blacklist map[string]memoryTokenEntry
}

// NewMemoryTokenStore 创建内存令牌黑名单存储
func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{
		blacklist: make(map[string]memoryTokenEntry),
	}
}

// Blacklist 将令牌加入内存黑名单
// 自动清理过期条目（惰性删除 + 写入时清理）
func (s *MemoryTokenStore) Blacklist(_ context.Context, token string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 写入时清理过期条目（避免内存泄漏）
	s.cleanupLocked()

	s.blacklist[token] = memoryTokenEntry{
		expiry: time.Now().Add(ttl),
	}
	return nil
}

// IsBlacklisted 检查令牌是否在内存黑名单中
// 过期条目视为不在黑名单中（自动清理）
func (s *MemoryTokenStore) IsBlacklisted(_ context.Context, token string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.blacklist[token]
	if !exists {
		return false, nil
	}

	// 已过期的条目视为不存在
	if time.Now().After(entry.expiry) {
		return false, nil
	}

	return true, nil
}

// cleanupLocked 清理过期条目（必须在写锁保护下调用）
func (s *MemoryTokenStore) cleanupLocked() {
	now := time.Now()
	for token, entry := range s.blacklist {
		if now.After(entry.expiry) {
			delete(s.blacklist, token)
		}
	}
}
