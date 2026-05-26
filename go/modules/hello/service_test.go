package hello

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	commonlib "github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/common-lib/module"
	"github.com/jimiechen/mineplanet/go/common-lib/sdk/configsdk"
	"github.com/jimiechen/mineplanet/go/common-lib/sdk/i18nsdk"
)

// TestHelloService_SayHello_OK 测试正常问候场景
func TestHelloService_SayHello_OK(t *testing.T) {
	ctx := context.Background()

	req := &pb.HelloWorldRequest{Name: "CaiRobot"}
	reqData, _ := proto.Marshal(req)

	cfg := configsdk.NewFakeClient()
	cfg.Set("hello_cfg", "server_name", "CaiRobot")
	cfg.Set("hello_cfg", "max_name_length", int64(32))
	i18n := i18nsdk.NewFakeClient()
	i18n.SetTranslation("en", "svc_hello_greeting", "Hello, {name}! Welcome to {server_name}.")

	svc := New(module.Deps{
		Config: cfg,
		I18n:   i18n,
		Logger: &mockLogger{},
	})

	respData, err := svc.SayHello(ctx, reqData)
	if err != nil {
		t.Fatalf("SayHello returned error: %v", err)
	}

	var resp pb.HelloWorldResponse
	_ = proto.Unmarshal(respData, &resp)

	if resp.Result.Code != commonlib.CodeSuccess {
		t.Errorf("Expected result code %d, got %d", commonlib.CodeSuccess, resp.Result.Code)
	}
}

// TestHelloService_SayHello_InvalidRequest 测试非法请求
func TestHelloService_SayHello_InvalidRequest(t *testing.T) {
	ctx := context.Background()

	svc := New(module.Deps{
		Config: configsdk.NewFakeClient(),
		Logger: &mockLogger{},
	})

	_, err := svc.SayHello(ctx, []byte("this is not valid protobuf"))
	if err == nil {
		t.Error("Expected error for invalid request, got nil")
	}
}
