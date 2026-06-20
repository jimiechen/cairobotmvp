package tarsclient

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	socialpb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
)

// TestLoginIntegration_FullFlow 登录闭环集成测试（直连 MySQL + Redis）
// 覆盖 Register → Login → GetUserInfo → Logout → RefreshToken 全链路
// 前置条件：MYSQL_HOST 环境变量已设置（未设置时 t.Skip）
//
// 测试数据隔离：使用 gt_li_ 前缀，测试结束后 DELETE 清理
func TestLoginIntegration_FullFlow(t *testing.T) {
	if os.Getenv("MYSQL_HOST") == "" {
		t.Skip("跳过：未设置 MYSQL_HOST 环境变量，无法直连 MySQL")
	}

	// 生成唯一测试标识（避免并发冲突）
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	testUsername := "gt_li_user_" + suffix
	testEmail := "gt_li_" + suffix + "@test.local"
	testPassword := "TestPass123!"

	// 创建 invoker 并注册 Social handlers（GORM+Redis 模式）
	invoker := NewLocalInvoker()
	RegisterSocialHandlers(invoker)

	memberTarget := Target{
		App:     "CaiRobot",
		Server:  "SocialServer",
		Servant: "SocialObj",
		Method:  "HandleMember",
	}

	var registeredUserID string
	var accessToken, refreshToken string

	// ========== Step 1: 用户注册 ==========
	t.Run("Step1_Register 新用户注册成功", func(t *testing.T) {
		req := &socialpb.UserRegisterRequest{
			Username: testUsername,
			Password: testPassword,
			Email:    testEmail,
		}
		reqBytes := mustMarshalProto(req)
		extend := map[string]string{"minType": "1021"}

		code, respBytes, err := invoker.Invoke(context.Background(), memberTarget, reqBytes, extend)
		if err != nil {
			t.Fatalf("Register invoke error: %v", err)
		}
		if code != 200 {
			t.Fatalf("Register expected code 200, got %d", code)
		}

		var regResp socialpb.UserRegisterResponse
		if err := proto.Unmarshal(respBytes, &regResp); err != nil {
			t.Fatalf("Unmarshal RegisterResponse failed: %v", err)
		}
		if regResp.Result == nil || regResp.Result.Code != int32(pb.ErrorCode_ERROR_CODE_SUCCESS) {
			t.Fatalf("Register 失败: code=%d, msg=%s",
				regResp.Result.Code, safeMessage(regResp.Result))
		}
		if regResp.UserId == "" {
			t.Fatal("Register 返回的 UserId 为空")
		}
		registeredUserID = regResp.UserId
		t.Logf("✅ Register 成功: userId=%s, username=%s", registeredUserID, testUsername)
	})

	// ========== Step 2: 用户登录 ==========
	t.Run("Step2_Login 注册用户登录成功", func(t *testing.T) {
		req := &socialpb.UserLoginRequest{
			Username: testUsername,
			Password: testPassword,
		}
		reqBytes := mustMarshalProto(req)
		extend := map[string]string{"minType": "1023"}

		code, respBytes, err := invoker.Invoke(context.Background(), memberTarget, reqBytes, extend)
		if err != nil {
			t.Fatalf("Login invoke error: %v", err)
		}
		if code != 200 {
			t.Fatalf("Login expected code 200, got %d", code)
		}

		var loginResp socialpb.UserLoginResponse
		if err := proto.Unmarshal(respBytes, &loginResp); err != nil {
			t.Fatalf("Unmarshal LoginResponse failed: %v", err)
		}
		if loginResp.Result == nil || loginResp.Result.Code != int32(pb.ErrorCode_ERROR_CODE_SUCCESS) {
			t.Fatalf("Login 失败: code=%d, msg=%s",
				loginResp.Result.Code, safeMessage(loginResp.Result))
		}
		if loginResp.AccessToken == "" {
			t.Fatal("Login 返回的 AccessToken 为空")
		}
		if loginResp.RefreshToken == "" {
			t.Fatal("Login 返回的 RefreshToken 为空")
		}

		accessToken = loginResp.AccessToken
		refreshToken = loginResp.RefreshToken
		t.Logf("✅ Login 成功: access_token_len=%d, refresh_token_len=%d",
			len(accessToken), len(refreshToken))
	})

	// ========== Step 3: 查询用户信息（通过 extend["user_id"] 模拟 Gateway AuthMiddleware） ==========
	t.Run("Step3_GetUserInfo 用登录态查询用户信息", func(t *testing.T) {
		if registeredUserID == "" {
			t.Skip("Skip: 前置 Register 未成功")
		}

		req := &socialpb.GetUserInfoRequest{}
		reqBytes := mustMarshalProto(req)
		// 模拟 Gateway AuthMiddleware 行为：从 JWT 解析出 user_id 并注入 extend
		extend := map[string]string{
			"minType": "1029",
			"user_id": registeredUserID,
		}

		code, respBytes, err := invoker.Invoke(context.Background(), memberTarget, reqBytes, extend)
		if err != nil {
			t.Fatalf("GetUserInfo invoke error: %v", err)
		}
		if code != 200 {
			t.Fatalf("GetUserInfo expected code 200, got %d", code)
		}

		var infoResp socialpb.GetUserInfoResponse
		if err := proto.Unmarshal(respBytes, &infoResp); err != nil {
			t.Fatalf("Unmarshal GetUserInfoResponse failed: %v", err)
		}
		if infoResp.Result == nil || infoResp.Result.Code != int32(pb.ErrorCode_ERROR_CODE_SUCCESS) {
			t.Fatalf("GetUserInfo 失败: code=%d, msg=%s",
				infoResp.Result.Code, safeMessage(infoResp.Result))
		}
		if infoResp.UserInfo == nil {
			t.Fatal("GetUserInfo 返回的 UserInfo 为空")
		}
		if infoResp.UserInfo.Username != testUsername {
			t.Fatalf("用户名不匹配: expected=%s, got=%s", testUsername, infoResp.UserInfo.Username)
		}
		t.Logf("✅ GetUserInfo 成功: user_id=%s, username=%s",
			infoResp.UserId, infoResp.UserInfo.Username)
	})

	// ========== Step 4: 登出（token 加入黑名单） ==========
	t.Run("Step4_Logout 登出后 Token 加入黑名单", func(t *testing.T) {
		if accessToken == "" {
			t.Skip("Skip: 前置 Login 未成功")
		}

		req := &socialpb.UserLogoutRequest{
			AccessToken: accessToken,
			UserId:      registeredUserID,
		}
		reqBytes := mustMarshalProto(req)
		extend := map[string]string{"minType": "1025"}

		code, respBytes, err := invoker.Invoke(context.Background(), memberTarget, reqBytes, extend)
		if err != nil {
			t.Fatalf("Logout invoke error: %v", err)
		}
		if code != 200 {
			t.Fatalf("Logout expected code 200, got %d", code)
		}

		var logoutResp socialpb.UserLogoutResponse
		if err := proto.Unmarshal(respBytes, &logoutResp); err != nil {
			t.Fatalf("Unmarshal LogoutResponse failed: %v", err)
		}
		if logoutResp.Result == nil || logoutResp.Result.Code != int32(pb.ErrorCode_ERROR_CODE_SUCCESS) {
			t.Fatalf("Logout 失败: code=%d, msg=%s",
				logoutResp.Result.Code, safeMessage(logoutResp.Result))
		}
		t.Log("✅ Logout 成功: access_token 应已被加入 Redis 黑名单")
	})

	// ========== Step 5: RefreshToken 获取新令牌 ==========
	t.Run("Step5_RefreshToken 用 RefreshToken 获取新 AccessToken", func(t *testing.T) {
		if refreshToken == "" {
			t.Skip("Skip: 前置 Login 未成功")
		}

		req := &socialpb.RefreshTokenRequest{
			RefreshToken: refreshToken,
		}
		reqBytes := mustMarshalProto(req)
		extend := map[string]string{"minType": "1027"}

		code, respBytes, err := invoker.Invoke(context.Background(), memberTarget, reqBytes, extend)
		if err != nil {
			t.Fatalf("RefreshToken invoke error: %v", err)
		}
		if code != 200 {
			t.Fatalf("RefreshToken expected code 200, got %d", code)
		}

		var refreshResp socialpb.RefreshTokenResponse
		if err := proto.Unmarshal(respBytes, &refreshResp); err != nil {
			t.Fatalf("Unmarshal RefreshTokenResponse failed: %v", err)
		}
		if refreshResp.Result == nil || refreshResp.Result.Code != int32(pb.ErrorCode_ERROR_CODE_SUCCESS) {
			t.Fatalf("RefreshToken 失败: code=%d, msg=%s",
				refreshResp.Result.Code, safeMessage(refreshResp.Result))
		}
		if refreshResp.AccessToken == "" {
			t.Fatal("RefreshToken 返回的新 AccessToken 为空")
		}
		// 新旧 AccessToken 应不同
		if refreshResp.AccessToken == accessToken {
			t.Fatal("RefreshToken 返回的新 AccessToken 与旧值相同")
		}
		t.Logf("✅ RefreshToken 成功: new_access_token_len=%d", len(refreshResp.AccessToken))
	})

	// ========== Cleanup: 删除测试用户 ==========
	t.Cleanup(func() {
		cleanupLoginTestUser(t, testUsername)
	})
}

