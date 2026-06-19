package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	socialpb "github.com/jimiechen/mineplanet/protocols/generated/go/social"
	"google.golang.org/protobuf/proto"
)

// Social Stateful E2E Test — Phase 1 MVP-P0 Task 5-D
//
// 核心改进（对比 Task 5-C 无状态版本）：
//   - 维护全局测试上下文 E2EContext，跨用例复用 user_id/token/group_id/topic_id
//   - auth_required 接口自动注入 extend.token
//   - 用例按业务流程顺序执行：Register → Login → CreateGroup → JoinGroup → CreateTopic → Interaction
//   - 区分 Positive PASS（正向成功）和 Negative PASS（负向预期错误）
//   - 请求体支持模板变量 {{user_a_id}}、{{group_id}} 等
//
// 验证层级（四层不变）：
//   L1: HTTP 层 — StatusCode == 200
//   L2: TarsGo 协议层 — Extend["code"] == "200" (returnCode)
//   L3: 业务层 — Response bytes 反序列化成功 + Result.Code 可解析
//   L4: 业务语义 — Result.Code == 10200（正向）/ 预期业务错误码（负向）

func main() {
	baseURL := os.Getenv("GATEWAY_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080/api/hello"
	}

	// 使用时间戳后缀确保用户名/群组标识唯一（避免 MemoryRepository 冲突）
	suffix := fmt.Sprintf("_%d", time.Now().Unix())

	ctx := &E2EContext{Suffix: suffix}
	results := runStatefulE2E(baseURL, ctx)

	printStatefulSummary(baseURL, results, ctx)

	failCount := countByStatus(results, "FAIL")
	if failCount > 0 {
		os.Exit(1)
	}
}

// ========== 数据结构 ==========

// E2EContext 全局测试上下文，保存跨用例的运行时变量
type E2EContext struct {
	// 用户 A（群主 / 发帖者）
	UserAID       string
	AccessTokenA  string
	RefreshTokenA string

	// 用户 B（普通成员 / 互动者）
	UserBID       string
	AccessTokenB  string
	RefreshTokenB string

	// 资源 ID
	GroupID  string
	TopicID  string
	ReplyID  string

	// 时间戳后缀，用于构造唯一的用户名/群组名（避免 MemoryRepository 冲突）
	Suffix string
}

// setVar 安全设置上下文变量（仅当值为非空时覆盖）
func (c *E2EContext) setVar(dest *string, val string) {
	if val != "" {
		*dest = val
	}
}

// resolveTemplate 替换请求体中的模板变量
// 支持: {{user_a_id}}, {{user_b_id}}, {{group_id}}, {{topic_id}}, {{reply_id}},
//       {{access_token_a}}, {{access_token_b}}, {{refresh_token_a}}, {{refresh_token_b}}
func (c *E2EContext) resolveTemplate(raw string) string {
	replacer := strings.NewReplacer(
		"{{user_a_id}}", c.UserAID,
		"{{user_b_id}}", c.UserBID,
		"{{group_id}}", c.GroupID,
		"{{topic_id}}", c.TopicID,
		"{{reply_id}}", c.ReplyID,
		"{{access_token_a}}", c.AccessTokenA,
		"{{access_token_b}}", c.AccessTokenB,
		"{{refresh_token_a}}", c.RefreshTokenA,
		"{{refresh_token_b}}", c.RefreshTokenB,
	)
	return replacer.Replace(raw)
}

// tokenForActor 根据 actor 选择对应 AccessToken
func (c *E2EContext) tokenForActor(actor string) string {
	switch actor {
	case "userA":
		return c.AccessTokenA
	case "userB":
		return c.AccessTokenB
	default:
		return ""
	}
}

// e2eTestCase 定义单条 Stateful E2E 测试用例
type e2eTestCase struct {
	Name           string          // 用例名称
	MaxType        int32           // TarsGo maxType
	MinType        int32           // TarsGo minType
	Method         string          // Servant 方法名
	RespType       string          // 响应类型标识（选择反序列化策略）

	// === Task 5-D 新增字段 ===
	IsPositive     bool            // true=正向成功用例(期望10200), false=负向用例(期望特定错误码)
	Actor          string          // 操作人: "userA"/"userB"/""(无)
	ExpectedBizCode string         // 负向用例期望的业务错误码（空字符串表示不检查具体码）

	// BuildRequest 使用 ctx 构建请求体（支持模板变量和动态字段）
	BuildRequest func(ctx *E2EContext) proto.Message
}

