package sdk

import (
	"context"
	"errors"
	"time"

	"github.com/jimiechen/mineplanet/go/common-lib/config"
)

// 远程模式错误定义
var (
	// ErrRemoteClientRequired Remote 模式需要 RemoteClient
	ErrRemoteClientRequired = errors.New("remote client required for remote mode")
)

// RemoteClient 远程服务客户端接口
// 抽象 TarsGo 或其他 RPC 框架的具体实现
type RemoteClient interface {
	// Invoke 调用远程方法
	// method: 方法名（GetAppLanguage / GetLangPack / GetLangDifference）
	// request: Protobuf 序列化的请求 bytes
	// 返回: 响应 bytes 或错误
	Invoke(ctx context.Context, method string, request []byte) ([]byte, error)
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

// LoadRemoteConfig 从统一配置加载远程客户端配置
func LoadRemoteConfig(cfg *config.ServerConfig) *Options {
	opts := &Options{
		Timeout:       5 * time.Second,
		RetryCount:    3,
		RetryInterval: 1 * time.Second,
	}

	if cfg.Cache.I18nTTLSeconds > 0 {
		opts.Timeout = time.Duration(cfg.Cache.I18nTTLSeconds) * time.Second
	}

	return opts
}
