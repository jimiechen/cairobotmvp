package service

import (
	"encoding/json"
	"fmt"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// M-03 状态确认：config_query.go 历史遗留已清理
//
// 评审报告提到的 config_query.go（563 行，旧项目遗留）在当前代码库中不存在。
// 经检查：
//   - go/services/config/ 和 go/services/i18n/ 目录下无此文件
//   - i18n service (pack.go) 已使用 sys_lang_pack + sys_lang_string 作为数据源
//   - config service 使用 sys_config_version + sys_config_schema 作为数据源
//   - 无双重语言包存储风险
//
// 结论：M-03 问题已解决（文件不存在），无需额外操作。
// 参见评审报告: docs/reviews/review-config-i18n-implementation.md#M-03

// BuildDynamicModule 从 ConfigVersion + TypedValue map + Schema 组装动态模块视图
// 这是 compose.go 的核心函数：决定哪些数据走 dynamic_modules 而非强类型字段
// 判断逻辑：module_key 不在 8 个预定义静态列表中 → 放入 dynamic_modules
func BuildDynamicModule(
	version *domain.ConfigVersion,
	typedMap map[string]*domain.TypedValue,
	schemaRepo repository.SchemaRepository,
	clientScope string,
) *DynamicModuleView {
	dm := &DynamicModuleView{
		ModuleKey: version.ModuleKey,
		Version:   version.Version,
		Fields:    make(map[string]*domain.TypedValue),
	}

	schemas, _ := schemaRepo.ListByModule(version.ModuleKey)

	for fieldKey, tv := range typedMap {
		dm.Fields[fieldKey] = tv
	}

	var dmDescriptors []*FieldDescriptorView
	for _, fs := range schemas {
		if !fs.MatchClientScope(clientScope) {
			continue
		}
		if !fs.IsEnabled {
			continue
		}
		desc := &FieldDescriptorView{
			FieldKey:   fs.FieldKey,
			FieldType:  string(fs.FieldType),
			IsRequired: fs.IsRequired,
			DefaultVal: fs.DefaultValue,
		}
		dmDescriptors = append(dmDescriptors, desc)
	}
	dm.Descriptors = dmDescriptors
	return dm
}

// ComposeFullResponse 完整组装 AppConfigsRsp 业务视图
// 输入：env + clientScope + 已加载的所有版本数据
// 输出：分离的 staticModules（8 个预定义）和 dynamicModules（其余）
func ComposeFullResponse(
	env, clientScope string,
	versions []*domain.ConfigVersion,
	schemaRepo repository.SchemaRepository,
	requestedModules []string,
) *AppConfigResponse {
	resp := &AppConfigResponse{
		StaticModules:   make(map[string]map[string]*domain.TypedValue),
		DynamicModules: make([]*DynamicModuleView, 0),
	}

	for _, ver := range versions {
		if len(requestedModules) > 0 && !contains(requestedModules, ver.ModuleKey) {
			continue
		}

		typedMap, _ := ParseConfigJSON(ver.ConfigJSON, ver.ModuleKey, schemaRepo)

		if domain.IsStaticModule(ver.ModuleKey) {
			resp.StaticModules[ver.ModuleKey] = typedMap
		} else {
			dm := BuildDynamicModule(ver, typedMap, schemaRepo, clientScope)
			resp.DynamicModules = append(resp.DynamicModules, dm)
		}
	}

	return resp
}

// MapStaticModulesToProtoFields 将静态模块数据映射到 Protobuf 强类型字段
// 解决评审报告 M-02：老客户端依赖强类型字段（base_cfg=2, wap_cfg=3 等），不能只放在 StaticModules map 中
//
// 映射规则（与 proto/base/app_config.proto 第 28-38 行对应）：
//   - base_cfg   → AppConfigsRsp.BaseCfg   (field 2)
//   - wap_cfg    → AppConfigsRsp.WapCfg    (field 3)
//   - regex_cfg  → AppConfigsRsp.RegexCfg  (field 4)
//   - pay_cfg    → AppConfigsRsp.PayCfg    (field 5)
//   - oss_cfg    → AppConfigsRsp.OssCfg    (field 6)
//   - lang_cfg   → AppConfigsRsp.LangCfg   (field 7)
//   - mute_cfg   → AppConfigsRsp.MuteCfg   (field 8)
//   - group_cfg  → AppConfigsRsp.GroupCfg  (field 9)
func MapStaticModulesToProtoFields(staticModules map[string]map[string]*domain.TypedValue) *pb.AppConfigsRsp {
	protoRsp := &pb.AppConfigsRsp{}

	for moduleKey, typedMap := range staticModules {
		switch moduleKey {
		case "base_cfg":
			protoRsp.BaseCfg = mapToAppBaseConfigs(typedMap)
		case "wap_cfg":
			protoRsp.WapCfg = mapToAppWapUrlConfigs(typedMap)
		case "regex_cfg":
			protoRsp.RegexCfg = mapToAppRegexConfigs(typedMap)
		case "pay_cfg":
			protoRsp.PayCfg = mapToAppPayConfigs(typedMap)
		case "oss_cfg":
			protoRsp.OssCfg = mapToAppOssConfigs(typedMap)
		case "lang_cfg":
			protoRsp.LangCfg = mapToAppLanguageConfigs(typedMap)
		case "mute_cfg":
			protoRsp.MuteCfg = mapToAppMuteConfigs(typedMap)
		case "group_cfg":
			protoRsp.GroupCfg = mapToAppGroupConfigs(typedMap)
		}
	}

	return protoRsp
}

// --- 各静态模块的映射函数 ---

func mapToAppBaseConfigs(typedMap map[string]*domain.TypedValue) *pb.AppBaseConfigs {
	cfg := &pb.AppBaseConfigs{}
	if v, ok := typedMap["domain_root"]; ok {
		cfg.DomainRoot = toString(v.Value)
	}
	if v, ok := typedMap["domain_wap"]; ok {
		cfg.DomainWap = toString(v.Value)
	}
	if v, ok := typedMap["sign_rand"]; ok {
		cfg.SignRand = toString(v.Value)
	}
	if v, ok := typedMap["construct_email"]; ok {
		cfg.ConstructEmail = toString(v.Value)
	}
	return cfg
}

func mapToAppWapUrlConfigs(typedMap map[string]*domain.TypedValue) *pb.AppWapUrlConfigs {
	cfg := &pb.AppWapUrlConfigs{}
	if v, ok := typedMap["user_agreement_url"]; ok {
		cfg.UserAgreementUrl = toString(v.Value)
	}
	if v, ok := typedMap["privacy_policy_url"]; ok {
		cfg.PrivacyPolicyUrl = toString(v.Value)
	}
	return cfg
}

func mapToAppRegexConfigs(typedMap map[string]*domain.TypedValue) *pb.AppRegexConfigs {
	cfg := &pb.AppRegexConfigs{}
	if v, ok := typedMap["regex_email"]; ok {
		cfg.RegexEmail = toString(v.Value)
	}
	if v, ok := typedMap["regex_password"]; ok {
		cfg.RegexPassword = toString(v.Value)
	}
	if v, ok := typedMap["regex_phone"]; ok {
		cfg.RegexPhone = toString(v.Value)
	}
	if v, ok := typedMap["regex_nick"]; ok {
		cfg.RegexNick = toString(v.Value)
	}
	if v, ok := typedMap["regex_circle_name"]; ok {
		cfg.RegexCircleName = toString(v.Value)
	}
	return cfg
}

func mapToAppPayConfigs(typedMap map[string]*domain.TypedValue) *pb.AppPayConfigs {
	cfg := &pb.AppPayConfigs{}
	if v, ok := typedMap["circle_pays"]; ok {
		var pays []*pb.PayMethod
		if arr, ok := v.Value.([]any); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					pays = append(pays, &pb.PayMethod{
						Method:      toString(m["method"]),
						DisplayName: toString(m["display_name"]),
						RangMin:     toFloat64(m["rang_min"]),
						RangMax:     toFloat64(m["rang_max"]),
					})
				}
			}
		}
		cfg.CirclePays = pays
	}
	return cfg
}