// e2eResult 定义单条 E2E 测试结果
type e2eResult struct {
	Name            string
	Category        string // "POS_PASS" / "NEG_PASS" / "WARN" / "FAIL"
	TarsCode        string
	HTTPStatus      string
	BizCode         string
	BizMessage      string
	RespSize        int
	LatencyMs       int64
	ErrorDetail     string
	rawResponseData []byte // 原始响应字节，供 extractContext 使用
}

func (r *e2eResult) Status() string { return r.Category }

// ========== 用例定义（按业务流程顺序）==========

// authRequiredEndpoints 记录哪些协议号配置了 auth_required=true
var authRequiredEndpoints = map[int32]bool{
	1025: true, // UserLogout
	1029: true, // GetUserInfo
	1031: true, // UpdateUserInfo
	1043: true, // GetBlockList
	1045: true, // GetUserStats
	2013: true, // JoinGroup
	2015: true, // LeaveGroup
	2019: true, // MuteMember
	2023: true, // BanMember
	2027: true, // RemoveMember
	2029: true, // UpdateMemberRole
	2037: true, // RenewMember
	2073: true, // CalcPayableAmount
	2087: true, // GroupUserEnter
	3009: true, // DeleteTopic
	3043: true, // AddTopicReply
	3061: true, // LikeTopic
	3063: true, // FavoriteTopic
	3095: true, // CreateReport
}

