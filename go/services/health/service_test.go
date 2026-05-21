package health

import (
	"context"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// TestHealthService_Check_OK 测试正常健康检查
// Given: 有效的 HealthCheckRequest (serviceName="test-service")
// When: 调用 Check
// Then: 返回成功的 HealthCheckResponse (status="OK")
func TestHealthService_Check_OK(t *testing.T) {
	ctx := context.Background()

	req := &pb.ServiceHealthCheckRequest{
		ServiceName: "test-service",
	}
	reqData, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	svc := NewService()
	respData, err := svc.Check(ctx, reqData)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var resp pb.ServiceHealthCheckResponse
	if err = proto.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Result == nil || resp.Result.Code != 10200 {
		t.Errorf("Expected result code 10200, got %v", resp.Result)
	}
	if resp.Status != "OK" {
		t.Errorf("Expected status 'OK', got '%s'", resp.Status)
	}
	if resp.Timestamp == 0 {
		t.Error("Expected non-zero timestamp")
	}
}

// TestHealthService_Check_WithEmptyServiceName 测试空服务名称
// Given: serviceName 为空的请求
// When: 调用 Check
// Then: 返回成功响应（服务名称可选）
func TestHealthService_Check_WithEmptyServiceName(t *testing.T) {
	ctx := context.Background()

	req := &pb.ServiceHealthCheckRequest{
		ServiceName: "",
	}
	reqData, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	svc := NewService()
	respData, err := svc.Check(ctx, reqData)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var resp pb.ServiceHealthCheckResponse
	if err = proto.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Status != "OK" {
		t.Errorf("Expected status 'OK' for empty service name, got '%s'", resp.Status)
	}
}

// TestHealthService_Check_InvalidRequest 测试非法请求
// Given: 无效的 Protobuf bytes
// When: 调用 Check
// Then: 返回错误
func TestHealthService_Check_InvalidRequest(t *testing.T) {
	ctx := context.Background()

	svc := NewService()

	invalidData := []byte("invalid protobuf data")

	_, err := svc.Check(ctx, invalidData)
	if err == nil {
		t.Error("Expected error for invalid request, got nil")
	}
}

// TestHealthService_Check_TimestampMonotonic 验证时间戳单调递增
// Given: 连续调用两次 Check
// When: 比较两次响应的时间戳
// Then: 第二次时间戳 >= 第一次
func TestHealthService_Check_TimestampMonotonic(t *testing.T) {
	ctx := context.Background()

	svc := NewService()

	req := &pb.ServiceHealthCheckRequest{ServiceName: "monotonic-test"}
	reqData, _ := proto.Marshal(req)

	respData1, _ := svc.Check(ctx, reqData)
	time.Sleep(10 * time.Millisecond)
	respData2, _ := svc.Check(ctx, reqData)

	var resp1, resp2 pb.ServiceHealthCheckResponse
	proto.Unmarshal(respData1, &resp1)
	proto.Unmarshal(respData2, &resp2)

	if resp2.Timestamp < resp1.Timestamp {
		t.Errorf("Timestamp should be monotonic: first=%d, second=%d", resp1.Timestamp, resp2.Timestamp)
	}
}
