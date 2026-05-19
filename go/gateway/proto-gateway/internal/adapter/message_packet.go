package adapter

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// MessagePacket 是 proto/base/message.proto 的 Go 结构体别名
// 使用 protoc 生成的真实类型，不再使用临时结构体
type MessagePacket = pb.MessagePacket

// Platform 枚举值别名
const (
	PlatformUnknown = pb.Platform_UNKNOWN
	PlatformWeb     = pb.Platform_WEB
	PlatformPC      = pb.Platform_PC
	PlatformAndroid = pb.Platform_ANDROID
	PlatformIOS     = pb.Platform_IOS
	PlatformOther   = pb.Platform_OTHER
)

// BuildErrorPacket 构造错误响应 MessagePacket
func BuildErrorPacket(req *MessagePacket, code int32, message string) *MessagePacket {
	resp := &MessagePacket{
		MaxType: 0,
		MinType: 0,
		Extend:  make(map[string]string),
		Data:    []byte{},
	}
	if req != nil {
		resp.MaxType = req.MaxType
		resp.MinType = req.MinType
		if req.Extend != nil {
			for k, v := range req.Extend {
				resp.Extend[k] = v
			}
		}
	}
	resp.Extend["code"] = fmt.Sprintf("%d", code)
	resp.Extend["message"] = message
	return resp
}

// BuildResponsePacket 构造成功响应 MessagePacket
func BuildResponsePacket(req *MessagePacket, responseMax int32, responseMin int32, data []byte, code int32) *MessagePacket {
	resp := &MessagePacket{
		MaxType: responseMax,
		MinType: responseMin,
		Extend:  make(map[string]string),
		Data:    data,
	}
	if req != nil && req.Extend != nil {
		for k, v := range req.Extend {
			resp.Extend[k] = v
		}
	}
	resp.Extend["code"] = fmt.Sprintf("%d", code)
	return resp
}

// EnsureTraceId 确保 extend 中有 traceId
func EnsureTraceId(extend map[string]string) string {
	if extend == nil {
		extend = make(map[string]string)
	}
	if extend["traceId"] == "" {
		extend["traceId"] = uuid.New().String()
	}
	return extend["traceId"]
}

// EnsureRequestId 确保 extend 中有 requestId
func EnsureRequestId(extend map[string]string) string {
	if extend == nil {
		extend = make(map[string]string)
	}
	if extend["requestId"] == "" {
		extend["requestId"] = uuid.New().String()
	}
	return extend["requestId"]
}

// BuildTarsExtend 构造 Tars 调用所需的 extend map
func BuildTarsExtend(req *MessagePacket, routeKey string, requestProto string, responseProto string, authRequired bool, auditRequired bool) map[string]string {
	extend := make(map[string]string)
	if req.Extend != nil {
		for k, v := range req.Extend {
			extend[k] = v
		}
	}

	EnsureTraceId(extend)
	EnsureRequestId(extend)

	extend["maxType"] = fmt.Sprintf("%d", req.MaxType)
	extend["minType"] = fmt.Sprintf("%d", req.MinType)
	extend["platform"] = fmt.Sprintf("%d", req.Platform)
	extend["routeKey"] = routeKey
	extend["requestProto"] = requestProto
	extend["responseProto"] = responseProto
	extend["authRequired"] = fmt.Sprintf("%t", authRequired)
	extend["auditRequired"] = fmt.Sprintf("%t", auditRequired)

	return extend
}

// SerializeMessagePacket 序列化 MessagePacket 为 bytes
// 使用 protobuf Marshal
func SerializeMessagePacket(packet *MessagePacket) ([]byte, error) {
	if packet == nil {
		return nil, fmt.Errorf("packet is nil")
	}
	return proto.Marshal(packet)
}

// DeserializeMessagePacket 从 bytes 反序列化 MessagePacket
// 使用 protobuf Unmarshal
func DeserializeMessagePacket(data []byte) (*MessagePacket, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty body")
	}
	packet := &MessagePacket{}
	if err := proto.Unmarshal(data, packet); err != nil {
		return nil, fmt.Errorf("unmarshal message packet failed: %w", err)
	}
	return packet, nil
}
