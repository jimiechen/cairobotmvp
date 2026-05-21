package health

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// HealthService HealthCheck 模块接口
// 业务模块不依赖 MessagePacket，只接收 Protobuf bytes 并返回 Protobuf bytes
type HealthService interface {
	Check(ctx context.Context, request []byte) ([]byte, error)
}

// Service Health 模块的具体实现
type Service struct{}

// NewService 创建 Health Service 实例
func NewService() *Service {
	return &Service{}
}

// Check 执行健康检查
// ctx: 上下文（可携带 traceId、requestId 等链路信息）
// request: Protobuf 序列化的 HealthCheckRequest bytes
// 返回: Protobuf 序列化的 HealthCheckResponse bytes
func (s *Service) Check(ctx context.Context, request []byte) ([]byte, error) {
	var req pb.ServiceHealthCheckRequest

	if err := proto.Unmarshal(request, &req); err != nil {
		return nil, err
	}

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
