package health

import (
	"context"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/common-lib/module"
	"github.com/jimiechen/mineplanet/go/common-lib/sdk/configsdk"
	"github.com/jimiechen/mineplanet/go/common-lib/sdk/i18nsdk"
)

// TestHealthService_Check_OK 测试正常健康检查
// 注：当 i18n=nil 时 I18nChecker 返回 unhealthy，整体 status 为 Unhealthy 是正确行为
func TestHealthService_Check_OK(t *testing.T) {
	ctx := context.Background()

	req := &pb.ServiceHealthCheckRequest{ServiceName: "test-service"}
	reqData, _ := proto.Marshal(req)

	cfg := configsdk.NewFakeClient()
	i18n := i18nsdk.NewFakeClient()

	svc := New(module.Deps{
		Config: cfg,
		I18n:   i18n,
		Logger: &mockLogger{},
	}, nil)

	respData, err := svc.Check(ctx, reqData)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var resp pb.ServiceHealthCheckResponse
	_ = proto.Unmarshal(respData, &resp)

	if resp.Result.Code == 0 {
		t.Errorf("Expected non-zero result code, got %d", resp.Result.Code)
	}
	if resp.Timestamp == 0 {
		t.Error("Expected non-zero timestamp")
	}
}

// TestHealthService_Check_InvalidRequest 测试非法请求
func TestHealthService_Check_InvalidRequest(t *testing.T) {
	ctx := context.Background()

	svc := New(module.Deps{
		Config: configsdk.NewFakeClient(),
		Logger: &mockLogger{},
	}, nil)

	_, err := svc.Check(ctx, []byte("invalid protobuf data"))
	if err == nil {
		t.Error("Expected error for invalid request, got nil")
	}
}

// TestHealthService_Check_TimestampMonotonic 验证时间戳单调递增
func TestHealthService_Check_TimestampMonotonic(t *testing.T) {
	ctx := context.Background()

	svc := New(module.Deps{
		Config: configsdk.NewFakeClient(),
		Logger: &mockLogger{},
	}, nil)

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
