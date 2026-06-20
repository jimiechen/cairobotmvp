package member

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ========== JWT 配置常量 ==========
// 密钥从环境变量或配置注入，禁止硬编码

const (
	// DefaultAccessTTL 访问令牌默认有效期（24 小时）
	DefaultAccessTTL = 24 * time.Hour
	// DefaultRefreshTTL 刷新令牌默认有效期（7 天）
	DefaultRefreshTTL = 7 * 24 * time.Hour
	// DefaultJWTIssuer JWT 签发者标识
	DefaultJWTIssuer = "cairobot-social"
)

// ========== JWT 错误定义 ==========

var (
	// ErrTokenInvalid 令牌无效（格式错误、签名失败等）
	ErrTokenInvalid = errors.New("token invalid")
	// ErrTokenExpired 令牌已过期
	ErrTokenExpired = errors.New("token expired")
)

// JWTConfig JWT 签名和验证配置
// 从外部注入，禁止在代码中硬编码密钥
type JWTConfig struct {
	// SecretKey HS256 签名密钥，从环境变量或配置中心读取
	SecretKey string
	// Issuer 签发者标识
	Issuer string
	// AccessTTL 访问令牌有效期
	AccessTTL time.Duration
	// RefreshTTL 刷新令牌有效期
	RefreshTTL time.Duration
}

// DefaultJWTConfig 返回默认 JWT 配置
// 生产环境必须通过 SetSecretKey() 注入真实密钥
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		SecretKey:  "", // 必须通过 SetSecretKey 设置
		Issuer:     DefaultJWTIssuer,
		AccessTTL:  DefaultAccessTTL,
		RefreshTTL: DefaultRefreshTTL,
	}
}

// SetSecretKey 设置签名密钥（链式调用）
func (c *JWTConfig) SetSecretKey(key string) *JWTConfig {
	c.SecretKey = key
	return c
}

// Validate 验证配置有效性
func (c *JWTConfig) Validate() error {
	if c.SecretKey == "" {
		return errors.New("JWT SecretKey 不能为空")
	}
	if len(c.SecretKey) < 32 {
		return fmt.Errorf("JWT SecretKey 长度不足：当前 %d 字节，建议 >= 32 字节", len(c.SecretKey))
	}
	return nil
}

// JWTManager JWT 令牌管理器
// 负责 access_token 和 refresh_token 的签发与解析
type JWTManager struct {
	config *JWTConfig
}

// NewJWTManager 创建 JWT 管理器实例
func NewJWTManager(config *JWTConfig) (*JWTManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("JWT 配置无效: %w", err)
	}
	return &JWTManager{config: config}, nil
}

// GenerateAccessToken 签发访问令牌
// Claims 包含 user_id、token_type、iat、exp、iss
func (m *JWTManager) GenerateAccessToken(userID string) (string, int64, error) {
	now := time.Now()
	exp := now.Add(m.config.AccessTTL)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    userID,
		"token_type": "access",
		"jti":        fmt.Sprintf("acc-%d", now.UnixNano()), // 唯一 ID，确保每次签发不同令牌
		"iat":        now.Unix(),
		"exp":        exp.Unix(),
		"iss":        m.config.Issuer,
	})

	tokenStr, err := token.SignedString([]byte(m.config.SecretKey))
	if err != nil {
		return "", 0, fmt.Errorf("签发 access_token 失败: %w", err)
	}

	return tokenStr, exp.UnixMilli(), nil
}

// GenerateRefreshToken 签发刷新令牌
// refresh_token 的 TTL 比 access_token 更长，用于换取新的 access_token
func (m *JWTManager) GenerateRefreshToken(userID string) (string, int64, error) {
	now := time.Now()
	exp := now.Add(m.config.RefreshTTL)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    userID,
		"token_type": "refresh",
		"jti":        fmt.Sprintf("ref-%d", now.UnixNano()), // 唯一 ID，确保每次签发不同令牌
		"iat":        now.Unix(),
		"exp":        exp.Unix(),
		"iss":        m.config.Issuer,
	})

	tokenStr, err := token.SignedString([]byte(m.config.SecretKey))
	if err != nil {
		return "", 0, fmt.Errorf("签发 refresh_token 失败: %w", err)
	}

	return tokenStr, exp.UnixMilli(), nil
}

// ParseToken 解析并验证令牌
// 返回 claims 中的 user_id 和 token_type；过期返回 ErrTokenExpired
func (m *JWTManager) ParseToken(tokenStr string) (userID string, tokenType string, err error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// 强制要求 HS256 签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", "", ErrTokenExpired
		}
		return "", "", fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", ErrTokenInvalid
	}

	userID, _ = claims["user_id"].(string)
	tokenType, _ = claims["token_type"].(string)

	return userID, tokenType, nil
}
