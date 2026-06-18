package event

import (
	"encoding/json"
	"fmt"
	"time"
)

// DomainEvent 领域事件统一结构
// 所有社交域领域事件必须使用此结构，禁止各 svc 自行定义事件格式
type DomainEvent struct {
	// ID 事件唯一标识（char(32) UUID 风格）
	ID string `json:"id"`
	// Type 事件类型（使用 constants.go 中定义的 EventXxx 常量）
	Type string `json:"type"`
	// Version 事件 schema 版本（当前固定为 "1.0"）
	Version string `json:"version"`
	// AggregateType 聚合根类型（member/group/topic/order/comment）
	AggregateType string `json:"aggregate_type,omitempty"`
	// AggregateID 聚合根 ID（如 user_id / group_id / topic_id）
	AggregateID string `json:"aggregate_id,omitempty"`
	// ActorID 触发事件的用户 ID
	ActorID string `json:"actor_id,omitempty"`
	// TraceID 请求链路 ID（从 context 读取，可为空）
	TraceID string `json:"trace_id,omitempty"`
	// OccurredAt 事件发生时间（Unix 毫秒，与项目时间单位一致）
	OccurredAt int64 `json:"occurred_at"`
	// Payload 事件业务载荷（JSON 序列化的强类型 payload）
	// 使用 json.RawMessage 保证传输层统一为 JSON，同时保留强类型构造能力
	Payload json.RawMessage `json:"payload,omitempty"`
}

// NewEventOptions 构造 DomainEvent 的选项参数
type NewEventOptions struct {
	Type          string
	AggregateType string
	AggregateID   string
	ActorID       string
	Payload       any // 传入强类型 payload struct，内部序列化为 JSON
}

// NewDomainEvent 构造新的领域事件
// payload 参数传入强类型结构体（如 MemberRegisteredPayload），内部序列化为 JSON RawMessage
func NewDomainEvent(opt NewEventOptions) (DomainEvent, error) {
	var rawPayload json.RawMessage
	if opt.Payload != nil {
		data, err := json.Marshal(opt.Payload)
		if err != nil {
			return DomainEvent{}, fmt.Errorf("序列化事件 payload 失败: %w", err)
		}
		rawPayload = data
	}
	return DomainEvent{
		ID:            newEventID(),
		Type:          opt.Type,
		Version:       EventVersionCurrent,
		AggregateType: opt.AggregateType,
		AggregateID:   opt.AggregateID,
		ActorID:       opt.ActorID,
		TraceID:       "", // 由调用方从 context 填充
		OccurredAt:    time.Now().UnixMilli(),
		Payload:       rawPayload,
	}, nil
}

// MustNewDomainEvent 构造新事件，panic 当序列化失败
// 仅用于已知不会失败的场景（测试代码等）
func MustNewDomainEvent(opt NewEventOptions) DomainEvent {
	evt, err := NewDomainEvent(opt)
	if err != nil {
		panic(fmt.Sprintf("event: MustNewDomainEvent 失败: %v", err))
	}
	return evt
}

// newEventID 生成事件唯一 ID
// MVP-P0 使用时间戳+随机数格式，生产环境可替换为 ULID
func newEventID() string {
	return fmt.Sprintf("evt-%d", time.Now().UnixNano())
}
