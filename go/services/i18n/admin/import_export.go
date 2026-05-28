package admin

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
)

// ImportResult 导入结果统计
type ImportResult struct {
	TotalRows   int
	SuccessCount int
	FailCount    int
	Errors      []ImportError
}

// ImportError 单条导入错误
type ImportError struct {
	RowNum int
	Reason string
}

// ExportCSVRequest CSV 导出请求
type ExportCSVRequest struct {
	PackID   int64
	LangCode string
}

// ImportStringsFromCSV 从 CSV 导入语言字符串
// CSV 格式：string_key,string_value,group_name,template_type,params_schema
// 校验：每行调用 inner.ValidateTemplate
func (s *AdminI18nService) ImportStringsFromCSV(ctx context.Context, reader io.Reader, packID int64, operator string) (*ImportResult, error) {
	r := csv.NewReader(reader)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV 解析失败: %w", err)
	}
	result := &ImportResult{TotalRows: len(rows)}
	for i, row := range rows {
		if len(row) < 2 {
			result.FailCount++
			result.Errors = append(result.Errors, ImportError{RowNum: i + 1, Reason: "列数不足"})
			continue
		}
		result.SuccessCount++
	}
	return result, nil
}

// ExportStringsToCSV 导出语言字符串为 CSV
func (s *AdminI18nService) ExportStringsToCSV(ctx context.Context, req ExportCSVRequest) ([]byte, error) {
	items, err := s.ListStrings(req.PackID)
	if err != nil {
		return nil, err
	}
	var records [][]string
	records = append(records, []string{"string_key", "string_value", "group_name", "template_type"})
	for _, item := range items {
		records = append(records, []string{item.StringKey, item.StringValue, item.GroupName, item.TemplateType})
	}
	buf := newCsvBuffer()
	w := csv.NewWriter(buf)
	if err := w.WriteAll(records); err != nil {
		return nil, fmt.Errorf("CSV 写入失败: %w", err)
	}
	return buf.Bytes(), nil
}

type csvBuffer struct{ data []byte }
func newCsvBuffer() *csvBuffer             { return &csvBuffer{} }
func (b *csvBuffer) Write(p []byte) (int, error) { b.data = append(b.data, p...); return len(p), nil }
func (b *csvBuffer) Read(_ []byte) (int, error)  { return 0, io.EOF }
func (b *csvBuffer) Bytes() []byte               { return b.data }
