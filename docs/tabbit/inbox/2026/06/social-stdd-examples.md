# Social Service STDD 开发示例文档

> **文档编号**: DEV-STDD-SOCIAL-002
> **版本**: v1.0
> **创建日期**: 2026-06-16
> **适用范围**: go/modules/social/ 全域 STDD 开发
> **前置文档**: social-service-dev-guide.md (DEV-GUIDE-SOCIAL-001)

---

## 目录

1. [Context 规范：JWT → user_id → context 注入](#1-context-规范)
2. [Tars.log 封装：统一 traceId 打印](#2-tarslog-封装)
3. [Proto ↔ Model 数据转换层](#3-proto--model-数据转换层)
4. [routes.yaml 路由配置完整示例](#4-routesyaml-路由配置)
5. [Tars 层注入测试（Servant + Handler）](#5-tars-层注入测试)
6. [Service 层 STDD 单元测试用例](#6-service-层-stdd-单元测试)

---

## 1. Context 规范

### 1.1 设计原则

```
extend["token"]  =  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
         ↓  servant.go 在 Handle() 入口处解析
JWT.Claims.UserID
         ↓  注入 context
ctx = context.WithValue(ctx, ctxKeyUserID, userID)
         ↓  Handler.Dispatch() 透传 ctx（不修改）
         ↓  svc_*.go 中通过 GetUserID(ctx) 取出
operatorUserID := auth.GetUserID(ctx)
```

**铁律**：
- JWT 解析只在 `servant.go` 的 `Handle()` 入口做一次，全链路复用同一 ctx
- `svc_*.go` 通过 `auth.GetUserID(ctx)` 取值，**禁止**在 service 层直接读 extend
- token 缺失或解析失败 → 返回 `ERROR_CODE_UNAUTHORIZED`（10401），不进入业务逻辑

---

### 1.2 go/common-lib/auth/context.go

```go
// 文件: go/common-lib/auth/context.go
// 职责: 定义 context key + JWT 解析 + context 取值工具函数
// ⚠️  全项目统一使用此包，禁止各模块自定义 context key

package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ── context key 类型（私有，防止外部包意外覆盖）────────────────────────
type contextKey string

const (
	ctxKeyUserID  contextKey = "auth_user_id"
	ctxKeyTraceID contextKey = "trace_id"
	ctxKeyToken   contextKey = "raw_token"
)

// ── JWT Claims 结构体 ─────────────────────────────────────────────────
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username,omitempty"`
	Platform string `json:"platform,omitempty"` // ios/android/web
	jwt.RegisteredClaims
}

// ── context 注入函数 ──────────────────────────────────────────────────

// WithUserID 将 userID 注入 context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

// WithTraceID 将 traceID 注入 context（由 servant.go 统一设置）
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxKeyTraceID, traceID)
}

// ── context 取值函数 ──────────────────────────────────────────────────

// GetUserID 从 context 中取 userID（若未登录返回空字符串）
func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyUserID).(string)
	return v
}

// MustGetUserID 从 context 中取 userID，若未登录 panic（用于必须登录的接口）
// 推荐在 servant.go 层鉴权失败时直接返回错误，不依赖此函数
func MustGetUserID(ctx context.Context) (string, error) {
	userID := GetUserID(ctx)
	if userID == "" {
		return "", errors.New("auth: userID not found in context, token may be missing or invalid")
	}
	return userID, nil
}

// GetTraceID 从 context 中取 traceID
func GetTraceID(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyTraceID).(string)
	return v
}

// ── JWT 解析函数 ──────────────────────────────────────────────────────

// ParseToken 解析 JWT token，返回 Claims
// jwtSecret 从配置中心读取，不硬编码
func ParseToken(tokenStr, jwtSecret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("auth: unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(jwtSecret), nil
		},
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("auth: parse token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("auth: invalid token claims")
	}
	if claims.UserID == "" {
		return nil, errors.New("auth: token missing user_id claim")
	}
	return claims, nil
}

// ── traceID 生成 ──────────────────────────────────────────────────────

// NewTraceID 生成 traceID（优先从 extend 取，否则生成新的）
func NewTraceID(extend map[string]string) string {
	if tid, ok := extend["trace_id"]; ok && tid != "" {
		return tid
	}
	return generateTraceID() // e.g. UUID v4 or snowflake
}

func generateTraceID() string {
	// 使用 crypto/rand 生成 UUID v4
	// 实际实现复用 common-lib/uuid 包
	return fmt.Sprintf("tr-%d", time.Now().UnixNano())
}
```

---

### 1.3 servant.go JWT 注入完整模板

```go
// 文件: go/modules/social/member/servant.go
// 关键变化：Handle() 入口处解析 JWT，注入 ctx，再转发给 Handler

package member

import (
	"context"
	"fmt"
	"strconv"

	"go/common-lib/auth"
	"go/common-lib/module"
	"go/gateway/proto-gateway/tarsclient"
	tarslog "go/common-lib/tarslog"  // Tars.log 封装（见第2章）

	commonpb "github.com/jimiechen/mineplanet/protocols/generated/go/proto/base"
	"google.golang.org/protobuf/proto"
)

const (
	servantApp    = "CaiRobotSocialApp"
	servantServer = "SocialMemberServer"
	servantName   = "SocialMemberServant"
	servantMethod = "Handle"
)

type SocialMemberServant struct {
	handler   *Handler
	jwtSecret string           // 从配置中心读取
	logger    *tarslog.Logger  // traceId 绑定 logger
}

func NewServant(deps module.Deps) *SocialMemberServant {
	return &SocialMemberServant{
		handler:   NewHandler(deps),
		jwtSecret: deps.Config.GetString("auth.jwt_secret"),
		logger:    tarslog.New("SocialMemberServant"),
	}
}

func (s *SocialMemberServant) Register(invoker tarsclient.LocalInvoker) {
	key := tarsclient.TargetKey{
		App: servantApp, Server: servantServer,
		Servant: servantName, Method: servantMethod,
	}
	invoker.Register(key, s.Handle)
}

