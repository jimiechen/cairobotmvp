package sdk

import (
	"context"
	"fmt"
	"time"

	"github.com/jimiechen/mineplanet/go/common-lib/config"
)

// TarsRemoteClient TarsGo 远程客户端实现
type TarsRemoteClient struct {
	servantName  string
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
	return nil, fmt.Errorf("TarsGo integration not yet implemented, servant: %s, method: %s", c.servantName, method)
}

// LoadRemoteConfig 从统一配置加载远程客户端配置
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
