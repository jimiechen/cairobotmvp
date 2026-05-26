package i18n

import (
	"fmt"
)

const MaxErrorLength = 512

// TruncateError 截断错误信息到指定长度
// 避免 mysql/redis 长错误堆栈把响应包撑爆
// 默认截断长度为 512 字符
//
// 截断规则：
// - 如果错误信息 ≤ maxLen，原样返回
// - 如果错误信息 > maxLen，截断并追加 "...(truncated)"
// - 截断以 UTF-8 字符为单位，不破坏多字节字符
func TruncateError(err error, maxLen int) string {
	if err == nil {
		return ""
	}

	msg := err.Error()
	if maxLen <= 0 {
		maxLen = MaxErrorLength
	}

	runes := []rune(msg)
	if len(runes) <= maxLen {
		return msg
	}

	truncated := string(runes[:maxLen])
	return fmt.Sprintf("%s...(truncated, original %d chars)", truncated, len(runes))
}

// TruncateErrorDefault 使用默认最大长度 (512) 截断错误
func TruncateErrorDefault(err error) string {
	return TruncateError(err, MaxErrorLength)
}