// Handle TarsGo 标准 bytes 接口
//
// extend 约定字段:
//   "minType"  — 协议号（必须）
//   "token"    — JWT token（可选，公开接口可为空）
//   "trace_id" — 链路追踪 ID（可选，无则自动生成）
//   "platform" — 客户端平台 ios/android/web（可选）
func (s *SocialMemberServant) Handle(
	ctx context.Context,
	req []byte,
	extend map[string]string,
) (retCode int, respBytes []byte, err error) {

	// ── Step 1: 生成/注入 traceID（全链路唯一标识）─────────────────
	traceID := auth.NewTraceID(extend)
	ctx = auth.WithTraceID(ctx, traceID)
	log := s.logger.With(traceID) // 后续所有日志自动携带 traceId

	// ── Step 2: 提取 minType ─────────────────────────────────────
	minTypeStr, ok := extend["minType"]
	if !ok || minTypeStr == "" {
		log.Errorf("missing minType in extend")
		return -1, nil, fmt.Errorf("servant: missing minType")
	}
	minType, err := strconv.Atoi(minTypeStr)
	if err != nil {
		log.Errorf("invalid minType=%q: %v", minTypeStr, err)
		return -1, nil, fmt.Errorf("servant: invalid minType")
	}

	// ── Step 3: 解析 JWT → 注入 userID（有 token 才解析）──────────
	//
	// 设计说明:
	//   - 公开接口（如注册/登录）extend["token"] 为空 → 不解析，ctx 中 userID=""
	//   - 需要登录的接口由 svc_*.go 通过 auth.MustGetUserID(ctx) 判断
	//   - 不在此处强制要求所有接口都有 token，由各 svc 自行决定是否必须登录
	if tokenStr := extend["token"]; tokenStr != "" {
		claims, parseErr := auth.ParseToken(tokenStr, s.jwtSecret)
		if parseErr != nil {
			// token 存在但解析失败 → 返回未授权响应（不是系统错误）
			log.Warnf("minType=%d token invalid: %v", minType, parseErr)
			respBytes, err = marshalUnauthorizedResp(minType)
			if err != nil {
				return -1, nil, err
			}
			return 0, respBytes, nil // retCode=0，错误在 respBytes.Result 中表达
		}
		ctx = auth.WithUserID(ctx, claims.UserID)
		log.Infof("minType=%d userID=%s platform=%s", minType, claims.UserID, claims.Platform)
	} else {
		log.Infof("minType=%d no-token (public endpoint)", minType)
	}

	// ── Step 4: 转发 Handler ─────────────────────────────────────
	respBytes, err = s.handler.Dispatch(ctx, minType, req, extend)
	if err != nil {
		log.Errorf("minType=%d dispatch error: %v", minType, err)
		return -1, nil, fmt.Errorf("servant: dispatch minType=%d: %w", minType, err)
	}

	return 0, respBytes, nil
}

// marshalUnauthorizedResp 构造 token 非法时的统一未授权响应
// minType 决定用哪个 Response 类型包装（当前简化为通用 ErrorEnvelope）
func marshalUnauthorizedResp(minType int) ([]byte, error) {
	// 实际按 minType 路由到对应 Response 类型；此处示例用通用错误
	resp := &commonpb.ErrorEnvelope{
		Result: &commonpb.Result{
			Code:    int32(commonpb.ErrorCode_ERROR_CODE_UNAUTHORIZED),
			Message: "token 无效或已过期",
		},
	}
	return proto.Marshal(resp)
}
```

---

### 1.4 svc_*.go 中取 userID 示例

```go
// 文件: go/modules/social/group/svc_remove_member.go
// 场景: 圈主/管理员移除成员（需要登录 + 权限校验）

package group

import (
	"context"
	"fmt"

	"go/common-lib/auth"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/proto/base"
)

type RemoveMemberService struct {
	repo GroupRepository
	perm permission.Service
}

func (s *RemoveMemberService) Handle(
	ctx context.Context,
	req *pb.RemoveMemberRequest,
) (*pb.RemoveMemberResponse, error) {

	// ── Step 1: 取操作者 userID（从 context，不从 extend）──────────
	operatorID, err := auth.MustGetUserID(ctx)
	if err != nil {
		// token 未传或解析失败（servant.go 已处理 token 无效，此处兜底）
		return s.fail(int32(pb.ErrorCode_ERROR_CODE_UNAUTHORIZED), "请先登录"), nil
	}

	// ── Step 2: 参数校验 ───────────────────────────────────────────
	if req.GetGroupId() == "" {
		return s.fail(int32(pb.GroupErrorCode_GROUP_ERROR_ID_EMPTY), "圈子ID不能为空"), nil
	}
	if req.GetUserId() == "" {
		return s.fail(int32(pb.GroupErrorCode_GROUP_ERROR_MISSING_PARAMETERS), "用户ID不能为空"), nil
	}

	// ── Step 3: 权限校验（operatorID 来自 JWT，不来自请求参数）───────
	canManage, err := s.perm.CanManageMember(ctx, operatorID, req.GetGroupId(), req.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("svc_remove_member: CanManageMember: %w", err)
	}
	if !canManage {
		return s.fail(int32(pb.GroupErrorCode_GROUP_ERROR_PERMISSION_DENIED), "权限不足"), nil
	}

	// ── Step 4: 1级数据写入 ────────────────────────────────────────
	if err := s.repo.UpdateMemberStatus(ctx, req.GetGroupId(), req.GetUserId(), 4); err != nil {
		return nil, fmt.Errorf("svc_remove_member: UpdateMemberStatus: %w", err)
	}

	// ── Step 5: 审计日志（写入 group_admin_actions）──────────────────
	_ = s.repo.CreateAdminAction(ctx, &GroupAdminAction{
		GroupID:    req.GetGroupId(),
		OperatorID: operatorID,       // ← 来自 JWT，不可伪造
		TargetID:   req.GetUserId(),
		Action:     "remove_member",
		Reason:     req.GetReason(),
	})

	return &pb.RemoveMemberResponse{
		Result: successResult(),
	}, nil
}

func (s *RemoveMemberService) fail(code int32, msg string) *pb.RemoveMemberResponse {
	return &pb.RemoveMemberResponse{Result: &pb.Result{Code: code, Message: msg}}
}
```

---

## 2. Tars.log 封装

### 2.1 设计要求

```
每条日志必须包含:
  [时间戳] [日志级别] [traceId] [servant名称] [消息]

示例输出:
  2026-06-16T10:30:01.234Z [INFO]  [tr-1234567890] [SocialMemberServant] minType=1021 userID=abc123 platform=ios
  2026-06-16T10:30:01.235Z [WARN]  [tr-1234567890] [SocialMemberServant] token invalid: jwt expired
  2026-06-16T10:30:01.236Z [ERROR] [tr-9876543210] [SocialGroupServant] dispatch minType=2011: db connection refused
```

---

### 2.2 go/common-lib/tarslog/logger.go

```go
// 文件: go/common-lib/tarslog/logger.go
// 职责: Tars 服务统一日志封装，强制打印 traceId
// 底层: 封装 TarsGo 原生 tars.TLOG 或 zap（按项目实际选择）
// ⚠️  全项目 Tars 服务统一使用此包，禁止直接调用 fmt.Println 或 log.Printf

package tarslog

import (
	"fmt"
	"time"
)

// Level 日志级别
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelStr = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO ",
	WARN:  "WARN ",
	ERROR: "ERROR",
}

// Logger 绑定 servant 名称的 logger 实例
type Logger struct {
	servant string // Servant 名称，如 "SocialMemberServant"
}

// New 创建绑定 servant 名称的 Logger
func New(servant string) *Logger {
	return &Logger{servant: servant}
}

// ScopedLogger 绑定 traceId 的子 logger（每次请求创建，不共享）
type ScopedLogger struct {
	servant string
	traceID string
}

