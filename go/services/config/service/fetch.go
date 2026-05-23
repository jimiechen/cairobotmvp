package service

import (
	"fmt"

	"github.com/jimiechen/mineplanet/go/services/config/domain"
)

// GetAppConfigs 主入口：读缓存 → miss 则查 repo → 组装响应
// 缓存键格式: cfg:{env}:{module_key}，缓存值为 *domain.ConfigVersion
func (s *AppConfigService) GetAppConfigs(req *AppConfigRequest) (*AppConfigResponse, error) {
	if req.Env == "" {
		req.Env = "dev"
	}

	versions, err := s.loadAllPublishedVersions(req.Env)
	if err != nil {
		return nil, fmt.Errorf("加载已发布版本失败: %w", err)
	}

	response := &AppConfigResponse{
		StaticModules:   make(map[string]map[string]*domain.TypedValue),
		DynamicModules: make([]*DynamicModuleView, 0),
	}

	for _, ver := range versions {
		if len(req.RequestedModules) > 0 && !contains(req.RequestedModules, ver.ModuleKey) {
			continue
		}

		typedMap, err := ParseConfigJSON(ver.ConfigJSON, ver.ModuleKey, s.schemaRepo)
		if err != nil {
			return nil, fmt.Errorf("解析 %s config_json 失败: %w", ver.ModuleKey, err)
		}

		if domain.IsStaticModule(ver.ModuleKey) {
			response.StaticModules[ver.ModuleKey] = typedMap
		} else {
			dm := BuildDynamicModule(ver, typedMap, s.schemaRepo, req.ClientScope)
			response.DynamicModules = append(response.DynamicModules, dm)
		}
	}

	return response, nil
}

// loadAllPublishedVersions 先查缓存，miss 时回退到 repo 全量查询
// 每个模块独立缓存，避免单 key 过大
func (s *AppConfigService) loadAllPublishedVersions(env string) ([]*domain.ConfigVersion, error) {
	versions, err := s.configRepo.ListPublishedVersions(env)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

// contains 判断字符串切片是否包含目标值
func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

// GetVersionInfo 版本轮询：对比客户端已知版本，返回最新版本和是否有变更
// 用于 6009/6010 协议的轻量变更检测
func (s *AppConfigService) GetVersionInfo(env string, knownVersions map[string]int64) (*VersionInfoResponse, error) {
	if env == "" {
		env = "dev"
	}

	published, err := s.configRepo.ListPublishedVersions(env)
	if err != nil {
		return nil, fmt.Errorf("查询已发布版本失败: %w", err)
	}

	configVersions := make(map[string]int64)
	hasChanges := false

	for _, ver := range published {
		configVersions[ver.ModuleKey] = ver.Version
		if knownVer, ok := knownVersions[ver.ModuleKey]; !ok || ver.Version > knownVer {
			hasChanges = true
		}
	}

	return &VersionInfoResponse{
		ConfigVersions: configVersions,
		HasChanges:     hasChanges,
	}, nil
}
