package member

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcUnblock_正常解除拉黑 当参数合法且存在拉黑关系时_应删除记录并返回成功
func TestSvcUnblock_正常解除拉黑(t *testing.T) {
	// Arrange — 先创建一条拉黑记录
	mockRepo := newMockRepository()
	blockSvc := NewSvcBlock(mockRepo)
	blockReq := &pb.BlockUserRequest{
		BlockedBy: "user-001",
		UserId:    "user-002",
		GroupId:   "group-001",
	}
	blockResp, err := blockSvc.Handle(context.Background(), blockReq)
	if err != nil || blockResp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Fatalf("前置：创建拉黑记录失败, code=%d, err=%v", blockResp.Result.Code, err)
	}

	svc := NewSvcUnblock(mockRepo)
	req := &pb.UnblockUserRequest{
		UserId:      "user-002",
		UnblockedBy: "user-001",
		GroupId:     "group-001",
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
	if resp.UserId != "user-002" {
		t.Errorf("期望 UserId=user-002，实际得到 %s", resp.UserId)
	}

	// 验证拉黑关系已删除
	isBlocked, _ := mockRepo.IsBlocked(context.Background(), "user-001", "user-002")
	if isBlocked {
		t.Error("期望拉黑关系已被删除")
	}
}

// TestSvcUnblock_目标未被拉黑 当不存在拉黑关系时_应幂等返回成功
func TestSvcUnblock_目标未被拉黑(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcUnblock(mockRepo)

	req := &pb.UnblockUserRequest{
		UserId:      "user-999",
		UnblockedBy: "user-001",
		GroupId:     "group-001",
	}

	// Act
	resp, err := svc.Handle(context.Background(), req)

	// Assert — 幂等：未拉黑也返回成功
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码（幂等）%d，实际得到 %d: %s",
			base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
}

// TestSvcUnblock_缺少用户ID 当userId为空时_应返回参数校验错误
func TestSvcUnblock_缺少用户ID(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcUnblock(mockRepo)

	req := &pb.UnblockUserRequest{
		// UserId 为空 — 必填字段缺失
		UnblockedBy: "user-001",
		GroupId:     "group-001",
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

// TestSvcUnblock_缺少圈子ID 当groupId为空时_应返回参数校验错误
func TestSvcUnblock_缺少圈子ID(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcUnblock(mockRepo)

	req := &pb.UnblockUserRequest{
		UserId:      "user-002",
		UnblockedBy: "user-001",
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