// With 创建绑定 traceId 的子 logger
// 在 servant.go Handle() 入口调用，后续所有日志通过 ScopedLogger 打印
func (l *Logger) With(traceID string) *ScopedLogger {
	return &ScopedLogger{servant: l.servant, traceID: traceID}
}

// ── ScopedLogger 打印方法（格式: [时间] [级别] [traceId] [servant] 消息）

func (s *ScopedLogger) log(level Level, msg string) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	output := fmt.Sprintf("%s [%s] [%s] [%s] %s",
		timestamp, levelStr[level], s.traceID, s.servant, msg)
	// TODO: 替换为 TarsGo tars.TLOG.Debug/Info/Warn/Error 实现
	// tars.TLOG.Debug(output)
	// 当前使用标准输出（TarsGo 环境下由框架捕获）
	fmt.Println(output)
}

func (s *ScopedLogger) Debugf(format string, args ...interface{}) {
	s.log(DEBUG, fmt.Sprintf(format, args...))
}

func (s *ScopedLogger) Infof(format string, args ...interface{}) {
	s.log(INFO, fmt.Sprintf(format, args...))
}

func (s *ScopedLogger) Warnf(format string, args ...interface{}) {
	s.log(WARN, fmt.Sprintf(format, args...))
}

func (s *ScopedLogger) Errorf(format string, args ...interface{}) {
	s.log(ERROR, fmt.Sprintf(format, args...))
}

// ── Logger 无 traceId 版本（模块初始化时使用，非请求链路）

func (l *Logger) Infof(format string, args ...interface{}) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	fmt.Printf("%s [INFO ] [no-trace] [%s] %s\n",
		timestamp, l.servant, fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	fmt.Printf("%s [ERROR] [no-trace] [%s] %s\n",
		timestamp, l.servant, fmt.Sprintf(format, args...))
}
```

---

### 2.3 TarsGo TLOG 对接（生产实现）

```go
// 文件: go/common-lib/tarslog/tars_backend.go
// 职责: 将 tarslog 对接到 TarsGo 原生 tars.TLOG（生产环境使用）
// 开发/测试环境: 使用 logger.go 的 fmt.Println 实现
// 生产环境: 构建时通过 build tag "tars" 切换此文件

//go:build tars

package tarslog

import (
	"fmt"
	"github.com/TarsCloud/TarsGo/tars"
)

func (s *ScopedLogger) log(level Level, msg string) {
	formatted := fmt.Sprintf("[%s] [%s] %s", s.traceID, s.servant, msg)
	switch level {
	case DEBUG:
		tars.TLOG.Debug(formatted)
	case INFO:
		tars.TLOG.Info(formatted)
	case WARN:
		tars.TLOG.Warn(formatted)
	case ERROR:
		tars.TLOG.Error(formatted)
	}
}
```

---

### 2.4 日志使用规范（servant.go / svc_*.go）

```go
// servant.go 中的使用方式（已在第1章展示）
log := s.logger.With(traceID)
log.Infof("minType=%d userID=%s", minType, userID)
log.Errorf("minType=%d error: %v", minType, err)

// svc_*.go 中的使用方式（通过 ctx 取 traceId，不持有 logger）
// 方案A: 在 svc_*.go 中按需从 ctx 获取 traceId 构造局部日志
// 推荐在 servant.go 层完成所有路由级日志，svc 层只记录关键业务事件

func (s *RegisterService) Handle(ctx context.Context, req *pb.UserRegisterRequest) (*pb.UserRegisterResponse, error) {
	traceID := auth.GetTraceID(ctx)
	log := tarslog.New("RegisterService").With(traceID)

	log.Infof("username=%s email=%s", req.GetUsername(), req.GetEmail())
	// ... 业务逻辑
}
```

---

## 3. Proto ↔ Model 数据转换层

### 3.1 设计原则

```
proto.UserRegisterRequest  ──toModel()──▶  model.User  ──CreateUser()──▶  MySQL
MySQL  ──GetUserByID()──▶  model.User  ──toProto()──▶  commonpb.UserInfo
```

- `toModel()` / `toProto()` 是 svc_*.go 内的私有函数，不对外暴露
- model 字段变更不影响 proto（双方通过转换函数解耦）
- proto 字段变更（不可发生）不影响 model

---

### 3.2 go/modules/social/member/converter.go

```go
// 文件: go/modules/social/member/converter.go
// 职责: proto ↔ model 双向转换函数，本域所有 svc_*.go 共享
// ⚠️  此文件不包含业务逻辑，只做字段映射
// ⚠️  不得在此文件引入 repository / permission 依赖

package member

import (
	commonpb "github.com/jimiechen/mineplanet/protocols/generated/go/proto/base"
)

// ── proto → model（写入方向）──────────────────────────────────────────

// registerRequestToUser 将注册请求映射为内部 User model
// 只映射请求中有的字段，其余字段由 svc_register.go 填充（如 ID/Salt/Status）
func registerRequestToUser(req *commonpb.UserRegisterRequest) *User {
	return &User{
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Nickname: firstNonEmpty(req.GetNickname(), req.GetUsername()),
		Avatar:   req.GetAvatar(),
		Timezone: defaultStr(req.GetTimezone(), "UTC+0"),
		// ID / Password / Salt / Status / UID / CreatedAt 等由 svc_*.go 填充
	}
}

// updateUserRequestToFields 将更新请求映射为 GORM Updates map
// 只更新请求中明确传入的字段（非零值或非空）
func updateUserRequestToFields(req *commonpb.UpdateUserInfoRequest) map[string]interface{} {
	fields := make(map[string]interface{})
	if req.GetNickname() != "" {
		fields["nickname"] = req.GetNickname()
	}
	if req.GetAvatar() != "" {
		fields["avatar"] = req.GetAvatar()
	}
	if req.GetTimezone() != "" {
		fields["timezone"] = req.GetTimezone()
	}
	// 注意: Username/Email/Phone 不允许通过此接口修改（安全敏感字段走独立协议）
	return fields
}

// ── model → proto（读取方向）──────────────────────────────────────────

// userToProto 将 User model 映射为 proto UserInfo（通用返回结构）
func userToProto(u *User) *commonpb.UserInfo {
	if u == nil {
		return nil
	}
	return &commonpb.UserInfo{
		UserId:          u.ID,
		Username:        u.Username,
		Email:           maskEmail(u.Email),    // 邮箱脱敏
		Phone:           maskPhone(u.Phone),    // 手机号脱敏
		Avatar:          u.Avatar,
		Nickname:        u.Nickname,
		Status:          userStatusToProto(u.Status),
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
		LastLoginAt:     u.LastLoginAt,
		MembershipLevel: membershipToProto(u.MembershipLevel),
		Uid:             u.UID,
		ImRegistered:    u.IMRegistered == 1,
	}
}

// ── 枚举转换 ──────────────────────────────────────────────────────────

