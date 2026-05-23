package domain

// ModuleKey 预定义模块键常量集合
// 与 proto AppConfigsRsp 中强类型字段(2-9) 一一对应
// 不负责模块注册或发现，只提供常量定义
const (
	ModuleKeyBase    = "base_cfg"
	ModuleKeyWap     = "wap_cfg"
	ModuleKeyRegex   = "regex_cfg"
	ModuleKeyPay     = "pay_cfg"
	ModuleKeyOss     = "oss_cfg"
	ModuleKeyLang    = "lang_cfg"
	ModuleKeyMute    = "mute_cfg"
	ModuleKeyGroup   = "group_cfg"
)

// AllStaticModuleKeys 返回所有预定义强类型模块键列表
// 用于 compose.go 判断某 module_key 是否应走强类型字段还是 dynamic_modules
func AllStaticModuleKeys() []string {
	return []string{
		ModuleKeyBase, ModuleKeyWap, ModuleKeyRegex,
		ModuleKeyPay, ModuleKeyOss, ModuleKeyLang,
		ModuleKeyMute, ModuleKeyGroup,
	}
}

// IsStaticModule 判断给定 module_key 是否属于预定义的 8 个强类型模块
// 动态新增的模块（不在本列表中的）将放入 AppConfigsRsp.dynamic_modules
func IsStaticModule(moduleKey string) bool {
	for _, key := range AllStaticModuleKeys() {
		if key == moduleKey {
			return true
		}
	}
	return false
}