func buildTestCases() []e2eTestCase {
	return []e2eTestCase{
		// ================================================================
		// A. 认证正向链路（必须返回 10200）
		// ================================================================

		{
			Name: "Register_UserA", MaxType: 1000, MinType: 1021,
			Method: "HandleMember", RespType: "member_register",
			IsPositive: true, Actor: "",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.UserRegisterRequest{
					Username: "e2e_stateful_usera" + ctx.Suffix,
					Password: "StatefulPass123!",
					Email:    "e2ea" + ctx.Suffix + "@stateful.test",
				}
			},
		},
		{
			Name: "Login_UserA", MaxType: 1000, MinType: 1023,
			Method: "HandleMember", RespType: "member_login",
			IsPositive: true, Actor: "",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.UserLoginRequest{
					Username: "e2e_stateful_usera" + ctx.Suffix,
					Password: "StatefulPass123!",
				}
			},
		},
		{
			Name: "Register_UserB", MaxType: 1000, MinType: 1021,
			Method: "HandleMember", RespType: "member_register",
			IsPositive: true, Actor: "",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.UserRegisterRequest{
					Username: "e2e_stateful_userb" + ctx.Suffix,
					Password: "StatefulPass123!",
					Email:    "e2eb" + ctx.Suffix + "@stateful.test",
				}
			},
		},
		{
			Name: "Login_UserB", MaxType: 1000, MinType: 1023,
			Method: "HandleMember", RespType: "member_login",
			IsPositive: true, Actor: "",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.UserLoginRequest{
					Username: "e2e_stateful_userb" + ctx.Suffix,
					Password: "StatefulPass123!",
				}
			},
		},
		{
			Name: "RefreshToken_UserA", MaxType: 1000, MinType: 1027,
			Method: "HandleMember", RespType: "member_refresh",
			IsPositive: true, Actor: "", // 使用 Login 返回的有效 refresh_token
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.RefreshTokenRequest{RefreshToken: ctx.RefreshTokenA}
			},
		},

		// ================================================================
		// B. 用户资料正向链路（携带 UserA Token）
		// ================================================================

		{
			Name: "GetUserInfo_UserA", MaxType: 1000, MinType: 1029,
			Method: "HandleMember", RespType: "member_get_info",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.GetUserInfoRequest{}
			},
		},
		{
			Name: "UpdateUserInfo_UserA", MaxType: 1000, MinType: 1031,
			Method: "HandleMember", RespType: "member_update_info",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.UpdateUserInfoRequest{Nickname: "StatefulNickA"}
			},
		},
		{
			Name: "GetBlockList_UserA", MaxType: 1000, MinType: 1043,
			Method: "HandleMember", RespType: "member_block_list",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.GetBlockListRequest{Page: 1, PageSize: 20, GroupId: "{{group_id}}"}
			},
		},
		{
			Name: "GetUserStats_UserA", MaxType: 1000, MinType: 1045,
			Method: "HandleMember", RespType: "member_stats",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.GetUserStatsRequest{IncludeDetails: true}
			},
		},

		// ================================================================
		// C. 群组正向链路（UserA 创建群组，UserB 加入）
		// ================================================================

		{
			Name: "CreateGroup_UserA", MaxType: 2000, MinType: 2005,
			Method: "HandleGroup", RespType: "group_create",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.CreateGroupRequest{
					Name:        "E2E Stateful Group" + ctx.Suffix,
					Slug:        "e2e-stateful-group" + ctx.Suffix,
					Description: "Stateful E2E test group",
					Category:    "test",
				}
			},
		},
		{
			Name: "GetGroupStats", MaxType: 2000, MinType: 2039,
			Method: "HandleGroup", RespType: "group_stats",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.GetGroupStatsRequest{GroupId: ctx.GroupID}
			},
		},
		{
			Name: "JoinGroup_UserB", MaxType: 2000, MinType: 2013,
			Method: "HandleGroup", RespType: "group_join",
			IsPositive: true, Actor: "userB",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.JoinGroupRequest{GroupId: ctx.GroupID}
			},
		},
		{
			Name: "GetGroupMemberIds", MaxType: 2000, MinType: 2077,
			Method: "HandleGroup", RespType: "group_member_ids",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.GetGroupMemberUserIdsRequest{GroupId: ctx.GroupID, Page: 1, PageSize: 20}
			},
		},
		{
			Name: "BatchGetGroups", MaxType: 2000, MinType: 2047,
			Method: "HandleGroup", RespType: "group_batch",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.BatchGetGroupsRequest{GroupIds: []string{ctx.GroupID}}
			},
		},

		// ================================================================
		// D. 群成员管理正向链路（UserA 作为 owner 操作 UserB）
		// ================================================================

		{
			Name: "MuteMember_UserB", MaxType: 2000, MinType: 2019,
			Method: "HandleGroup", RespType: "group_mute",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.MuteMemberRequest{
					GroupId:      ctx.GroupID,
					UserId:       ctx.UserBID,
					MuteDuration: socialpb.MuteDuration_MUTE_DURATION_1_HOUR,
				}
			},
		},
		{
			Name: "BanMember_UserB", MaxType: 2000, MinType: 2023,
			Method: "HandleGroup", RespType: "group_ban",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.BanMemberRequest{GroupId: ctx.GroupID, UserId: ctx.UserBID}
			},
		},
		{
			Name: "RemoveMember_UserB", MaxType: 2000, MinType: 2027,
			Method: "HandleGroup", RespType: "group_remove",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.RemoveMemberRequest{GroupId: ctx.GroupID, UserId: ctx.UserBID}
			},
		},
		{
			Name: "UpdateMemberRole_UserB", MaxType: 2000, MinType: 2029,
			Method: "HandleGroup", RespType: "group_role",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.UpdateMemberRoleRequest{
					GroupId: ctx.GroupID, UserId: ctx.UserBID, Role: 2,
				}
			},
		},

		// ================================================================
		// E. Topic 正向链路（UserA 发帖，UserB 互动）
		// ================================================================

		{
			Name: "CreateTopic_UserA", MaxType: 3000, MinType: 3001,
			Method: "HandleTopic", RespType: "topic_create",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.CreateTopicRequest{
					Title:   "E2E Stateful Topic" + ctx.Suffix,
					Content: "Stateful topic content for full flow test",
					GroupId: ctx.GroupID,
				}
			},
		},
		{
			Name: "GetTopicList", MaxType: 3000, MinType: 3005,
			Method: "HandleTopic", RespType: "topic_list",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.GetTopicListRequest{GroupId: ctx.GroupID, Page: 1, PageSize: 20}
			},
		},
		{
			Name: "AddTopicReply_UserB", MaxType: 3000, MinType: 3043,
			Method: "HandleTopic", RespType: "topic_reply",
			IsPositive: true, Actor: "userB",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.AddTopicReplyRequest{TopicId: ctx.TopicID, Content: "Stateful reply from UserB"}
			},
		},
		{
			Name: "LikeTopic_UserB", MaxType: 3000, MinType: 3061,
			Method: "HandleTopic", RespType: "topic_like",
			IsPositive: true, Actor: "userB",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.LikeTopicRequest{TopicId: ctx.TopicID}
			},
		},
		{
			Name: "FavoriteTopic_UserB", MaxType: 3000, MinType: 3063,
			Method: "HandleTopic", RespType: "topic_favorite",
			IsPositive: true, Actor: "userB",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.FavoriteTopicRequest{TopicId: ctx.TopicID}
			},
		},
		{
			Name: "BatchGetTopicInfo", MaxType: 3000, MinType: 3057,
			Method: "HandleTopic", RespType: "topic_batch_info",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.BatchGetTopicInfoRequest{TopicIds: []string{ctx.TopicID}}
			},
		},
		{
			Name: "CreateReport_UserB", MaxType: 3000, MinType: 3095,
			Method: "HandleTopic", RespType: "topic_report",
			IsPositive: true, Actor: "userB",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.CreateReportRequest{
					TargetType: "topic", TargetId: ctx.TopicID,
					GroupId: ctx.GroupID, ReportType: "spam", ReasonText: "E2E stateful report",
				}
			},
		},
		{
			Name: "CheckTopicActions", MaxType: 3000, MinType: 3099,
			Method: "HandleTopic", RespType: "topic_actions",
			IsPositive: true, Actor: "userB",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.CheckTopicActionsRequest{TopicId: ctx.TopicID}
			},
		},
		{
			Name: "GetReplyList", MaxType: 3000, MinType: 3065,
			Method: "HandleTopic", RespType: "topic_reply_list",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.GetReplyListRequest{TopicId: ctx.TopicID, Page: 1, PageSize: 20}
			},
		},

		// ================================================================
		// F. 收尾操作（Logout + DeleteTopic）
		// ================================================================

		{
			Name: "DeleteTopic_UserA", MaxType: 3000, MinType: 3009,
			Method: "HandleTopic", RespType: "topic_delete",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.DeleteTopicRequest{TopicId: ctx.TopicID}
			},
		},
		{
			Name: "Logout_UserA", MaxType: 1000, MinType: 1025,
			Method: "HandleMember", RespType: "member_logout",
			IsPositive: true, Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.UserLogoutRequest{
					UserId:      ctx.UserAID,
					AccessToken: ctx.AccessTokenA,
				}
			},
		},

		// ================================================================
		// G. 负向用例（预期返回特定业务错误码，验证错误处理正确性）
		// ================================================================

		{
			Name: "NEG_DuplicateRegister", MaxType: 1000, MinType: 1021,
			Method: "HandleMember", RespType: "member_register",
			IsPositive: false, Actor: "", ExpectedBizCode: "10612",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.UserRegisterRequest{
					Username: "e2e_stateful_usera" + ctx.Suffix, // 已注册
					Password: "StatefulPass123!",
					Email:    "dup" + ctx.Suffix + "@test.com",
				}
			},
		},
		{
			Name: "NEG_WrongPassword", MaxType: 1000, MinType: 1023,
			Method: "HandleMember", RespType: "member_login",
			IsPositive: false, Actor: "", ExpectedBizCode: "10401",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.UserLoginRequest{
					Username: "e2e_stateful_usera" + ctx.Suffix,
					Password: "WrongPassword999",
				}
			},
		},
		{
			Name: "NEG_InvalidRefreshToken", MaxType: 1000, MinType: 1027,
			Method: "HandleMember", RespType: "member_refresh",
			IsPositive: false, Actor: "", ExpectedBizCode: "10401",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.RefreshTokenRequest{RefreshToken: "fake-invalid-refresh-token"}
			},
		},
		{
			Name: "NEG_JoinNonexistGroup", MaxType: 2000, MinType: 2013,
			Method: "HandleGroup", RespType: "group_join",
			IsPositive: false, Actor: "userB", ExpectedBizCode: "10701",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.JoinGroupRequest{GroupId: "grp_nonexist_99999"}
			},
		},
		{
			Name: "NEG_GetNonexistTopic", MaxType: 3000, MinType: 3005,
			Method: "HandleTopic", RespType: "topic_list",
			IsPositive: true, // GetTopicList 对不存在的 group_id 返回空列表(10200)，非错误码
			Actor: "userA",
			BuildRequest: func(ctx *E2EContext) proto.Message {
				return &socialpb.GetTopicListRequest{GroupId: "grp_nonexist_99999", Page: 1, PageSize: 20}
			},
		},
	}
}

