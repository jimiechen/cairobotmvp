package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestSvcCreate_正常创建 当名称和slug合法时_应返回成功响应包含圈子信息
func TestSvcCreate_正常创建(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcCreate(mockRepo, nil)
	ctx := WithOwnerID(context.Background(), "owner-001")

	req := &pb.CreateGroupRequest{
		Name:     "测试圈子",
		Slug:     "test-circle",
		Category: "技术",
		JoinMode: pb.JoinMode_JOIN_MODE_OPEN,
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Errorf("期望成功码 %d，实际得到 %d: %s", base.ErrorCode_ERROR_CODE_SUCCESS, resp.Result.Code, resp.Result.Message)
	}
	if resp.GroupId == "" {
		t.Error("期望 GroupId 不为空")
	}
	if resp.GroupInfo == nil {
		t.Error("期望 GroupInfo 不为空")
	}
}

// TestSvcCreate_Slug重复 当slug已存在时_应返回标识符已被占用错误
func TestSvcCreate_Slug重复(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcCreate(mockRepo, nil)
	ctx := WithOwnerID(context.Background(), "owner-001")

	// 预先创建一个圈子占用 slug
	existingGroup := &Group{
		ID:   "group-001",
		Name: "已有圈子",
		Slug: "test-slug-dup",
	}
	mockRepo.groups[existingGroup.ID] = existingGroup
	mockRepo.groups[existingGroup.Slug] = existingGroup

	req := &pb.CreateGroupRequest{
		Name:     "新圈子",
		Slug:     "test-slug-dup", // 重复 slug
		JoinMode: pb.JoinMode_JOIN_MODE_OPEN,
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_NAME_ALREADY_EXISTS) {
		t.Errorf("期望标识符已被占用错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_NAME_ALREADY_EXISTS, resp.Result.Code)
	}
}

// TestSvcCreate_缺少必填字段 当名称为空时_应返回参数校验错误
func TestSvcCreate_缺少必填字段(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcCreate(mockRepo, nil)
	ctx := context.Background()

	req := &pb.CreateGroupRequest{
		Name: "", // 空名称
		Slug: "test-slug",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.GroupErrorCode_GROUP_ERROR_NAME_EMPTY) {
		t.Errorf("期望名称不能为空错误码 %d，实际得到 %d", base.GroupErrorCode_GROUP_ERROR_NAME_EMPTY, resp.Result.Code)
	}
}
