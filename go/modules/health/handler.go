package health

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/common-lib/health"
	"github.com/jimiechen/mineplanet/go/common-lib/module"
)

// Handler 负责 Protobuf bytes ↔ 业务逻辑转换
type Handler struct {
	usecase   *Usecase
	logger    module.Logger
	checkersMu sync.RWMutex
}

// NewHandler 创建 Handler 实例
func NewHandler(usecase *Usecase, logger module.Logger) *Handler {
	return &Handler{
		usecase: usecase,
		logger:  logger,
	}
}

// Register 动态注册额外 Checker（线程安全）
func (h *Handler) Register(checker health.Checker) {
	h.checkersMu.Lock()
	defer h.checkersMu.Unlock()
	h.usecase.RegisterChecker(checker)
}

// HandleCheck 处理健康检查请求
func (h *Handler) HandleCheck(ctx context.Context, request []byte) ([]byte, error) {
	var req pb.ServiceHealthCheckRequest

	if err := proto.Unmarshal(request, &req); err != nil {
		h.logger.Error(ctx, "请求解析失败", "error", err)
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	rsp, err := h.usecase.DoCheck(ctx, &req)
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
