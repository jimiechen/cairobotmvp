package service

import (
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// ProtoMessage Protobuf 消息的统一抽象
// 由于当前生成的 pb 代码未导出 Message 接口，使用 any 作为通用类型
// 在 assignProtoField 中通过 type switch 进行具体类型匹配
type ProtoMessage = any

// StaticModuleMapper 静态模块 → Protobuf 映射函数签名
// 将 TypedValue map 转换为对应的 Protobuf 强类型消息
type StaticModuleMapper func(map[string]*domain.TypedValue) ProtoMessage

// staticMapperRegistry 静态模块映射注册表
// 用 map 替代 switch/case，消除硬编码 module_key 字符串字面量
// 新增静态模块只需在此注册，无需修改 MapStaticModulesToProtoFields 函数
var staticMapperRegistry = map[string]StaticModuleMapper{
	domain.ModuleKeyBase:  newAppBaseConfigsMapper,
	domain.ModuleKeyWap:   newAppWapUrlConfigsMapper,
	domain.ModuleKeyRegex: newAppRegexConfigsMapper,
	domain.ModuleKeyPay:   newAppPayConfigsMapper,
	domain.ModuleKeyOss:   newAppOssConfigsMapper,
	domain.ModuleKeyLang:  newAppLanguageConfigsMapper,
	domain.ModuleKeyMute:  newAppMuteConfigsMapper,
	domain.ModuleKeyGroup: newAppGroupConfigsMapper,
}

// MapStaticModulesToProtoFields 使用注册表将静态模块数据映射到 Protobuf 强类型字段
// 输入：从 DB 加载的 StaticModules（map[moduleKey]TypedValueMap）
// 输出：填充了强类型字段的 AppConfigsRsp proto 消息
func MapStaticModulesToProtoFields(staticModules map[string]map[string]*domain.TypedValue) *pb.AppConfigsRsp {
	protoRsp := &pb.AppConfigsRsp{}

	for moduleKey, typedMap := range staticModules {
		mapper, ok := staticMapperRegistry[moduleKey]
		if !ok {
			continue
		}

		msg := mapper(typedMap)
		assignProtoField(protoRsp, moduleKey, msg)
	}

	return protoRsp
}

// assignProtoField 将已构造的 proto Message 赋值到 AppConfigsRsp 对应字段
// 使用 type switch 确定具体消息类型并写入对应字段
func assignProtoField(rsp *pb.AppConfigsRsp, moduleKey string, msg ProtoMessage) {
	switch v := msg.(type) {
	case *pb.AppBaseConfigs:
		rsp.BaseCfg = v
	case *pb.AppWapUrlConfigs:
		rsp.WapCfg = v
	case *pb.AppRegexConfigs:
		rsp.RegexCfg = v
	case *pb.AppPayConfigs:
		rsp.PayCfg = v
	case *pb.AppOssConfigs:
		rsp.OssCfg = v
	case *pb.AppLanguageConfigs:
		rsp.LangCfg = v
	case *pb.AppMuteConfigs:
		rsp.MuteCfg = v
	case *pb.AppGroupConfigs:
		rsp.GroupCfg = v
	}
}
