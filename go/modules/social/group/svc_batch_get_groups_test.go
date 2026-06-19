package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcBatchGetGroups_正常请求_返回圈子列表 提供有效group_ids_应返回成功响应包含圈子信息
func TestSvcBatchGetGroups_正常请求_返回圈子列表(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcBatchGetGroups(mockRepo)

	// 预先创建两个圈子
	group1 := &Group{
		ID:   "group-001",
		Name: "测试圈子1",
		Slug: "test-slug-1",
	}
	group2 := &Group{
		ID:   "group-002",
		Name: "测试圈子2",
		Slug: "test-slug-2",
	}
	mockRepo.groups[group1.ID] = group1
	mockRepo.groups[group1.Slug] = group1
	mockRepo.groups[group2.ID] = group2
	mockRepo.groups[group2.Slug] = group2

	req := &pb.BatchGetGroupsRequest{
		GroupIds: []string{"group-001", "group-002"},
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if len(resp.Groups) != 2 {
		t.Errorf("期望返回 2 个圈子，实际得到 %d", len(resp.Groups))
	}
	if len(resp.GroupIds) != 2 {
		t.Errorf("期望 GroupIds 长度为 2，实际得到 %d", len(resp.GroupIds))
	}
}

// TestSvcBatchGetGroups_空group_ids_返回参数错误 group_ids为空_应返回参数校验错误
func TestSvcBatchGetGroups_空group_ids_返回参数错误(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcBatchGetGroups(mockRepo)

	req := &pb.BatchGetGroupsRequest{
		GroupIds: []string{}, // 空 group_ids
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_INVALID_REQUEST) {
		t.Errorf("期望参数校验错误码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_INVALID_REQUEST, resp.Result.Code, resp.Result.Message)
	}
}
