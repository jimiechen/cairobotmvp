// Package auth 提供 Gateway 层的 JWT 认证能力
// 职责：Token 签发、校验、解析
// 不负责：用户查找（S1 阶段不做数据库用户验证）、权限 RBAC（后续 S2 阶段实现）
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Auth 错误码（与 design spec §5.4 一致）
const (
	ErrCodeMissingToken = "40101" // 缺少 Token
	ErrCodeInvalidToken = "40102" // Token 无效/过期
	ErrCodeForbidden    = "40103" // 权限不足
)

// TokenClaims 定义 JWT 载荷，包含用户身份和角色信息
type TokenClaims struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`     // parent / child / admin
	DeviceID string `json:"device_id"` // 可选
	jwt.RegisteredClaims
}

// AuthService JWT 认证服务，负责签发和校验 Token
type AuthService struct {
	secret     []byte        // HMAC 签名密钥
	issuer     string        // 签发者标识
	expiration time.Duration // Token 有效期
}

// NewAuthService 创建 AuthService 实例
//
// 参数：
//   - secret: HMAC 签名密钥，建议至少 32 字节
//   - issuer: 签发者标识，如 "cairobot"
//   - expiration: Token 有效期，如 24*time.Hour
func NewAuthService(secret []byte, issuer string, expiration time.Duration) *AuthService {
	if len(secret) == 0 {
		secret = []byte("default-secret-change-in-production")
	}
	if issuer == "" {
		issuer = "cairobot"
	}
	if expiration == 0 {
		expiration = 24 * time.Hour
	}
	return &AuthService{
		secret:     secret,
		issuer:     issuer,
		expiration: expiration,
	}
}

// GenerateToken 签发 JWT Token
//
// 输入：
//   - userID: 用户唯一标识
//   - role: 用户角色（parent/child/admin）
//
// 输出：JWT Token 字符串
func (s *AuthService) GenerateToken(userID, role string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("userID 不能为空")
	}
	if role == "" {
		role = "parent" // 默认角色
	}

	now := time.Now()
	claims := TokenClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken 校验 JWT Token 并返回载荷
//
// 输入：tokenStr JWT 字符串
// 输出：*TokenClaims 解析后的载荷信息
// 错误：Token 格式错误、签名不匹配、过期等
func (s *AuthService) ValidateToken(tokenStr string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 校验签名算法防止算法混淆攻击
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
