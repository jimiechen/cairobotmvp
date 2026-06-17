package member

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcUpdateUserInfo_正常更新昵称 当传入合法昵称时_应更新并返回新信息
func TestSvcUpdateUserInfo_正常更新昵称(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	originalUser := &User{
		ID:       "user-001",
		Username: "testuser",
		Nickname: "旧昵称",
		Email:    "old@example.com",
		Status:   UserStatusActive,
	}
	mockRepo.users[originalUser.ID] = originalUser

	svc := NewSvcUpdateUserInfo(mockRepo)
	ctx := context.WithValue(context.Background(), ctxKeyUserID, "user-001")

	req := &pb.UpdateUserInfoRequest{
		Nickname: "新昵称",
	}

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
	if resp.UserInfo.Nickname != "新昵称" {
		t.Errorf("期望 Nickname=新昵称，实际得到 %s", resp.UserInfo.Nickname)
	}
	// 验证 repo 中的数据已更新
	updatedUser, _ := mockRepo.GetUserByID(context.Background(), "user-001")
	if updatedUser.Nickname != "新昵称" {
		t.Errorf("期望 repo 中 Nickname 已更新为 新昵称，实际得到 %s", updatedUser.Nickname)
	}
}

// TestSvcUpdateUserInfo_全可选字段为空 当不传任何字段时_应仍返回成功
func TestSvcUpdateUserInfo_全可选字段为空(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	originalUser := &User{
		ID:       "user-001",
		Username: "testuser",
		Nickname: "原始昵称",
		Avatar:   "https://example.com/old.png",
		Email:    "old@example.com",
		Phone:    "13800138000",
		Status:   UserStatusActive,
	}
	mockRepo.users[originalUser.ID] = originalUser

	svc := NewSvcUpdateUserInfo(mockRepo)
	ctx := context.WithValue(context.Background(), ctxKeyUserID, "user-001")

	req := &pb.UpdateUserInfoRequest{}

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
}

// TestSvcUpdateUserInfo_用户不存在 当userId对应记录不存在时_应返回错误
func TestSvcUpdateUserInfo_用户不存在(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcUpdateUserInfo(mockRepo)
	ctx := context.WithValue(context.Background(), ctxKeyUserID, "nonexistent")

	req := &pb.UpdateUserInfoRequest{
		Nickname: "新昵称",
	}

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

// TestSvcUpdateUserInfo_缺少用户ID 当上下文无userId时_应返回参数校验错误
func TestSvcUpdateUserInfo_缺少用户ID(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcUpdateUserInfo(mockRepo)

	req := &pb.UpdateUserInfoRequest{
		Nickname: "新昵称",
	}

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
