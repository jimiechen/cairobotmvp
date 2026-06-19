package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	socialpb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	"google.golang.org/protobuf/proto"
)

func main() {
	baseURL := "http://localhost:8080/api/hello"
	fmt.Printf("=== UserLogin P0 诊断（含 UserRegister 对照）===\n")
	fmt.Printf("Target: %s | DisableKeepAlives: true\n\n", baseURL)

	// 对照: UserRegister (1021)
	testOne(baseURL, "UserRegister", 1000, 1021, "HandleMember",
		&socialpb.UserRegisterRequest{Username: "diag_user_001", Password: "DiagPass123!", Email: "diag@test.com"})

	fmt.Println()

	// 目标: UserLogin (1023) × 3 次
	loginPacket := buildPacket(1000, 1023, "HandleMember",
		&socialpb.UserLoginRequest{Username: "e2e_full_001", Password: "TestPass123!"})
	for i := 1; i <= 3; i++ {
		testOneRaw(baseURL, fmt.Sprintf("UserLogin#%d", i), loginPacket)
		time.Sleep(300 * time.Millisecond)
	}
}

func testOne(baseURL, name string, maxType, minType int32, method string, req proto.Message) {
	packetData := buildPacket(maxType, minType, method, req)
	testOneRaw(baseURL, name, packetData)
}

func buildPacket(maxType, minType int32, method string, req proto.Message) []byte {
	reqData, _ := proto.Marshal(req)
	packet := &pb.MessagePacket{
		MaxType:  maxType,
		MinType:  minType,
		Extend: map[string]string{
			"method":  method,
			"minType": fmt.Sprintf("%d", minType),
			"traceId": fmt.Sprintf("diag-%s-%d", method, time.Now().UnixNano()),
		},
		Platform: pb.Platform_WEB,
		Data:      reqData,
	}
	data, _ := proto.Marshal(packet)
	return data
}

func testOneRaw(baseURL, name string, packetData []byte) {
	fmt.Printf("--- %s (%d bytes) ---\n", name, len(packetData))
	start := time.Now()

	httpReq, err := http.NewRequest("POST", baseURL, bytes.NewReader(packetData))
	if err != nil {
		fmt.Printf("  FAIL NewRequest: %v\n", err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	resp, err := client.Do(httpReq)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("  FAIL | %v | ERROR: %v\n", elapsed, err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("  HTTP %d | %v | Body %d bytes\n", resp.StatusCode, elapsed, len(body))

	if len(body) == 0 {
		fmt.Printf("  (empty response body)\n")
		return
	}

	respPb := &pb.MessagePacket{}
	if uerr := proto.Unmarshal(body, respPb); uerr != nil {
		fmt.Printf("  PacketUnmarshalError: %v\n", uerr)
		return
	}

	tarsCode := respPb.Extend["code"]
	fmt.Printf("  TarsCode=%s | Data=%d bytes\n", tarsCode, len(respPb.Data))

	if len(respPb.Data) == 0 {
		return
	}

	// 尝试解析为 UserLoginResponse（仅对 1023 有效）
	var loginRsp socialpb.UserLoginResponse
	if perr := proto.Unmarshal(respPb.Data, &loginRsp); perr == nil && loginRsp.Result != nil {
		fmt.Printf("  BizCode=%d | Msg=%s\n", loginRsp.Result.Code, loginRsp.Result.Message)
		at := trunc(loginRsp.AccessToken, 20)
		rt := trunc(loginRsp.RefreshToken, 20)
		fmt.Printf("  Token: access=%s refresh=%s userId=%s\n", at, rt, loginRsp.UserId)
	} else {
		// 通用 Result 解析
		var regRsp socialpb.UserRegisterResponse
		if perr2 := proto.Unmarshal(respPb.Data, &regRsp); perr2 == nil && regRsp.Result != nil {
			fmt.Printf("  BizCode=%d | Msg=%s\n", regRsp.Result.Code, regRsp.Result.Message)
		} else {
			n := len(respPb.Data)
			if n > 32 { n = 32 }
			fmt.Printf("  RawBizData(hex): %x\n", respPb.Data[:n])
		}
	}
}

func trunc(s string, max int) string {
	if len(s) > max { return s[:max] + "..." }
	return s
}
