package tarsclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/common-lib/module"
	"github.com/jimiechen/mineplanet/go/modules/health"
	"github.com/jimiechen/mineplanet/go/modules/hello"
	socialmodule "github.com/jimiechen/mineplanet/go/modules/social"
	"github.com/jimiechen/mineplanet/go/modules/social/group"
	"github.com/jimiechen/mineplanet/go/modules/social/member"
	"github.com/jimiechen/mineplanet/go/modules/social/topic"
	configdomain "github.com/jimiechen/mineplanet/go/services/config/domain"
	configservice "github.com/jimiechen/mineplanet/go/services/config/service"
	i18ndomain "github.com/jimiechen/mineplanet/go/services/i18n/domain"
	i18nservice "github.com/jimiechen/mineplanet/go/services/i18n/service"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Target 定义 Tars 调用目标
type Target struct {
	App       string
	Server    string
	Servant   string
	Module    string
	Interface string
	Method    string
}

// TargetKey 生成目标唯一键
type TargetKey struct {
	App     string
	Server  string
	Servant string
	Method  string
}

// String 返回目标键字符串
func (tk TargetKey) String() string {
	return tk.App + "." + tk.Server + "." + tk.Servant + "." + tk.Method
}

// ToTargetKey 从 Target 生成 TargetKey
func ToTargetKey(t Target) TargetKey {
	return TargetKey{
		App:     t.App,
		Server:  t.Server,
		Servant: t.Servant,
		Method:  t.Method,
	}
}

// TarsInvoker 定义统一调用接口
type TarsInvoker interface {
	Invoke(ctx context.Context, target Target, request []byte, extend map[string]string) (returnCode int, response []byte, err error)
}

// LocalHandler 定义本地 handler 接口
type LocalHandler interface {
	Invoke(ctx context.Context, request []byte, extend map[string]string) (returnCode int, response []byte, err error)
}

// LocalInvoker 单体部署（monolith）模式下的本进程 TarsGo servant adapter
// 不绕过 Tars 框架，而是在同一部署单元内通过进程内调用转发到 TarsGo servant
// 严格遵守 Tars bytes 契约：request/response 均为 Protobuf bytes
type LocalInvoker struct {
	handlers map[string]LocalHandler
}

// NewLocalInvoker 创建 LocalInvoker
func NewLocalInvoker() *LocalInvoker {
	return &LocalInvoker{
		handlers: make(map[string]LocalHandler),
	}
}

// HandlersForTest 返回已注册的 handler key 列表（仅用于测试诊断）
func (li *LocalInvoker) HandlersForTest() map[string]LocalHandler {
	return li.handlers
}

// buildMinimalDeps 构建最小依赖集，用于 Gateway 单体部署模式的模块初始化
// 仅提供 no-op 的 Config 和 Logger 实现，I18n/DB/Cache 为 nil
func buildMinimalDeps() module.Deps {
	return module.Deps{
		Config: &noopConfigReader{},
		Logger: &noopLogger{},
	}
}

// noopConfigReader 空实现的 ConfigReader，用于无外部配置中心的本地开发
type noopConfigReader struct{}

func (n *noopConfigReader) GetString(_ context.Context, _, _ string) (string, error) { return "", nil }
func (n *noopConfigReader) GetInt(_ context.Context, _, _ string) (int64, error)     { return 0, nil }
func (n *noopConfigReader) GetBool(_ context.Context, _, _ string) (bool, error)     { return false, nil }
func (n *noopConfigReader) Watch(_ context.Context, _ string, _ func(string, interface{}, interface{})) error {
	return nil
}
func (n *noopConfigReader) Ping(_ context.Context) error { return nil }

// noopLogger 空实现的 Logger，输出到 stdout（仅用于本地开发联调）
type noopLogger struct{}