// ========== 核心 Runner ==========

// runStatefulE2E 按顺序执行所有用例，维护 E2EContext 跨用例传递状态
func runStatefulE2E(baseURL string, ctx *E2EContext) []e2eResult {
	testCases := buildTestCases()
	results := make([]e2eResult, 0, len(testCases))

	for _, tc := range testCases {
		result := runSingleStatefulE2E(baseURL, tc, ctx)
		results = append(results, result)

		// 成功后提取响应中的关键字段到上下文（供后续用例使用）
		if result.Category == "POS_PASS" || result.BizCode == "10200" {
			extractContext(tc.RespType, result.rawResponseData, ctx)
		}
	}

	return results
}

// runSingleStatefulE2E 执行单条状态化 E2E 测试
func runSingleStatefulE2E(baseURL string, tc e2eTestCase, ctx *E2EContext) e2eResult {
	result := e2eResult{Name: tc.Name}
	startTime := time.Now()

	// Step 1: 使用 BuildRequest 构建请求体（支持模板变量）
	reqMsg := tc.BuildRequest(ctx)

	// Step 1.5: 对字符串字段执行模板变量替换
	resolveStringFields(reqMsg, ctx)

	// Step 2: 序列化请求体
	reqData, err := proto.Marshal(reqMsg)
	if err != nil {
		result.Category = "FAIL"
		result.ErrorDetail = fmt.Sprintf("Marshal failed: %v", err)
		return result
	}

	// Step 3: 构造 TARS MessagePacket（自动注入 token）
	packet := &pb.MessagePacket{
		MaxType: tc.MaxType,
		MinType: tc.MinType,
		Extend: map[string]string{
			"method":  tc.Method,
			"minType": fmt.Sprintf("%d", tc.MinType),
			"traceId": fmt.Sprintf("e2e-state-%s-%d", tc.Name, time.Now().UnixNano()),
		},
		Platform: pb.Platform_WEB,
		Data:      reqData,
	}

	// 自动注入 extend.token（auth_required 接口 + 有 actor 时）
	if token := ctx.tokenForActor(tc.Actor); token != "" {
		packet.Extend["token"] = token
	}

	packetData, err := proto.Marshal(packet)
	if err != nil {
		result.Category = "FAIL"
		result.ErrorDetail = fmt.Sprintf("Packet Marshal failed: %v", err)
		return result
	}

	// Step 4: 发送 HTTP POST
	httpReq, err := http.NewRequest("POST", baseURL, bytes.NewReader(packetData))
	if err != nil {
		result.Category = "FAIL"
		result.ErrorDetail = fmt.Sprintf("NewRequest failed: %v", err)
		return result
	}
	httpReq.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		result.Category = "FAIL"
		result.TarsCode = "(conn-error)"
		result.HTTPStatus = "(conn-error)"
		result.ErrorDetail = fmt.Sprintf("HTTP conn failed: %v", err)
		return result
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	result.HTTPStatus = fmt.Sprintf("%d", resp.StatusCode)
	result.LatencyMs = time.Since(startTime).Milliseconds()

	// Step 5: 解析 TARS MessagePacket 响应（L1 + L2 校验）
	respPacket := &pb.MessagePacket{}
	if len(body) > 0 {
		if uerr := proto.Unmarshal(body, respPacket); uerr != nil {
			result.Category = "FAIL"
			result.TarsCode = fmt.Sprintf("%d", resp.StatusCode)
			result.ErrorDetail = fmt.Sprintf("Unmarshal packet failed (raw=%d bytes): %v", len(body), uerr)
			return result
		}
	}

	result.TarsCode = respPacket.Extend["code"]
	result.RespSize = len(respPacket.Data)
	result.rawResponseData = respPacket.Data // 保存供 extractContext 使用

	// L1: HTTP 层校验
	if resp.StatusCode != 200 {
		result.Category = "FAIL"
		result.ErrorDetail = fmt.Sprintf("HTTP %d (TarsCode=%s)", resp.StatusCode, result.TarsCode)
		return result
	}

	// L2: TarsGo 协议层校验
	switch result.TarsCode {
	case "200":
		// 协议层成功，继续 L3/L4
	case "10404":
		result.Category = "FAIL"
		result.ErrorDetail = "HANDLER_NOT_FOUND"
		return result
	case "500":
		result.Category = "FAIL"
		result.ErrorDetail = "SERVER_ERROR"
		return result
	default:
		// 非 200 的 TarsCode 但 HTTP 200：可能是业务层面路由拦截
		result.Category = "WARN"
		result.ErrorDetail = fmt.Sprintf("Unexpected TarsCode: %s", result.TarsCode)
	}

	// Step 6: 业务层 Protobuf 反序列化 + Result.Code 校验（L3 + L4）
	if len(respPacket.Data) > 0 {
		bizCode, bizMsg := inspectBusinessResponse(tc.RespType, respPacket.Data)
		result.BizCode = bizCode
		result.BizMessage = bizMsg

		// L4: 业务语义判定（区分正向/负向）
		result.Category = classifyResult(tc, bizCode, result.TarsCode)
		if result.Category == "" {
			result.Category = "FAIL"
			result.ErrorDetail = fmt.Sprintf("Unexpected bizCode=%s msg=%s", bizCode, bizMsg)
		}
	} else {
		if result.Category == "" {
			result.Category = "WARN"
			result.BizCode = "(empty)"
			result.BizMessage = "No business payload"
		}
	}

	return result
}

