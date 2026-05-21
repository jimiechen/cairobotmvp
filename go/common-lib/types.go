package commonlib

import "context"

// ModuleInvokeFunc 模块服务调用函数签名
// 业务模块统一使用 Protobuf bytes 作为输入输出，不依赖 MessagePacket
type ModuleInvokeFunc func(ctx context.Context, request []byte) ([]byte, error)

// ModuleHandler 模块处理器接口
// 所有业务模块必须实现此接口或通过 Adapter 适配
type ModuleHandler interface {
	Invoke(ctx context.Context, request []byte) ([]byte, error)
}
