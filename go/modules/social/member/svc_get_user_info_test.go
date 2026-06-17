package member

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcGetUserInfo_正常查询 当用户存在时_应返回完整用户信息
func TestSvcGetUserInfo_正常查询(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	testUser := &User{
		ID:       "user-001",
		UID:      "100000001",
		Username: "testuser",
		Email:    "test@example.com",
		Phone:    "13800138000",
		Nickname: "测试用户",
		Avatar:   "https://example.com/avatar.png",
		Status:   UserStatusActive,
	}
	mockRepo.users[testUser.ID] = testUser

	svc := NewSvcGetUserInfo(mockRepo)

	// 将 userId 放入上下文（MVP-P0 简化：模拟认证中间件注入）
	ctx := context.WithValue(context.Background(), ctxKeyUserID, "user-001")
	req := &pb.GetUserInfoRequest{}

	// Act
	resp, err := svc.Handle(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s",
			base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if resp.UserInfo == nil {
		t.Fatal("期望 UserInfo 不为 nil")
	}
	if resp.UserInfo.UserId != "user-001" {
		t.Errorf("期望 UserId=user-001，实际得到 %s", resp.UserInfo.UserId)
	}
	if resp.UserInfo.Username != "testuser" {
		t.Errorf("期望 Username=testuser，实际得到 %s", resp.UserInfo.Username)
	}
	if resp.UserInfo.Nickname != "测试用户" {
		t.Errorf("期望 Nickname=测试用户，实际得到 %s", resp.UserInfo.Nickname)
	}
	if resp.UserId != "user-001" {
		t.Errorf("期望响应 UserId=user-001，实际得到 %s", resp.UserId)
	}
}

// TestSvcGetUserInfo_用户不存在 当userId对应记录不存在时_应返回错误
func TestSvcGetUserInfo_用户不存在(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcGetUserInfo(mockRepo)

	ctx := context.WithValue(context.Background(), ctxKeyUserID, "nonexistent")
	req := &pb.GetUserInfoRequest{}

	// Act
	resp, err := svc.Handle(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code >= int32(base.ErrorCode_ERROR_CODE_SUCCESS) &&
		resp.Result.Code < int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST) {
		t.Errorf("期望失败类错误码，实际得到成功类码 %d", resp.Result.Code)
	}
}

// TestSvcGetUserInfo_缺少用户ID 当上下文无userId时_应返回参数校验错误
func TestSvcGetUserInfo_缺少用户ID(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcGetUserInfo(mockRepo)

	req := &pb.GetUserInfoRequest{}

	// Act — 不注入 userId 到上下文
	resp, err := svc.Handle(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST) {
		t.Errorf("期望参数校验错误 %d，实际得到 %d",
			base.ErrorCode_ERROR_CODE_INVALID_REQUEST, resp.Result.Code)
	}
}
