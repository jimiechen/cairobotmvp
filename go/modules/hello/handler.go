package hello

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/common-lib/module"
)

// Handler 负责 Protobuf bytes ↔ 业务逻辑转换
// 不包含业务规则，仅做协议适配层
type Handler struct {
	usecase *Usecase
	logger  module.Logger
}

// NewHandler 创建 Handler 实例
func NewHandler(usecase *Usecase, logger module.Logger) *Handler {
	return &Handler{
		usecase: usecase,
		logger:  logger,
	}
}

// HandleSayHello 处理问候请求
func (h *Handler) HandleSayHello(ctx context.Context, request []byte) ([]byte, error) {
	var req pb.HelloWorldRequest

	if err := proto.Unmarshal(request, &req); err != nil {
		h.logger.Error(ctx, "请求解析失败", "error", err)
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	rsp, err := h.usecase.Greet(ctx, &req)
	if err != nil {
		h.logger.Error(ctx, "业务处理失败", "error", err)
		return nil, err
	}

	respBytes, err := proto.Marshal(rsp)
	if err != nil {
		h.logger.Error(ctx, "响应序列化失败", "error", err)
		return nil, fmt.Errorf("marshal response failed: %w", err)
	}

	return respBytes, nil
}
