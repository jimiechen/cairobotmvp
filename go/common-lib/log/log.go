package log

import (
	"github.com/TarsCloud/TarsGo/contrib/log"
)

// Logger 日志接口，复用 TarsCloud contrib/log 的实现
// 所有业务模块统一通过此类型接收日志依赖
type Logger = log.Logger

// Default 返回默认日志实例
func Default() *Logger {
	return log.GetCtxLogger("cairobot")
}