func (l *noopLogger) Info(_ context.Context, v ...interface{})             { /* no-op for local dev */ }
func (l *noopLogger) Infof(_ context.Context, _ string, _ ...interface{})  {}
func (l *noopLogger) Error(_ context.Context, v ...interface{})            { /* no-op for local dev */ }
func (l *noopLogger) Errorf(_ context.Context, _ string, _ ...interface{}) {}
func (l *noopLogger) Warn(_ context.Context, v ...interface{})             { /* no-op for local dev */ }
func (l *noopLogger) Debug(_ context.Context, v ...interface{})            { /* no-op for local dev */ }

// noopConfigService 空实现的 ConfigService，用于无外部依赖的本地开发模式
// 返回空配置响应，不连接 SQLite / 外部配置中心
type noopConfigService struct{}

func (s *noopConfigService) GetAppConfigs(req *configservice.AppConfigRequest) (*configservice.AppConfigResponse, error) {
	return &configservice.AppConfigResponse{
		StaticModules:  make(map[string]map[string]*configdomain.TypedValue),
		DynamicModules: []*configservice.DynamicModuleView{},
	}, nil
}

func (s *noopConfigService) GetVersionInfo(env string, knownVersions map[string]int64) (*configservice.VersionInfoResponse, error) {
	return &configservice.VersionInfoResponse{
		ConfigVersions: make(map[string]int64),
		HasChanges:     false,
	}, nil
}

// noopI18nService 空实现的 I18nService，用于无外部依赖的本地开发模式
// 返回默认语言列表和空语言包，不连接 SQLite
type noopI18nService struct{}

func (s *noopI18nService) GetLanguages(clientVersion string) ([]i18nservice.LanguageMeta, error) {
	return []i18nservice.LanguageMeta{
		{Code: "zh-CN", Name: "简体中文", NativeName: "简体中文", IsDefault: true},
		{Code: "en-US", Name: "English", NativeName: "English", IsDefault: false},
	}, nil
}

func (s *noopI18nService) GetLangPack(langCode, clientVersion, env string) (*i18nservice.LangPackResponse, error) {
	return &i18nservice.LangPackResponse{
		PackVersion: 1,
		Strings:     []i18nservice.LangStringEntry{},
	}, nil
}

func (s *noopI18nService) GetLangDifference(langCode string, sinceVersion int64, clientVersion, env string) (*i18nservice.LangDiffResponse, error) {
	return &i18nservice.LangDiffResponse{
		CurrentVersion: sinceVersion,
		Additions:      []i18nservice.LangStringEntry{},
		Deletions:      []string{},
	}, nil
}

func (s *noopI18nService) ValidateTemplate(value string, templateType i18ndomain.TemplateType, params []i18ndomain.LangParam) error {
	return nil // noop: always pass
}

// Register 注册本地 handler
func (li *LocalInvoker) Register(key TargetKey, handler LocalHandler) {
	li.handlers[key.String()] = handler
}

// Invoke 执行本地调用
func (li *LocalInvoker) Invoke(ctx context.Context, target Target, request []byte, extend map[string]string) (int, []byte, error) {
	key := ToTargetKey(target).String()
	handler, ok := li.handlers[key]
	if !ok {
		return 10404, nil, errors.New("local handler not found: " + key)
	}
	return handler.Invoke(ctx, request, extend)
}

// ModuleInvokeFunc 模块服务调用函数签名
// 业务模块统一使用 Protobuf bytes 作为输入输出，不依赖 MessagePacket
//
// Deprecated: 请使用 commonlib.ModuleInvokeFunc 替代。
type ModuleInvokeFunc = commonlib.ModuleInvokeFunc

// moduleHandler 将模块服务适配为 LocalHandler 接口
// 负责返回码转换：模块成功→10200，模块失败→10500
type moduleHandler struct {
	fn ModuleInvokeFunc
}

// NewModuleHandler 创建模块适配 handler
func NewModuleHandler(fn ModuleInvokeFunc) LocalHandler {
	return &moduleHandler{fn: fn}
}

