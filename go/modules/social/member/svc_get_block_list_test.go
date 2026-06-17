package member

import (
	"context"
	"fmt"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcGetBlockList_正常分页查询 当存在拉黑记录时_应返回分页列表和总数
func TestSvcGetBlockList_正常分页查询(t *testing.T) {
	// Arrange — 先创建多条拉黑记录
	mockRepo := newMockRepository()
	blockSvc := NewSvcBlock(mockRepo)

	// 创建 3 条拉黑记录
	for i := 1; i <= 3; i++ {
		blockedID := fmt.Sprintf("user-%03d", i+10)
		blockReq := &pb.BlockUserRequest{
			BlockedBy: "user-001",
			UserId:    blockedID,
			GroupId:   "group-001",
		}
		resp, err := blockSvc.Handle(context.Background(), blockReq)
		if err != nil || resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
			t.Fatalf("前置：创建拉黑记录 %d 失败, err=%v", i, err)
		}
	}

	svc := NewSvcGetBlockList(mockRepo)
	ctx := context.WithValue(context.Background(), ctxKeyUserID, "user-001")

	req := &pb.GetBlockListRequest{
		Page:     1,
		PageSize: 10,
		GroupId:  "group-001",
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
	if resp.Total != 3 {
		t.Errorf("期望 Total=3，实际得到 %d", resp.Total)
	}
	if len(resp.Blocks) != 3 {
		t.Errorf("期望 Blocks 长度=3，实际得到 %d", len(resp.Blocks))
	}
	if resp.Page != 1 {
		t.Errorf("期望 Page=1，实际得到 %d", resp.Page)
	}
	if resp.PageSize != 10 {
		t.Errorf("期望 PageSize=10，实际得到 %d", resp.PageSize)
	}
}

// TestSvcGetBlockList_空列表 当无拉黑记录时_应返回total=0
func TestSvcGetBlockList_空列表(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcGetBlockList(mockRepo)
	ctx := context.WithValue(context.Background(), ctxKeyUserID, "user-001")

	req := &pb.GetBlockListRequest{
		Page:     1,
		PageSize: 10,
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
	if resp.Total != 0 {
		t.Errorf("期望 Total=0，实际得到 %d", resp.Total)
	}
	if resp.Blocks != nil && len(resp.Blocks) != 0 {
		t.Errorf("期望 Blocks 为空列表，实际长度=%d", len(resp.Blocks))
	}
}

// TestSvcGetBlockList_缺少用户ID 当上下文无userId时_应返回参数校验错误
func TestSvcGetBlockList_缺少用户ID(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcGetBlockList(mockRepo)

	req := &pb.GetBlockListRequest{
		Page:     1,
		PageSize: 10,
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
