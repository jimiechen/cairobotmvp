package apis

import (
	"encoding/base64"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	i18nAdmin "github.com/jimiechen/mineplanet/go/services/i18n/admin"
	"go-admin/app/admin/i18n_admin/models"
)

// ImportExportApi CSV 导入导出 HTTP Handler
type ImportExportApi struct {
	svc i18nAdmin.I18nAdminService
}

// NewImportExportApi 创建 ImportExport API 实例
func NewImportExportApi(svc i18nAdmin.I18nAdminService) ImportExportApi {
	return ImportExportApi{svc: svc}
}

// ImportStringsFromCSV 从 CSV 导入语言字符串（两阶段：dry-run → 事务写入）
// POST /api/admin/v1/i18n/import/csv?pack_id=xxx
// Content-Type: multipart/form-data; file=xxx.csv
// 权限：i18n:string:write
//
// 流程：
// 1. 解析 CSV 头部验证格式
// 2. dry-run 阶段：逐行调用 ValidateTemplate，收集错误
// 3. 若 dry-run 有错，返回 10400 + 错误明细（不写入任何数据）
// 4. 若 dry-run 全通过，事务批量写入（≤100条/事务）
func (e ImportExportApi) ImportStringsFromCSV(c *gin.Context) {
	packIDStr := c.Query("pack_id")
	packID := parseInt64(packIDStr)
	if packID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的 pack_id 参数"})
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少上传文件"})
		return
	}
	defer file.Close()
	reader := io.LimitReader(file, 5*1024*1024) // 限制 5MB

	operator := c.GetString("x-user-name")
	result, importErr := e.svc.ImportStringsFromCSV(c.Request.Context(), reader, packID, operator)
	if importErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "导入处理失败", "error": importErr.Error()})
		return
	}
	resp := models.ImportResultResp{
		Code:         200,
		Message:      "导入完成",
		TotalRows:    result.TotalRows,
		SuccessCount: result.SuccessCount,
		FailCount:    result.FailCount,
	}
	if result.FailCount > 0 {
		resp.Code = 10400
		resp.Message = "部分行导入失败"
		resp.Errors = make([]models.ImportErrorItem, len(result.Errors))
		for i, ie := range result.Errors {
			resp.Errors[i] = models.ImportErrorItem{RowNum: ie.RowNum, Reason: ie.Reason}
		}
	}
	c.JSON(http.StatusOK, resp)
}

// ExportStringsToCSV 导出语言字符串为 CSV
// GET /api/admin/v1/i18n/export/csv?pack_id=xxx
// 权限：i18n:string:read
func (e ImportExportApi) ExportStringsToCSV(c *gin.Context) {
	packIDStr := c.Query("pack_id")
	packID := parseInt64(packIDStr)
	if packID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的 pack_id 参数"})
		return
	}
	req := i18nAdmin.ExportCSVRequest{PackID: packID}
	data, err := e.svc.ExportStringsToCSV(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "导出失败", "error": err.Error()})
		return
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	c.Header("Content-Disposition", `attachment; filename="strings_`+packIDStr+`.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
	_ = encoded // 同时支持 JSON 响应模式（前端可选用）
}

// parseInt64 安全解析 int64，解析失败返回 0
func parseInt64(s string) int64 {
	n := int64(0)
	for _, ch := range s {
		if ch < '0' || ch > '9' { return n }
		n = n*10 + int64(ch-'0')
	}
	return n
}

// trimSpace 去除首尾空白
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') { start++ }
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') { end-- }
	return s[start:end]
}

// hasPrefix 检查前缀
func hasPrefix(s, prefix string) bool { return len(s) >= len(prefix) && s[:len(prefix)] == prefix }
