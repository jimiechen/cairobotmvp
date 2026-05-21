package adapter

import (
	"context"
	"fmt"

	"github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/services/hello"
	"github.com/jimiechen/mineplanet/go/services/health"
)

// SystemAdapter Tars 调用适配器
// 将 Tars servant 接口适配到模块化业务服务
// 不包含业务逻辑，只做方法分发和返回码转换
type SystemAdapter struct {
	helloSvc  hello.HelloService
	healthSvc health.HealthService
}

// NewSystemAdapter 创建 SystemAdapter 实例
func NewSystemAdapter() *SystemAdapter {
	return &SystemAdapter{
		helloSvc:  hello.NewService(),
		healthSvc: health.NewService(),
	}
}

// Invoke 执行 Tars 调用分发
// method: 从 extend["method"] 获取目标方法名
// request: Protobuf 序列化的请求 bytes
// 返回: returnCode, responseBytes, error
func (a *SystemAdapter) Invoke(ctx context.Context, request []byte, extend map[string]string) (int, []byte, error) {
	method := extend["method"]
	if method == "" {
		method = "HealthCheck"
	}

	switch method {
	case "HealthCheck":
		resp, err := a.healthSvc.Check(ctx, request)
		if err != nil {
			return commonlib.CodeInternalError, nil, err
		}
		return commonlib.CodeSuccess, resp, nil

	case "HelloWorld":
		resp, err := a.helloSvc.SayHello(ctx, request)
		if err != nil {
			return commonlib.CodeInternalError, nil, err
		}
		return commonlib.CodeSuccess, resp, nil

	default:
		return commonlib.CodeNotFound, nil, fmt.Errorf("unknown method: %s", method)
	}
}
