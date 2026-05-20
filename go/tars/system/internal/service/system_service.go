package service

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// SystemService 系统服务业务逻辑
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