// Invoke 实现 LocalHandler 接口，调用底层模块服务并转换返回码
func (h *moduleHandler) Invoke(ctx context.Context, request []byte, extend map[string]string) (int, []byte, error) {
	resp, err := h.fn(ctx, request)
	if err != nil {
		return commonlib.CodeInternalError, nil, err
	}
	return commonlib.CodeSuccess, resp, nil
}

// RegisterModuleHandlers 注册模块化业务服务的本地 handler
// 每个模块独立注册到对应的 TargetKey，通过 NewModuleHandler 适配接口
// 注册 System 模块（HealthCheck + HelloWorld）的本地 handler
func RegisterModuleHandlers(invoker *LocalInvoker) {
	deps := buildMinimalDeps()

	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "SystemServer",
		Servant: "SystemObj",
		Method:  "HealthCheck",
	}, NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		healthSvc := health.New(deps, nil)
		return healthSvc.Check(ctx, req)
	}))

	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "SystemServer",
		Servant: "SystemObj",
		Method:  "HelloWorld",
	}, NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		helloSvc := hello.New(deps)
		return helloSvc.SayHello(ctx, req)
	}))
}

// TarsGoInvoker 微服务部署（microservice）模式下的远程 TarsGo client invoker
// 通过 TarsGo client 远程调用独立部署的 TarsCloud servant
// 与 LocalInvoker 共享同一 TarsInvoker 接口，严格遵守 Tars bytes 契约
// S1 阶段未实现，当前调用返回 10500 错误
type TarsGoInvoker struct{}

// NewTarsGoInvoker 创建 TarsGoInvoker
func NewTarsGoInvoker() *TarsGoInvoker {
	return &TarsGoInvoker{}
}

// Invoke 执行远程 TarsGo 调用（S1 未实现）
func (ti *TarsGoInvoker) Invoke(ctx context.Context, target Target, request []byte, extend map[string]string) (int, []byte, error) {
	return 10500, nil, errors.New("tars invoker is not implemented yet")
}

// RegisterConfigI18nHandlers 注册 Config 和 I18n 模块的本地 TarsGo servant handler
// Gateway 单体部署模式启动时必须调用，注册 6001-6010 协议的所有方法
//
// 注册的路由：
// - 6001/6002: GetAppConfigs (ConfigServer.ConfigObj)
// - 6009/6010: AppConfigVersion (ConfigServer.ConfigObj)
// - 6003/6004: GetAppLanguage (I18nServer.I18nObj)
// - 6005/6006: GetLangPack (I18nServer.I18nObj)
// - 6007/6008: GetLangDifference (I18nServer.I18nObj)
func RegisterConfigI18nHandlers(invoker *LocalInvoker, configSvc configservice.ConfigService, i18nSvc i18nservice.I18nService) {
	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "ConfigServer",
		Servant: "ConfigObj",
		Method:  "GetAppConfigs",
	}, NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		var appReq configservice.AppConfigRequest
		if err := json.Unmarshal(req, &appReq); err != nil {
			return nil, err
		}
		resp, err := configSvc.GetAppConfigs(&appReq)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}))

	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "ConfigServer",
		Servant: "ConfigObj",
		Method:  "AppConfigVersion",
	}, NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		var versionReq struct {
			Env           string           `json:"env"`
			KnownVersions map[string]int64 `json:"known_versions"`
		}
		if err := json.Unmarshal(req, &versionReq); err != nil {
			return nil, err
		}
		resp, err := configSvc.GetVersionInfo(versionReq.Env, versionReq.KnownVersions)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}))

	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "I18nServer",
		Servant: "I18nObj",
		Method:  "GetAppLanguage",
	}, NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		var langReq struct {
			ClientVersion string `json:"client_version"`
		}
		if err := json.Unmarshal(req, &langReq); err != nil {
			return nil, err
		}
		languages, err := i18nSvc.GetLanguages(langReq.ClientVersion)
		if err != nil {
			return nil, err
		}
		return json.Marshal(languages)
	}))

	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "I18nServer",
		Servant: "I18nObj",
		Method:  "GetLangPack",
	}, NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		var packReq struct {
			LangCode      string `json:"lang_code"`
			ClientVersion string `json:"client_version"`
			Env           string `json:"env"`
		}
		if err := json.Unmarshal(req, &packReq); err != nil {
			return nil, err
		}
		resp, err := i18nSvc.GetLangPack(packReq.LangCode, packReq.ClientVersion, packReq.Env)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}))

	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "I18nServer",
		Servant: "I18nObj",
		Method:  "GetLangDifference",
	}, NewModuleHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		var diffReq struct {
			LangCode      string `json:"lang_code"`
			SinceVersion  int64  `json:"since_version"`
			ClientVersion string `json:"client_version"`
			Env           string `json:"env"`
		}
		if err := json.Unmarshal(req, &diffReq); err != nil {
			return nil, err
		}
		resp, err := i18nSvc.GetLangDifference(diffReq.LangCode, diffReq.SinceVersion, diffReq.ClientVersion, diffReq.Env)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}))
}

