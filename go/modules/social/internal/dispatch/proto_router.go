// Package dispatch 提供基于 Go 泛型的 Protobuf 协议路由器
// 用于消除 Handler.Dispatch 中重复的 Unmarshal→Handle→Marshal 样板代码
//
// 使用方式：
//
//	router := dispatch.NewProtoRouter()
//	dispatch.Register(router, "1021", registerSvc)  // 注册路由
//	resp, err := router.Dispatch(ctx, "1021", reqBytes) // 分发请求
package dispatch

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
)

// ProtoService 定义 Protobuf 服务处理器的统一接口
// Req 为请求消息类型，Resp 为响应消息类型（均需为 proto.Message）
type ProtoService[Req any, Resp any] interface {
	Handle(context.Context, *Req) (*Resp, error)
}

// ProtoRouter 协议路由器，按 minType 字符串路由到对应的 ProtoService 处理器
type ProtoRouter struct {
	handlers map[string]func(context.Context, []byte) ([]byte, error)
}

// NewProtoRouter 创建空的 ProtoRouter
func NewProtoRouter() *ProtoRouter {
	return &ProtoRouter{
		handlers: make(map[string]func(context.Context, []byte) ([]byte, error)),
	}
}

// Register 将 ProtoService 处理器注册到指定 minType 路由
// 内部封装 Unmarshal → Handle → Marshal 的标准协议处理流程
func Register[Req any, Resp any](router *ProtoRouter, minType string, handler ProtoService[Req, Resp]) {
	router.handlers[minType] = func(ctx context.Context, reqBytes []byte) ([]byte, error) {
		var req Req
		if err := proto.Unmarshal(reqBytes, &req); err != nil {
			return nil, fmt.Errorf("unmarshal %T failed: %w", req, err)
		}
		resp, err := handler.Handle(ctx, &req)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(resp)
	}
}

// Dispatch 根据 minType 分发请求到已注册的处理器
// 当 minType 未注册时返回 "unsupported minType" 错误
func (r *ProtoRouter) Dispatch(ctx context.Context, minType string, reqBytes []byte) ([]byte, error) {
	h, ok := r.handlers[minType]
	if !ok {
		return nil, fmt.Errorf("unsupported minType: %s", minType)
	}
	return h(ctx, reqBytes)
}
