package hello

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	commonlib "github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/common-lib/sdk/configsdk"
	"github.com/jimiechen/mineplanet/go/common-lib/sdk/i18nsdk"
)

func TestUsecase_Greet_NormalCase(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	cfg.Set("hello_cfg", "server_name", "CaiRobot")
	cfg.Set("hello_cfg", "max_name_length", int64(32))

	i18n := i18nsdk.NewFakeClient()
	i18n.SetTranslation("zh-CN", "svc_hello_greeting", "你好，{name}！欢迎使用 {server_name}。")

	usecase := NewUsecase(cfg, i18n)

	req := &pb.HelloWorldRequest{
		Name: "张三",
	}

	rsp, err := usecase.Greet(context.Background(), req)
	if err != nil {
		t.Fatalf("期望成功，实际失败: %v", err)
	}

	if rsp.Result.Code != commonlib.CodeSuccess {
		t.Fatalf("期望 code=%d，实际 %d", commonlib.CodeSuccess, rsp.Result.Code)
	}
	expectedGreeting := "你好，张三！欢迎使用 CaiRobot。"
	if rsp.Message != expectedGreeting {
		t.Fatalf("期望 message='%s'，实际 '%s'", expectedGreeting, rsp.Message)
	}
}

func TestUsecase_Greet_EnglishLang(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	cfg.Set("hello_cfg", "server_name", "CaiRobot")
	cfg.Set("hello_cfg", "max_name_length", int64(32))

	i18n := i18nsdk.NewFakeClient()
	i18n.SetTranslation("en", "svc_hello_greeting", "Hello, {name}! Welcome to {server_name}.")

	usecase := NewUsecase(cfg, i18n)

	req := &pb.HelloWorldRequest{
		Name: "Alice",
	}

	rsp, err := usecase.Greet(context.Background(), req)
	if err != nil {
		t.Fatalf("期望成功，实际失败: %v", err)
	}

	expectedGreeting := "Hello, Alice! Welcome to CaiRobot."
	if rsp.Message != expectedGreeting {
		t.Fatalf("期望 message='%s'，实际 '%s'", expectedGreeting, rsp.Message)
	}
}

func TestUsecase_Greet_NameTooLong(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	cfg.Set("hello_cfg", "server_name", "CaiRobot")
	cfg.Set("hello_cfg", "max_name_length", int64(5))

	i18n := i18nsdk.NewFakeClient()

	usecase := NewUsecase(cfg, i18n)

	req := &pb.HelloWorldRequest{
		Name: "VeryLongName",
	}

	rsp, err := usecase.Greet(context.Background(), req)
	if err != nil {
		t.Fatalf("期望成功（业务错误），实际失败: %v", err)
	}

	if rsp.Result.Code != commonlib.CodeBadRequest {
		t.Fatalf("期望 code=%d（名称过长），实际 %d", commonlib.CodeBadRequest, rsp.Result.Code)
	}
}

func TestUsecase_Greet_EmptyName(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	cfg.Set("hello_cfg", "server_name", "CaiRobot")
	cfg.Set("hello_cfg", "max_name_length", int64(32))

	i18n := i18nsdk.NewFakeClient()
	i18n.SetTranslation("zh-CN", "svc_hello_greeting", "你好，{name}！欢迎使用 {server_name}。")

	usecase := NewUsecase(cfg, i18n)

	req := &pb.HelloWorldRequest{
		Name: "",
	}

	rsp, err := usecase.Greet(context.Background(), req)
	if err != nil {
		t.Fatalf("期望成功，实际失败: %v", err)
	}

	expectedGreeting := "你好，World！欢迎使用 CaiRobot。"
	if rsp.Message != expectedGreeting {
		t.Fatalf("期望使用默认名 'World'，实际 '%s'", rsp.Message)
	}
}

