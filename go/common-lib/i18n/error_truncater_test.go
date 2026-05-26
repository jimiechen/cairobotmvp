package i18n

import (
	"errors"
	"testing"
)

func TestTruncateError_NilError(t *testing.T) {
	result := TruncateError(nil, 512)
	if result != "" {
		t.Fatalf("期望 nil error 返回空串，实际 '%s'", result)
	}
}

func TestTruncateError_ShortError(t *testing.T) {
	err := errors.New("connection refused")
	result := TruncateError(err, 512)

	if result != "connection refused" {
		t.Fatalf("期望短错误原样返回，实际 '%s'", result)
	}
}

func TestTruncateError_ExactlyMaxLength(t *testing.T) {
	msg := string(make([]rune, 512))
	err := errors.New(msg)
	result := TruncateError(err, 512)

	runes := []rune(result)
	if len(runes) > 600 {
		t.Fatalf("期望截断后合理长度，实际 %d 字符", len(runes))
	}
}

func TestTruncateError_OverMaxLength(t *testing.T) {
	longMsg := string(make([]rune, 1024))
	err := errors.New(longMsg)
	result := TruncateError(err, 512)

	runes := []rune(result)

	expectedSuffix := "...(truncated, original 1024 chars)"
	hasSuffix := len(result) >= len(expectedSuffix) &&
		result[len(result)-len(expectedSuffix):] == expectedSuffix
	if !hasSuffix {
		t.Fatalf("期望截断后缀 '%s'，实际尾部无此后缀", expectedSuffix)
	}

	if len(runes) > 550 {
		t.Fatalf("期望截断后 ≤550 rune（含后缀），实际 %d", len(runes))
	}
}

func TestTruncateError_CustomMaxLength(t *testing.T) {
	longMsg := string(make([]rune, 100))
	err := errors.New(longMsg)
	result := TruncateError(err, 50)

	runes := []rune(result)
	if len(runes) > 90 {
		t.Fatalf("期望截断到 50 rune 后合理长度，实际 %d", len(runes))
	}

	expectedSuffix := "...(truncated, original 100 chars)"
	hasSuffix := len(result) >= len(expectedSuffix) &&
		result[len(result)-len(expectedSuffix):] == expectedSuffix
	if !hasSuffix {
		t.Fatalf("期望含截断后缀")
	}
}

func TestTruncateError_ZeroMaxLength(t *testing.T) {
	err := errors.New("test error")
	result := TruncateError(err, 0)

	if result == "" {
		t.Fatal("期望使用默认最大长度")
	}
}

func TestTruncateError_DefaultMaxLength(t *testing.T) {
	longMsg := string(make([]rune, 800))
	err := errors.New(longMsg)
	result := TruncateErrorDefault(err)

	runes := []rune(result)
	if len(runes) > MaxErrorLength+50 {
		t.Fatalf("期望使用默认 %d 最大长度，实际超长", MaxErrorLength)
	}
}

func TestTruncateError_MultiByteChars(t *testing.T) {
	chineseErr := errors.New("数据库连接失败：网络超时，请检查防火墙设置和MySQL端口是否正确配置")
	result := TruncateError(chineseErr, 30)

	runes := []rune(result)
	if len(runes) > 70 {
		t.Fatalf("多字节字符截断异常，实际 %d rune", len(runes))
	}
}

func TestTruncateError_PreservesShortMessage(t *testing.T) {
	shortErrors := []string{
		"ok",
		"timeout",
		"connection refused",
		"permission denied",
		"not found",
	}

	for _, msg := range shortErrors {
		err := errors.New(msg)
		result := TruncateError(err, 512)
		if result != msg {
			t.Fatalf("短错误信息应原样保留：期望 '%s'，实际 '%s'", msg, result)
		}
	}
}
