package adapter

import (
	"bytes"
	"testing"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

func TestBuildErrorPacket(t *testing.T) {
	t.Run("with request", func(t *testing.T) {
		req := &MessagePacket{
			MaxType:  2100,
			MinType:  2097,
			Extend:   map[string]string{"traceId": "abc"},
			Platform: pb.Platform_ANDROID,
		}
		resp := BuildErrorPacket(req, 10404, "not found")
		if resp.MaxType != 2100 {
			t.Fatalf("expected MaxType 2100, got %d", resp.MaxType)
		}
		if resp.MinType != 2097 {
			t.Fatalf("expected MinType 2097, got %d", resp.MinType)
		}
		if resp.Extend["code"] != "10404" {
			t.Fatalf("expected code 10404, got %s", resp.Extend["code"])
		}
		if resp.Extend["message"] != "not found" {
			t.Fatalf("expected message 'not found', got %s", resp.Extend["message"])
		}
		if resp.Extend["traceId"] != "abc" {
			t.Fatalf("expected traceId 'abc', got %s", resp.Extend["traceId"])
		}
	})

	t.Run("without request", func(t *testing.T) {
		resp := BuildErrorPacket(nil, 10400, "bad request")
		if resp.MaxType != 0 {
			t.Fatalf("expected MaxType 0, got %d", resp.MaxType)
		}
		if resp.MinType != 0 {
			t.Fatalf("expected MinType 0, got %d", resp.MinType)
		}
		if resp.Extend["code"] != "10400" {
			t.Fatalf("expected code 10400, got %s", resp.Extend["code"])
		}
	})
}

func TestBuildResponsePacket(t *testing.T) {
	req := &MessagePacket{
		MaxType:  2100,
		MinType:  2097,
		Extend:   map[string]string{"traceId": "abc", "requestId": "def"},
		Platform: pb.Platform_WEB,
	}
	resp := BuildResponsePacket(req, 2100, 2098, []byte("ok"), 10200)
	if resp.MaxType != 2100 {
		t.Fatalf("expected MaxType 2100, got %d", resp.MaxType)
	}
	if resp.MinType != 2098 {
		t.Fatalf("expected MinType 2098, got %d", resp.MinType)
	}
	if string(resp.Data) != "ok" {
		t.Fatalf("expected data 'ok', got %s", string(resp.Data))
	}
	if resp.Extend["code"] != "10200" {
		t.Fatalf("expected code 10200, got %s", resp.Extend["code"])
	}
	if resp.Extend["traceId"] != "abc" {
		t.Fatalf("expected traceId 'abc', got %s", resp.Extend["traceId"])
	}
	if resp.Extend["requestId"] != "def" {
		t.Fatalf("expected requestId 'def', got %s", resp.Extend["requestId"])
	}
}

func TestEnsureTraceId(t *testing.T) {
	t.Run("generate new", func(t *testing.T) {
		extend := make(map[string]string)
		traceId := EnsureTraceId(extend)
		if traceId == "" {
			t.Fatal("expected non-empty traceId")
		}
		if extend["traceId"] != traceId {
			t.Fatalf("expected extend traceId %s, got %s", traceId, extend["traceId"])
		}
	})

	t.Run("keep existing", func(t *testing.T) {
		extend := map[string]string{"traceId": "existing"}
		traceId := EnsureTraceId(extend)
		if traceId != "existing" {
			t.Fatalf("expected traceId 'existing', got %s", traceId)
		}
	})
}

func TestEnsureRequestId(t *testing.T) {
	t.Run("generate new", func(t *testing.T) {
		extend := make(map[string]string)
		requestId := EnsureRequestId(extend)
		if requestId == "" {
			t.Fatal("expected non-empty requestId")
		}
	})

	t.Run("keep existing", func(t *testing.T) {
		extend := map[string]string{"requestId": "existing"}
		requestId := EnsureRequestId(extend)
		if requestId != "existing" {
			t.Fatalf("expected requestId 'existing', got %s", requestId)
		}
	})
}

func TestBuildTarsExtend(t *testing.T) {
	req := &MessagePacket{
		MaxType:  2100,
		MinType:  2097,
		Platform: pb.Platform_ANDROID,
		Extend:   map[string]string{"traceId": "abc", "requestId": "def"},
	}
	extend := BuildTarsExtend(req, "2100:2097", "proto.Req", "proto.Resp", true, false)
	if extend["traceId"] != "abc" {
		t.Fatalf("expected traceId 'abc', got %s", extend["traceId"])
	}
	if extend["requestId"] != "def" {
		t.Fatalf("expected requestId 'def', got %s", extend["requestId"])
	}
	if extend["maxType"] != "2100" {
		t.Fatalf("expected maxType '2100', got %s", extend["maxType"])
	}
	if extend["minType"] != "2097" {
		t.Fatalf("expected minType '2097', got %s", extend["minType"])
	}
	if extend["platform"] != "3" {
		t.Fatalf("expected platform '3', got %s", extend["platform"])
	}
	if extend["routeKey"] != "2100:2097" {
		t.Fatalf("expected routeKey '2100:2097', got %s", extend["routeKey"])
	}
	if extend["requestProto"] != "proto.Req" {
		t.Fatalf("expected requestProto 'proto.Req', got %s", extend["requestProto"])
	}
	if extend["responseProto"] != "proto.Resp" {
		t.Fatalf("expected responseProto 'proto.Resp', got %s", extend["responseProto"])
	}
	if extend["authRequired"] != "true" {
		t.Fatalf("expected authRequired 'true', got %s", extend["authRequired"])
	}
	if extend["auditRequired"] != "false" {
		t.Fatalf("expected auditRequired 'false', got %s", extend["auditRequired"])
	}
}

func TestSerializeDeserializeMessagePacket(t *testing.T) {
	t.Run("encode decode success", func(t *testing.T) {
		original := &MessagePacket{
			MaxType:  2100,
			MinType:  2097,
			Platform: pb.Platform_ANDROID,
			Extend:   map[string]string{"traceId": "abc", "requestId": "def"},
			Data:     []byte("hello world"),
		}

		encoded, err := SerializeMessagePacket(original)
		if err != nil {
			t.Fatalf("serialize failed: %v", err)
		}
		if len(encoded) == 0 {
			t.Fatal("expected non-empty encoded bytes")
		}

		decoded, err := DeserializeMessagePacket(encoded)
		if err != nil {
			t.Fatalf("deserialize failed: %v", err)
		}

		if decoded.MaxType != original.MaxType {
			t.Fatalf("expected MaxType %d, got %d", original.MaxType, decoded.MaxType)
		}
		if decoded.MinType != original.MinType {
			t.Fatalf("expected MinType %d, got %d", original.MinType, decoded.MinType)
		}
		if decoded.Platform != original.Platform {
			t.Fatalf("expected Platform %d, got %d", original.Platform, decoded.Platform)
		}
		if !bytes.Equal(decoded.Data, original.Data) {
			t.Fatalf("expected Data %v, got %v", original.Data, decoded.Data)
		}
		if decoded.Extend["traceId"] != "abc" {
			t.Fatalf("expected traceId 'abc', got %s", decoded.Extend["traceId"])
		}
		if decoded.Extend["requestId"] != "def" {
			t.Fatalf("expected requestId 'def', got %s", decoded.Extend["requestId"])
		}
	})

	t.Run("empty body", func(t *testing.T) {
		_, err := DeserializeMessagePacket([]byte{})
		if err == nil {
			t.Fatal("expected error for empty body")
		}
	})

	t.Run("invalid bytes", func(t *testing.T) {
		_, err := DeserializeMessagePacket([]byte{0xFF, 0xFF, 0xFF})
		if err == nil {
			t.Fatal("expected error for invalid bytes")
		}
	})

	t.Run("nil packet", func(t *testing.T) {
		_, err := SerializeMessagePacket(nil)
		if err == nil {
			t.Fatal("expected error for nil packet")
		}
	})
}
