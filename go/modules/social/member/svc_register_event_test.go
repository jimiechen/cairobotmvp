package member

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	base "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// TestSvcRegister_注册成功_应发布MemberRegistered事件 验证注册成功后发布 MemberRegistered 领域事件
func TestSvcRegister_注册成功_应发布MemberRegistered事件(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	publisher := &event.FakePublisher{}
	svc := NewSvcRegister(mockRepo, publisher)

	req := &pb.UserRegisterRequest{
		Username: "eventuser",
		Password: "password123",
		Email:    "event@test.com",
		Nickname: "事件测试用户",
	}

	// Act
	resp, err := svc.Handle(context.Background(), req)

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
	if publishedEvt.Type != event.EventMemberRegistered {
		t.Errorf("期望事件类型 %s，实际得到 %s", event.EventMemberRegistered, publishedEvt.Type)
	}
	if publishedEvt.AggregateID != resp.UserId {
		t.Errorf("期望 AggregateID 为用户 ID %s，实际得到 %s", resp.UserId, publishedEvt.AggregateID)
	}
	if publishedEvt.ActorID != resp.UserId {
		t.Errorf("期望 ActorID 为用户 ID %s，实际得到 %s", resp.UserId, publishedEvt.ActorID)
	}
}

// TestSvcRegister_传入nil_publisher_不发布事件 验证 publisher 为 nil 时不 panic 且不发布任何事件
func TestSvcRegister_传入nil_publisher_不发布事件(t *testing.T) {
	// Arrange
	mockRepo := newMockRepository()
	svc := NewSvcRegister(mockRepo, nil) // nil publisher

	req := &pb.UserRegisterRequest{
		Username: "nilpubuser",
		Password: "password123",
	}

	// Act — 不应 panic
	resp, err := svc.Handle(context.Background(), req)

	// Assert — 业务正常完成
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != int32(base.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Fatalf("期望成功码，实际得到 %d", resp.Result.Code)
	}
	// nil publisher 下无法断言事件数量（没有 FakePublisher 实例），
	// 但只要不 panic 且业务正常即通过
}
