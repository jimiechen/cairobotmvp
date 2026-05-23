package adapter

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
	"github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/services/i18n/service"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// I18nAdapter I18n 模块的 Tars servant 适配器
// 负责将 Tars 调用分发到 i18n service 的具体方法
// 不包含业务逻辑，只做方法分发和返回码转换
type I18nAdapter struct {
	i18nSvc service.I18nService
}

// NewI18nAdapter 创建 I18nAdapter 实例
func NewI18nAdapter(i18nSvc service.I18nService) *I18nAdapter {
	return &I18nAdapter{
		i18nSvc: i18nSvc,
	}
}

// Invoke 执行 Tars 调用分发
// method: 从 extend["method"] 获取目标方法名（GetAppLanguage / GetLangPack / GetLangDifference）
// request: Protobuf 序列化的请求 bytes（符合 routes.yaml 中的 tars_request_type: vector<byte>）
func (a *I18nAdapter) Invoke(ctx context.Context, request []byte, extend map[string]string) (int, []byte, error) {
	method := extend["method"]
	switch method {
	case "GetAppLanguage":
		return a.handleGetAppLanguage(ctx, request)
	case "GetLangPack":
		return a.handleGetLangPack(ctx, request)
	case "GetLangDifference":
		return a.handleGetLangDifference(ctx, request)
	default:
		return commonlib.CodeNotFound, nil, fmt.Errorf("unknown method: %s", method)
	}
}

// handleGetAppLanguage 处理获取语言列表请求
// 流程：Protobuf 反序列化 → 调用 service → Protobuf 序列化响应
func (a *I18nAdapter) handleGetAppLanguage(ctx context.Context, req []byte) (int, []byte, error) {
	var protoReq pb.AppFetchLanguageReq
	if err := proto.Unmarshal(req, &protoReq); err != nil {
		return commonlib.CodeBadRequest, nil, fmt.Errorf("protobuf unmarshal failed: %w", err)
	}

	languages, err := a.i18nSvc.GetLanguages(protoReq.GetClientVersion())
	if err != nil {
		return commonlib.CodeInternalError, nil, fmt.Errorf("get languages failed: %w", err)
	}

	protoResp := &pb.AppFetchLanguageRsp{
		Result: &pb.Result{
			Code:    commonlib.CodeSuccess,
			Message: "success",
		},
	}

	for _, lang := range languages {
		protoResp.Languages = append(protoResp.Languages, &pb.I18nLanguageMeta{
			Code:       lang.Code,
			Name:       lang.Name,
			NativeName: lang.NativeName,
			IsDefault:  lang.IsDefault,
		})
	}

	respBytes, err := proto.Marshal(protoResp)
	if err != nil {
		return commonlib.CodeInternalError, nil, fmt.Errorf("protobuf marshal failed: %w", err)
	}

	return commonlib.CodeSuccess, respBytes, nil
}

// handleGetLangPack 处理获取全量语言包请求
// 流程：Protobuf 反序列化 → 调用 service → Protobuf 序列化响应
func (a *I18nAdapter) handleGetLangPack(ctx context.Context, req []byte) (int, []byte, error) {
	var protoReq pb.AppFetchLangPackReq
	if err := proto.Unmarshal(req, &protoReq); err != nil {
		return commonlib.CodeBadRequest, nil, fmt.Errorf("protobuf unmarshal failed: %w", err)
	}

	resp, err := a.i18nSvc.GetLangPack(protoReq.GetLangCode(), protoReq.GetClientVersion(), "")
	if err != nil {
		return commonlib.CodeInternalError, nil, fmt.Errorf("get lang pack failed: %w", err)
	}

	protoResp := &pb.AppFetchLangPackRsp{
		Result: &pb.Result{
			Code:    commonlib.CodeSuccess,
			Message: "success",
		},
		PackVersion: resp.PackVersion,
	}

	for _, s := range resp.Strings {
		protoResp.Strings = append(protoResp.Strings, &pb.LangStringEntry{
			Key:            s.Key,
			Value:          s.Value,
			TemplateType:   s.TemplateType,
			OperationType:  s.OperationType,
		})
	}

	respBytes, err := proto.Marshal(protoResp)
	if err != nil {
		return commonlib.CodeInternalError, nil, fmt.Errorf("protobuf marshal failed: %w", err)
	}

	return commonlib.CodeSuccess, respBytes, nil
}

// handleGetLangDifference 处理获取增量语言包请求
// 流程：Protobuf 反序列化 → 调用 service → Protobuf 序列化响应
func (a *I18nAdapter) handleGetLangDifference(ctx context.Context, req []byte) (int, []byte, error) {
	var protoReq pb.AppFetchLangDifferenceReq
	if err := proto.Unmarshal(req, &protoReq); err != nil {
		return commonlib.CodeBadRequest, nil, fmt.Errorf("protobuf unmarshal failed: %w", err)
	}

	resp, err := a.i18nSvc.GetLangDifference(
		protoReq.GetLangCode(),
		protoReq.GetSinceVersion(),
		protoReq.GetClientVersion(),
		"",
	)
	if err != nil {
		return commonlib.CodeInternalError, nil, fmt.Errorf("get lang difference failed: %w", err)
	}

	protoResp := &pb.AppFetchLangDifferenceRsp{
		Result: &pb.Result{
			Code:    commonlib.CodeSuccess,
			Message: "success",
		},
		CurrentVersion: resp.CurrentVersion,
		Deletions:      resp.Deletions,
	}

	for _, s := range resp.Additions {
		protoResp.Additions = append(protoResp.Additions, &pb.LangStringEntry{
			Key:            s.Key,
			Value:          s.Value,
			TemplateType:   s.TemplateType,
			OperationType:  s.OperationType,
		})
	}

	respBytes, err := proto.Marshal(protoResp)
	if err != nil {
		return commonlib.CodeInternalError, nil, fmt.Errorf("protobuf marshal failed: %w", err)
	}

	return commonlib.CodeSuccess, respBytes, nil
}