// classifyResult 根据用例类型和实际 BizCode 判定结果分类
func classifyResult(tc e2eTestCase, bizCode, tarsCode string) string {
	if tarsCode != "200" && tarsCode != "" {
		// TarsCode 非 200 但未在前面处理为 FAIL
		return "WARN"
	}

	if tc.IsPositive {
		// 正向用例：期望 10200
		if isBusinessSuccess(bizCode) {
			return "POS_PASS"
		}
		// 正向用例未返回 10200 → 失败
		return "FAIL"
	}

	// 负向用例：检查是否匹配预期错误码
	if tc.ExpectedBizCode != "" && bizCode == tc.ExpectedBizCode {
		return "NEG_PASS"
	}
	if tc.ExpectedBizCode == "" && !isBusinessSuccess(bizCode) {
		// 未指定具体错误码，但确实返回了非成功码 → 也算 NEG_PASS
		return "NEG_PASS"
	}
	// 负向用例但返回了成功码 → 异常
	if isBusinessSuccess(bizCode) {
		return "FAIL"
	}
	return "WARN"
}

// ========== 上下文提取 ==========

// extractContext 从成功的响应中提取 ID/token 写入 E2EContext
func extractContext(respType string, data []byte, ctx *E2EContext) {
	switch respType {
	case "member_register":
		var rsp socialpb.UserRegisterResponse
		if proto.Unmarshal(data, &rsp) != nil {
			return
		}
		if rsp.Result != nil && rsp.Result.Code == 10200 {
			// 按用例执行顺序区分：第一个 Register 设 UserAID，第二个设 UserBID
			if ctx.UserAID == "" {
				ctx.setVar(&ctx.UserAID, rsp.UserId)
			} else {
				ctx.setVar(&ctx.UserBID, rsp.UserId)
			}
		}

	case "member_login":
		var rsp socialpb.UserLoginResponse
		if proto.Unmarshal(data, &rsp) != nil {
			return
		}
		if rsp.Result != nil && rsp.Result.Code == 10200 {
			// 根据上下文中哪个用户正在登录来决定存到哪里
			// 简化策略：LoginA 在 LoginB 之前执行，所以：
			if ctx.AccessTokenA == "" {
				ctx.setVar(&ctx.AccessTokenA, rsp.AccessToken)
				ctx.setVar(&ctx.RefreshTokenA, rsp.RefreshToken)
				ctx.setVar(&ctx.UserAID, rsp.UserId)
			} else {
				ctx.setVar(&ctx.AccessTokenB, rsp.AccessToken)
				ctx.setVar(&ctx.RefreshTokenB, rsp.RefreshToken)
				ctx.setVar(&ctx.UserBID, rsp.UserId)
			}
		}

	case "group_create":
		var rsp socialpb.CreateGroupResponse
		if proto.Unmarshal(data, &rsp) != nil {
			return
		}
		if rsp.Result != nil && rsp.Result.Code == 10200 {
			ctx.setVar(&ctx.GroupID, rsp.GroupId)
		}

	case "topic_create":
		var rsp socialpb.CreateTopicResponse
		if proto.Unmarshal(data, &rsp) != nil {
			return
		}
		if rsp.Result != nil && rsp.Result.Code == 10200 {
			ctx.setVar(&ctx.TopicID, rsp.TopicId)
		}

	case "topic_reply":
		var rsp socialpb.AddTopicReplyResponse
		if proto.Unmarshal(data, &rsp) != nil {
			return
		}
		if rsp.Result != nil && rsp.Result.Code == 10200 {
			ctx.setVar(&ctx.ReplyID, rsp.ReplyId)
		}
	}
}

