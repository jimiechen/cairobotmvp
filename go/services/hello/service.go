package hello

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// HelloService HelloWorld 模块接口
// 业务模块不依赖 MessagePacket，只接收 Protobuf bytes 并返回 Protobuf bytes
type HelloService interface {
	SayHello(ctx context.Context, request []byte) ([]byte, error)
}

// Service Hello 模块的具体实现
type Service struct{}

// NewService 创建 Hello Service 实例
func NewService() *Service {
	return &Service{}
}

// SayHello 执行问候操作
// ctx: 上下文（可携带 traceId、requestId 等链路信息）
// request: Protobuf 序列化的 HelloWorldRequest bytes
// 返回: Protobuf 序列化的 HelloWorldResponse bytes
func (s *Service) SayHello(ctx context.Context, request []byte) ([]byte, error) {
	var req pb.HelloWorldRequest

	if err := proto.Unmarshal(request, &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	name := req.Name
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
