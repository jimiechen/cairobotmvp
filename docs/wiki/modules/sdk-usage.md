# SDK 使用规范

## 1. 引言

本文档定义 configsdk 和 i18nsdk 的使用方式、接入约束和最佳实践。

## 2. 引入方式（使用完整模块路径）

```go
// Config SDK
import configsdk "github.com/jimiechen/mineplanet/go/services/config/sdk"

// I18N SDK
import i18nsdk "github.com/jimiechen/mineplanet/go/services/i18n/sdk"
```

## 3. 初始化模式

### 3.1 InProcess 模式（开发/测试）

```go
client, err := configsdk.New(configsdk.Options{
    Mode:     configsdk.ModeInProcess,
    DBPath:   ":memory:",
})
```

### 3.2 Remote 模式（生产）

```go
client, err := configsdk.New(configsdk.Options{
    Mode:      configsdk.ModeRemote,
    TarsAddr:  "tcp -h 127.0.0.1 -p 10001",
})
```

## 4. 核心接口

### 4.1 ConfigSDK

| 方法 | 说明 |
|------|------|
| GetString | 获取字符串配置 |
| GetInt | 获取整数配置 |
| GetBool | 获取布尔配置 |
| GetJSON | 获取 JSON 配置 |
| GetModule | 获取整个模块快照 |
| Watch | 监听变更 |
| Ping | 健康检查 |

### 4.2 I18NSDK

| 方法 | 说明 |
|------|------|
| T | 翻译（支持 named 占位符） |
| Raw | 获取原始模板 |
| BatchT | 批量翻译 |
| Watch | 监听语言包变更 |

## 5. 三层缓存模型

L1(进程LRU) → L2(Redis) → L3(Service)

## 6. Watch 与失效语义

订阅 cairobot.config.invalidate / cairobot.i18n.invalidate channel

## 7. 业务接入示例

### OpenAPI 服务

### 设备网关服务

### 用户中台服务

## 8. 禁止事项（铁律）

❌ 不得 import services/config/service 或 repository（必须只通过 sdk 包访问）
❌ 不得调用 admin-server API（铁律，业务服务 0 引用）
❌ 不得使用相对路径导入（必须使用完整 module path）
❌ 不得在业务代码中直接操作 Redis 缓存 key
❌ 不得绕过 SDK 直接连接 MySQL

## 9. 故障排查

常见错误码及解决方案