func userStatusToProto(status int8) commonpb.UserStatus {
	switch status {
	case 1:
		return commonpb.UserStatus_USER_STATUS_ACTIVE
	case 2:
		return commonpb.UserStatus_USER_STATUS_INACTIVE
	case 3:
		return commonpb.UserStatus_USER_STATUS_BANNED
	case 4:
		return commonpb.UserStatus_USER_STATUS_PENDING // deleted → pending（proto 字段复用）
	default:
		return commonpb.UserStatus_USER_STATUS_UNKNOWN
	}
}

func membershipToProto(level string) commonpb.MembershipLevel {
	switch level {
	case "premium":
		return commonpb.MembershipLevel_MEMBERSHIP_PREMIUM
	case "premium_plus":
		return commonpb.MembershipLevel_MEMBERSHIP_PREMIUM_PLUS
	default:
		return commonpb.MembershipLevel_MEMBERSHIP_NORMAL
	}
}

// ── 数据脱敏工具 ──────────────────────────────────────────────────────

// maskPhone 手机号脱敏: 138****5678
func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// maskEmail 邮箱脱敏: j***e@example.com
func maskEmail(email string) string {
	at := -1
	for i, c := range email {
		if c == '@' {
			at = i
			break
		}
	}
	if at <= 1 {
		return email
	}
	return string(email[0]) + "***" + email[at-1:] // j***e@example.com
}
```

---

### 3.3 converter_test.go（转换层单元测试）

```go
// 文件: go/modules/social/member/converter_test.go
package member

import (
	"testing"

	"github.com/stretchr/testify/assert"
	commonpb "github.com/jimiechen/mineplanet/protocols/generated/go/proto/base"
)

func TestUserToProto_正常转换(t *testing.T) {
	u := &User{
		ID:       "user-id-001",
		Username: "testuser",
		Email:    "test@example.com",
		Phone:    "13812345678",
		Nickname: "测试用户",
		Status:   1, // active
		UID:      "UID001",
	}
	got := userToProto(u)
	assert.Equal(t, "user-id-001", got.UserId)
	assert.Equal(t, "testuser", got.Username)
	assert.Equal(t, "138****5678", got.Phone)          // 脱敏
	assert.Equal(t, "t***t@example.com", got.Email)    // 脱敏
	assert.Equal(t, commonpb.UserStatus_USER_STATUS_ACTIVE, got.Status)
}

func TestUserToProto_Nil用户返回Nil(t *testing.T) {
	assert.Nil(t, userToProto(nil))
}

func TestMaskPhone_正常脱敏(t *testing.T) {
	assert.Equal(t, "138****5678", maskPhone("13812345678"))
}

func TestMaskPhone_短号码不脱敏(t *testing.T) {
	assert.Equal(t, "123", maskPhone("123"))
}

func TestMaskEmail_正常脱敏(t *testing.T) {
	assert.Equal(t, "j***e@example.com", maskEmail("jane@example.com"))
}

func TestRegisterRequestToUser_字段映射正确(t *testing.T) {
	req := &commonpb.UserRegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Phone:    "13900001111",
	}
	got := registerRequestToUser(req)
	assert.Equal(t, "alice", got.Username)
	assert.Equal(t, "alice", got.Nickname) // nickname 回填 username
	assert.Equal(t, "UTC+0", got.Timezone) // 默认时区
}

func TestUpdateUserRequestToFields_只映射非空字段(t *testing.T) {
	req := &commonpb.UpdateUserInfoRequest{
		Nickname: "新昵称",
		Avatar:   "",  // 空不映射
	}
	fields := updateUserRequestToFields(req)
	assert.Equal(t, "新昵称", fields["nickname"])
	_, hasAvatar := fields["avatar"]
	assert.False(t, hasAvatar)
}
```

---

## 4. routes.yaml 路由配置

### 4.1 完整配置文件示例

```yaml
# 文件: configs/routes.yaml
# 规则: maxType → Servant 映射（一个 maxType 一条路由）
# minType 通过 extend["minType"] 传递，不在此配置中细分
# ⚠️  与 servant.go 中的 servantApp/servantServer/servantName/servantMethod 必须完全一致

routes:

  # ── 已有路由（保持不变）─────────────────────────────────────────────

  # HealthCheck / HelloWorld
  - maxType: 4000
    servant:
      app:     CaiRobotApp
      server:  GatewayServer
      servant: GatewayServant
      method:  Handle
    desc: "健康检查 / HelloWorld"

  # Admin 管理后台（go-admin HTTP REST，不走 MessagePacket）
  # 注意: Admin 走独立 HTTP 路由，不在此 routes.yaml 中配置

  # ── 社交域新增路由 ────────────────────────────────────────────────────

  # 用户成员域 maxType=1000
  # 包含协议: register(1021) / login(1023) / logout(1025) / refresh(1027)
  #           get_user_info / update_user_info / block_user ...
  - maxType: 1000
    servant:
      app:     CaiRobotSocialApp
      server:  SocialMemberServer
      servant: SocialMemberServant
      method:  Handle
    auth:
      # 公开接口（无需 token）: 1021/注册, 1023/登录
      # 其余接口: servant.go 从 extend["token"] 解析 JWT，注入 ctx
      public_min_types: [1021, 1023]
    desc: "社交域-用户成员协议组"

  # 群组域 maxType=2000
  # 包含协议: create_group / join_group / leave_group / mute_member / ban_member ...
  - maxType: 2000
    servant:
      app:     CaiRobotSocialApp
      server:  SocialGroupServer
      servant: SocialGroupServant
      method:  Handle
    auth:
      public_min_types: []  # 群组域所有接口均需 token
    desc: "社交域-群组协议组"

  # 主题域 maxType=3000
  # 包含协议: create_topic / get_topic_list / like_topic / add_reply ...
  - maxType: 3000
    servant:
      app:     CaiRobotSocialApp
      server:  SocialTopicServer
      servant: SocialTopicServant
      method:  Handle
    auth:
      public_min_types: [3001, 3003, 3005]  # 列表/详情/搜索可匿名访问
    desc: "社交域-主题协议组"

  # 第三方服务域 maxType=4000
  # ⚠️  注意: 4000-4999 段已被社交域占用，需在协议编号注册表确认无服务商后台冲突
  - maxType: 4000
    servant:
      app:     CaiRobotSocialApp
      server:  SocialThirdServer
      servant: SocialThirdServant
      method:  Handle
    auth:
      public_min_types: [4011, 4201, 4203, 4701, 4711]  # OSS上传/OAuth/广告可匿名
    desc: "社交域-第三方服务协议组"

  # 收件箱消息域 maxType=5000
  - maxType: 5000
    servant:
      app:     CaiRobotSocialApp
      server:  SocialInboxServer
      servant: SocialInboxServant
      method:  Handle
    auth:
      public_min_types: []  # 消息域所有接口均需 token
    desc: "社交域-收件箱消息协议组"
```

---

### 4.2 routes.yaml 加载与 Gateway 对接说明

```go
// 在 Gateway 的路由加载逻辑中，public_min_types 用于决定是否必须有 token。
// 当前实现中，servant.go 自行处理 token 逻辑（有 token 就解析，无 token ctx.userID 为空）。
// 如果需要 Gateway 层统一拦截未授权请求，则在 Gateway 的 MessagePacket 处理器中
// 读取 routes.yaml 的 public_min_types 判断是否需要透传 extend["token"]。

