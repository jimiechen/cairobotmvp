package health

import (
	"context"
	"time"
)

// Checker 健康检查器接口
// 每个涉及外部依赖的业务模块必须实现此接口并注册到 health 模块
// 后续模块只需实现 Check() 方法，不再各自重复造轮子
type Checker interface {
	Name() string
	Check(ctx context.Context) ComponentStatus
}

// ComponentStatus 组件健康状态
// 描述单个依赖的健康情况，用于汇总和展示
type ComponentStatus struct {
	Name      string // 组件名称，如 mysql、redis、config、i18n
	Healthy   bool   // 是否健康
	LatencyMs int64  // 检查耗时（毫秒）
	Error     string // 错误信息，healthy=true 时为空
}

// NewComponentStatus 创建健康的组件状态
func NewComponentStatus(name string, latencyMs int64) ComponentStatus {
	return ComponentStatus{
		Name:      name,
		Healthy:   true,
		LatencyMs: latencyMs,
	}
}

// NewUnhealthyComponentStatus 创建不健康的组件状态
func NewUnhealthyComponentStatus(name string, latencyMs int64, err error) ComponentStatus {
 errorMsg := ""
	if err != nil {
	 errorMsg = err.Error()
 }
 return ComponentStatus{
 Name:      name,
 Healthy:   false,
 LatencyMs: latencyMs,
 Error:     errorMsg,
 }
}

const DefaultCheckTimeout = 1 * time.Second

// CheckWithTimeout 带超时的健康检查
// 确保单个 checker 不会拖垮整个 health 协议
func CheckWithTimeout(ctx context.Context, checker Checker) ComponentStatus {
 ctx, cancel := context.WithTimeout(ctx, DefaultCheckTimeout)
 defer cancel()

 type result struct {
 status ComponentStatus
 }

 ch := make(chan result, 1)

 go func() {
 ch <- result{status: checker.Check(ctx)}
 }()

 select {
 case res := <-ch:
 return res.status
 case <-ctx.Done():
 return NewUnhealthyComponentStatus(checker.Name(), DefaultCheckTimeout.Milliseconds(), ctx.Err())
 }
}
