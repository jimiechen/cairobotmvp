package sdk

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"github.com/jimiechen/mineplanet/go/common-lib/config"
	"github.com/jimiechen/mineplanet/go/services/config/domain"
	"github.com/jimiechen/mineplanet/go/services/config/service"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// fetchModule 从数据源获取模块配置
// InProcess 模式：直接调用 ConfigService.GetAppConfigs
// Remote 模式：通过 TarsGo 调用远程服务
func (c *configClient) fetchModule(ctx context.Context, moduleKey string) (*ModuleSnapshot, error) {
	switch c.options.Mode {
	case ModeInProcess:
		return c.fetchInProcess(ctx, moduleKey)
	case ModeRemote:
		return c.fetchRemote(ctx, moduleKey)
	default:
		return nil, ErrUnsupportedMode
	}
}

// fetchInProcess 进程内模式：直接调用本地 ConfigService
func (c *configClient) fetchInProcess(ctx context.Context, moduleKey string) (*ModuleSnapshot, error) {
	if c.options.Service == nil {
		return nil, ErrServiceRequired
	}
	req := &service.AppConfigRequest{
		Env:               c.options.Env,
		ClientScope:       c.options.ClientScope,
		RequestedModules:  []string{moduleKey},
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

	// 构建 Protobuf 请求
	protoReq := &pb.AppConfigsReq{
		Env:               c.options.Env,
		ClientScope:       c.options.ClientScope,
		RequestedModules:  []string{moduleKey},
	}

	reqBytes, err := proto.Marshal(protoReq)
	if err != nil {
		return nil, fmt.Errorf("protobuf marshal failed: %w", err)
	}

	// 设置超时
	if c.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.options.Timeout)
		defer cancel()
	}

	// 调用远程服务
	respBytes, err := c.options.RemoteClient.Invoke(ctx, "GetAppConfigs", reqBytes)
	if err != nil {
		return nil, fmt.Errorf("remote invoke failed: %w", err)
	}

	// 反序列化响应
	var protoResp pb.AppConfigsRsp
	if err := proto.Unmarshal(respBytes, &protoResp); err != nil {
		return nil, fmt.Errorf("protobuf unmarshal failed: %w", err)
	}

	// 检查响应码
	if protoResp.GetResult().GetCode() != 0 {
		return nil, fmt.Errorf("remote service error: code=%d, message=%s",
			protoResp.GetResult().GetCode(),
			protoResp.GetResult().GetMessage())
	}

	// 转换为 service 层响应
	resp := &service.AppConfigResponse{
		StaticModules:  make(map[string]map[string]*domain.TypedValue),
		DynamicModules: make([]*service.DynamicModuleView, 0),
	}

	// 提取静态模块
	if protoResp.BaseCfg != nil {
		resp.StaticModules["base_cfg"] = map[string]*domain.TypedValue{
			"domain_root":     {Value: protoResp.BaseCfg.GetDomainRoot()},
			"domain_wap":      {Value: protoResp.BaseCfg.GetDomainWap()},
			"sign_rand":       {Value: protoResp.BaseCfg.GetSignRand()},
			"construct_email": {Value: protoResp.BaseCfg.GetConstructEmail()},
		}
	}

	// 提取动态模块
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

// extractModule 从 AppConfigResponse 中提取指定模块的快照
// 静态模块从 StaticModules map 中取，动态模块从 DynamicModules 列表中找
func extractModule(resp *service.AppConfigResponse, moduleKey string) (*ModuleSnapshot, error) {
	snapshot := &ModuleSnapshot{
		ModuleKey: moduleKey,
		Fields:    make(map[string]*domain.TypedValue),
	}
	if domain.IsStaticModule(moduleKey) {
		if fields, exists := resp.StaticModules[moduleKey]; exists {
			snapshot.Fields = fields
		}
	} else {
		for _, mod := range resp.DynamicModules {
			if mod.ModuleKey == moduleKey {
				snapshot.Fields = mod.Fields
				break
			}
		}
	}
	return snapshot, nil
}

// pingService 检查配置服务是否可用
// InProcess 模式下调用 GetVersionInfo 验证连接
// Remote 模式下发送心跳包
func (c *configClient) pingService(ctx context.Context) error {
	switch c.options.Mode {
	case ModeInProcess:
		if c.options.Service == nil {
			return ErrServiceRequired
		}
		_, err := c.options.Service.GetVersionInfo(c.options.Env, nil)
		return err
	case ModeRemote:
		if c.options.RemoteClient == nil {
			return fmt.Errorf("remote client not initialized")
		}
		// 发送版本轮询请求作为心跳
		protoReq := &pb.AppConfigVersionReq{
			Env: c.options.Env,
		}
		reqBytes, err := proto.Marshal(protoReq)
		if err != nil {
			return fmt.Errorf("protobuf marshal failed: %w", err)
		}

		if c.options.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, c.options.Timeout)
			defer cancel()
		}

		_, err = c.options.RemoteClient.Invoke(ctx, "AppConfigVersion", reqBytes)
		return err
	default:
		return ErrUnsupportedMode
	}
}

// RemoteClient 远程服务客户端接口
// 抽象 TarsGo 或其他 RPC 框架的具体实现
type RemoteClient interface {
	// Invoke 调用远程方法
	// method: 方法名（GetAppConfigs / AppConfigVersion）
	// request: Protobuf 序列化的请求 bytes
	// 返回: 响应 bytes 或错误
	Invoke(ctx context.Context, method string, request []byte) ([]byte, error)
}

// Options 扩展选项
type RemoteOptions struct {
	// RemoteClient 远程客户端实例
	RemoteClient RemoteClient
	// Timeout 远程调用超时
	Timeout time.Duration
	// RetryCount 重试次数
	RetryCount int
	// RetryInterval 重试间隔
	RetryInterval time.Duration
}

// WithRemoteClient 设置远程客户端
func WithRemoteClient(client RemoteClient) Option {
	return func(o *Options) {
		o.RemoteClient = client
	}
}

// WithTimeout 设置超时
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithRetry 设置重试策略
func WithRetry(count int, interval time.Duration) Option {
	return func(o *Options) {
		o.RetryCount = count
		o.RetryInterval = interval
	}
}

// TarsRemoteClient TarsGo 远程客户端实现
type TarsRemoteClient struct {
	// servantName Tars servant 名称
	servantName string
	// communicator Tars 通信器
	communicator interface{}
}

// NewTarsRemoteClient 创建 Tars 远程客户端
func NewTarsRemoteClient(servantName string) *TarsRemoteClient {
	return &TarsRemoteClient{
		servantName: servantName,
	}
}

// Invoke 调用远程方法
func (c *TarsRemoteClient) Invoke(ctx context.Context, method string, request []byte) ([]byte, error) {
	// TODO: 集成 TarsGo 框架
	// 1. 初始化 Tars 通信器
	// 2. 查找 servant 代理
	// 3. 调用远程方法
	// 4. 返回响应
	return nil, fmt.Errorf("TarsGo integration not yet implemented, servant: %s, method: %s", c.servantName, method)
}

// LoadConfig 从统一配置加载远程客户端配置
func LoadRemoteConfig(cfg *config.ServerConfig) *RemoteOptions {
	opts := &RemoteOptions{
		Timeout:       5 * time.Second,
		RetryCount:    3,
		RetryInterval: 1 * time.Second,
	}

	if cfg.Cache.ConfigTTLSeconds > 0 {
		opts.Timeout = time.Duration(cfg.Cache.ConfigTTLSeconds) * time.Second
	}

	return opts
}
