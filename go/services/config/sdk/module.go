package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// GetModule 获取整个模块的配置快照
// 返回包含所有字段的 ModuleSnapshot，可用于批量读取
func (c *configClient) GetModule(ctx context.Context, moduleKey string) (*ModuleSnapshot, error) {
	cacheKey := buildCacheKey(moduleKey)
	snapshot, ok := c.lruCache.get(cacheKey)
	if ok {
		return snapshot, nil
	}
	snapshot, err := c.fetchModule(ctx, moduleKey)
	if err != nil {
		return nil, fmt.Errorf("fetch module %s failed: %w", moduleKey, err)
	}
	c.lruCache.set(cacheKey, snapshot)
	return snapshot, nil
}

// Bind 将模块字段绑定到结构体 out
// out 必须是指向结构体的指针，结构体字段名与 field_key 对应（支持 tag: `config:"field_key"`）
// 支持的类型映射：string → string, int/int64 → int64, bool → bool, float64 → float64, json.RawMessage/any → JSON
func (c *configClient) Bind(ctx context.Context, moduleKey string, out any) error {
	snapshot, err := c.GetModule(ctx, moduleKey)
	if err != nil {
		return fmt.Errorf("get module failed: %w", err)
	}
	return bindStruct(snapshot, out)
}

// bindStruct 将 ModuleSnapshot 绑定到结构体
// 通过 reflect 实现运行时字段映射和类型转换
func bindStruct(snapshot *ModuleSnapshot, out any) error {
	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr || outVal.IsNil() {
		return ErrBindFailed
	}
	outVal = outVal.Elem()
	if outVal.Kind() != reflect.Struct {
		return ErrBindFailed
	}
	outType := outVal.Type()
	for i := 0; i < outType.NumField(); i++ {
		field := outType.Field(i)
		fieldVal := outVal.Field(i)
		if !fieldVal.CanSet() {
			continue
		}
		fieldKey := resolveFieldKey(field)
		tv := snapshot.GetField(fieldKey)
		if tv == nil {
			continue
		}
		if err := setFieldValue(fieldVal, tv); err != nil {
			return fmt.Errorf("bind field %s failed: %w", fieldKey, err)
		}
	}
	return nil
}

// resolveFieldKey 解析结构体字段的配置 key
// 优先使用 `config` tag，否则使用字段名的 snake_case 形式
func resolveFieldKey(field reflect.StructField) string {
	if tag := field.Tag.Get("config"); tag != "" {
		return tag
	}
	return camelToSnake(field.Name)
}

// setFieldValue 将 TypedValue 设置到 reflect.Value
// 根据目标字段类型调用对应的类型转换方法
func setFieldValue(fieldVal reflect.Value, tv interface{ String() string; Int() int64; Bool() bool; Float() float64; JSON() json.RawMessage }) error {
	switch fieldVal.Kind() {
	case reflect.String:
		fieldVal.SetString(tv.String())
	case reflect.Int, reflect.Int64:
		fieldVal.SetInt(tv.Int())
	case reflect.Bool:
		fieldVal.SetBool(tv.Bool())
	case reflect.Float32, reflect.Float64:
		fieldVal.SetFloat(tv.Float())
	case reflect.Interface, reflect.Ptr:
		raw := tv.JSON()
		if raw == nil {
			return ErrTypeMismatch
		}
		var target any
		if err := json.Unmarshal(raw, &target); err != nil {
			return err
		}
		fieldVal.Set(reflect.ValueOf(target))
	default:
		return ErrTypeMismatch
	}
	return nil
}

// camelToSnake 将 CamelCase 转换为 snake_case
// 用于自动将 Go 字段名映射为配置 field_key
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
