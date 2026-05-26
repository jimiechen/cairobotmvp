package hello

import (
	"context"
	"fmt"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	commonlib "github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/common-lib/i18n"
	"github.com/jimiechen/mineplanet/go/common-lib/module"
)

// Usecase Hello 模块核心业务逻辑
// 组合 configsdk + i18nsdk，演示 SDK 接入规范
type Usecase struct {
	cfg  module.ConfigReader
	i18n module.I18nRenderer
}

// NewUsecase 创建 Usecase 实例
func NewUsecase(cfg module.ConfigReader, i18n module.I18nRenderer) *Usecase {
	return &Usecase{
		cfg:  cfg,
		i18n: i18n,
	}
}

// Greet 执行问候业务逻辑
func (u *Usecase) Greet(ctx context.Context, req *pb.HelloWorldRequest) (*pb.HelloWorldResponse, error) {
	name := req.Name
	if name == "" {
		name = "World"
	}

	maxNameLength, err := u.cfg.GetInt(ctx, "hello_cfg", "max_name_length")
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}

	if int64(len(name)) > maxNameLength {
		return &pb.HelloWorldResponse{
			Result: &pb.Result{
				Code:    commonlib.CodeBadRequest,
				Message: "name too long",
			},
			Message: fmt.Sprintf("Hello, %s!", name),
		}, nil
	}

	serverName, err := u.cfg.GetString(ctx, "hello_cfg", "server_name")
	if err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}
	if serverName == "" {
		serverName = "CaiRobot"
	}

	lang := i18n.ResolveLang(ctx, "", "", u.cfg)

	greeting := u.renderGreeting(ctx, lang, name, serverName)

	return &pb.HelloWorldResponse{
		Result: &pb.Result{
			Code:    commonlib.CodeSuccess,
			Message: "success",
		},
		Message: greeting,
	}, nil
}

// renderGreeting 渲染多语言问候语
func (u *Usecase) renderGreeting(ctx context.Context, lang string, name string, serverName string) string {
	if u.i18n == nil {
		return fmt.Sprintf("Hello, %s! Welcome to %s.", name, serverName)
	}

	greeting, err := u.i18n.T(ctx, lang, "svc_hello_greeting", map[string]any{
		"name":        name,
		"server_name": serverName,
	})
	if err != nil {
		return fmt.Sprintf("Hello, %s! Welcome to %s.", name, serverName)
	}

	return greeting
}
