package health

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	commonlib "github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/common-lib/health"
	"github.com/jimiechen/mineplanet/go/common-lib/sdk/configsdk"
	"github.com/jimiechen/mineplanet/go/common-lib/sdk/i18nsdk"
)

func TestUsecase_DoCheck_BasicOK(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	cfg.Set("system_cfg", "build_version", "1.0.0-test")
	cfg.Set("health_cfg", "max_depth", int64(2))

	usecase := NewUsecase(cfg, nil, nil)

	req := &pb.ServiceHealthCheckRequest{}

	rsp, err := usecase.DoCheck(context.Background(), req)
	if err != nil {
		t.Fatalf("期望成功，实际失败: %v", err)
	}

	if rsp.Result.Code != commonlib.CodeSuccess {
		t.Fatalf("期望 code=%d，实际 %d", commonlib.CodeSuccess, rsp.Result.Code)
	}
	if rsp.Status != "OK" {
		t.Fatalf("期望 status='OK'，实际 '%s'", rsp.Status)
	}
}

func TestUsecase_DoCheck_WithCheckers(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	cfg.Set("system_cfg", "build_version", "1.0.0")

	i18n := i18nsdk.NewFakeClient()
	i18n.SetTranslation("zh-CN", "svc_health_status_summary", "{healthy} / {total} 项依赖正常")

	checkers := []health.Checker{
		NewConfigChecker(cfg),
		NewI18nChecker(i18n),
	}

	usecase := NewUsecase(cfg, i18n, checkers)

	req := &pb.ServiceHealthCheckRequest{}

	rsp, err := usecase.DoCheck(context.Background(), req)
	if err != nil {
		t.Fatalf("期望成功，实际失败: %v", err)
	}

	if rsp.Status != "OK" {
		t.Fatalf("期望所有组件健康时 status='OK'，实际 '%s'", rsp.Status)
	}
}

func TestUsecase_DoCheck_DefaultVersion(t *testing.T) {
	cfg := configsdk.NewFakeClient()

	usecase := NewUsecase(cfg, nil, nil)

	req := &pb.ServiceHealthCheckRequest{}

	rsp, err := usecase.DoCheck(context.Background(), req)
	if err != nil {
		t.Fatalf("期望成功，实际失败: %v", err)
	}

	version := usecase.GetVersion(context.Background())
	if version != "0.0.0-dev" {
		t.Fatalf("期望默认 version='0.0.0-dev'，实际 '%s'", version)
	}
	_ = rsp
}

func TestUsecase_DoCheck_NilI18nClient(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	checkers := []health.Checker{NewConfigChecker(cfg)}

	usecase := NewUsecase(cfg, nil, checkers)

	req := &pb.ServiceHealthCheckRequest{}

	rsp, err := usecase.DoCheck(context.Background(), req)
	if err != nil {
		t.Fatalf("期望成功，实际失败: %v", err)
	}

	message := usecase.GetMessage(context.Background())
	if message == "" {
		t.Fatal("期望非空状态摘要")
	}
	_ = rsp
}

func TestUsecase_DoCheck_I18nRenderFailed(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	i18n := i18nsdk.NewFakeClient()
	i18n.SetErrorOnKey("svc_health_status_summary")

	checkers := []health.Checker{NewConfigChecker(cfg)}

	usecase := NewUsecase(cfg, i18n, checkers)

	req := &pb.ServiceHealthCheckRequest{}

	rsp, err := usecase.DoCheck(context.Background(), req)
	if err != nil {
		t.Fatalf("期望成功（降级），实际失败: %v", err)
	}

	if rsp.Status != "OK" {
		t.Fatalf("期望降级时仍返回 OK status，实际 '%s'", rsp.Status)
	}
	_ = rsp
}

func TestHandler_HandleCheck_ValidRequest(t *testing.T) {
	logger := &mockLogger{}
	cfg := configsdk.NewFakeClient()
	i18n := i18nsdk.NewFakeClient()

	usecase := NewUsecase(cfg, i18n, nil)
	handler := NewHandler(usecase, logger)

	req := &pb.ServiceHealthCheckRequest{}
	reqBytes, _ := proto.Marshal(req)

	respBytes, err := handler.HandleCheck(context.Background(), reqBytes)
	if err != nil {
		t.Fatalf("期望成功，实际失败: %v", err)
	}

	var rsp pb.ServiceHealthCheckResponse
	if err := proto.Unmarshal(respBytes, &rsp); err != nil {
		t.Fatalf("响应解析失败: %v", err)
	}

	if rsp.Result.Code != commonlib.CodeSuccess {
		t.Fatalf("期望成功响应，实际 code=%d", rsp.Result.Code)
	}
}

func TestHandler_HandleCheck_InvalidRequest(t *testing.T) {
	logger := &mockLogger{}
	usecase := NewUsecase(configsdk.NewFakeClient(), nil, nil)
	handler := NewHandler(usecase, logger)

	invalidBytes := []byte("invalid protobuf data")

	_, err := handler.HandleCheck(context.Background(), invalidBytes)
	if err == nil {
		t.Fatal("期望解析失败，实际成功")
	}
}

func TestConcurrentCheckersTimeout(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	slowChecker := &slowChecker{name: "slow-checker"}

	checkers := []health.Checker{slowChecker}
	usecase := NewUsecase(cfg, nil, checkers)

	req := &pb.ServiceHealthCheckRequest{}

	start := time.Now()
	rsp, err := usecase.DoCheck(context.Background(), req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("期望成功，实际失败: %v", err)
	}

	statuses := usecase.GetComponentStatuses(context.Background())
	if len(statuses) != 1 {
		t.Fatalf("期望 1 个组件状态，实际 %d", len(statuses))
	}

	if statuses[0].Healthy {
		t.Fatal("期望超时时组件不健康")
	}

	if elapsed > 2*time.Second {
		t.Fatalf("期望在 1s 超时内完成，实际 %.2f 秒", elapsed.Seconds())
	}
	_ = rsp
}

type slowChecker struct {
	name string
}

func (s *slowChecker) Name() string { return s.name }

func (s *slowChecker) Check(ctx context.Context) health.ComponentStatus {
	time.Sleep(5 * time.Second)
	return health.NewComponentStatus(s.name, 5000)
}

type mockLogger struct{}

func (l *mockLogger) Info(ctx context.Context, v ...interface{})   {}
func (l *mockLogger) Infof(ctx context.Context, f string, v ...interface{}) {}
func (l *mockLogger) Error(ctx context.Context, v ...interface{})  {}
func (l *mockLogger) Errorf(ctx context.Context, f string, v ...interface{}) {}
func (l *mockLogger) Warn(ctx context.Context, v ...interface{})  {}
func (l *mockLogger) Debug(ctx context.Context, v ...interface{}) {}

func init() {
	fmt.Println("health tests loaded")
}
