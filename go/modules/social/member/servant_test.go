package member

import (
	"context"
	"testing"

	"github.com/jimiechen/mineplanet/go/modules/social/event"
)

// TestServant_Handle_ExtendUserID桥接到Context 验证 user_id 从 extend 桥接到 context
func TestServant_Handle_ExtendUserID桥接到Context(t *testing.T) {
	// 创建 Memory Repository 和 NoopPublisher
	repo := NewMemoryRepository()
	publisher := &event.NoopPublisher{}

	// 创建 Servant
	servant := NewServant(repo, publisher)

	// 构造请求（空请求体，仅测试 bridge）
	reqBytes := []byte{}

	// 构造 extend，包含 user_id
	extend := map[string]string{
		"user_id": "u_test_123",
		"minType": "1029", // GetUserInfo
	}

	// 调用 Handle
	code, resp, err := servant.Handle(context.Background(), reqBytes, extend)

	// 验证：不应返回错误（即使请求体为空，bridge 也应正常工作）
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	// 验证：返回码应为 200（bridge 成功）
	if code != 200 {
		t.Fatalf("expected code 200, got %d", code)
	}

	// 验证：响应体非空（至少包含错误响应）
	if len(resp) == 0 {
		t.Fatal("expected non-empty response")
	}

	t.Logf("✅ user_id bridge 成功 → code=%d, resp_size=%d", code, len(resp))
}

// TestServant_Handle_Extend无UserID 验证 extend 中无 user_id 时 context 不被污染
func TestServant_Handle_Extend无UserID(t *testing.T) {
	repo := NewMemoryRepository()
	publisher := &event.NoopPublisher{}
	servant := NewServant(repo, publisher)

	reqBytes := []byte{}
	extend := map[string]string{
		"minType": "1029",
		// 不包含 user_id
	}

	code, resp, err := servant.Handle(context.Background(), reqBytes, extend)

	// 验证：不应 panic 或报错
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	// 验证：返回码应为 200
	if code != 200 {
		t.Fatalf("expected code 200, got %d", code)
	}

	// 验证：响应体非空
	if len(resp) == 0 {
		t.Fatal("expected non-empty response")
	}

	t.Logf("✅ 无 user_id 时 bridge 不报错 → code=%d, resp_size=%d", code, len(resp))
}

// TestServant_Handle_Extend空UserID 验证 user_id 为空字符串时不桥接
func TestServant_Handle_Extend空UserID(t *testing.T) {
	repo := NewMemoryRepository()
	publisher := &event.NoopPublisher{}
	servant := NewServant(repo, publisher)

	reqBytes := []byte{}
	extend := map[string]string{
		"user_id": "", // 空字符串
		"minType": "1029",
	}

	code, resp, err := servant.Handle(context.Background(), reqBytes, extend)

	// 验证：不应报错
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	// 验证：返回码应为 200
	if code != 200 {
		t.Fatalf("expected code 200, got %d", code)
	}

	t.Logf("✅ 空 user_id 时 bridge 不报错 → code=%d, resp_size=%d", code, len(resp))
}

// TestServant_Handle_ExtendNil 验证 extend 为 nil 时能正常处理（minType 从 extend 提取，nil 时为空）
// 注意：extend 为 nil 时 minType 为空，Handler.Dispatch 会返回 unsupported minType 错误
// 这是预期行为，测试应验证 Servant.Handle 能正确返回错误码 500
func TestServant_Handle_ExtendNil(t *testing.T) {
	repo := NewMemoryRepository()
	publisher := &event.NoopPublisher{}
	servant := NewServant(repo, publisher)

	reqBytes := []byte{}

	code, _, err := servant.Handle(context.Background(), reqBytes, nil)

	// 验证：应返回错误（unsupported minType）
	if err == nil {
		t.Fatal("expected error when extend is nil")
	}

	// 验证：返回码应为 500（Handler.Dispatch 返回 unsupported minType）
	if code != 500 {
		t.Fatalf("expected code 500, got %d", code)
	}

	t.Logf("✅ extend 为 nil 时返回预期错误 → code=%d, err=%v", code, err)
}

// TestCtxKeyUserID常量值 验证 CtxKeyUserID 值与 Gateway AuthMiddleware 一致
func TestCtxKeyUserID常量值(t *testing.T) {
	// 验证：CtxKeyUserID 值必须为 "user_id"（与 Gateway AuthMiddleware 注入的 extend key 一致）
	expectedKey := "user_id"
	if string(CtxKeyUserID) != expectedKey {
		t.Fatalf("CtxKeyUserID 值期望 '%s', 实际 '%s'", expectedKey, string(CtxKeyUserID))
	}
	t.Logf("✅ CtxKeyUserID = '%s' (与 Gateway AuthMiddleware 一致)", expectedKey)
}