package admin

import (
	"encoding/json"
	"time"

	"github.com/jimiechen/mineplanet/go/services/config/sdk"
)

// adminInvalidateEvent 构造 admin 发出的 InvalidateEvent
func sdkInvalidateEvent(moduleKeys []string) sdk.InvalidateEvent {
	return sdk.InvalidateEvent{
		TenantID:   "default",
		Scope:      "config",
		Env:        "dev",
		ModuleKeys: moduleKeys,
		Version:    time.Now().UnixMilli(),
		Timestamp:  time.Now().Unix(),
	}
}

// marshalEvent 将 InvalidateEvent 序列化为 JSON
func marshalEvent(evt sdk.InvalidateEvent) (string, error) {
	data, err := json.Marshal(evt)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
