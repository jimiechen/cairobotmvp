package apis

import (
	"encoding/base64"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	i18nAdmin "github.com/jimiechen/mineplanet/go/services/i18n/admin"
	"go-admin/app/admin/i18n_admin/models"
)

type ImportExportApi struct {
	svc i18nAdmin.I18nAdminService
}

func NewImportExportApi(svc i18nAdmin.I18nAdminService) ImportExportApi {
	return ImportExportApi{svc: svc}
}

func (e *ImportExportApi) requireSvc(c *gin.Context) bool {
	if e.svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "多语言服务未初始化"})
		return false
	}
	return true
}

func (e ImportExportApi) ImportStringsFromCSV(c *gin.Context) {
	if !e.requireSvc(c) { return }
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
	reader := io.LimitReader(file, 5*1024*1024)
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

func (e ImportExportApi) ExportStringsToCSV(c *gin.Context) {
	if !e.requireSvc(c) { return }
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
	_ = encoded
}

func parseInt64(s string) int64 {
	n := int64(0)
	for _, ch := range s {
		if ch < '0' || ch > '9' { return n }
		n = n*10 + int64(ch-'0')
	}
	return n
}
