package repository

import (
	"github.com/jimiechen/mineplanet/go/services/i18n/domain"
)

// MockRepo 用于单元测试的 Mock 实现
// 实现了 I18nRepository 接口
type MockRepo struct {
	Packs   map[string]*domain.LangPack
	Strings map[int64][]domain.LangString
}

// NewMockRepo 创建 Mock 仓库实例
func NewMockRepo() *MockRepo {
	return &MockRepo{
		Packs:   make(map[string]*domain.LangPack),
		Strings: make(map[int64][]domain.LangString),
	}
}

// GetPackByLangCode 根据语言代码查询语言包
func (r *MockRepo) GetPackByLangCode(langCode string, env string) (*domain.LangPack, error) {
	key := langCode + ":" + env
	if pack, exists := r.Packs[key]; exists {
		return pack, nil
	}
	return nil, nil
}

// GetStringsByPackID 根据语言包 ID 查询所有字符串
func (r *MockRepo) GetStringsByPackID(packID int64) ([]domain.LangString, error) {
	if strings, exists := r.Strings[packID]; exists {
		return strings, nil
	}
	return []domain.LangString{}, nil
}

// GetDiffSince 查询指定版本之后的增量变更
func (r *MockRepo) GetDiffSince(packID int64, sinceVersion int) ([]domain.LangString, error) {
	if strings, exists := r.Strings[packID]; exists {
		var result []domain.LangString
		for _, s := range strings {
			if s.Version > sinceVersion && (s.OperationType == domain.OperationAdd || s.OperationType == domain.OperationMod) {
				result = append(result, s)
			}
		}
		return result, nil
	}
	return []domain.LangString{}, nil
}

// ListPacks 列出所有已发布的语言包
func (r *MockRepo) ListPacks(env string) ([]domain.LangPack, error) {
	var packs []domain.LangPack
	for _, pack := range r.Packs {
		if pack.Env == env && pack.IsPublished {
			packs = append(packs, *pack)
		}
	}
	return packs, nil
}

// SetupMockRepoWithSeedData 创建带种子数据的 Mock 仓库
// 用于单元测试，预填充 zh-CN 和 en 两种语言的示例数据
func SetupMockRepoWithSeedData() *MockRepo {
	repo := NewMockRepo()

	zhPack := &domain.LangPack{
		ID:          1,
		PackName:    "webp",
		Env:         "dev",
		Version:     1,
		LangCode:    "zh-CN",
		Description: "简体中文",
		IsPublished: true,
	}
	enPack := &domain.LangPack{
		ID:          2,
		PackName:    "webp",
		Env:         "dev",
		Version:     1,
		LangCode:    "en",
		Description: "English",
		IsPublished: true,
	}

	repo.Packs["zh-CN:dev"] = zhPack
	repo.Packs["en:dev"] = enPack

	repo.Strings[1] = []domain.LangString{
		{ID: 1, PackID: 1, StringKey: domain.StringKey("svc_common_ok"), StringValue: "确定", TemplateType: domain.TemplatePlain, OperationType: domain.OperationAdd, Version: 1},
		{ID: 2, PackID: 1, StringKey: domain.StringKey("svc_common_cancel"), StringValue: "取消", TemplateType: domain.TemplatePlain, OperationType: domain.OperationAdd, Version: 1},
		{ID: 3, PackID: 1, StringKey: domain.StringKey("svc_msg_welcome"), StringValue: "欢迎 {name}，你有 {count} 条新消息", TemplateType: domain.TemplateNamed, ParamsSchema: `[{"name":"name","type":"string","required":true},{"name":"count","type":"int","required":true}]`, OperationType: domain.OperationAdd, Version: 1},
	}

	repo.Strings[2] = []domain.LangString{
		{ID: 4, PackID: 2, StringKey: domain.StringKey("svc_common_ok"), StringValue: "OK", TemplateType: domain.TemplatePlain, OperationType: domain.OperationAdd, Version: 1},
		{ID: 5, PackID: 2, StringKey: domain.StringKey("svc_common_cancel"), StringValue: "Cancel", TemplateType: domain.TemplatePlain, OperationType: domain.OperationAdd, Version: 1},
		{ID: 6, PackID: 2, StringKey: domain.StringKey("svc_msg_welcome"), StringValue: "Welcome {name}, you have {count} new messages", TemplateType: domain.TemplateNamed, ParamsSchema: `[{"name":"name","type":"string","required":true},{"name":"count","type":"int","required":true}]`, OperationType: domain.OperationAdd, Version: 1},
	}

	return repo
}
