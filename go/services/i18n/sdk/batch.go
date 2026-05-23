package sdk

import (
	"context"
)

// BatchT 批量翻译多个 key
//
// 优化策略：
// 1. 一次拉取整个语言包（避免逐 key 查询）
// 2. 在本地缓存中批量查找
// 3. 对每个 key 执行翻译渲染
//
// 返回 map[key]result，某个 key 翻译失败时返回该 key 的原始值
func (c *clientImpl) BatchT(ctx context.Context, langCode string, keys []string, params map[string]any) (map[string]string, error) {
	results := make(map[string]string, len(keys))

	templates, err := c.loadPack(ctx, langCode)
	if err != nil {
		for _, key := range keys {
			results[key] = key // fallback: 加载失败时返回 key 本身
		}
		return results, err
	}

	for _, key := range keys {
		tmpl, exists := templates[key]
		if !exists {
			results[key] = key // fallback: key 不存在时返回 key 本身
			continue
		}

		switch tmpl.TemplateType {
		case "plain":
			results[key] = tmpl.Value
		case "named":
			rendered, err := renderNamedTemplate(tmpl.Value, params, tmpl.Params)
			if err != nil {
				results[key] = key // 渲染失败时返回 key
				continue
			}
			results[key] = rendered
		case "icu":
			results[key] = key // ICU 不支持时返回 key
		default:
			results[key] = tmpl.Value
		}
	}

	return results, nil
}
