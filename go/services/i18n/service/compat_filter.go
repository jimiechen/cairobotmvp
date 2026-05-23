package service

import (
	"strings"
)

// ApplyCompatFilter 按客户端版本过滤模板类型
// 保证老客户端不会因新模板类型崩溃
//
// 规则：
// - client_version >= 2.0.0: 返回所有类型（plain, named, icu）
// - client_version < 2.0.0: 只返回 template_type=plain 的条目
//
// Args:
//   - entries: 原始语言字符串列表
//   - clientVersion: 客户端版本号
//
// Returns:
//   - []LangStringEntry: 过滤后的字符串列表
func ApplyCompatFilter(entries []LangStringEntry, clientVersion string) []LangStringEntry {
	if isClientSupportsAdvancedTemplates(clientVersion) {
		return entries
	}

	var filtered []LangStringEntry
	for _, entry := range entries {
		if entry.TemplateType == "plain" || entry.TemplateType == "" {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// isClientSupportsAdvancedTemplates 判断客户端是否支持高级模板类型
// 版本号 >= 2.0.0 支持 named 和 icu 类型
func isClientSupportsAdvancedTemplates(clientVersion string) bool {
	return compareVersions(clientVersion, "2.0.0") >= 0
}

// compareVersions 简单的版本号比较
// 返回值：-1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		if i < len(parts1) {
			num1 = parseVersionPart(parts1[i])
		}
		if i < len(parts2) {
			num2 = parseVersionPart(parts2[i])
		}
		if num1 < num2 {
			return -1
		}
		if num1 > num2 {
			return 1
		}
	}

	return 0
}

// parseVersionPart 解析版本号的单个部分
func parseVersionPart(part string) int {
	num := 0
	for _, ch := range part {
		if ch >= '0' && ch <= '9' {
			num = num*10 + int(ch-'0')
		} else {
			break
		}
	}
	return num
}
