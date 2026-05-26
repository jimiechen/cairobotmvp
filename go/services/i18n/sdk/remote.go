package sdk

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
	"github.com/jimiechen/mineplanet/go/services/i18n/service"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
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

	protoReq := &pb.AppFetchLangPackReq{
		LangCode:      langCode,
		ClientVersion: r.options.ClientVersion,
	}

	reqBytes, err := proto.Marshal(protoReq)
	if err != nil {
		return nil, fmt.Errorf("protobuf marshal failed: %w", err)
	}

	if r.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.options.Timeout)
		defer cancel()
	}

	respBytes, err := r.options.RemoteClient.Invoke(ctx, "GetLangPack", reqBytes)
	if err != nil {
		return nil, fmt.Errorf("remote invoke failed: %w", err)
	}

	var protoResp pb.AppFetchLangPackRsp
	if err := proto.Unmarshal(respBytes, &protoResp); err != nil {
		return nil, fmt.Errorf("protobuf unmarshal failed: %w", err)
	}

	if protoResp.GetResult().GetCode() != 0 {
		return nil, fmt.Errorf("remote service error: code=%d, message=%s",
			protoResp.GetResult().GetCode(),
			protoResp.GetResult().GetMessage())
	}

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

	protoReq := &pb.AppFetchLangDifferenceReq{
		LangCode:      langCode,
		SinceVersion:  sinceVersion,
		ClientVersion: r.options.ClientVersion,
	}

	reqBytes, err := proto.Marshal(protoReq)
	if err != nil {
		return nil, fmt.Errorf("protobuf marshal failed: %w", err)
	}

	if r.options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.options.Timeout)
		defer cancel()
	}

	respBytes, err := r.options.RemoteClient.Invoke(ctx, "GetLangDifference", reqBytes)
	if err != nil {
		return nil, fmt.Errorf("remote invoke failed: %w", err)
	}

	var protoResp pb.AppFetchLangDifferenceRsp
	if err := proto.Unmarshal(respBytes, &protoResp); err != nil {
		return nil, fmt.Errorf("protobuf unmarshal failed: %w", err)
	}

	if protoResp.GetResult().GetCode() != 0 {
		return nil, fmt.Errorf("remote service error: code=%d, message=%s",
			protoResp.GetResult().GetCode(),
			protoResp.GetResult().GetMessage())
	}

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
