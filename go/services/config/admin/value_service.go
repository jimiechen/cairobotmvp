package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/service"
)

// PublishValueRequest 发布配置值请求 DTO
type PublishValueRequest struct {
	ModuleKey string
	Env       string
	Fields    map[string]*domain.TypedValue
	Operator  string
}

// ValueVersion 配置版本返回值
type ValueVersion struct {
	Version    int64
	ModuleKey  string
	Env        string
	FieldCount int
}

// PublishValue 发布配置值（创建新版本）
// 流程：DTO → ValidateConfigMap → 序列化 ConfigJSON → repo.Save → 审计 → Invalidate + Publish
func (s *AdminConfigService) PublishValue(ctx context.Context, req PublishValueRequest) (*ValueVersion, error) {
	if req.ModuleKey == "" {
		return nil, fmt.Errorf("module_key 不能为空")
	}
	if len(req.Fields) == 0 {
		return nil, fmt.Errorf("fields 不能为空")
	}
	moduleSchema, err := s.innerSchema.GetModuleSchema(req.ModuleKey)
	if err != nil {
		return nil, fmt.Errorf("获取模块 schema 失败: %w", err)
	}
	valErrors := service.ValidateConfigMap(req.Fields, moduleSchema)
	if len(valErrors) > 0 {
		return nil, &ValidationError{Errors: valErrors}
	}
	configJSON, err := json.Marshal(req.Fields)
	if err != nil {
		return nil, fmt.Errorf("序列化配置值失败: %w", err)
	}
	version := &domain.ConfigVersion{
		ModuleKey:  req.ModuleKey,
		Env:        req.Env,
		ConfigJSON: string(configJSON),
		Version:    time.Now().UnixMilli(),
		IsPublished: true,
		PublishedAt: ptrTime(time.Now()),
		CreateBy:   req.Operator,
		UpdateBy:   req.Operator,
	}
	if err := s.configRepo.Save(version); err != nil {
		return nil, fmt.Errorf("保存配置版本失败: %w", err)
	}
	s.writeAudit(ctx, "publish_value", "value", fmt.Sprintf("%d", version.Version), req.Operator, map[string]string{
		"module_key":  req.ModuleKey,
		"env":         req.Env,
		"field_count": fmt.Sprintf("%d", len(req.Fields)),
	})
	if err := s.invalidateAndBroadcast(ctx, []string{req.ModuleKey}); err != nil {
		return &ValueVersion{
			Version:    version.Version,
			ModuleKey:  req.ModuleKey,
			Env:        req.Env,
			FieldCount: len(req.Fields),
		}, nil
	}
	return &ValueVersion{
		Version:    version.Version,
		ModuleKey:  req.ModuleKey,
		Env:        req.Env,
		FieldCount: len(req.Fields),
	}, nil
}

// ValidationError 校验错误汇总（携带所有字段级错误）
type ValidationError struct {
	Errors []service.ValidationError
}

func (e *ValidationError) Error() string {
	msg := "配置值校验失败:"
	for _, ve := range e.Errors {
		msg += " [" + ve.Field + ": " + ve.Reason + "]"
	}
	return msg
}

func ptrTime(t time.Time) *time.Time { return &t }
