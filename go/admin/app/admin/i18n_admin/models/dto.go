package models

// StringItem 语言字符串返回项
type StringItem struct {
	ID            int64  `json:"id"`
	PackID        int64  `json:"pack_id"`
	StringKey     string `json:"string_key"`
	StringValue   string `json:"string_value"`
	GroupName     string `json:"group_name"`
	TemplateType  string `json:"template_type"`
	OperationType string `json:"operation_type"`
	Version       int    `json:"version"`
}

// CreateStringReq 新增语言字符串请求体
type CreateStringReq struct {
	PackID       int64  `json:"pack_id" binding:"required,gt=0"`
	StringKey    string `json:"string_key" binding:"required"`
	StringValue  string `json:"string_value" binding:"required"`
	GroupName    string `json:"group_name"`
	TemplateType string `json:"template_type"` // plain | named | icu
	ParamsSchema string `json:"params_schema"`
	PreviewSample string `json:"preview_sample"`
}

// UpdateStringReq 更新语言字符串请求体
type UpdateStringReq struct {
	ID           int64  `json:"id" binding:"required,gt=0"`
	StringValue  string `json:"string_value"`
	GroupName    string `json:"group_name"`
	ParamsSchema string `json:"params_schema"`
	PreviewSample string `json:"preview_sample"`
}

// DeleteStringReq 删除语言字符串请求体
type DeleteStringReq struct {
	ID int64 `json:"id" binding:"required,gt=0"`
}

// ListStringsReq 查询语言字符串列表参数
type ListStringsReq struct {
	PackID int64 `form:"pack_id" binding:"required,gt=0"`
}

// PackVersion 语言包版本信息
type PackVersion struct {
	PackID   int64  `json:"pack_id"`
	LangCode string `json:"lang_code"`
	Version  int    `json:"version"`
}

// PublishPackReq 发布语言包请求体
type PublishPackReq struct {
	PackID   int64  `json:"pack_id" binding:"required,gt=0"`
	LangCode string `json:"lang_code" binding:"required"`
	Env      string `json:"env" binding:"required"`
}

// RollbackPackReq 回滚语言包请求体
type RollbackPackReq struct {
	PackID        int64  `json:"pack_id" binding:"required,gt=0"`
	TargetVersion int    `json:"target_version" binding:"required,gte=0"`
}

// ImportCSVReq CSV 导入请求（multipart form）
type ImportCSVReq struct {
	PackID int64 `form:"pack_id" binding:"required,gt=0"`
}

// ImportResultResp CSV 导入结果响应
type ImportResultResp struct {
	Code        int            `json:"code"`
	Message     string         `json:"message"`
	TotalRows   int            `json:"total_rows"`
	SuccessCount int           `json:"success_count"`
	FailCount   int            `json:"fail_count"`
	Errors      []ImportErrorItem `json:"errors,omitempty"`
}

// ImportErrorItem 单条导入错误
type ImportErrorItem struct {
	RowNum int    `json:"row_num"`
	Reason string `json:"reason"`
}

// ExportCSVResp CSV 导出响应
type ExportCSVResp struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    string `json:"data"` // base64 编码的 CSV 内容
}
