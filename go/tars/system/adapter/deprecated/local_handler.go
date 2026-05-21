package localhandler

import (
	"context"
	"fmt"

	"github.com/jimiechen/mineplanet/go/tars/system/internal/service"
)

// Handler 提供 System 模块的本地 bytes handler
// 这是 System 模块允许外部调用的适配层，不暴露业务内部实现
type Handler struct {
	service *service.SystemService
}

// NewHandler 创建 Handler
func NewHandler() *Handler {
	return &Handler{
		service: service.NewSystemService(),
	}
}

// Invoke 实现统一 bytes 调用接口
// 返回：returnCode, responseBytes, error
func (h *Handler) Invoke(ctx context.Context, request []byte, extend map[string]string) (int, []byte, error) {
	method := extend["method"]
	if method == "" {
		method = "HealthCheck"
	}

	switch method {
	case "HealthCheck":
		respBytes, err := h.service.HealthCheck(ctx, "")
		if err != nil {
			return 10500, nil, err
		}
		return 10200, respBytes, nil

	case "HelloWorld":
		respBytes, err := h.service.HelloWorld(ctx, "")
		if err != nil {
			return 10500, nil, err
		}
		return 10200, respBytes, nil

	default:
		return 10404, nil, fmt.Errorf("unknown method: %s", method)
	}
}