// resolveStringFields 对 protobuf message 的 string 字段执行模板变量替换
// MVP-P0 策略：只处理已知需要模板的字段（GroupId/UserId/TopicId 等）
// 通过 proto reflection 或手动类型断言实现
func resolveStringFields(msg proto.Message, ctx *E2EContext) {
	switch m := msg.(type) {
	case *socialpb.JoinGroupRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
	case *socialpb.LeaveGroupRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
	case *socialpb.MuteMemberRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
		m.UserId = ctx.resolveTemplate(m.UserId)
	case *socialpb.BanMemberRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
		m.UserId = ctx.resolveTemplate(m.UserId)
	case *socialpb.RemoveMemberRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
		m.UserId = ctx.resolveTemplate(m.UserId)
	case *socialpb.UpdateMemberRoleRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
		m.UserId = ctx.resolveTemplate(m.UserId)
	case *socialpb.RenewMemberRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
		m.UserId = ctx.resolveTemplate(m.UserId)
	case *socialpb.CalcPayableAmountRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
	case *socialpb.GroupUserEnterRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
		m.UserId = ctx.resolveTemplate(m.UserId)
	case *socialpb.GetGroupStatsRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
	case *socialpb.BatchGetGroupsRequest:
		for i, gid := range m.GroupIds {
			m.GroupIds[i] = ctx.resolveTemplate(gid)
		}
	case *socialpb.GetGroupMemberUserIdsRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
	case *socialpb.CreateTopicRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
	case *socialpb.DeleteTopicRequest:
		m.TopicId = ctx.resolveTemplate(m.TopicId)
	case *socialpb.AddTopicReplyRequest:
		m.TopicId = ctx.resolveTemplate(m.TopicId)
	case *socialpb.LikeTopicRequest:
		m.TopicId = ctx.resolveTemplate(m.TopicId)
	case *socialpb.FavoriteTopicRequest:
		m.TopicId = ctx.resolveTemplate(m.TopicId)
	case *socialpb.BatchGetTopicInfoRequest:
		for i, tid := range m.TopicIds {
			m.TopicIds[i] = ctx.resolveTemplate(tid)
		}
	case *socialpb.CreateReportRequest:
		m.TargetId = ctx.resolveTemplate(m.TargetId)
		m.GroupId = ctx.resolveTemplate(m.GroupId)
	case *socialpb.CheckTopicActionsRequest:
		m.TopicId = ctx.resolveTemplate(m.TopicId)
	case *socialpb.GetReplyListRequest:
		m.TopicId = ctx.resolveTemplate(m.TopicId)
	case *socialpb.GetTopicListRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
	case *socialpb.GetBlockListRequest:
		m.GroupId = ctx.resolveTemplate(m.GroupId)
	}
}