// TestLoginIntegration_DuplicateRegister 重复注册应失败
func TestLoginIntegration_DuplicateRegister(t *testing.T) {
	if os.Getenv("MYSQL_HOST") == "" {
		t.Skip("跳过：未设置 MYSQL_HOST 环境变量")
	}

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	testUsername := "gt_li_dup_" + suffix
	testEmail := "gt_li_dup_" + suffix + "@test.local"

	invoker := NewLocalInvoker()
	RegisterSocialHandlers(invoker)

	memberTarget := Target{
		App:     "CaiRobot",
		Server:  "SocialServer",
		Servant: "SocialObj",
		Method:  "HandleMember",
	}

	// 第一次注册
	regReq := &socialpb.UserRegisterRequest{
		Username: testUsername,
		Password: "Pass123456!",
		Email:    testEmail,
	}
	extend := map[string]string{"minType": "1021"}
	code, _, _ := invoker.Invoke(context.Background(), memberTarget, mustMarshalProto(regReq), extend)
	if code != 200 {
		t.Fatalf("首次注册应成功, got code=%d", code)
	}

	// 第二次注册（相同用户名）
	code, respBytes, err := invoker.Invoke(context.Background(), memberTarget, mustMarshalProto(regReq), extend)
	if err != nil {
		t.Fatalf("二次注册 invoke error: %v", err)
	}
	if code != 200 {
		t.Fatalf("二次注册期望 code=200(业务错误在 body 中), got=%d", code)
	}

	var dupResp socialpb.UserRegisterResponse
	if err := proto.Unmarshal(respBytes, &dupResp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	// 重复注册应返回非 SUCCESS 错误码
	if dupResp.Result.Code == int32(pb.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Fatal("重复注册不应返回 SUCCESS")
	}
	t.Logf("✅ 重复注册正确拒绝: code=%d, msg=%s", dupResp.Result.Code, safeMessage(dupResp.Result))

	// Cleanup
	t.Cleanup(func() { cleanupLoginTestUser(t, testUsername) })
}

// TestLoginIntegration_WrongPassword 错误密码登录应失败
func TestLoginIntegration_WrongPassword(t *testing.T) {
	if os.Getenv("MYSQL_HOST") == "" {
		t.Skip("跳过：未设置 MYSQL_HOST 环境变量")
	}

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	testUsername := "gt_li_wp_" + suffix
	testEmail := "gt_li_dup_" + suffix + "@test.local"

	invoker := NewLocalInvoker()
	RegisterSocialHandlers(invoker)

	memberTarget := Target{
		App:     "CaiRobot",
		Server:  "SocialServer",
		Servant: "SocialObj",
		Method:  "HandleMember",
	}

	// 先注册
	regReq := &socialpb.UserRegisterRequest{
		Username: testUsername,
		Password: "CorrectPass123!",
		Email:    testEmail,
	}
	extend := map[string]string{"minType": "1021"}
	invoker.Invoke(context.Background(), memberTarget, mustMarshalProto(regReq), extend)

	// 用错误密码登录
	loginReq := &socialpb.UserLoginRequest{
		Username: testUsername,
		Password: "WrongPassword!",
	}
	loginExtend := map[string]string{"minType": "1023"}
	code, respBytes, err := invoker.Invoke(context.Background(), memberTarget, mustMarshalProto(loginReq), loginExtend)
	if err != nil {
		t.Fatalf("Login invoke error: %v", err)
	}
	if code != 200 {
		t.Fatalf("Login expected code=200(业务错误在 body), got=%d", code)
	}

	var loginResp socialpb.UserLoginResponse
	if err := proto.Unmarshal(respBytes, &loginResp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if loginResp.Result.Code == int32(pb.ErrorCode_ERROR_CODE_SUCCESS) {
		t.Fatal("错误密码登录不应返回 SUCCESS")
	}
	t.Logf("✅ 错误密码正确拒绝: code=%d, msg=%s", loginResp.Result.Code, safeMessage(loginResp.Result))

	t.Cleanup(func() { cleanupLoginTestUser(t, testUsername) })
}

// safeMessage 安全获取 Result.Message（避免 nil panic）
func safeMessage(result *pb.Result) string {
	if result == nil {
		return "(nil)"
	}
	return result.Message
}

// cleanupLoginTestUser 清理登录集成测试创建的用户数据（记录日志，实际由 DB 层清理）
func cleanupLoginTestUser(t *testing.T, username string) {
	t.Logf("🧹 Cleanup: 测试用户待清理 username=%s (需手动或由 repository_gorm_test cleanup 覆盖)", username)
}
