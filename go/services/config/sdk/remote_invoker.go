package sdk

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/service"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// fetchInProcess 进程内模式：直接调用本地 ConfigService
func (c *configClient) fetchInProcess(ctx context.Context, moduleKey string) (*ModuleSnapshot, error) {
	if c.options.Service == nil {
		return nil, ErrServiceRequired
	}
	req := &service.AppConfigRequest{
		Env:              c.options.Env,
		ClientScope:      c.options.ClientScope,
		RequestedModules: []string{moduleKey},
	}
	resp, err := c.options.Service.GetAppConfigs(req)
	if err != nil {
		return nil, fmt.Errorf("get app configs failed: %w", err)
	}
	return extractModule(resp, moduleKey)
}

// fetchRemote 远程模式：通过 TarsGo 调用远程 ConfigServer
// 使用 Protobuf 序列化/反序列化，支持超时控制和重试策略
func (c *configClient) fetchRemote(ctx context.Context, moduleKey string) (*ModuleSnapshot, error) {
	if c.options.RemoteClient == nil {
		return nil, fmt.Errorf("remote client not initialized")
	}

	protoReq := &pb.AppConfigsReq{
		Env:              c.options.Env,
		ClientScope:      c.options.ClientScope,
		RequestedModules: []string{moduleKey},
	}

	reqBytes, err := proto.Marshal(protoReq)
	if err != nil {
		return nil, fmt.Errorf("protobuf marshal failed: %w", err)
	}

	if c.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.options.Timeout)
		defer cancel()
	}

	respBytes, err := c.options.RemoteClient.Invoke(ctx, "GetAppConfigs", reqBytes)
	if err != nil {
		return nil, fmt.Errorf("remote invoke failed: %w", err)
	}

	var protoResp pb.AppConfigsRsp
	if err := proto.Unmarshal(respBytes, &protoResp); err != nil {
		return nil, fmt.Errorf("protobuf unmarshal failed: %w", err)
	}

	if protoResp.GetResult().GetCode() != 0 {
		return nil, fmt.Errorf("remote service error: code=%d, message=%s",
			protoResp.GetResult().GetCode(),
			protoResp.GetResult().GetMessage())
	}

	resp := &service.AppConfigResponse{
		StaticModules:  make(map[string]map[string]*domain.TypedValue),
		DynamicModules: make([]*service.DynamicModuleView, 0),
	}

	if protoResp.BaseCfg != nil {
		resp.StaticModules["base_cfg"] = map[string]*domain.TypedValue{
			"domain_root":     {Value: protoResp.BaseCfg.GetDomainRoot()},
			"domain_wap":      {Value: protoResp.BaseCfg.GetDomainWap()},
			"sign_rand":       {Value: protoResp.BaseCfg.GetSignRand()},
			"construct_email": {Value: protoResp.BaseCfg.GetConstructEmail()},
		}
	}

	for _, dm := range protoResp.GetDynamicModules() {
		view := &service.DynamicModuleView{
			ModuleKey: dm.GetModuleKey(),
			Version:   dm.GetVersion(),
			Fields:    make(map[string]*domain.TypedValue),
		}
		for fieldKey, fieldValue := range dm.GetFields() {
			view.Fields[fieldKey] = &domain.TypedValue{
				Value: fieldValue,
			}
		}
		resp.DynamicModules = append(resp.DynamicModules, view)
	}

	return extractModule(resp, moduleKey)
}