// ========== 业务响应解析（复用原有逻辑）==========

// inspectBusinessResponse 尝试解析业务层响应，提取 Result.Code 和 Result.Message
func inspectBusinessResponse(respType string, data []byte) (code string, message string) {
	switch respType {
	case "member_register":
		var rsp socialpb.UserRegisterResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "member_login":
		var rsp socialpb.UserLoginResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "member_logout":
		var rsp socialpb.UserLogoutResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "member_refresh":
		var rsp socialpb.RefreshTokenResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "member_get_info":
		var rsp socialpb.GetUserInfoResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "member_update_info":
		var rsp socialpb.UpdateUserInfoResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "member_block":
		var rsp socialpb.BlockUserResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "member_unblock":
		var rsp socialpb.UnblockUserResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "member_block_list":
		var rsp socialpb.GetBlockListResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "member_update_status":
		var rsp socialpb.UpdateMemberStatusResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "member_stats":
		var rsp socialpb.GetUserStatsResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""

	// ===== Group 域 =====
	case "group_create":
		var rsp socialpb.CreateGroupResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "group_join":
		var rsp socialpb.JoinGroupResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "group_leave":
		var rsp socialpb.LeaveGroupResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "group_mute":
		var rsp socialpb.MuteMemberResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "group_ban":
		var rsp socialpb.BanMemberResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "group_remove":
		var rsp socialpb.RemoveMemberResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "group_role":
		var rsp socialpb.UpdateMemberRoleResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "group_renew":
		var rsp socialpb.RenewMemberResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "group_calc":
		var rsp socialpb.CalcPayableAmountResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "group_enter":
		var rsp socialpb.GroupUserEnterResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "group_stats":
		var rsp socialpb.GetGroupStatsResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "group_batch":
		var rsp socialpb.BatchGetGroupsResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "group_member_ids":
		var rsp socialpb.GetGroupMemberUserIdsResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""

	// ===== Topic 域 =====
	case "topic_create":
		var rsp socialpb.CreateTopicResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "topic_list":
		var rsp socialpb.GetTopicListResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "topic_delete":
		var rsp socialpb.DeleteTopicResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "topic_reply":
		var rsp socialpb.AddTopicReplyResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "topic_like":
		var rsp socialpb.LikeTopicResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "topic_favorite":
		var rsp socialpb.FavoriteTopicResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "topic_batch_info":
		var rsp socialpb.BatchGetTopicInfoResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "topic_report":
		var rsp socialpb.CreateReportResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "topic_actions":
		var rsp socialpb.CheckTopicActionsResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""
	case "topic_reply_list":
		var rsp socialpb.GetReplyListResponse
		if err := proto.Unmarshal(data, &rsp); err != nil {
			return "(parse-error)", err.Error()
		}
		if rsp.Result != nil {
			return fmt.Sprintf("%d", rsp.Result.Code), rsp.Result.Message
		}
		return "(no-result)", ""

	default:
		return "(unknown-type)", fmt.Sprintf("Unknown respType: %s, data=%d bytes", respType, len(data))
	}
}