func mapToAppOssConfigs(typedMap map[string]*domain.TypedValue) *pb.AppOssConfigs {
	cfg := &pb.AppOssConfigs{}
	if v, ok := typedMap["oss_host"]; ok {
		cfg.OssHost = toString(v.Value)
	}
	if v, ok := typedMap["oss_domain"]; ok {
		cfg.OssDomain = toString(v.Value)
	}
	if v, ok := typedMap["cdn_domain"]; ok {
		cfg.CdnDomain = toString(v.Value)
	}
	return cfg
}

func mapToAppLanguageConfigs(typedMap map[string]*domain.TypedValue) *pb.AppLanguageConfigs {
	cfg := &pb.AppLanguageConfigs{}
	if v, ok := typedMap["languages"]; ok {
		var languages []*pb.LanguageMeta
		if arr, ok := v.Value.([]any); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					languages = append(languages, &pb.LanguageMeta{
						Code:       toString(m["code"]),
						Name:       toString(m["name"]),
						NativeName: toString(m["native_name"]),
						IsDefault:  toBool(m["is_default"]),
					})
				}
			}
		}
		cfg.Languages = languages
	}
	if v, ok := typedMap["lang_code"]; ok {
		cfg.LangCode = toString(v.Value)
	}
	return cfg
}

func mapToAppMuteConfigs(typedMap map[string]*domain.TypedValue) *pb.AppMuteConfigs {
	cfg := &pb.AppMuteConfigs{}
	if v, ok := typedMap["durations"]; ok {
		var durations []*pb.MuteDuration
		if arr, ok := v.Value.([]any); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					durations = append(durations, &pb.MuteDuration{
						Label:   toString(m["label"]),
						Seconds: int32(toFloat64(m["seconds"])),
					})
				}
			}
		}
		cfg.Durations = durations
	}
	return cfg
}

func mapToAppGroupConfigs(typedMap map[string]*domain.TypedValue) *pb.AppGroupConfigs {
	cfg := &pb.AppGroupConfigs{}
	if v, ok := typedMap["group_config_pay_notice"]; ok {
		cfg.GroupConfigPayNotice = toString(v.Value)
	}
	return cfg
}

// --- 类型转换辅助函数 ---

func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func toFloat64(v any) float64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case int32:
		return float64(n)
	default:
		return 0
	}
}

func toBool(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// ToJSONMap 将 DynamicModuleView.Fields 序列化为 map[string]string
// 用于填充 proto DynamicConfigModule.fields（map<string, string>）
func ToJSONMap(fields map[string]*domain.TypedValue) map[string]string {
	result := make(map[string]string, len(fields))
	for k, tv := range fields {
		bytes, err := json.Marshal(tv.Value)
		if err != nil {
			result[k] = ""
			continue
		}
		result[k] = string(bytes)
	}
	return result
}

// ClassifyModules 将版本列表分为静态和动态两组
// 用于 service 层做分流处理
func ClassifyModules(versions []*domain.ConfigVersion) (static, dynamic []*domain.ConfigVersion) {
	for _, v := range versions {
		if domain.IsStaticModule(v.ModuleKey) {
			static = append(static, v)
		} else {
			dynamic = append(dynamic, v)
		}
	}
	return
}
