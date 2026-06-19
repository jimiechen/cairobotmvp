package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	socialpb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	"google.golang.org/protobuf/proto"
)

// TestSocialE2E_Smoke Phase 1 MVP-P0 最小冒烟测试（5 条核心协议）
// 验证 Gateway → LocalInvoker → Social Servant → Handler → SVC → MemoryRepository 全链路
func TestSocialE2E_Smoke(t *testing.T) {
	baseURL := "http://localhost:8080/api/hello"

	tests := []struct {
		name    string
		maxType int32
		minType int32
		method  string
		req     proto.Message
	}{
		{
			name:    "UserRegister(1021)",
			maxType: 1000,
			minType: 1021,
			method:  "HandleMember",
			req: &socialpb.UserRegisterRequest{
				Username: "e2e_user_001",
				Password: "TestPass123!",
				Email:    "e2e001@test.com",
			},
		},
		{
			name:    "UserLogin(1025)",
			maxType: 1000,
			minType: 1025,
			method:  "HandleMember",
			req: &socialpb.UserLoginRequest{
				Username: "e2e_user_001",
				Password: "TestPass123!",
			},
		},
		{
			name:    "CreateGroup(2005)",
			maxType: 2000,
			minType: 2005,
			method:  "HandleGroup",
			req: &socialpb.CreateGroupRequest{
				Name:        "E2E Test Group",
				Slug:        "e2e-test-group",
				Description: "E2E smoke test group",
				Category:    "test",
			},
		},
		{
			name:    "JoinGroup(2013)",
			maxType: 2000,
			minType: 2013,
			method:  "HandleGroup",
			req: &socialpb.JoinGroupRequest{
				GroupId: "grp_e2e_001",
			},
		},
		{
			name:    "CreateTopic(3001)",
			maxType: 3000,
			minType: 3001,
			method:  "HandleTopic",
			req: &socialpb.CreateTopicRequest{
				Title:   "E2E Test Topic",
				Content: "This is an E2E smoke test topic",
				GroupId: "grp_e2e_001",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			respCode, httpStatus, respBody := callGateway(baseURL, tc.maxType, tc.minType, tc.method, tc.req)

			t.Logf("HTTP=%d | Code=%s | Body=%s", httpStatus, respCode, respBody)

			// 核心断言：绝不能返回 10404（handler not found）
			if respCode == "10404" {
				t.Fatalf("❌ HANDLER_NOT_FOUND (10404) — 协议 %d 未注册到 LocalInvoker", tc.minType)
			}

			// HTTP 必须为 200
			if httpStatus != 200 {
				t.Fatalf("❌ HTTP 状态码期望 200, 实际 %d", httpStatus)
			}

			// 业务成功码 10200 或业务错误码 10400（参数校验等）均可接受
			if respCode == "10200" {
				t.Logf("✅ 业务成功")
			} else if respCode == "10400" {
				t.Logf("⚠️ 业务错误（可能为预期行为：参数校验、依赖数据不存在等）")
			} else {
				t.Logf("⚠️ 未知响应码: %s", respCode)
			}
		})
	}
}

func callGateway(baseURL string, maxType, minType int32, method string, req proto.Message) (respCode string, httpStatus int, respBody string) {
	// 序列化请求体
	reqData, _ := proto.Marshal(req)

	// 构造 TARS MessagePacket
	packet := &pb.MessagePacket{
		MaxType: maxType,
		MinType: minType,
		Extend: map[string]string{
			"method":  method,
			"minType": fmt.Sprintf("%d", minType),
			"traceId": fmt.Sprintf("e2e-%d-%d", maxType, minType),
		},
		Platform: pb.Platform_WEB,
		Data:      reqData,
	}

	packetData, _ := proto.Marshal(packet)

	// 发送 HTTP POST
	httpReq, _ := http.NewRequest("POST", baseURL, bytes.NewReader(packetData))
	httpReq.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "ERR_HTTP", 0, fmt.Sprintf("HTTP failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 解析响应
	respPacket := &pb.MessagePacket{}
	if len(body) > 0 {
		proto.Unmarshal(body, respPacket)
	}

	return respPacket.Extend["code"], resp.StatusCode, formatRespSummary(respPacket.Data)
}

func formatRespSummary(data []byte) string {
	if len(data) == 0 {
		return "(empty)"
	}
	// 尝试 JSON 格式化
	if len(data) > 2 && (data[0] == '{' || data[0] == '[') {
		var pretty json.RawMessage
		if json.Unmarshal(data, &pretty) == nil {
			s := string(pretty)
			if len(s) > 200 {
				return s[:200] + "..."
			}
			return s
		}
	}
	return fmt.Sprintf("%d bytes (raw)", len(data))
}

// TestMain 入口：确保 Gateway 正在运行
func TestMain(m *testing.M) {
	fmt.Println("=" + fmt_repeat("=", 50) + "=")
	fmt.Println("  Social E2E Smoke Test - Phase 1 MVP-P0")
	fmt.Printf("  时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("=" + fmt_repeat("=", 50) + "=")

	// 检查 Gateway 是否在运行
	resp, err := http.Get("http://localhost:8080/api/hello")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Gateway 未运行 (localhost:8080): %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	fmt.Printf("✅ Gateway 运行中 (HTTP %d)\n", resp.StatusCode)

	exitCode := m.Run()

	fmt.Println("\n" + fmt_repeat("=", 55))
	fmt.Println("  E2E Smoke Test Complete")
	fmt.Println(fmt_repeat("=", 55))
	os.Exit(exitCode)
}

func fmt_repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
