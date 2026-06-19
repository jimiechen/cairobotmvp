package group

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// TestSvcJoin_加入群组成功_应发布GroupJoined事件 验证加入群组成功后发布 GroupJoined 领域事件
func TestSvcJoin_加入群组成功_应发布GroupJoined事件(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	publisher := &event.FakePublisher{}
	svc := NewSvcJoin(mockRepo, publisher)

	// 预先创建一个圈子
	testGroup := &Group{ID: "group-event-001", Slug: "event-join-group"}
	mockRepo.groups[testGroup.ID] = testGroup
	mockRepo.groups[testGroup.Slug] = testGroup

	ctx := WithUserID(context.Background(), "user-event-join-001")
	req := &pb.JoinGroupRequest{
		GroupId:    "group-event-001",
		JoinReason: "想加入学习",
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
	if publishedEvt.Type != event.EventGroupJoined {
		t.Errorf("期望事件类型 %s，实际得到 %s", event.EventGroupJoined, publishedEvt.Type)
	}
	if publishedEvt.AggregateID != "group-event-001" {
		t.Errorf("期望 AggregateID 为 group-event-001，实际得到 %s", publishedEvt.AggregateID)
	}
	if publishedEvt.ActorID != "user-event-join-001" {
		t.Errorf("期望 ActorID 为 user-event-join-001，实际得到 %s", publishedEvt.ActorID)
	}
}

// TestSvcJoin_nil_publisher_不发布事件 验证 publisher 为 nil 时不 panic 且业务正常
func TestSvcJoin_nil_publisher_不发布事件(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcJoin(mockRepo, nil) // nil publisher

	testGroup := &Group{ID: "group-nil-join-001", Slug: "nil-join-group"}
	mockRepo.groups[testGroup.ID] = testGroup
	mockRepo.groups[testGroup.Slug] = testGroup

	ctx := WithUserID(context.Background(), "user-nil-join-001")
	req := &pb.JoinGroupRequest{
		GroupId: "group-nil-join-001",
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
