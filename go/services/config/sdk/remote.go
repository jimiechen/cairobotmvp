package sdk

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
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
	Invoke(ctx context.Context, method string, request []byte) ([]byte, error)
}

// RemoteOptions 远程模式扩展选项
type RemoteOptions struct {
	RemoteClient  RemoteClient
	Timeout       time.Duration
	RetryCount    int
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
