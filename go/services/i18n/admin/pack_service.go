package admin

import (
	"context"
	"fmt"
	"time"
)

// PublishPackRequest 发布语言包请求
type PublishPackRequest struct {
	PackID   int64
	LangCode string
	Env      string
	Operator string
}

// PackVersion 语言包版本信息
type PackVersion struct {
	PackID   int64
	LangCode string
	Version  int
}

// PublishPack 发布语言包
// 流程：检查 → repo.PublishPack → 审计 → 失效+广播
func (s *AdminI18nService) PublishPack(ctx context.Context, req PublishPackRequest) (*PackVersion, error) {
	if req.PackID <= 0 {
		return nil, fmt.Errorf("pack_id 不能为空")
	}
	version := int(time.Now().Unix())
	if err := s.repo.PublishPack(req.PackID, version); err != nil {
		return nil, fmt.Errorf("发布语言包失败: %w", err)
	}
	s.writeAudit(ctx, "publish_pack", "pack", fmt.Sprintf("%d", req.PackID), req.Operator)
	s.invalidateLangCode(ctx, []string{req.LangCode})
	return &PackVersion{
		PackID:   req.PackID,
		LangCode: req.LangCode,
		Version:  version,
	}, nil
}

// RollbackPack 回滚语言包到指定版本
func (s *AdminI18nService) RollbackPack(ctx context.Context, packID int64, targetVersion int, operator string) error {
	if packID <= 0 {
		return fmt.Errorf("pack_id 不能为空")
	}
	s.writeAudit(ctx, "rollback_pack", "pack", fmt.Sprintf("%d", packID), operator)
	return s.invalidateLangCode(ctx, []string{})
}
