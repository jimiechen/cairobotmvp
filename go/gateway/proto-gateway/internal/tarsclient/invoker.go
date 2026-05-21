package tarsclient

import (
	"context"
	"errors"

	"github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/services/hello"
	"github.com/jimiechen/mineplanet/go/services/health"
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

// LocalInvoker 单体部署（monolith）模式下的本进程 TarsGo servant adapter
// 不绕过 Tars 框架，而是在同一部署单元内通过进程内调用转发到 TarsGo servant
// 严格遵守 Tars bytes 契约：request/response 均为 Protobuf bytes
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

// RegisterSystemHandlers 注册 System 模块的本地 TarsGo servant handler
// Gateway 单体部署模式启动时必须调用，注册 HealthCheck 和 HelloWorld
//
// Deprecated: 请使用 RegisterModuleHandlers 替代，后者直接使用模块化服务。
// 此函数保留用于向后兼容，未来版本将移除。
func RegisterSystemHandlers(invoker *LocalInvoker) {
	sysAdapter := NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		healthSvc := health.NewService()
		return healthSvc.Check(ctx, req)
	})

	helloAdapter := NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		helloSvc := hello.NewService()
		return helloSvc.SayHello(ctx, req)
	})

	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "SystemServer",
		Servant: "SystemObj",
		Method:  "HealthCheck",
	}, sysAdapter)

	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "SystemServer",
		Servant: "SystemObj",
		Method:  "HelloWorld",
	}, helloAdapter)
}

// ModuleInvokeFunc 模块服务调用函数签名
// 业务模块统一使用 Protobuf bytes 作为输入输出，不依赖 MessagePacket
//
// Deprecated: 请使用 commonlib.ModuleInvokeFunc 替代。
type ModuleInvokeFunc = commonlib.ModuleInvokeFunc

// moduleHandler 将模块服务适配为 LocalHandler 接口
// 负责返回码转换：模块成功→10200，模块失败→10500
type moduleHandler struct {
	fn ModuleInvokeFunc
}

// NewModuleHandler 创建模块适配 handler
func NewModuleHandler(fn ModuleInvokeFunc) LocalHandler {
	return &moduleHandler{fn: fn}
}

// Invoke 实现 LocalHandler 接口，调用底层模块服务并转换返回码
func (h *moduleHandler) Invoke(ctx context.Context, request []byte, extend map[string]string) (int, []byte, error) {
	resp, err := h.fn(ctx, request)
	if err != nil {
		return commonlib.CodeInternalError, nil, err
	}
	return commonlib.CodeSuccess, resp, nil
}

// RegisterModuleHandlers 注册模块化业务服务的本地 handler
// 每个模块独立注册到对应的 TargetKey，通过 NewModuleHandler 适配接口
// 替代 RegisterSystemHandlers，推荐新代码使用此函数
func RegisterModuleHandlers(invoker *LocalInvoker) {
	helloSvc := hello.NewService()
	healthSvc := health.NewService()

	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "SystemServer",
		Servant: "SystemObj",
		Method:  "HealthCheck",
	}, NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		return healthSvc.Check(ctx, req)
	}))

	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "SystemServer",
		Servant: "SystemObj",
		Method:  "HelloWorld",
	}, NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		return helloSvc.SayHello(ctx, req)
	}))
}

// TarsGoInvoker 微服务部署（microservice）模式下的远程 TarsGo client invoker
// 通过 TarsGo client 远程调用独立部署的 TarsCloud servant
// 与 LocalInvoker 共享同一 TarsInvoker 接口，严格遵守 Tars bytes 契约
// S1 阶段未实现，当前调用返回 10500 错误
type TarsGoInvoker struct{}

// NewTarsGoInvoker 创建 TarsGoInvoker
func NewTarsGoInvoker() *TarsGoInvoker {
	return &TarsGoInvoker{}
}

// Invoke 执行远程 TarsGo 调用（S1 未实现）
func (ti *TarsGoInvoker) Invoke(ctx context.Context, target Target, request []byte, extend map[string]string) (int, []byte, error) {
	return 10500, nil, errors.New("tars invoker is not implemented yet")
}
