package member

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// newTestLogoutDependencies 创建登出测试依赖（JWTManager + MemoryTokenStore）
func newTestLogoutDependencies(t *testing.T) (*JWTManager, TokenStore) {
	t.Helper()
	jwtCfg := DefaultJWTConfig().SetSecretKey("test-secret-key-for-jwt-unit-tests-32b")
	jwtManager, err := NewJWTManager(jwtCfg)
	if err != nil {
		t.Fatalf("创建 JWTManager 失败: %v", err)
	}
	tokenStore := NewMemoryTokenStore()
	return jwtManager, tokenStore
}

// TestSvcLogout_正常登出 当提供有效用户ID时_应返回成功响应
func TestSvcLogout_正常登出(t *testing.T) {
	// Arrange
	jwtManager, tokenStore := newTestLogoutDependencies(t)
	svc := NewSvcLogout(tokenStore, jwtManager)

	req := &pb.UserLogoutRequest{
		UserId:      "user-001",
		AccessToken: "access-token-123",
		DeviceId:    "device-abc",
	}

	// Act
	resp, err := svc.Handle(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp == nil {
		t.Fatal("期望返回 Response，实际为 nil")
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if resp.UserId != "user-001" {
		t.Errorf("期望 UserId=user-001，实际得到 %s", resp.UserId)
	}
}

// TestSvcLogout_缺少用户ID 当user_id为空时_应返回参数校验错误
func TestSvcLogout_缺少用户ID(t *testing.T) {
	// Arrange
	jwtManager, tokenStore := newTestLogoutDependencies(t)
	svc := NewSvcLogout(tokenStore, jwtManager)

	req := &pb.UserLogoutRequest{
		UserId: "",
	}

	// Act
	resp, err := svc.Handle(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code >= 10200 && resp.Result.Code < 10400 {
		t.Errorf("期望参数校验错误（>= 10400），实际得到成功类码 %d", resp.Result.Code)
	}
}
