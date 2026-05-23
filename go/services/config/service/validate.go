package service

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// ValidateFieldValue 按 schema.validator 校验单个字段的合法性
// validator 格式约定：
//   - 空/nil 表示跳过校验
//   - 以 "/" 包裹的视为正则表达式，如 "/^[a-z]+$/"
//   - "range:min,max" 表示数值范围校验
//   - "enum:v1,v2,v3" 表示枚举白名单
//   - "required" 表示非空检查（与 is_required 联动但可单独覆盖）
//
// 返回 nil 表示校验通过，返回 error 描述具体违规原因
func ValidateFieldValue(value *domain.TypedValue, schema *domain.FieldSchema) error {
	if schema == nil || schema.Validator == "" {
		return nil
	}

	validator := strings.TrimSpace(schema.Validator)

	if validator == "required" {
		if value == nil || value.Value == nil {
			return ValidationError{Field: schema.FieldKey, Reason: "字段为必填项"}
		}
		if strVal, ok := value.Value.(string); ok && strings.TrimSpace(strVal) == "" {
			return ValidationError{Field: schema.FieldKey, Reason: "字段为必填项，不能为空串"}
		}
		return nil
	}

	if strings.HasPrefix(validator, "regex:") {
		pattern := strings.TrimPrefix(validator, "regex:")
		return validateRegex(value, pattern, schema.FieldKey)
	}

	if strings.HasPrefix(validator, "range:") {
		rangeStr := strings.TrimPrefix(validator, "range:")
		return validateRange(value, rangeStr, schema.FieldKey)
	}

	if strings.HasPrefix(validator, "enum:") {
		enumStr := strings.TrimPrefix(validator, "enum:")
		return validateEnum(value, enumStr, schema.FieldKey)
	}

	return nil
}

// ValidateConfigMap 批量校验整个模块的 config_map
// 返回所有校验失败的字段错误列表
func ValidateConfigMap(typedMap map[string]*domain.TypedValue, moduleSchema *domain.ModuleSchema) []ValidationError {
	var errors []ValidationError
	for fieldKey, tv := range typedMap {
		fs := moduleSchema.FindField(fieldKey)
		if fs == nil {
			continue
		}
		if err := ValidateFieldValue(tv, fs); err != nil {
			if ve, ok := err.(ValidationError); ok {
				errors = append(errors, ve)
			}
		}
	}
	return errors
}

// ValidationError 字段级校验错误，携带具体字段名和原因
type ValidationError struct {
	Field  string
	Reason string
}

func (e ValidationError) Error() string {
	return "field [" + e.Field + "] " + e.Reason
}

func validateRegex(value *domain.TypedValue, pattern, fieldKey string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return RegexCompileError{Pattern: pattern, Inner: err}
	}
	strVal := value.String()
	if !re.MatchString(strVal) {
		return ValidationError{Field: fieldKey, Reason: "不匹配正则: " + pattern}
	}
	return nil
}

func validateRange(value *domain.TypedValue, rangeStr, fieldKey string) error {
	parts := strings.Split(rangeStr, ",")
	if len(parts) != 2 {
		return ValidationError{Field: fieldKey, Reason: "range 格式错误"}
	}
	minVal := parseFloat(parts[0])
	maxVal := parseFloat(parts[1])
	numVal := value.Float()
	if numVal < minVal || numVal > maxVal {
		return ValidationError{Field: fieldKey, Reason: "超出范围 [" + rangeStr + "]"}
	}
	return nil
}

func validateEnum(value *domain.TypedValue, enumStr, fieldKey string) error {
	allowed := strings.Split(enumStr, ",")
	strVal := value.String()
	for _, a := range allowed {
		if strings.TrimSpace(a) == strVal {
			return nil
		}
	}
	return ValidationError{Field: fieldKey, Reason: "不在枚举列表: " + enumStr}
}

// RegexCompileError 正则表达式编译错误
type RegexCompileError struct {
	Pattern string
	Inner   error
}

func (e RegexCompileError) Error() string {
	return "正则编译失败: " + e.Pattern + ": " + e.Inner.Error()
}

func parseFloat(s string) float64 {
	var f float64
	json.Unmarshal([]byte(s), &f)
	return f
}
