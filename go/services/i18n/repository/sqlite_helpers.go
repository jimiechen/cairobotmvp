package repository

import (
	"strings"
	"time"
)

// parseTimeStr 尝试多种时间格式解析字符串，兼容不同数据库的时间输出格式
func parseTimeStr(s *string) time.Time {
	if s == nil || *s == "" {
		return time.Time{}
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.DateTime,
	} {
		if t, err := time.Parse(layout, *s); err == nil {
			return t
		}
	}
	if strings.TrimSpace(*s) != "" {
		if t, err := time.Parse(time.Layout, *s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func init() {
	time.Local = time.UTC
}
