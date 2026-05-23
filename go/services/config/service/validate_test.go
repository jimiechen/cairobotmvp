package service

import (
	"strings"
	"testing"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

func TestValidateFieldValue_空validator应通过(t *testing.T) {
	tv := domain.NewTypedValue(domain.FieldTypeString, "any")
	fs := &domain.FieldSchema{FieldKey: "f1", Validator: ""}

	err := ValidateFieldValue(tv, fs)
	if err != nil {
		t.Error("空 validator 应通过")
	}
}

func TestValidateFieldValue_required检查(t *testing.T) {
	fs := &domain.FieldSchema{FieldKey: "name", Validator: "required"}

	err := ValidateFieldValue(nil, fs)
	if err == nil {
		t.Error("nil 值 + required 应失败")
	}
	err = ValidateFieldValue(domain.NewTypedValue(domain.FieldTypeString, ""), fs)
	if err == nil {
		t.Error("空串 + required 应失败")
	}
	tv := domain.NewTypedValue(domain.FieldTypeString, "value")
	err = ValidateFieldValue(tv, fs)
	if err != nil {
		t.Error("有值 + required 应通过")
	}
}

func TestValidateFieldValue_regex校验(t *testing.T) {
	fs := &domain.FieldSchema{FieldKey: "email", Validator: "regex:^\\w+@\\w+\\.com$"}

	validTV := domain.NewTypedValue(domain.FieldTypeString, "user@test.com")
	if err := ValidateFieldValue(validTV, fs); err != nil {
		t.Error("合法邮箱应通过正则校验")
	}

	invalidTV := domain.NewTypedValue(domain.FieldTypeString, "not-an-email")
	if err := ValidateFieldValue(invalidTV, fs); err == nil {
		t.Error("非法邮箱应不通过正则校验")
	}
}

func TestValidateFieldValue_range校验(t *testing.T) {
	fs := &domain.FieldSchema{FieldKey: "age", Validator: "range:0,150"}

	validTV := domain.NewTypedValue(domain.FieldTypeInt, int64(25))
	if err := ValidateFieldValue(validTV, fs); err != nil {
		t.Error("25 应在 [0,150] 范围内")
	}

	tooHigh := domain.NewTypedValue(domain.FieldTypeInt, int64(200))
	if err := ValidateFieldValue(tooHigh, fs); err == nil {
		t.Error("200 应超出范围")
	}
}

func TestValidateFieldValue_enum校验(t *testing.T) {
	fs := &domain.FieldSchema{FieldKey: "role", Validator: "enum:admin,user,guest"}

	validTV := domain.NewTypedValue(domain.FieldTypeEnum, "admin")
	if err := ValidateFieldValue(validTV, fs); err != nil {
		t.Error("admin 应在枚举列表中")
	}

	invalidTV := domain.NewTypedValue(domain.FieldTypeEnum, "hacker")
	if err := ValidateFieldValue(invalidTV, fs); err == nil {
		t.Error("hacker 不应在枚举列表中")
	}
}

func TestValidateConfigMap_批量校验(t *testing.T) {
	ms := &domain.ModuleSchema{
		ModuleKey: "test",
		Fields: []*domain.FieldSchema{
			{FieldKey: "email", Validator: "regex:^\\w+@\\w+$"},
			{FieldKey: "age", Validator: "range:0,120"},
		},
	}
	typedMap := map[string]*domain.TypedValue{
		"email": domain.NewTypedValue(domain.FieldTypeString, "bad"),
		"age":   domain.NewTypedValue(domain.FieldTypeInt, int64(200)),
	}

	errors := ValidateConfigMap(typedMap, ms)
	if len(errors) != 2 {
		t.Fatalf("期望 2 个校验错误, 实际 %d", len(errors))
	}
}

func TestValidationError_Error信息格式(t *testing.T) {
	err := ValidationError{Field: "email", Reason: "格式错误"}
	msg := err.Error()
	if !strings.Contains(msg, "[email]") || !strings.Contains(msg, "格式错误") {
		t.Errorf("错误信息格式异常: %s", msg)
	}
}

func TestRegexCompileError_非法正则(t *testing.T) {
	fs := &domain.FieldSchema{FieldKey: "f1", Validator: "regex:[invalid"}
	tv := domain.NewTypedValue(domain.FieldTypeString, "anything")

	err := ValidateFieldValue(tv, fs)
	if err == nil {
		t.Error("非法正则应返回错误")
	}
	if _, ok := err.(RegexCompileError); !ok {
		t.Error("应为 RegexCompileError 类型")
	}
}