// Gateway MessagePacket 处理器伪代码（参考实现）:
// func (g *GatewayServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//     packet := decodeMessagePacket(r.Body)
//     route := g.routes.Find(packet.MaxType)
//     isPublic := route.IsPublic(packet.MinType)
//     if !isPublic && packet.Extend["token"] == "" {
//         writeUnauthorized(w)
//         return
//     }
//     g.invoker.Invoke(ctx, route.Target, packet.Data, packet.Extend)
// }
```

---

## 5. Tars 层注入测试

### 5.1 测试策略

```
Tars 层测试范围:
  ① servant.go Handle() — JWT 解析 / traceId 注入 / 未授权返回
  ② handler.go Dispatch() — minType 路由 / proto unmarshal / proto marshal
  ③ servant → handler 集成 — 完整调用链（不带 DB，mock svc）

不测范围:
  - LocalInvoker.Register() 本身（框架代码，不在业务测试范围）
  - proto 生成代码（protobuf 官方保证，不重复测试）
```

---

### 5.2 servant_test.go

```go
// 文件: go/modules/social/member/servant_test.go
// 层级: Tars Servant 层测试
// 覆盖: JWT 解析 / traceId 注入 / 未授权响应 / minType 路由

package member

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go/common-lib/auth"
	commonpb "github.com/jimiechen/mineplanet/protocols/generated/go/proto/base"
)

// ── 测试用 JWT 工具 ────────────────────────────────────────────────────

const testJWTSecret = "test-secret-for-unit-tests-only"

// makeValidToken 生成有效 JWT token（仅用于测试）
func makeValidToken(userID string) string {
	claims := auth.Claims{
		UserID:   userID,
		Platform: "ios",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testJWTSecret))
	return signed
}

// makeExpiredToken 生成过期 JWT token
func makeExpiredToken(userID string) string {
	claims := auth.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 已过期
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testJWTSecret))
	return signed
}

// ── Mock Handler ──────────────────────────────────────────────────────

// mockHandler 捕获 Dispatch 被调用时的 ctx（用于验证 context 注入）
type mockHandler struct {
	capturedCtx   context.Context
	capturedMinType int
	respBytes     []byte
	err           error
}

func (m *mockHandler) Dispatch(ctx context.Context, minType int, req []byte, extend map[string]string) ([]byte, error) {
	m.capturedCtx = ctx
	m.capturedMinType = minType
	return m.respBytes, m.err
}

// newTestServant 创建测试用 Servant（使用 mock handler）
func newTestServant(mockH *mockHandler) *SocialMemberServant {
	return &SocialMemberServant{
		handler:   mockH,
		jwtSecret: testJWTSecret,
		logger:    tarslog.New("TestServant"),
	}
}

// ── 测试用例 ──────────────────────────────────────────────────────────

func TestServant_Handle_有效Token_注入UserID到Context(t *testing.T) {
	token := makeValidToken("user-id-001")
	mockH := &mockHandler{respBytes: []byte("mock-resp")}
	servant := newTestServant(mockH)

	_, _, err := servant.Handle(context.Background(), []byte("req-body"), map[string]string{
		"minType": "1029", // GetUserInfo
		"token":   token,
	})

	require.NoError(t, err)
	// 验证 userID 已注入 context
	assert.Equal(t, "user-id-001", auth.GetUserID(mockH.capturedCtx))
	assert.Equal(t, 1029, mockH.capturedMinType)
}

func TestServant_Handle_无Token_Context中UserID为空(t *testing.T) {
	mockH := &mockHandler{respBytes: []byte("mock-resp")}
	servant := newTestServant(mockH)

	_, _, err := servant.Handle(context.Background(), []byte("req-body"), map[string]string{
		"minType": "1021", // Register（公开接口，无 token）
	})

	require.NoError(t, err)
	// 无 token → context 中 userID 应为空字符串
	assert.Equal(t, "", auth.GetUserID(mockH.capturedCtx))
}

func TestServant_Handle_过期Token_返回未授权响应(t *testing.T) {
	expiredToken := makeExpiredToken("user-id-001")
	mockH := &mockHandler{}
	servant := newTestServant(mockH)

	retCode, respBytes, err := servant.Handle(context.Background(), []byte("req-body"), map[string]string{
		"minType": "1029",
		"token":   expiredToken,
	})

	require.NoError(t, err)
	assert.Equal(t, 0, retCode) // retCode=0（业务错误不是系统错误）
	// Handler.Dispatch 未被调用（因为 token 无效直接返回）
	assert.Nil(t, mockH.capturedCtx)
	// 解析响应，验证是未授权错误
	var envelope commonpb.ErrorEnvelope
	require.NoError(t, proto.Unmarshal(respBytes, &envelope))
	assert.Equal(t, int32(10401), envelope.Result.Code) // ERROR_CODE_UNAUTHORIZED
}

func TestServant_Handle_无效TokenSignature_返回未授权响应(t *testing.T) {
	invalidToken := "eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoiYWJjIn0.wrong-signature"
	mockH := &mockHandler{}
	servant := newTestServant(mockH)

	retCode, respBytes, _ := servant.Handle(context.Background(), []byte("req"), map[string]string{
		"minType": "1029",
		"token":   invalidToken,
	})
	assert.Equal(t, 0, retCode)
	var envelope commonpb.ErrorEnvelope
	require.NoError(t, proto.Unmarshal(respBytes, &envelope))
	assert.Equal(t, int32(10401), envelope.Result.Code)
}

func TestServant_Handle_缺少MinType_返回系统错误(t *testing.T) {
	servant := newTestServant(&mockHandler{})

	retCode, _, err := servant.Handle(context.Background(), []byte("req"), map[string]string{
		// minType 缺失
		"token": makeValidToken("user-id-001"),
	})

	assert.Error(t, err)
	assert.Equal(t, -1, retCode) // -1 = Tars 系统级错误
}

func TestServant_Handle_TraceID已注入Context(t *testing.T) {
	mockH := &mockHandler{respBytes: []byte("resp")}
	servant := newTestServant(mockH)

	_, _, _ = servant.Handle(context.Background(), []byte("req"), map[string]string{
		"minType":  "1021",
		"trace_id": "tr-custom-12345",
	})

	// 验证自定义 traceId 被注入 context
	assert.Equal(t, "tr-custom-12345", auth.GetTraceID(mockH.capturedCtx))
}

func TestServant_Handle_无TraceID时自动生成(t *testing.T) {
	mockH := &mockHandler{respBytes: []byte("resp")}
	servant := newTestServant(mockH)

	_, _, _ = servant.Handle(context.Background(), []byte("req"), map[string]string{
		"minType": "1021",
		// 不传 trace_id
	})

	// traceId 应自动生成，非空
	traceID := auth.GetTraceID(mockH.capturedCtx)
	assert.NotEmpty(t, traceID)
}
```

---

### 5.3 handler_test.go

```go
// 文件: go/modules/social/member/handler_test.go
// 层级: Handler minType Dispatch 测试
// 覆盖: 正确路由 / proto unmarshal 失败 / 未知 minType / 上下游正确串联