// RegisterAllLocalHandlers 注册所有本地 handler（System + Config + I18n + Social）
// 使用 noop stub 作为 Config/I18n 服务实现，无需外部依赖
// Social 模块使用 Memory Repository，无需 MySQL 连接
// Gateway 单体部署模式（local）启动时调用此函数即可完成全部注册
func RegisterAllLocalHandlers(invoker *LocalInvoker) {
	RegisterModuleHandlers(invoker)
	RegisterConfigI18nHandlers(invoker, &noopConfigService{}, &noopI18nService{})
	RegisterSocialHandlers(invoker)
}

// RegisterSocialHandlers 注册 Social 域（Member + Group + Topic）的本地 TarsGo servant handler
// 优先使用 GORM+Redis 直连真实数据库（通过 MYSQL_HOST 环境变量触发）
// 未设置 MYSQL_HOST 时 fallback 到 Memory Repository（保持单元测试兼容性）
//
// 环境变量：
//   - MYSQL_HOST / MYSQL_PORT / MYSQL_USER / MYSQL_PASSWORD / MYSQL_DATABASE：MySQL 连接
//   - REDIS_HOST / REDIS_PORT / REDIS_DB：Redis 连接（可选，未设置则用 MemoryTokenStore）
//
// 注册的 TargetKey：
//   - CaiRobot.SocialServer.SocialObj.HandleMember  → Member Servant（1000 段协议）
//   - CaiRobot.SocialServer.SocialObj.HandleGroup   → Group Servant（2000 段协议）
//   - CaiRobot.SocialServer.SocialObj.HandleTopic    → Topic Servant（3000 段协议）
func RegisterSocialHandlers(invoker *LocalInvoker) {
	// 根据环境变量选择 Repository 实现
	var memberRepo member.Repository
	var groupRepo group.Repository
	var topicRepo topic.Repository

	mysqlHost := os.Getenv("MYSQL_HOST")
	if mysqlHost != "" {
		// 直连 MySQL 模式
		db := connectMySQL()
		memberRepo = member.NewGormRepository(db)
		groupRepo = group.NewGormRepository(db)
		topicRepo = topic.NewGormRepository(db)
		fmt.Printf("social: 使用 GORM+MySQL 直连模式 (host=%s)\n", mysqlHost)
	} else {
		// Memory fallback（无 MYSQL_HOST 时的兼容模式）
		memberRepo = member.NewMemoryRepository()
		groupRepo = group.NewMemoryRepository()
		topicRepo = topic.NewMemoryRepository()
		fmt.Println("social: 使用 Memory Repository 模式（未设置 MYSQL_HOST）")
	}

	// 创建 JWT 管理器（UserLogin/UserLogout/RefreshToken 必需）
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "cairobot-mvp-p0-dev-secret-key-32bytes-min!!"
	}
	jwtMgr, err := member.NewJWTManager(
		member.DefaultJWTConfig().SetSecretKey(jwtSecret),
	)
	if err != nil {
		panic(fmt.Sprintf("social: 创建 JWTManager 失败: %v", err))
	}

	// 创建 Social 聚合模块（注入 JWTManager 解决 UserLogin nil pointer panic）
	socialMod := socialmodule.NewModule(
		memberRepo, groupRepo, topicRepo,
		socialmodule.WithJWTManager(jwtMgr),
	)

	// 注入 TokenStore（仅 MySQL 模式下连接 Redis）
	if mysqlHost != "" {
		redisClient := connectRedis()
		if redisClient != nil {
			tokenStore := member.NewRedisTokenStore(redisClient, "social:test:tl:")
			socialMod.MemberServant.InjectTokenStore(tokenStore)
			fmt.Println("social: RedisTokenStore 已注入 MemberServant")
		}
	}

	// 注册 Member 域 handler（1000 段：1021-1046）
	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "SocialServer",
		Servant: "SocialObj",
		Method:  "HandleMember",
	}, &servantHandler{servant: socialMod.MemberServant})

	// 注册 Group 域 handler（2000 段：2005-2078）
	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "SocialServer",
		Servant: "SocialObj",
		Method:  "HandleGroup",
	}, &servantHandler{servant: socialMod.GroupServant})

	// 注册 Topic 域 handler（3000 段：3001-3066）
	invoker.Register(TargetKey{
		App:     "CaiRobot",
		Server:  "SocialServer",
		Servant: "SocialObj",
		Method:  "HandleTopic",
	}, &servantHandler{servant: socialMod.TopicServant})
}

