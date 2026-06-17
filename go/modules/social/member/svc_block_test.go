package member

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcBlock_正常拉黑 当参数合法时_应创建拉黑记录并返回blockInfo
func TestSvcBlock_正常拉黑(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcBlock(mockRepo)

	req := &pb.BlockUserRequest{
		BlockedBy: "user-001",
		UserId:    "user-002",
		GroupId:   "group-001",
	}

	// Act
	resp, err := svc.Handle(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s",
			base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if resp.BlockInfo == nil {
		t.Fatal("期望 BlockInfo 不为 nil")
	}
	if resp.BlockInfo.UserId != "user-002" {
		t.Errorf("期望 BlockInfo.UserId=user-002，实际得到 %s", resp.BlockInfo.UserId)
	}
	if resp.BlockInfo.BlockedBy != "user-001" {
		t.Errorf("期望 BlockInfo.BlockedBy=user-001，实际得到 %s", resp.BlockInfo.BlockedBy)
	}
	if resp.BlockInfo.GroupId != "group-001" {
		t.Errorf("期望 BlockInfo.GroupId=group-001，实际得到 %s", resp.BlockInfo.GroupId)
	}
	if resp.BlockInfo.CreatedAt == 0 {
		t.Error("期望 CreatedAt 不为 0")
	}
}

// TestSvcBlock_缺少被拉黑用户ID 当userId为空时_应返回参数校验错误
func TestSvcBlock_缺少被拉黑用户ID(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcBlock(mockRepo)

	req := &pb.BlockUserRequest{
		BlockedBy: "user-001",
		GroupId:   "group-001",
		// UserId 为空 — 必填字段缺失
	}

	// Act
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

// TestSvcBlock_缺少圈子ID 当groupId为空时_应返回参数校验错误
func TestSvcBlock_缺少圈子ID(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcBlock(mockRepo)

	req := &pb.BlockUserRequest{
		BlockedBy: "user-001",
		UserId:    "user-002",
		// GroupId 为空 — 必填字段缺失
	}

	// Act
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

// TestSvcBlock_重复拉黑同一用户 当已存在拉黑关系时_应幂等返回成功
func TestSvcBlock_重复拉黑同一用户(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcBlock(mockRepo)

	req := &pb.BlockUserRequest{
		BlockedBy: "user-001",
		UserId:    "user-002",
		GroupId:   "group-001",
	}

	// 第一次拉黑
	resp1, err1 := svc.Handle(context.Background(), req)
	if err1 != nil || resp1.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Fatalf("第一次拉黑应成功，实际 code=%d, err=%v", resp1.Result.Code, err1)
	}

	// 第二次拉黑同一用户（幂等）
	resp2, err2 := svc.Handle(context.Background(), req)

	// Assert — 仍应返回成功（幂等语义）
	if err2 != nil {
		t.Fatalf("期望无错误，实际得到: %v", err2)
	}
	if resp2.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望重复拉黑仍返回成功码 %d，实际得到 %d: %s",
			base.ErrorCode_ERROR_CODE_SUCCESS, resp2.Result.Code, resp2.Result.Message)
	}
}
