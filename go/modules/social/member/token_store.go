package member

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// TokenStore 令牌黑名单存储接口
// 使用 JWT 的 jti (tokenID) 作为 key，不存储完整 token 字符串
// 支持 Memory 和 Redis 两种实现
type TokenStore interface {
	// Store 将 token 的 jti 存入黑名单，ttl 为过期时间（秒）
	Store(ctx context.Context, jti string, ttl int64) error
	// Delete 从黑名单移除指定 jti
	Delete(ctx context.Context, jti string) error
	// Exists 检查指定 jti 是否在黑名单中（即 token 是否已失效）
	Exists(ctx context.Context, jti string) (bool, error)
}

// ========== 内存实现（用于单测和本地开发）==========

// memoryTokenEntry 黑名单条目
type memoryTokenEntry struct {
	expiry time.Time
}

// MemoryTokenStore 基于内存的令牌黑名单存储
// 使用 map + sync.RWMutex 实现线程安全，key 为 jti
// 适用场景：单测试、本地开发、无 Redis 环境
type MemoryTokenStore struct {
	mu        sync.RWMutex
	blacklist map[string]memoryTokenEntry
}

// NewMemoryTokenStore 创建内存令牌黑名单存储
func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{
		blacklist: make(map[string]memoryTokenEntry),
	}
}

// Store 将 jti 加入内存黑名单，ttl 单位为秒
func (s *MemoryTokenStore) Store(_ context.Context, jti string, ttlSec int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupLocked()

	s.blacklist[jti] = memoryTokenEntry{
		expiry: time.Now().Add(time.Duration(ttlSec) * time.Second),
	}
	return nil
}

// Exists 检查 jti 是否在内存黑名单中
func (s *MemoryTokenStore) Exists(_ context.Context, jti string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.blacklist[jti]
	if !exists {
		return false, nil
	}

	if time.Now().After(entry.expiry) {
		return false, nil
	}

	return true, nil
}

// Delete 从内存黑名单移除指定 jti
func (s *MemoryTokenStore) Delete(_ context.Context, jti string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.blacklist, jti)
	return nil
}

// cleanupLocked 清理过期条目（必须在写锁保护下调用）
func (s *MemoryTokenStore) cleanupLocked() {
	now := time.Now()
	for jti, entry := range s.blacklist {
		if now.After(entry.expiry) {
			delete(s.blacklist, jti)
		}
	}
}

// ========== 兼容性别名（逐步废弃，供过渡期使用）==========

// Blacklist 兼容旧接口：将完整 token 解析出 jti 后存入黑名单
// 过渡期方法，新代码应直接使用 Store(jti, ttl)
func (s *MemoryTokenStore) Blacklist(ctx context.Context, token string, ttl time.Duration) error {
	jti := extractJTIFromToken(token)
	if jti == "" {
		jti = token // fallback: 无法解析 jti 时使用原始 token
	}
	return s.Store(ctx, jti, int64(ttl.Seconds()))
}

// IsBlacklisted 兼容旧接口
func (s *MemoryTokenStore) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	jti := extractJTIFromToken(token)
	if jti == "" {
		jti = token
	}
	return s.Exists(ctx, jti)
}

// extractJTIFromToken 从 JWT 字符串中提取 jti claim（不做签名验证）
// 仅用于从 token 字符串快速获取 jti，不用于鉴权
func extractJTIFromToken(tokenStr string) string {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return ""
	}
	// 解码 payload (第二段)，JWT 使用 URL-safe base64 without padding
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// 尝试标准 base64（兼容非标准实现）
		payload, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return ""
		}
	}
	// 手动解析 JSON 提取 jti（避免引入 full jwt.Parse 的开销）
	// 简化处理：查找 "jti":"xxx" 模式
	str := string(payload)
	// 快速查找 jti 字段
	var jtiPrefix = `"jti":"`
	start := strings.Index(str, jtiPrefix)
	if start == -1 {
		return ""
	}
	start += len(jtiPrefix)
	end := strings.Index(str[start:], `"`)
	if end == -1 {
		return ""
	}
	return str[start : start+end]
}

// ========== 三域公共黑名单检查 ==========

// TokenRevokedCode 令牌已撤销/失效的统一错误码
// 对应 HTTP 401 Unauthorized，在 TarsGo Servant.Handle 返回 (200, errorBody, nil)
const TokenRevokedCode = int32(10401)

// CheckTokenBlacklist 鉴权路径黑名单检查（三域公共逻辑）
//
// 在写入 CtxKeyUserID 到 context 之前调用。
// 从 extend["token"] 提取 JWT 字符串 → ParseJTI 提取 jti → 查询 TokenStore。
// 若 jti 在黑名单中，返回错误响应并阻断请求进入业务 SVC。
//
// 参数：
//   - ctx: 请求上下文
//   - extend: Gateway 传入的扩展字段，需包含 "token" key
//   - jwtMgr: JWT 管理器，用于从 token 提取 jti（可为 nil，nil 时跳过检查）
//   - tokenStore: 黑名单存储（可为 nil，nil 时跳过检查）
//
// 返回值：
//   - code, respBytes, err: 若 token 被撤销则返回 (200, errorBody, nil)，调用方应直接 return
//   - 若未命中黑名单或无需检查，返回 (0, nil, nil)，调用方应继续执行后续逻辑
func CheckTokenBlacklist(
	ctx context.Context,
	extend map[string]string,
	jwtMgr *JWTManager,
	tokenStore TokenStore,
) (int, []byte, error) {
	// 任一依赖为 nil 时跳过检查（降级放行）
	if jwtMgr == nil || tokenStore == nil {
		return 0, nil, nil
	}

	tokenStr := extend["token"]
	if tokenStr == "" {
		return 0, nil, nil
	}

	jti := jwtMgr.ParseJTI(tokenStr)
	if jti == "" {
		return 0, nil, nil
	}

	blacklisted, err := tokenStore.Exists(ctx, jti)
	if err != nil {
		// TokenStore 查询异常时放行（不因存储故障阻断业务）
		// 日志由上层记录
		return 0, nil, nil
	}

	if blacklisted {
		errResp := map[string]interface{}{
			"result": map[string]interface{}{
				"code":    TokenRevokedCode,
				"message": "token invalid or revoked",
			},
		}
		respBytes, _ := json.Marshal(errResp)
		return 200, respBytes, nil
	}

	return 0, nil, nil
}
