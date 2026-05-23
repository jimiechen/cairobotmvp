package domain

import (
	"encoding/json"
	"testing"
)

func TestTypedValue_String_string类型应原样返回(t *testing.T) {
	tv := NewTypedValue(FieldTypeString, "hello")
	if tv.String() != "hello" {
		t.Errorf("期望 hello, 实际 %s", tv.String())
	}
}

func TestTypedValue_String_nil应返回空串(t *testing.T) {
	tv := NewTypedValue(FieldTypeString, nil)
	if tv.String() != "" {
		t.Errorf("期望空串, 实际 %s", tv.String())
	}
}

func TestTypedValue_String非string类型应格式化(t *testing.T) {
	tv := NewTypedValue(FieldTypeInt, 42)
	if tv.String() != "42" {
		t.Errorf("期望 '42', 实际 '%s'", tv.String())
	}
}

func TestTypedValue_Int_int64类型(t *testing.T) {
	tv := NewTypedValue(FieldTypeInt, int64(100))
	if tv.Int() != 100 {
		t.Errorf("期望 100, 实际 %d", tv.Int())
	}
}

func TestTypedValue_Int_int类型自动转换(t *testing.T) {
	tv := NewTypedValue(FieldTypeInt, int(200))
	if tv.Int() != 200 {
		t.Errorf("期望 200, 实际 %d", tv.Int())
	}
}

func TestTypedValue_Int_float64截断(t *testing.T) {
	tv := NewTypedValue(FieldTypeInt, float64(3.14))
	if tv.Int() != 3 {
		t.Errorf("期望 3, 实际 %d", tv.Int())
	}
}

func TestTypedValue_Int_jsonNumber(t *testing.T) {
	tv := NewTypedValue(FieldTypeInt, json.Number("99"))
	if tv.Int() != 99 {
		t.Errorf("期望 99, 实际 %d", tv.Int())
	}
}

func TestTypedValue_Int_nil应返回0(t *testing.T) {
	tv := NewTypedValue(FieldTypeInt, nil)
	if tv.Int() != 0 {
		t.Errorf("期望 0, 实际 %d", tv.Int())
	}
}

func TestTypedValue_Bool_true(t *testing.T) {
	tv := NewTypedValue(FieldTypeBool, true)
	if !tv.Bool() {
		t.Error("期望 true")
	}
}

func TestTypedValue_Bool_false与nil(t *testing.T) {
	testCases := []struct {
		name  string
		value any
		want  bool
	}{
		{"false值", false, false},
		{"nil", nil, false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tv := NewTypedValue(FieldTypeBool, tc.value)
			if tv.Bool() != tc.want {
				t.Errorf("期望 %v, 实际 %v", tc.want, tv.Bool())
			}
		})
	}
}

func TestTypedValue_Float_float64类型(t *testing.T) {
	tv := NewTypedValue(FieldTypeFloat, 3.14)
	if tv.Float() != 3.14 {
		t.Errorf("期望 3.14, 实际 %f", tv.Float())
	}
}

func TestTypedValue_Float_int64转换(t *testing.T) {
	tv := NewTypedValue(FieldTypeFloat, int64(42))
	if tv.Float() != 42.0 {
		t.Errorf("期望 42.0, 实际 %f", tv.Float())
	}
}

func TestTypedValue_JSON_rawMessage(t *testing.T) {
	raw := json.RawMessage(`{"key":"val"}`)
	tv := NewTypedValue(FieldTypeJSON, raw)
	result := tv.JSON()
	if string(result) != `{"key":"val"}` {
		t.Errorf("期望原始JSON, 实际 %s", string(result))
	}
}

func TestTypedValue_JSON_对象序列化(t *testing.T) {
	tv := NewTypedValue(FieldTypeJSON, map[string]string{"k": "v"})
	result := tv.JSON()
	var m map[string]string
	err := json.Unmarshal(result, &m)
	if err != nil {
		t.Fatalf("JSON 反序列化失败: %v", err)
	}
	if m["k"] != "v" {
		t.Error("JSON 内容不一致")
	}
}

func TestTypedValue_JSON_nil应返回nil(t *testing.T) {
	tv := NewTypedValue(FieldTypeJSON, nil)
	if tv.JSON() != nil {
		t.Error("nil 值应返回 nil RawMessage")
	}
}
