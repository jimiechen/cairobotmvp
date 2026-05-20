package tarsclient

import (
	"context"
	"errors"

	"github.com/jimiechen/mineplanet/go/tars/system/localhandler"
)

// Target 定义 Tars 调用目标
type Target struct {
	App       string
	Server    string
	Servant   string
	Module    string
	Interface string
	Method    string
}

// TargetKey 生成目标唯一键
type TargetKey struct {
	App     string
	Server  string
	Servant string
	Method  string
}

// String 返回目标键字符串
func (tk TargetKey) String() string {
	return tk.App + "." + tk.Server + "." + tk.Servant + "." + tk.Method
}

// ToTargetKey 从 Target 生成 TargetKey
func ToTargetKey(t Target) TargetKey {
	return TargetKey{
		App:     t.App,
		Server:  t.Server,
		Servant: t.Servant,
		Method:  t.Method,
	}
}

// TarsInvoker 定义统一调用接口
type TarsInvoker interface {
	Invoke(ctx context.Context, target Target, request []byte, extend map[string]string) (returnCode int, response []byte, err error)
}

// LocalHandler 定义本地 handler 接口
type LocalHandler interface {
	Invoke(ctx context.Context, request []byte, extend map[string]string) (returnCode int, response []byte, err error)
}

// LocalInvoker 单体模式下的本地调用实现
type LocalInvoker struct {
	handlers map[string]LocalHandler
}

// NewLocalInvoker 创建 LocalInvoker
func NewLocalInvoker() *LocalInvoker {
	return &LocalInvoker{
		handlers: make(map[string]LocalHandler),
	}
}

// Register 注册本地 handler
func (li *LocalInvoker) Register(key TargetKey, handler LocalHandler) {
	li.handlers[key.String()] = handler
}

// Invoke 执行本地调用
func (li *LocalInvoker) Invoke(ctx context.Context, target Target, request []byte, extend map[string]string) (int, []byte, error) {
	key := ToTargetKey(target).String()
	handler, ok := li.handlers[key]
	if !ok {
		return 10404, nil, errors.New("local handler not found: " + key)
	}
	return handler.Invoke(ctx, request, extend)
}

// RegisterSystemHandlers 注册 System 模块的本地 handler
// Gateway local 模式启动时必须调用，注册 HealthCheck 和 HelloWorld
func RegisterSystemHandlers(invoker *LocalInvoker) {
	sysHandler := localhandler.NewHandler()

	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "SystemServer",
		Servant: "SystemObj",
		Method:  "HealthCheck",
	}, sysHandler)

	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "SystemServer",
		Servant: "SystemObj",
		Method:  "HelloWorld",
	}, sysHandler)
}

// TarsGoInvoker 微服务模式下的 TarsGo 调用实现（当前未实现）
type TarsGoInvoker struct{}

// NewTarsGoInvoker 创建 TarsGoInvoker
func NewTarsGoInvoker() *TarsGoInvoker {
	return &TarsGoInvoker{}
}

// Invoke 执行 TarsGo 调用（当前未实现）
func (ti *TarsGoInvoker) Invoke(ctx context.Context, target Target, request []byte, extend map[string]string) (int, []byte, error) {
	return 10500, nil, errors.New("tars invoker is not implemented yet")
}
