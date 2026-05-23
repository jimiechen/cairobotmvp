package sdk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// GetString 获取字符串类型配置值
// 查询路径：L1 LRU → L2 Redis → L3 远程服务
func (c *configClient) GetString(ctx context.Context, moduleKey, fieldKey string) (string, error) {
	tv, err := c.getField(ctx, moduleKey, fieldKey)
	if err != nil {
		return "", err
	}
	return tv.String(), nil
}

// GetInt 获取整数类型配置值（int64）
// 用于 FieldTypeInt / FieldTypeEnum 场景
func (c *configClient) GetInt(ctx context.Context, moduleKey, fieldKey string) (int64, error) {
	tv, err := c.getField(ctx, moduleKey, fieldKey)
	if err != nil {
		return 0, err
	}
	return tv.Int(), nil
}

// GetBool 获取布尔类型配置值
func (c *configClient) GetBool(ctx context.Context, moduleKey, fieldKey string) (bool, error) {
	tv, err := c.getField(ctx, moduleKey, fieldKey)
	if err != nil {
		return false, err
	}
	return tv.Bool(), nil
}

// GetFloat 获取浮点数类型配置值（float64）
// 用于 FieldTypeFloat 场景
func (c *configClient) GetFloat(ctx context.Context, moduleKey, fieldKey string) (float64, error) {
	tv, err := c.getField(ctx, moduleKey, fieldKey)
	if err != nil {
		return 0, err
	}
	return tv.Float(), nil
}

// GetJSON 获取 JSON 类型配置值并反序列化到 out
// 用于 FieldTypeJSON / FieldTypeList 场景
// out 必须是指向结构体或 map 的指针
func (c *configClient) GetJSON(ctx context.Context, moduleKey, fieldKey string, out any) error {
	tv, err := c.getField(ctx, moduleKey, fieldKey)
	if err != nil {
		return err
	}
	raw := tv.JSON()
	if raw == nil {
		return ErrTypeMismatch
	}
	return json.Unmarshal(raw, out)
}

// getField 获取指定模块字段的 TypedValue（核心查询逻辑）
// 1. 先查 L1 LRU 缓存
// 2. miss 则查 L2 Redis（如果配置了）
// 3. 再 miss 则查 L3 远程服务
// 4. 从 ModuleSnapshot 中按 fieldKey 取 TypedValue
func (c *configClient) getField(ctx context.Context, moduleKey, fieldKey string) (*domain.TypedValue, error) {
	cacheKey := buildCacheKey(moduleKey)
	snapshot, ok := c.lruCache.get(cacheKey)
	if ok {
		return extractField(snapshot, fieldKey)
	}
	snapshot, err := c.fetchModule(ctx, moduleKey)
	if err != nil {
		return nil, fmt.Errorf("fetch module %s failed: %w", moduleKey, err)
	}
	c.lruCache.set(cacheKey, snapshot)
	return extractField(snapshot, fieldKey)
}

// extractField 从快照中提取指定字段
// 找不到返回 ErrFieldNotFound
func extractField(snapshot *ModuleSnapshot, fieldKey string) (*domain.TypedValue, error) {
	tv := snapshot.GetField(fieldKey)
	if tv == nil {
		return nil, fmt.Errorf("%w: %s.%s", ErrFieldNotFound, snapshot.ModuleKey, fieldKey)
	}
	return tv, nil
}

// buildCacheKey 构建缓存 key
// 格式: {env}:{module_key}
func buildCacheKey(moduleKey string) string {
	return "sdk:" + moduleKey
}
