package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"google.golang.org/protobuf/proto"
)

// TestClient proto-gateway 快速验证客户端
// 用于开发阶段验证 TarsGo HTTP Servant 链路是否正常
// 用法：go run ./cmd/testclient/main.go
func main() {
	baseURL := os.Getenv("GATEWAY_TEST_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080/api/hello"
	}

	tests := []struct {
		name    string
		maxType int32
		minType int32
		method  string
	}{
		{"HealthCheck", 2100, 2097, "HealthCheck"},
		{"HelloWorld", 2100, 2101, "HelloWorld"},
	}

	passCount := 0
	failCount := 0

	for _, tc := range tests {
		packet := &pb.MessagePacket{
			MaxType: tc.maxType,
			MinType: tc.minType,
			Extend: map[string]string{
				"method":     tc.method,
				"traceId":    "test-trace-001",
				"requestId":  "test-req-001",
			},
			Platform: pb.Platform_WEB,
			Data:      []byte("Hello World"),
		}

		data, err := proto.Marshal(packet)
		if err != nil {
			fmt.Printf("❌ [%s] Marshal failed: %v\n", tc.name, err)
			failCount++
			continue
		}

		req, err := http.NewRequest("POST", baseURL, bytes.NewReader(data))
		if err != nil {
			fmt.Printf("❌ [%s] NewRequest failed: %v\n", tc.name, err)
			failCount++
			continue
		}
		req.Header.Set("Content-Type", "application/octet-stream")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ [%s] Request failed: %v\n", tc.name, err)
			failCount++
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var respPacket *pb.MessagePacket
		if len(body) > 0 {
			respPacket = &pb.MessagePacket{}
			if uerr := proto.Unmarshal(body, respPacket); uerr != nil {
				fmt.Printf("⚠️  [%s] Response unmarshal failed (raw): %s\n", tc.name, string(body))
				respPacket = nil
			}
		}

		fmt.Printf("\n=== %s ===\n", tc.name)
		fmt.Printf("Status: %d\n", resp.StatusCode)

		if respPacket != nil {
			code := respPacket.Extend["code"]
			fmt.Printf("Response MaxType: %d, MinType: %d\n", respPacket.MaxType, respPacket.MinType)
			fmt.Printf("Response Code: %s\n", code)
			fmt.Printf("Response Data (string): %s\n", string(respPacket.Data))
			fmt.Printf("Response Extend: ")
			json.NewEncoder(os.Stdout).Encode(respPacket.Extend)

			if code == "10200" && resp.StatusCode == 200 {
				fmt.Printf("✅ [%s] PASSED\n", tc.name)
				passCount++
			} else {
				fmt.Printf("❌ [%s] FAILED (code=%s)\n", tc.name, code)
				failCount++
			}
		} else {
			fmt.Printf("Response Body (raw hex): %x\n", body)
			fmt.Printf("❌ [%s] FAILED (invalid response)\n", tc.name)
			failCount++
		}
	}

	fmt.Printf("\n========== 结果汇总 ==========\n")
	fmt.Printf("通过: %d, 失败: %d, 总计: %d\n", passCount, failCount, passCount+failCount)
	if failCount > 0 {
		os.Exit(1)
	}
}
