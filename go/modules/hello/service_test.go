package hello

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestHelloService_SayHello_OK 测试正常问候场景
// Given: 有效的 HelloWorldRequest (name="CaiRobot")
// When: 调用 SayHello
// Then: 返回成功的 HelloWorldResponse (message="Hello, CaiRobot!")
func TestHelloService_SayHello_OK(t *testing.T) {
	ctx := context.Background()

	// 构造请求 Protobuf bytes
	req := &pb.HelloWorldRequest{
		Name: "CaiRobot",
	}
	reqData, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// 创建 Service 实例
	svc := NewService()

	// 调用 SayHello
	respData, err := svc.SayHello(ctx, reqData)
	if err != nil {
		t.Fatalf("SayHello returned error: %v", err)
	}

	// 解析响应 Protobuf bytes
	var resp pb.HelloWorldResponse
	if err = proto.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// 验证响应内容
	if resp.Result == nil || resp.Result.Code != 10200 {
		t.Errorf("Expected result code 10200, got %v", resp.Result)
	}
	if resp.Message != "Hello, CaiRobot!" {
		t.Errorf("Expected message 'Hello, CaiRobot!', got '%s'", resp.Message)
	}
}

// TestHelloService_SayHello_DefaultName 测试空名称使用默认值
// Given: name 为空的 HelloWorldRequest
// When: 调用 SayHello
// Then: 返回 "Hello, World!" （使用默认名称）
func TestHelloService_SayHello_DefaultName(t *testing.T) {
	ctx := context.Background()

	// 构造空 name 的请求
	req := &pb.HelloWorldRequest{
		Name: "",
	}
	reqData, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	svc := NewService()
	respData, err := svc.SayHello(ctx, reqData)
	if err != nil {
		t.Fatalf("SayHello returned error: %v", err)
	}

	var resp pb.HelloWorldResponse
	if err = proto.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// 验证默认名称
	if resp.Message != "Hello, World!" {
		t.Errorf("Expected message 'Hello, World!', got '%s'", resp.Message)
	}
}

// TestHelloService_SayHello_InvalidRequest 测试非法请求
// Given: 无效的 Protobuf bytes
// When: 调用 SayHello
// Then: 返回错误
func TestHelloService_SayHello_InvalidRequest(t *testing.T) {
	ctx := context.Background()

	svc := NewService()

	// 使用无效的 bytes
	invalidData := []byte("this is not valid protobuf")

	_, err := svc.SayHello(ctx, invalidData)
	if err == nil {
		t.Error("Expected error for invalid request, got nil")
	}
}
