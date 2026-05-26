package service

import (
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// 以下函数为各静态模块的 TypedValue → Protobuf 映射器
// 每个 mapper 从 TypedValue map 中提取字段值并构造对应的 proto message

func newAppBaseConfigsMapper(typedMap map[string]*domain.TypedValue) ProtoMessage {
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

func newAppWapUrlConfigsMapper(typedMap map[string]*domain.TypedValue) ProtoMessage {
	cfg := &pb.AppWapUrlConfigs{}
	if v, ok := typedMap["user_agreement_url"]; ok {
		cfg.UserAgreementUrl = toString(v.Value)
	}
	if v, ok := typedMap["privacy_policy_url"]; ok {
		cfg.PrivacyPolicyUrl = toString(v.Value)
	}
	return cfg
}

func newAppRegexConfigsMapper(typedMap map[string]*domain.TypedValue) ProtoMessage {
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

func newAppPayConfigsMapper(typedMap map[string]*domain.TypedValue) ProtoMessage {
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

func newAppOssConfigsMapper(typedMap map[string]*domain.TypedValue) ProtoMessage {
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

func newAppLanguageConfigsMapper(typedMap map[string]*domain.TypedValue) ProtoMessage {
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

func newAppMuteConfigsMapper(typedMap map[string]*domain.TypedValue) ProtoMessage {
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

func newAppGroupConfigsMapper(typedMap map[string]*domain.TypedValue) ProtoMessage {
	cfg := &pb.AppGroupConfigs{}
	if v, ok := typedMap["group_config_pay_notice"]; ok {
		cfg.GroupConfigPayNotice = toString(v.Value)
	}
	return cfg
}