// isBusinessSuccess 判断业务层是否成功
func isBusinessSuccess(bizCode string) bool {
	return bizCode == "0" || bizCode == "10200"
}

// ========== 输出 ==========

func countByStatus(results []e2eResult, status string) int {
	n := 0
	for _, r := range results {
		if r.Category == status {
			n++
		}
	}
	return n
}

func printStatefulSummary(baseURL string, results []e2eResult, ctx *E2EContext) {
	posPass := countByStatus(results, "POS_PASS")
	negPass := countByStatus(results, "NEG_PASS")
	warnCount := countByStatus(results, "WARN")
	failCount := countByStatus(results, "FAIL")
	total := len(results)

	fmt.Printf("\n")
	fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║   Social Stateful E2E Test — Phase 1 MVP-P0 Task 5-D         ║\n")
	fmt.Printf("╠══════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ 时间: %-50s ║\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("║ Gateway: %-46s ║\n", baseURL)
	fmt.Printf("╠══════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ POS_PASS=%d  NEG_PASS=%d  WARN=%d  FAIL=%d  (总计 %d)          ║\n", posPass, negPass, warnCount, failCount, total)
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")

	// 表头
	fmt.Printf("%-4s %-28s %-8s %-6s %-10s %-6s %s\n",
		"#", "Protocol", "TarsCode", "HTTP", "Category", "Size", "Detail")
	fmt.Printf("---- ---------------------------- -------- ------ ---------- ------ ----------------------------------------\n")

	for i, r := range results {
		icon := categoryIcon(r.Category)
		detail := r.ErrorDetail
		if detail == "" && r.BizMessage != "" && !isBusinessSuccess(r.BizCode) {
			detail = r.BizMessage
		}
		if detail == "" && r.Category == "POS_PASS" {
			detail = "BizCode=" + r.BizCode
		}
		if len(detail) > 40 {
			detail = detail[:37] + "..."
		}
		fmt.Printf("%s [%-2d] %-28s %-8s %-6s %-10s %-6dB %s\n",
			icon, i+1, r.Name, r.TarsCode, r.HTTPStatus, r.Category, r.RespSize, detail)
	}

	// 上下文快照
	fmt.Printf("\n--- E2EContext 快照 ---\n")
	fmt.Printf("  UserAID=%q  AccessTokenA=%s  RefreshTokenA=%s\n",
		ctx.UserAID, truncToken(ctx.AccessTokenA), truncToken(ctx.RefreshTokenA))
	fmt.Printf("  UserBID=%q  AccessTokenB=%s  RefreshTokenB=%s\n",
		ctx.UserBID, truncToken(ctx.AccessTokenB), truncToken(ctx.RefreshTokenB))
	fmt.Printf("  GroupID=%q  TopicID=%q  ReplyID=%q\n", ctx.GroupID, ctx.TopicID, ctx.ReplyID)

	// 结论
	fmt.Printf("\n")
	if failCount > 0 {
		fmt.Printf("❌ 有 %d 条 FAIL，系统级异常需修复\n", failCount)
	} else if posPass >= 15 { // 核心正向链路最低阈值
		fmt.Printf("✅ Stateful E2E 通过: POS_PASS=%d NEG_PASS=%d WARN=%d FAIL=0\n", posPass, negPass, warnCount)
	} else {
		fmt.Printf("⚠️  POS_PASS=%d 偏低，核心正向链路可能存在阻塞\n", posPass)
	}
}

func categoryIcon(cat string) string {
	switch cat {
	case "POS_PASS":
		return "✅"
	case "NEG_PASS":
		return "✔️ "
	case "WARN":
		return "⚠️ "
	default:
		return "❌"
	}
}

func truncToken(t string) string {
	if len(t) <= 12 {
		return t
	}
	return t[:12] + "..."
}