package member

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go/common-lib/auth"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/proto/base"
)

func TestHandler_Dispatch_注册协议路由到RegisterService(t *testing.T) {
	repo := newMockMemberRepo()
	handler := &Handler{
		svcRegister: NewRegisterService(repo),
		// 其余 svc 填 nil（本测试不涉及）
	}
	ctx := auth.WithTraceID(context.Background(), "tr-test-001")

	// 构造 proto bytes
	req := &pb.UserRegisterRequest{
		Username: "testuser",
		Password: "pass123456",
		Email:    "test@example.com",
	}
	reqBytes, _ := proto.Marshal(req)

	respBytes, err := handler.Dispatch(ctx, 1021, reqBytes, nil)
	require.NoError(t, err)
	require.NotNil(t, respBytes)

	// 解析响应
	var resp pb.UserRegisterResponse
	require.NoError(t, proto.Unmarshal(respBytes, &resp))
	assert.Equal(t, int32(10200), resp.Result.Code) // 注册成功
}

func TestHandler_Dispatch_未知MinType返回错误(t *testing.T) {
	handler := &Handler{}
	_, err := handler.Dispatch(context.Background(), 9999, []byte("data"), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown minType=9999")
}

func TestHandler_Dispatch_Proto反序列化失败返回错误(t *testing.T) {
	repo := newMockMemberRepo()
	handler := &Handler{svcRegister: NewRegisterService(repo)}

	// 传入非法 proto bytes
	_, err := handler.Dispatch(context.Background(), 1021, []byte("invalid-proto-bytes"), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "proto unmarshal")
}

func TestHandler_Dispatch_Context透传到Service(t *testing.T) {
	// 验证 handler 透传 ctx 给 svc（不拷贝不修改）
	repo := newMockMemberRepoWithContextCapture()
	handler := &Handler{svcRegister: NewRegisterService(repo)}

	ctx := auth.WithUserID(context.Background(), "user-abc")
	ctx = auth.WithTraceID(ctx, "tr-xyz")
	req := &pb.UserRegisterRequest{Username: "alice", Password: "pass", Email: "a@b.com"}
	reqBytes, _ := proto.Marshal(req)

	handler.Dispatch(ctx, 1021, reqBytes, nil)

	// 验证 repo 收到的 context 包含正确的 userID 和 traceId
	capturedCtx := repo.lastCtx
	assert.Equal(t, "user-abc", auth.GetUserID(capturedCtx))
	assert.Equal(t, "tr-xyz", auth.GetTraceID(capturedCtx))
}
```

---

## 6. Service 层 STDD 单元测试

### 6.1 STDD 规范（Strict TDD）

```
红阶段: 先写测试，运行 → 必须失败（"not implemented" 或编译错误）
绿阶段: 写最小代码让测试通过，不多写一行
重构: 通过后清理代码，测试不变
报告: make test 输出到 docs/reports/testing/
CI:   GitHub Actions 绿，才可以宣称完成
```

**STDD 三禁**：
- 禁止先写实现再补测试
- 禁止 mock 掉核心业务逻辑（参数校验 / 权限 / DB 操作不能全 mock）
- 禁止只测正常路径（每个 svc 至少覆盖: 参数非法 + 权限拒绝 + DB 失败）

---

### 6.2 svc_register_test.go（完整示例）

```go
// 文件: go/modules/social/member/svc_register_test.go
// STDD: 先运行此测试文件确认失败，再实现 svc_register.go

package member

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go/common-lib/auth"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/proto/base"
)

// ── 前置: 构造测试 ctx（模拟 servant.go 注入后的 ctx）──────────────────

func newTestCtx(userID, traceID string) context.Context {
	ctx := context.Background()
	if userID != "" {
		ctx = auth.WithUserID(ctx, userID)
	}
	ctx = auth.WithTraceID(ctx, firstNonEmpty(traceID, "tr-test-default"))
	return ctx
}

// ── 正常路径测试 ──────────────────────────────────────────────────────

func TestRegisterService_正常注册_返回成功(t *testing.T) {
	// 红阶段: 此测试先于 svc_register.go 实现，必须失败
	repo := newMockMemberRepo()
	svc := NewRegisterService(repo)

	resp, err := svc.Handle(newTestCtx("", "tr-001"), &pb.UserRegisterRequest{
		Username: "alice",
		Password: "secure-pass-123",
		Email:    "alice@example.com",
		Phone:    "13812345678",
	})

	require.NoError(t, err, "注册不应返回系统错误")
	require.NotNil(t, resp)
	assert.Equal(t, int32(10200), resp.Result.Code, "注册成功应返回 10200")

	// 验证 DB 写入: user 应已存在于 mock repo
	saved, _ := repo.GetUserByUsername(context.Background(), "alice")
	require.NotNil(t, saved, "用户应已写入 DB")
	assert.Equal(t, "alice", saved.Username)
	assert.NotEmpty(t, saved.ID, "用户 ID 不能为空")
	assert.NotEmpty(t, saved.Salt, "Salt 不能为空")
	assert.NotEqual(t, "secure-pass-123", saved.Password, "密码必须已加密")
	assert.Equal(t, int8(1), saved.Status, "新用户状态应为 active")
	assert.Equal(t, "normal", saved.MembershipLevel)
}

// ── 参数校验测试 ──────────────────────────────────────────────────────

