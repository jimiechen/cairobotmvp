// Package service 提供 System 模块的旧版单体服务实现
//
// Deprecated: 此包已废弃，请使用模块化替代方案：
//   - HealthCheck 功能已迁移至 go/modules/health
//   - HelloWorld 功能已迁移至 go/modules/hello
//   - Tars 调用适配层已迁移至 adapter.SystemAdapter
//
// 保留原因：向后兼容，给过渡期。新代码请勿使用此包。
package service

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SystemService 系统服务业务逻辑
//
// Deprecated: 请使用 modules/hello.HelloService 和 modules/health.HealthService 替代。
type SystemService struct{}

// NewSystemService 创建 SystemService
func NewSystemService() *SystemService {
	return &SystemService{}
}

// HealthCheck 健康检查
// 返回 ServiceHealthCheckResponse 的 Protobuf bytes
func (s *SystemService) HealthCheck(ctx context.Context, serviceName string) ([]byte, error) {
	resp := &pb.ServiceHealthCheckResponse{
		Result: &pb.Result{
			Code:    10200,
			Message: "success",
		},
		Status:    "OK",
		Timestamp: time.Now().Unix(),
	}
	return proto.Marshal(resp)
}

// HelloWorld 问候
// 返回 HelloWorldResponse 的 Protobuf bytes
func (s *SystemService) HelloWorld(ctx context.Context, name string) ([]byte, error) {
	if name == "" {
		name = "World"
	}
	resp := &pb.HelloWorldResponse{
		Result: &pb.Result{
			Code:    10200,
			Message: "success",
		},
		Message:   fmt.Sprintf("Hello, %s!", name),
		Timestamp: time.Now().Unix(),
	}
	return proto.Marshal(resp)
}
