package domain

// OperationType 表示语言字符串的操作类型
// 用于增量同步协议，标识字符串的变更状态
type OperationType string

const (
	// OperationAdd 新增操作
	OperationAdd OperationType = "ADD"
	// OperationMod 修改操作
	OperationMod OperationType = "MOD"
	// OperationDel 删除操作
	OperationDel OperationType = "DEL"
)

// IsValid 判断操作类型是否合法
func (op OperationType) IsValid() bool {
	switch op {
	case OperationAdd, OperationMod, OperationDel:
		return true
	default:
		return false
	}
}