// connectMySQL 从环境变量构建 MySQL DSN 并返回 GORM DB 实例
// 环境变量：MYSQL_HOST, MYSQL_PORT, MYSQL_USER, MYSQL_PASSWORD, MYSQL_DATABASE
func connectMySQL() *gorm.DB {
	host := getEnv("MYSQL_HOST", "127.0.0.1")
	port := getEnv("MYSQL_PORT", "3306")
	user := getEnv("MYSQL_USER", "root")
	pass := getEnv("MYSQL_PASSWORD", "")
	dbname := getEnv("MYSQL_DATABASE", "go_biz")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbname)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("social: MySQL 连接失败 (%s): %v", dsn, err))
	}
	return db
}

// connectRedis 从环境变量创建 Redis Client
// 环境变量：REDIS_HOST, REDIS_PORT, REDIS_DB
// 连接失败时返回 nil（不阻断启动，Logout/Refresh 将降级为内存行为）
func connectRedis() *redis.Client {
	host := getEnv("REDIS_HOST", "")
	if host == "" {
		return nil
	}
	port := getEnv("REDIS_PORT", "6379")
	dbNum := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		fmt.Sscanf(dbStr, "%d", &dbNum)
	}

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
		DB:   dbNum,
	})

	// 快速连通性检测（不阻断启动）
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Printf("social: 警告 - Redis 连通性检测失败: %v（Token 黑名单功能降级）\n", err)
		return nil
	}
	return client
}

// getEnv 读取环境变量，不存在时返回默认值
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// servantHandler 将 Social Servant 适配为 LocalHandler 接口
// Servant.Handle 签名与 LocalHandler.Invoke 完全匹配，直接委托调用
type servantHandler struct {
	servant interface {
		Handle(ctx context.Context, req []byte, extend map[string]string) (int, []byte, error)
	}
}

func (h *servantHandler) Invoke(ctx context.Context, request []byte, extend map[string]string) (int, []byte, error) {
	return h.servant.Handle(ctx, request, extend)
}
