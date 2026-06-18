package topic

import (
	"context"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestSvcCreateTopic_正常创建 当标题和内容合法时_应返回成功响应包含topic_id
func TestSvcCreateTopic_正常创建(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcCreateTopic(mockRepo, nil)
	ctx := WithUserID(context.Background(), "user-001")

	req := &pb.CreateTopicRequest{
		Title:   "测试帖子标题",
		Content: "测试帖子内容",
		GroupId: "group-001",
	}

	resp, err := svc.Handle(ctx, req)
	if err != nil {
		t.Fatalf("期望无错误，实际得到: %v", err)
	}
	if resp.Result.Code != 10200 {
		t.Errorf("期望成功码 10200，实际得到 %d: %s", resp.Result.Code, resp.Result.Message)
	}
	if resp.TopicId == "" {
		t.Error("期望 TopicId 不为空")
	}
	if resp.TopicInfo == nil {
		t.Error("期望 TopicInfo 不为空")
	}
}

// TestSvcCreateTopic_缺少标题 当title为空时_应返回参数校验错误
func TestSvcCreateTopic_缺少标题(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewSvcCreateTopic(mockRepo, nil)

	req := &pb.CreateTopicRequest{
		Title:   "",
		Content: "有内容",
	}

	resp, err := svc.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("期望无错误（业务错误通过 Result 表达），实际得到: %v", err)
	}
	if resp.Result.Code == 10200 {
		t.Error("期望参数校验失败，实际返回成功")
	}
}
