package adapter

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
	"github.com/jimiechen/mineplanet/go/common-lib"
	"github.com/jimiechen/mineplanet/go/services/config/service"
	pb "github.com/jimiechen/mineplanet/protocols/generated/go/base"
)

// ConfigAdapter Config 模块的 Tars servant 适配器
// 负责将 Tars 调用分发到 config service 的具体方法
// 不包含业务逻辑，只做方法分发和返回码转换
type ConfigAdapter struct {
	configSvc service.ConfigService
}

// NewConfigAdapter 创建 ConfigAdapter 实例
func NewConfigAdapter(configSvc service.ConfigService) *ConfigAdapter {
	return &ConfigAdapter{
		configSvc: configSvc,
	}
}

// Invoke 执行 Tars 调用分发
// method: 从 extend["method"] 获取目标方法名（GetAppConfigs / AppConfigVersion）
// request: Protobuf 序列化的请求 bytes（符合 routes.yaml 中的 tars_request_type: vector<byte>）
func (a *ConfigAdapter) Invoke(ctx context.Context, request []byte, extend map[string]string) (int, []byte, error) {
	method := extend["method"]
	switch method {
	case "GetAppConfigs":
		return a.handleGetAppConfigs(ctx, request)
	case "AppConfigVersion":
		return a.handleAppConfigVersion(ctx, request)
	default:
		return commonlib.CodeNotFound, nil, fmt.Errorf("unknown method: %s", method)
	}
}

// handleGetAppConfigs 处理获取全量应用配置请求
// 流程：Protobuf 反序列化 → 调用 service → Protobuf 序列化响应
func (a *ConfigAdapter) handleGetAppConfigs(ctx context.Context, req []byte) (int, []byte, error) {
	var protoReq pb.AppConfigsReq
	if err := proto.Unmarshal(req, &protoReq); err != nil {
		return commonlib.CodeBadRequest, nil, fmt.Errorf("protobuf unmarshal failed: %w", err)
	}

	// 转换为 service 层请求
	appReq := &service.AppConfigRequest{
		Env:               protoReq.GetEnv(),
		ClientScope:       protoReq.GetClientScope(),
		ClientVersion:     protoReq.GetClientVersion(),
		RequestedModules:  protoReq.GetRequestedModules(),
	}

	resp, err := a.configSvc.GetAppConfigs(appReq)
	if err != nil {
		return commonlib.CodeInternalError, nil, fmt.Errorf("get app configs failed: %w", err)
	}

	// 转换为 Protobuf 响应
	protoResp := &pb.AppConfigsRsp{
		Result: &pb.Result{
			Code:    commonlib.CodeSuccess,
			Message: "success",
		},
	}

	// 设置动态模块
	for _, dm := range resp.DynamicModules {
		protoResp.DynamicModules = append(protoResp.DynamicModules, &pb.DynamicConfigModule{
			ModuleKey:    dm.ModuleKey,
			ModuleName:   dm.ModuleName,
			Version:      dm.Version,
			Fields:       convertToProtoFields(dm.Fields),
		})
	}

	respBytes, err := proto.Marshal(protoResp)
	if err != nil {
		return commonlib.CodeInternalError, nil, fmt.Errorf("protobuf marshal failed: %w", err)
	}

	return commonlib.CodeSuccess, respBytes, nil
}

// handleAppConfigVersion 处理配置版本轮询请求
// 流程：Protobuf 反序列化 → 调用 service → Protobuf 序列化响应
func (a *ConfigAdapter) handleAppConfigVersion(ctx context.Context, req []byte) (int, []byte, error) {
	var protoReq pb.AppConfigVersionReq
	if err := proto.Unmarshal(req, &protoReq); err != nil {
		return commonlib.CodeBadRequest, nil, fmt.Errorf("protobuf unmarshal failed: %w", err)
	}

	resp, err := a.configSvc.GetVersionInfo(protoReq.GetEnv(), protoReq.GetKnownVersions())
	if err != nil {
		return commonlib.CodeInternalError, nil, fmt.Errorf("get version info failed: %w", err)
	}

	protoResp := &pb.AppConfigVersionRsp{
		Result: &pb.Result{
			Code:    commonlib.CodeSuccess,
			Message: "success",
		},
		ConfigVersions:     resp.ConfigVersions,
		LangPackVersions:   resp.LangPackVersions,
		HasChanges:         resp.HasChanges,
	}

	respBytes, err := proto.Marshal(protoResp)
	if err != nil {
		return commonlib.CodeInternalError, nil, fmt.Errorf("protobuf marshal failed: %w", err)
	}

	return commonlib.CodeSuccess, respBytes, nil
}

// convertToProtoFields 将 service 层字段转换为 Protobuf 字段
func convertToProtoFields(fields []service.ConfigField) []*pb.ConfigField {
	var protoFields []*pb.ConfigField
	for _, f := range fields {
		protoFields = append(protoFields, &pb.ConfigField{
			FieldKey:     f.FieldKey,
			FieldName:    f.FieldName,
			FieldType:    f.FieldType,
			Value:        fmt.Sprintf("%v", f.Value),
			DefaultValue: fmt.Sprintf("%v", f.DefaultValue),
			Required:     f.Required,
		})
	}
	return protoFields
}
