package member

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcLogin_正常登录 当用户名密码正确时_应返回成功响应包含令牌和用户信息
func TestSvcLogin_正常登录(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	// 预先创建一个用户（模拟已注册）
	hashedPwd, _ := hashPassword("password123")
	existingUser := &User{
		ID:       "user-001",
		Username: "testuser",
		Password: hashedPwd,
		Nickname: "测试用户",
		Email:    "test@example.com",
		Status:   UserStatusActive,
	}
	mockRepo.users[existingUser.ID] = existingUser

	svc := NewSvcLogin(mockRepo)

	req := &pb.UserLoginRequest{
		Username: "testuser",
		Password: "password123",
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
	if resp.Result.Code != 10200 { // ERROR_CODE_SUCCESS
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if resp.UserId == "" {
		t.Error("期望 UserId 不为空")
	}
	if resp.AccessToken == "" {
		t.Error("期望 AccessToken 不为空")
	}
}

// TestSvcLogin_用户不存在 当用户名未注册时_应返回用户不存在错误
func TestSvcLogin_用户不存在(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcLogin(mockRepo)

	req := &pb.UserLoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	// Act
	resp, err := svc.Handle(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code >= 10200 && resp.Result.Code < 10400 {
		t.Errorf("期望失败类错误码（>= 10400），实际得到成功类码 %d", resp.Result.Code)
	}
}

// TestSvcLogin_密码错误 当密码不匹配时_应返回认证失败错误
func TestSvcLogin_密码错误(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	existingUser := &User{
		ID:       "user-001",
		Username: "testuser",
		Password: mustHash("correct-password"),
		Status:   UserStatusActive,
	}
	mockRepo.users[existingUser.ID] = existingUser

	svc := NewSvcLogin(mockRepo)

	req := &pb.UserLoginRequest{
		Username: "testuser",
		Password: "wrong-password",
	}

	// Act
	resp, err := svc.Handle(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code != 10401 { // ERROR_CODE_UNAUTHORIZED
		t.Errorf("期望认证失败 10401，实际得到 %d", resp.Result.Code)
	}
}

// mustHash 测试辅助函数：bcrypt 哈希，失败则 panic
func mustHash(password string) string {
	h, err := hashPassword(password)
	if err != nil {
		panic("mustHash failed: " + err.Error())
	}
	return h
}
