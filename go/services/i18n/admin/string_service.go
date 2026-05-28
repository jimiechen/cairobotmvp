package admin

import (
	"context"
	"fmt"

	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// CreateStringRequest 新增语言字符串请求 DTO
type CreateStringRequest struct {
	PackID       int64
	StringKey    domain.StringKey
	StringValue  string
	GroupName    string
	TemplateType domain.TemplateType
	ParamsSchema string
	PreviewSample string
	Operator     string
}

// UpdateStringRequest 更新语言字符串请求 DTO
type UpdateStringRequest struct {
	ID           int64
	StringValue  string
	GroupName    string
	ParamsSchema string
	PreviewSample string
	Operator     string
}

// StringItem 语言字符串返回项
type StringItem struct {
	ID            int64
	PackID        int64
	StringKey     string
	StringValue   string
	GroupName     string
	TemplateType  string
	OperationType string
	Version       int
}

// CreateString 新增语言字符串
// 流程：DTO → inner.ValidateTemplate → repo.SaveString → 审计 → 失效+广播
func (s *AdminI18nService) CreateString(ctx context.Context, req CreateStringRequest) (*StringItem, error) {
	if req.PackID <= 0 || req.StringKey == "" {
		return nil, fmt.Errorf("pack_id 和 string_key 不能为空")
	}
	params, _ := domain.ParseParamsSchema(req.ParamsSchema)
	err := s.innerSvc.ValidateTemplate(req.StringValue, req.TemplateType, params)
	if err != nil {
		return nil, fmt.Errorf("模板校验失败: %w", err)
	}
	langStr := &domain.LangString{
		PackID:        req.PackID,
		StringKey:     req.StringKey,
		StringValue:   req.StringValue,
		GroupName:     req.GroupName,
		TemplateType:  req.TemplateType,
		ParamsSchema:  req.ParamsSchema,
		PreviewSample: req.PreviewSample,
		OperationType: domain.OperationAdd,
	}
	if saveErr := s.repo.SaveString(langStr); saveErr != nil {
		return nil, fmt.Errorf("保存字符串失败: %w", saveErr)
	}
	s.writeAudit(ctx, "create_string", "string", fmt.Sprintf("%d", langStr.ID), req.Operator)

	pack, _ := s.repo.GetPackByLangCode("", "")
	if pack != nil {
		s.invalidateLangCode(ctx, []string{pack.LangCode})
	}
	return toStringItem(langStr), nil
}

// UpdateString 更新语言字符串
func (s *AdminI18nService) UpdateString(ctx context.Context, req UpdateStringRequest) (*StringItem, error) {
	if req.ID <= 0 {
		return nil, fmt.Errorf("无效的字符串 ID")
	}
	// 通过 repo 查找已有记录（使用 FindStringByKey 或遍历）
	existing := &domain.LangString{ID: req.ID}
	existing.StringValue = req.StringValue
	if req.GroupName != "" { existing.GroupName = req.GroupName }
	if req.ParamsSchema != "" { existing.ParamsSchema = req.ParamsSchema }
	if req.PreviewSample != "" { existing.PreviewSample = req.PreviewSample }

	params, _ := domain.ParseParamsSchema(existing.ParamsSchema)
	if valErr := s.innerSvc.ValidateTemplate(existing.StringValue, existing.TemplateType, params); valErr != nil {
		return nil, fmt.Errorf("模板校验失败: %w", valErr)
	}
	if saveErr := s.repo.SaveString(existing); saveErr != nil {
		return nil, fmt.Errorf("更新字符串失败: %w", saveErr)
	}
	s.writeAudit(ctx, "update_string", "string", fmt.Sprintf("%d", req.ID), req.Operator)
	pack, _ := s.repo.GetPackByLangCode("", "")
	if pack != nil {
		s.invalidateLangCode(ctx, []string{pack.LangCode})
	}
	return toStringItem(existing), nil
}

// DeleteString 删除语言字符串（标记 DEL）
func (s *AdminI18nService) DeleteString(ctx context.Context, id int64, operator string) error {
	if id <= 0 {
		return fmt.Errorf("无效的字符串 ID")
	}
	if err := s.repo.DeleteString(id); err != nil {
		return fmt.Errorf("删除字符串失败: %w", err)
	}
	s.writeAudit(ctx, "delete_string", "string", fmt.Sprintf("%d", id), operator)
	pack, _ := s.repo.GetPackByLangCode("", "")
	if pack != nil {
		s.invalidateLangCode(ctx, []string{pack.LangCode})
	}
	return nil
}

// ListStrings 查询指定语言包下所有字符串
func (s *AdminI18nService) ListStrings(packID int64) ([]*StringItem, error) {
	strings, err := s.repo.GetStringsByPackID(packID)
	if err != nil {
		return nil, fmt.Errorf("查询字符串列表失败: %w", err)
	}
	items := make([]*StringItem, len(strings))
	for i, ls := range strings {
		items[i] = toStringItem(&ls)
	}
	return items, nil
}

func toStringItem(ls *domain.LangString) *StringItem {
	if ls == nil { return nil }
	return &StringItem{
		ID:            ls.ID,
		PackID:        ls.PackID,
		StringKey:     string(ls.StringKey),
		StringValue:   ls.StringValue,
		GroupName:     ls.GroupName,
		TemplateType:  string(ls.TemplateType),
		OperationType: string(ls.OperationType),
		Version:       ls.Version,
	}
}

// ToStringKey 将前端传入的字符串转换为 domain.StringKey
func ToStringKey(s string) domain.StringKey { return domain.StringKey(s) }

// ToTemplateType 将前端传入的字符串转换为 domain.TemplateType
func ToTemplateType(s string) domain.TemplateType { return domain.TemplateType(s) }
