package sdk

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"github.com/jimiechen/mineplanet/go/common-lib/config"
	"github.com/jimiechen/mineplanet/go/services/i18n/service"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// 远程模式错误定义
var (
	// ErrRemoteClientRequired Remote 模式需要 RemoteClient
	ErrRemoteClientRequired = errors.New("remote client required for remote mode")
)

// remoteClient 远程调用客户端
//
// 职责：
// - 通过 TarsGo 协议调用远程 I18nServer
// - 提供与 InProcess 模式相同的接口语义
// - 使用 Protobuf 序列化/反序列化
type remoteClient struct {
	options *Options
}

func newRemoteClient(opts *Options) *remoteClient {
	return &remoteClient{
		options: opts,
	}
}

// getLangPack 从远程服务获取全量语言包
// 使用 Protobuf 序列化/反序列化，支持超时控制
func (r *remoteClient) getLangPack(ctx context.Context, langCode string) (*service.LangPackResponse, error) {
	if r.options.RemoteClient == nil {
		return nil, fmt.Errorf("remote client not initialized")
	}

	// 构建 Protobuf 请求
	protoReq := &pb.AppFetchLangPackReq{
		LangCode:      langCode,
		ClientVersion: r.options.ClientVersion,
	}

	reqBytes, err := proto.Marshal(protoReq)
	if err != nil {
		return nil, fmt.Errorf("protobuf marshal failed: %w", err)
	}

	// 设置超时
	if r.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.options.Timeout)
		defer cancel()
	}

	// 调用远程服务
	respBytes, err := r.options.RemoteClient.Invoke(ctx, "GetLangPack", reqBytes)
	if err != nil {
		return nil, fmt.Errorf("remote invoke failed: %w", err)
	}

	// 反序列化响应
	var protoResp pb.AppFetchLangPackRsp
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
	resp := &service.LangPackResponse{
		PackVersion: protoResp.GetPackVersion(),
		Strings:     make([]service.LangStringEntry, 0, len(protoResp.GetStrings())),
	}

	for _, s := range protoResp.GetStrings() {
		resp.Strings = append(resp.Strings, service.LangStringEntry{
			Key:           s.GetKey(),
			Value:         s.GetValue(),
			TemplateType:  s.GetTemplateType(),
			OperationType: s.GetOperationType(),
		})
	}

	return resp, nil
}

// getLangDifference 从远程服务获取增量语言包
// 使用 Protobuf 序列化/反序列化，支持超时控制
func (r *remoteClient) getLangDifference(ctx context.Context, langCode string, sinceVersion int64) (*service.LangDiffResponse, error) {
	if r.options.RemoteClient == nil {
		return nil, fmt.Errorf("remote client not initialized")
	}

	// 构建 Protobuf 请求
	protoReq := &pb.AppFetchLangDifferenceReq{
		LangCode:      langCode,
		SinceVersion:  sinceVersion,
		ClientVersion: r.options.ClientVersion,
	}

	reqBytes, err := proto.Marshal(protoReq)
	if err != nil {
		return nil, fmt.Errorf("protobuf marshal failed: %w", err)
	}

	// 设置超时
	if r.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.options.Timeout)
		defer cancel()
	}

	// 调用远程服务
	respBytes, err := r.options.RemoteClient.Invoke(ctx, "GetLangDifference", reqBytes)
	if err != nil {
		return nil, fmt.Errorf("remote invoke failed: %w", err)
	}

	// 反序列化响应
	var protoResp pb.AppFetchLangDifferenceRsp
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
	resp := &service.LangDiffResponse{
		CurrentVersion: protoResp.GetCurrentVersion(),
		Additions:      make([]service.LangStringEntry, 0, len(protoResp.GetAdditions())),
		Deletions:      protoResp.GetDeletions(),
	}

	for _, s := range protoResp.GetAdditions() {
		resp.Additions = append(resp.Additions, service.LangStringEntry{
			Key:           s.GetKey(),
			Value:         s.GetValue(),
			TemplateType:  s.GetTemplateType(),
			OperationType: s.GetOperationType(),
		})
	}

	return resp, nil
}

// ping 检查远程服务可用性
// 发送语言列表请求作为心跳包
func (r *remoteClient) ping(ctx context.Context) error {
	if r.options.RemoteClient == nil {
		return fmt.Errorf("remote client not initialized")
	}

	// 发送语言列表请求作为心跳
	protoReq := &pb.AppFetchLanguageReq{
		ClientVersion: r.options.ClientVersion,
	}

	reqBytes, err := proto.Marshal(protoReq)
	if err != nil {
		return fmt.Errorf("protobuf marshal failed: %w", err)
	}

	if r.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.options.Timeout)
		defer cancel()
	}

	_, err = r.options.RemoteClient.Invoke(ctx, "GetAppLanguage", reqBytes)
	return err
}

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

// Invoke 调用远程方法
func (c *TarsRemoteClient) Invoke(ctx context.Context, method string, request []byte) ([]byte, error) {
	// TODO(tars, MVP2): 集成 TarsGo 框架
	// 1. 初始化 Tars 通信器
	// 2. 查找 servant 代理
	// 3. 调用远程方法
	// 4. 返回响应
	return nil, fmt.Errorf("TarsGo integration not yet implemented, servant: %s, method: %s", c.servantName, method)
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
