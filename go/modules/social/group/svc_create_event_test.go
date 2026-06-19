package group

import (
	"context"
	"encoding/json"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// TestSvcCreateGroup_创建成功_应发布GroupCreated事件 验证创建圈子成功后发布 GroupCreated 领域事件
func TestSvcCreateGroup_创建成功_应发布GroupCreated事件(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	publisher := &event.FakePublisher{}
	svc := NewSvcCreate(mockRepo, publisher)
	ctx := WithOwnerID(context.Background(), "owner-event-001")

	req := &pb.CreateGroupRequest{
		Name:     "事件测试圈子",
		Slug:     "event-test-circle",
		Category: "技术",
		JoinMode: pb.JoinMode_JOIN_MODE_OPEN,
	}

	// Act
	resp, err := svc.Handle(ctx, req)

	// Assert — 业务结果正确
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Fatalf("期望成功码，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}

	// Assert — 事件已发布
	if publisher.EventCount() != 1 {
		t.Fatalf("期望发布 1 个事件，实际发布 %d 个", publisher.EventCount())
	}

	publishedEvt := publisher.Events()[0]
	if publishedEvt.Type != event.EventGroupCreated {
		t.Errorf("期望事件类型 %s，实际得到 %s", event.EventGroupCreated, publishedEvt.Type)
	}

	// Assert — payload 字段校验
	var payload event.GroupCreatedPayload
	if err := json.Unmarshal(publishedEvt.Payload, &payload); err != nil {
		t.Fatalf("反序列化 payload 失败: %v", err)
	}
	if payload.GroupID != resp.GroupId {
		t.Errorf("期望 payload.GroupID 为 %s，实际得到 %s", resp.GroupId, payload.GroupID)
	}
	if payload.OwnerID != "owner-event-001" {
		t.Errorf("期望 payload.OwnerID 为 owner-event-001，实际得到 %s", payload.OwnerID)
	}
}

// TestSvcCreateGroup_nil_publisher_不发布事件 验证 publisher 为 nil 时不 panic 且业务正常
func TestSvcCreateGroup_nil_publisher_不发布事件(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcCreate(mockRepo, nil) // nil publisher
	ctx := WithOwnerID(context.Background(), "owner-nil-001")

	req := &pb.CreateGroupRequest{
		Name:     "NilPub圈子",
		Slug:     "nilpub-slug",
		JoinMode: pb.JoinMode_JOIN_MODE_OPEN,
	}

	// Act — 不应 panic
	resp, err := svc.Handle(ctx, req)

	// Assert — 业务正常完成
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Fatalf("期望成功码，实际得到 %d", resp.Result.Code)
	}
}
