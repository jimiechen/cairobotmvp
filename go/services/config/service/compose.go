package service

import (
	"encoding/json"
	"fmt"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/repository"
)

// BuildDynamicModule 从 ConfigVersion + TypedValue map + Schema 组装动态模块视图
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

// ToJSONMap 将 DynamicModuleView.Fields 序列化为 map[string]string
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

// ClassifyModules 将版本列表分为静态和动态两组
func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// toFloat64 将 any 转为 float64
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

// toBool 将 any 转为 bool
func toBool(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