func TestUsecase_Greet_I18nRenderFailed(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	cfg.Set("hello_cfg", "server_name", "CaiRobot")
	cfg.Set("hello_cfg", "max_name_length", int64(32))

	i18n := i18nsdk.NewFakeClient()
	i18n.SetErrorOnKey("svc_hello_greeting")

	usecase := NewUsecase(cfg, i18n)

	req := &pb.HelloWorldRequest{
		Name: "TestUser",
	}

	rsp, err := usecase.Greet(context.Background(), req)
	if err != nil {
		t.Fatalf("期望成功（降级），实际失败: %v", err)
	}

	expectedFallback := "Hello, TestUser! Welcome to CaiRobot."
	if rsp.Message != expectedFallback {
		t.Fatalf("期望降级文案 '%s'，实际 '%s'", expectedFallback, rsp.Message)
	}
}

func TestUsecase_Greet_NilI18nClient(t *testing.T) {
	cfg := configsdk.NewFakeClient()
	cfg.Set("hello_cfg", "server_name", "TestServer")
	cfg.Set("hello_cfg", "max_name_length", int64(100))

	usecase := NewUsecase(cfg, nil)

	req := &pb.HelloWorldRequest{
		Name: "NoI18n",
	}

	rsp, err := usecase.Greet(context.Background(), req)
	if err != nil {
		t.Fatalf("期望成功，实际失败: %v", err)
	}

	expectedFallback := "Hello, NoI18n! Welcome to TestServer."
	if rsp.Message != expectedFallback {
		t.Fatalf("期望 fallback '%s'，实际 '%s'", expectedFallback, rsp.Message)
	}
}

func TestUsecase_Greet_ConfigReadError(t *testing.T) {
	cfg := configsdk.NewFakeClient()

	usecase := NewUsecase(cfg, nil)

	req := &pb.HelloWorldRequest{
		Name: "ConfigError",
	}

	rsp, err := usecase.Greet(context.Background(), req)
	if err != nil {
		t.Fatalf("期望成功（FakeClient 对缺失 key 返回空串），实际失败: %v", err)
	}
	if rsp.Result.Code == commonlib.CodeSuccess && rsp.Message == "" {
		t.Fatalf("期望配置缺失时返回非空 message 或业务错误码")
	}
	_ = rsp
}

func TestHandler_HandleSayHello_ValidRequest(t *testing.T) {
	logger := &mockLogger{}
	cfg := configsdk.NewFakeClient()
	cfg.Set("hello_cfg", "server_name", "CaiRobot")
	cfg.Set("hello_cfg", "max_name_length", int64(32))
	i18n := i18nsdk.NewFakeClient()
	i18n.SetTranslation("zh-CN", "svc_hello_greeting", "你好，{name}！欢迎使用 {server_name}。")

	usecase := NewUsecase(cfg, i18n)
	handler := NewHandler(usecase, logger)

	req := &pb.HelloWorldRequest{
		Name: "HandlerTest",
	}
	reqBytes, _ := proto.Marshal(req)

	respBytes, err := handler.HandleSayHello(context.Background(), reqBytes)
	if err != nil {
		t.Fatalf("期望成功，实际失败: %v", err)
	}

	var rsp pb.HelloWorldResponse
	if err := proto.Unmarshal(respBytes, &rsp); err != nil {
		t.Fatalf("响应解析失败: %v", err)
	}

	if rsp.Result.Code != commonlib.CodeSuccess {
		t.Fatalf("期望成功响应，实际 code=%d", rsp.Result.Code)
	}
}

func TestHandler_HandleSayHello_InvalidRequest(t *testing.T) {
	logger := &mockLogger{}
	usecase := NewUsecase(configsdk.NewFakeClient(), nil)
	handler := NewHandler(usecase, logger)

	invalidBytes := []byte("invalid protobuf data")

	_, err := handler.HandleSayHello(context.Background(), invalidBytes)
	if err == nil {
		t.Fatal("期望解析失败，实际成功")
	}
}

type mockLogger struct{}

func (l *mockLogger) Info(ctx context.Context, v ...interface{})   {}
func (l *mockLogger) Infof(ctx context.Context, f string, v ...interface{}) {}
func (l *mockLogger) Error(ctx context.Context, v ...interface{})  {}
func (l *mockLogger) Errorf(ctx context.Context, f string, v ...interface{}) {}
func (l *mockLogger) Warn(ctx context.Context, v ...interface{})  {}
func (l *mockLogger) Debug(ctx context.Context, v ...interface{}) {}