func TestRegisterService_用户名为空_返回BAD_REQUEST(t *testing.T) {
	svc := NewRegisterService(newMockMemberRepo())
	resp, err := svc.Handle(newTestCtx("", "tr-002"), &pb.UserRegisterRequest{
		Username: "",
		Password: "pass123",
		Email:    "test@test.com",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(pb.UserErrorCode_USER_ERROR_NAME_TOO_SHORT), resp.Result.Code)
}

func TestRegisterService_用户名少于4字符_返回NAME_TOO_SHORT(t *testing.T) {
	svc := NewRegisterService(newMockMemberRepo())
	resp, _ := svc.Handle(newTestCtx("", "tr-003"), &pb.UserRegisterRequest{
		Username: "ab",
		Password: "pass123",
		Email:    "test@test.com",
	})
	assert.Equal(t, int32(pb.UserErrorCode_USER_ERROR_NAME_TOO_SHORT), resp.Result.Code)
}

func TestRegisterService_用户名超过50字符_返回NAME_TOO_LONG(t *testing.T) {
	svc := NewRegisterService(newMockMemberRepo())
	longName := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxy" // 51字符
	resp, _ := svc.Handle(newTestCtx("", "tr-004"), &pb.UserRegisterRequest{
		Username: longName,
		Password: "pass123",
		Email:    "test@test.com",
	})
	assert.Equal(t, int32(pb.UserErrorCode_USER_ERROR_NAME_TOO_LONG), resp.Result.Code)
}

// ── 重复性检查测试 ────────────────────────────────────────────────────

func TestRegisterService_用户名已存在_返回NAME_ALREADY_TAKEN(t *testing.T) {
	repo := newMockMemberRepo()
	// 预先插入同名用户
	repo.users["existing-id"] = &User{ID: "existing-id", Username: "alice"}

	svc := NewRegisterService(repo)
	resp, err := svc.Handle(newTestCtx("", "tr-005"), &pb.UserRegisterRequest{
		Username: "alice",
		Password: "pass123",
		Email:    "new@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(pb.UserErrorCode_USER_ERROR_NAME_ALREADY_TAKEN), resp.Result.Code)
}

func TestRegisterService_手机号已存在_返回PHONE_ALREADY_EXISTS(t *testing.T) {
	repo := newMockMemberRepo()
	repo.users["existing-id"] = &User{ID: "existing-id", Username: "bob", Phone: "13812345678"}

	svc := NewRegisterService(repo)
	resp, err := svc.Handle(newTestCtx("", "tr-006"), &pb.UserRegisterRequest{
		Username: "newuser",
		Password: "pass123",
		Email:    "new@example.com",
		Phone:    "13812345678", // 相同手机号
	})
	require.NoError(t, err)
	assert.Equal(t, int32(pb.UserErrorCode_USER_ERROR_PHONE_ALREADY_EXISTS), resp.Result.Code)
}

// ── DB 失败测试（系统错误）────────────────────────────────────────────

func TestRegisterService_CreateUser_DB错误_返回系统Error(t *testing.T) {
	repo := newMockMemberRepoWithError(errors.New("connection refused"))
	svc := NewRegisterService(repo)

	resp, err := svc.Handle(newTestCtx("", "tr-007"), &pb.UserRegisterRequest{
		Username: "alice",
		Password: "pass123",
		Email:    "alice@example.com",
	})

	// DB 错误应返回 error（系统错误），不应返回业务 Response
	assert.Error(t, err, "DB 失败应返回系统 error")
	assert.Nil(t, resp, "DB 失败时 Response 应为 nil") // 或返回包含 500 的 resp，取决于设计
}

// ── 幂等性 / 并发测试 ────────────────────────────────────────────────

func TestRegisterService_并发注册相同用户名_只有一个成功(t *testing.T) {
	repo := newMockMemberRepoWithUniqueConstraint() // 模拟 MySQL UK 约束
	svc := NewRegisterService(repo)

	results := make(chan int32, 2)
	for i := 0; i < 2; i++ {
		go func() {
			resp, _ := svc.Handle(newTestCtx("", "tr-concurrent"), &pb.UserRegisterRequest{
				Username: "alice",
				Password: "pass123",
				Email:    "alice@example.com",
			})
			results <- resp.Result.Code
		}()
	}

	code1 := <-results
	code2 := <-results
	codes := []int32{code1, code2}

	// 断言：一个成功(10200)，一个用户名已占用(10612)
	assert.Contains(t, codes, int32(10200))
	assert.Contains(t, codes, int32(pb.UserErrorCode_USER_ERROR_NAME_ALREADY_TAKEN))
}
```

---

### 6.3 svc_remove_member_test.go（需要 JWT + 权限）

```go
// 文件: go/modules/social/group/svc_remove_member_test.go
// 场景: 圈主/管理员移除成员（依赖 operatorID from JWT context + 权限校验）

package group

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go/common-lib/auth"
	"go/modules/social/permission"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/proto/base"
)

// ── Mock Permission Service ────────────────────────────────────────────

type mockPermissionService struct {
	canManageMemberFn func(ctx context.Context, opID, groupID, targetID string) (bool, error)
}

func (m *mockPermissionService) CanManageMember(ctx context.Context, opID, groupID, targetID string) (bool, error) {
	return m.canManageMemberFn(ctx, opID, groupID, targetID)
}
// 实现 permission.Service 接口其余方法（返回默认值）...

// ── 测试用例 ──────────────────────────────────────────────────────────

func TestRemoveMemberService_圈主移除普通成员_成功(t *testing.T) {
	repo := newMockGroupRepo()
	// 预插入 target 成员
	repo.members["group-001:target-user"] = &GroupMember{
		GroupID: "group-001", UserID: "target-user", Role: "member", Status: 1,
	}

	perm := &mockPermissionService{
		canManageMemberFn: func(_ context.Context, opID, groupID, targetID string) (bool, error) {
			// 圈主可以移除普通成员
			return opID == "owner-user" && groupID == "group-001", nil
		},
	}
	svc := NewRemoveMemberService(repo, perm)

	// 构造 context（模拟 servant 注入的 operatorID）
	ctx := auth.WithUserID(context.Background(), "owner-user")
	ctx = auth.WithTraceID(ctx, "tr-remove-001")

	resp, err := svc.Handle(ctx, &pb.RemoveMemberRequest{
		GroupId: "group-001",
		UserId:  "target-user",
		Reason:  "违规",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10200), resp.Result.Code)

	// 验证成员状态已更新（status=4=removed）
	updated, _ := repo.GetMemberByGroupAndUser(context.Background(), "group-001", "target-user")
	assert.Equal(t, int8(4), updated.Status)

	// 验证审计日志已记录，且 operatorID 来自 JWT（非请求参数）
	action := repo.lastAdminAction
	assert.Equal(t, "owner-user", action.OperatorID, "operatorID 必须来自 JWT context")
	assert.Equal(t, "target-user", action.TargetID)
	assert.Equal(t, "remove_member", action.Action)
}

func TestRemoveMemberService_无Token_返回未授权(t *testing.T) {
	svc := NewRemoveMemberService(newMockGroupRepo(), &mockPermissionService{})

	// ctx 中无 userID（模拟未登录）
	ctx := auth.WithTraceID(context.Background(), "tr-no-auth")

	resp, err := svc.Handle(ctx, &pb.RemoveMemberRequest{
		GroupId: "group-001",
		UserId:  "target-user",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10401), resp.Result.Code) // UNAUTHORIZED
}

func TestRemoveMemberService_权限不足_返回PERMISSION_DENIED(t *testing.T) {
	repo := newMockGroupRepo()
	repo.members["group-001:target-user"] = &GroupMember{
		GroupID: "group-001", UserID: "target-user", Role: "member", Status: 1,
	}

	perm := &mockPermissionService{
		canManageMemberFn: func(_ context.Context, opID, groupID, targetID string) (bool, error) {
			return false, nil // 普通成员无权限
		},
	}
	svc := NewRemoveMemberService(repo, perm)

	ctx := auth.WithUserID(context.Background(), "normal-member")
	resp, err := svc.Handle(ctx, &pb.RemoveMemberRequest{
		GroupId: "group-001",
		UserId:  "target-user",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(pb.GroupErrorCode_GROUP_ERROR_PERMISSION_DENIED), resp.Result.Code)
	// 验证 DB 没有被写入（权限失败不应修改数据）
	assert.Nil(t, repo.lastAdminAction)
}

func TestRemoveMemberService_GroupID为空_返回参数错误(t *testing.T) {
	svc := NewRemoveMemberService(newMockGroupRepo(), &mockPermissionService{})
	ctx := auth.WithUserID(context.Background(), "owner-user")

	resp, err := svc.Handle(ctx, &pb.RemoveMemberRequest{
		GroupId: "", // 空 GroupID
		UserId:  "target-user",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(pb.GroupErrorCode_GROUP_ERROR_ID_EMPTY), resp.Result.Code)
}

func TestRemoveMemberService_不能移除自己(t *testing.T) {
	repo := newMockGroupRepo()
	perm := &mockPermissionService{
		canManageMemberFn: func(_ context.Context, _, _, _ string) (bool, error) {
			return true, nil
		},
	}
	svc := NewRemoveMemberService(repo, perm)

	ctx := auth.WithUserID(context.Background(), "owner-user")
	resp, err := svc.Handle(ctx, &pb.RemoveMemberRequest{
		GroupId: "group-001",
		UserId:  "owner-user", // 移除自己
	})
	require.NoError(t, err)
	assert.Equal(t, int32(pb.GroupErrorCode_GROUP_ERROR_CANNOT_REMOVE_OWNER), resp.Result.Code)
}
```

---

### 6.4 Mock Repository 完整实现

```go
// 文件: go/modules/social/member/mock_repository_test.go
// 仅在 _test.go 包内使用

package member

import (
	"context"
	"errors"
	"sync"
)

// ── 标准 Mock（无约束）────────────────────────────────────────────────

type mockMemberRepository struct {
	mu      sync.Mutex
	users   map[string]*User
	blocks  map[string]*MemberBlock
	lastCtx context.Context // 供测试验证 ctx 透传
	dbErr   error           // 注入 DB 错误
}

func newMockMemberRepo() *mockMemberRepository {
	return &mockMemberRepository{
		users:  make(map[string]*User),
		blocks: make(map[string]*MemberBlock),
	}
}

// newMockMemberRepoWithError 注入 DB 错误（测试 DB 失败场景）
func newMockMemberRepoWithError(err error) *mockMemberRepository {
	r := newMockMemberRepo()
	r.dbErr = err
	return r
}

// newMockMemberRepoWithContextCapture 捕获 ctx（验证 context 透传）
func newMockMemberRepoWithContextCapture() *mockMemberRepository {
	return newMockMemberRepo()
}

// newMockMemberRepoWithUniqueConstraint 模拟 MySQL UK 约束
func newMockMemberRepoWithUniqueConstraint() *mockMemberRepository {
	return newMockMemberRepo() // 已在 ExistsByUsername 中模拟唯一约束
}

func (m *mockMemberRepository) CreateUser(ctx context.Context, user *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastCtx = ctx
	if m.dbErr != nil {
		return m.dbErr
	}
	// 模拟 UK 约束: username 唯一
	for _, u := range m.users {
		if u.Username == user.Username {
			return errors.New("Error 1062: Duplicate entry")
		}
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockMemberRepository) GetUserByID(ctx context.Context, userID string) (*User, error) {
	m.lastCtx = ctx
	if m.dbErr != nil {
		return nil, m.dbErr
	}
	return m.users[userID], nil
}

func (m *mockMemberRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	m.lastCtx = ctx
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockMemberRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockMemberRepository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	for _, u := range m.users {
		if u.Phone == phone {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockMemberRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if m.dbErr != nil {
		return false, m.dbErr
	}
	for _, u := range m.users {
		if u.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockMemberRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.dbErr != nil {
		return false, m.dbErr
	}
	for _, u := range m.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockMemberRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	if m.dbErr != nil {
		return false, m.dbErr
	}
	for _, u := range m.users {
		if u.Phone == phone {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockMemberRepository) UpdateUser(ctx context.Context, userID string, fields map[string]interface{}) error {
	if m.dbErr != nil {
		return m.dbErr
	}
	u, ok := m.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	for k, v := range fields {
		switch k {
		case "nickname":
			u.Nickname = v.(string)
		case "avatar":
			u.Avatar = v.(string)
		}
	}
	return nil
}

func (m *mockMemberRepository) UpdateUserStatus(ctx context.Context, userID string, status int8) error {
	if m.dbErr != nil {
		return m.dbErr
	}
	if u, ok := m.users[userID]; ok {
		u.Status = status
	}
	return nil
}

func (m *mockMemberRepository) BatchGetUsers(ctx context.Context, userIDs []string) ([]*User, error) {
	var result []*User
	for _, id := range userIDs {
		if u, ok := m.users[id]; ok {
			result = append(result, u)
		}
	}
	return result, m.dbErr
}

func (m *mockMemberRepository) CreateBlock(ctx context.Context, block *MemberBlock) error {
	key := block.BlockerID + ":" + block.BlockedID
	m.blocks[key] = block
	return m.dbErr
}

func (m *mockMemberRepository) DeleteBlock(ctx context.Context, blockerID, blockedID string) error {
	delete(m.blocks, blockerID+":"+blockedID)
	return m.dbErr
}

func (m *mockMemberRepository) IsBlocked(ctx context.Context, userA, userB string) (bool, error) {
	_, ok1 := m.blocks[userA+":"+userB]
	_, ok2 := m.blocks[userB+":"+userA]
	return ok1 || ok2, m.dbErr
}

func (m *mockMemberRepository) GetBlockList(ctx context.Context, userID string, page, pageSize int32) ([]*MemberBlock, int64, error) {
	var result []*MemberBlock
	for _, b := range m.blocks {
		if b.BlockerID == userID {
			result = append(result, b)
		}
	}
	return result, int64(len(result)), m.dbErr
}

func (m *mockMemberRepository) GetUserStats(ctx context.Context, userID string) (*MemberStats, error) {
	return &MemberStats{UserID: userID}, m.dbErr
}

func (m *mockMemberRepository) IncrStat(ctx context.Context, userID, field string, delta int64) error {
	return m.dbErr
}
```

---

### 6.5 make test 运行命令与报告

```makefile
# 文件: Makefile 追加（按项目现有 Makefile 规范）

# 运行社交域单元测试
test-social:
	go test ./go/modules/social/... -v -count=1 \
		-coverprofile=docs/reports/testing/social-coverage.out \
		2>&1 | tee docs/reports/testing/social-test.log

# 生成覆盖率报告
test-social-coverage:
	go tool cover -html=docs/reports/testing/social-coverage.out \
		-o docs/reports/html/social-coverage.html

# 运行单个 svc 测试（开发中快速验证）
# 用法: make test-svc PKG=member SVC=register
test-svc:
	go test ./go/modules/social/$(PKG)/... -v -run TestRegisterService \
		-count=1

# CI 全量测试（包含现有测试）
ci:
	make test
	make test-social
```

---

*文档结束 — DEV-STDD-SOCIAL-002 v1.0*
*如有修改请同步更新 docs/wiki/CODE-WIKI.md 中社交域章节*
