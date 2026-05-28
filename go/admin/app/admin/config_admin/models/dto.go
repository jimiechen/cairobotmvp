package models

// SchemaItem Schema 列表返回项（给前端渲染用）
type SchemaItem struct {
	ID           int64  `json:"id"`
	ModuleKey    string `json:"module_key"`
	FieldKey     string `json:"field_key"`
	FieldType    string `json:"field_type"`
	DefaultValue string `json:"default_value"`
	Validator    string `json:"validator"`
	IsRequired   bool   `json:"is_required"`
	IsSecret     bool   `json:"is_secret"`
	Description  string `json:"description"`
	ClientScope  string `json:"client_scope"`
	SortOrder    int    `json:"sort_order"`
	IsEnabled    bool   `json:"is_enabled"`
}

// CreateSchemaReq 新增 Schema 请求体
type CreateSchemaReq struct {
	ModuleKey    string `json:"module_key" binding:"required"`
	FieldKey     string `json:"field_key" binding:"required"`
	FieldType    string `json:"field_type" binding:"required"`
	DefaultValue string `json:"default_value"`
	Validator    string `json:"validator"`
	IsRequired   bool   `json:"is_required"`
	IsSecret     bool   `json:"is_secret"`
	Description  string `json:"description"`
	ClientScope  string `json:"client_scope"`
	SortOrder    int    `json:"sort_order"`
}

// UpdateSchemaReq 更新 Schema 请求体
type UpdateSchemaReq struct {
	ID           int64  `json:"id" binding:"required,gt=0"`
	FieldType    string `json:"field_type"`
	DefaultValue string `json:"default_value"`
	Validator    string `json:"validator"`
	IsRequired   bool   `json:"is_required"`
	IsSecret     bool   `json:"is_secret"`
	Description  string `json:"description"`
	ClientScope  string `json:"client_scope"`
	SortOrder    int    `json:"sort_order"`
}

// DeleteSchemaReq 删除 Schema 请求体
type DeleteSchemaReq struct {
	ID int64 `json:"id" binding:"required,gt=0"`
}

// ListSchemaReq 查询 Schema 列表请求参数
type ListSchemaReq struct {
	ModuleKey string `form:"module_key" binding:"required"`
}

// ValueVersion 配置版本发布结果
type ValueVersion struct {
	Version    int64  `json:"version"`
	ModuleKey  string `json:"module_key"`
	Env        string `json:"env"`
	FieldCount int    `json:"field_count"`
}

// PublishValueReq 发布配置值请求体
type PublishValueReq struct {
	ModuleKey string            `json:"module_key" binding:"required"`
	Env       string            `json:"env" binding:"required"`
	Fields    []PublishFieldItem `json:"fields" binding:"required,min=1"`
}

// PublishFieldItem 发布配置字段项
type PublishFieldItem struct {
	FieldKey string      `json:"field_key" binding:"required"`
	Value    interface{} `json:"value"`
}

// ValidationErrorResp 校验错误响应（HTTP 10400）
type ValidationErrorResp struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Errors  []ValidationErrorItem `json:"errors,omitempty"`
}

// ValidationErrorItem 单字段校验错误
type ValidationErrorItem struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}
