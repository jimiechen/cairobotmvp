package commonlib

// 统一业务错误码定义
// 所有模块共享同一套错误码，确保跨层一致性

const (
	CodeSuccess            = 10200 // 操作成功
	CodeBadRequest         = 10400 // 请求参数错误
	CodeUnauthorized       = 10401 // 未授权
	CodeNotFound           = 10404 // 资源未找到
	CodeInternalError      = 10500 // 内部错误
	CodeTarsNotImplemented = 10501 // Tars 远程调用未实现（S1 阶段）
)
