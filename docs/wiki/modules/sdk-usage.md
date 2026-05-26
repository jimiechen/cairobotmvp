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

### 7.1 Hello 模块（configsdk + i18nsdk 完整范例）✅ **推荐参考**

**演示能力**：
- ✅ 强类型配置读取（GetString / GetInt）
- ✅ 配置驱动校验（max_name_length）
- ✅ 服务端 i18n 渲染（named 模板）
- ✅ 失败降级机制

**代码位置**：[go/modules/hello/](../../../go/modules/hello/)
- [usecase.go (97行)](../../../go/modules/hello/usecase.go) - 核心业务逻辑
- [usecase_test.go (247行)](../../../go/modules/hello/usecase_test.go) - 单元测试

### 7.2 Health 模块（i18nsdk ICU plural + Checker 范例）✅ **推荐参考**

**演示能力**：
- ✅ ICU plural 模板渲染
- ✅ Checker 抽象与复用
- ✅ 并发健康检查（超时控制）

**代码位置**：[go/modules/health/](../../../go/modules/health/)
- [usecase.go (144行)](../../../go/modules/health/usecase.go) - 核心业务逻辑
- [checker.go (96行)](../../../go/modules/health/checker.go) - Checker 实现

### 7.3 OpenAPI 服务（待实现）

### 7.4 设备网关服务（待实现）

### 7.5 用户中台服务（待实现）

## 8. 禁止事项（铁律）

❌ 不得 import services/config/service 或 repository（必须只通过 sdk 包访问）
❌ 不得调用 admin-server API（铁律，业务服务 0 引用）
❌ 不得使用相对路径导入（必须使用完整 module path）
❌ 不得在业务代码中直接操作 Redis 缓存 key
❌ 不得绕过 SDK 直接连接 MySQL

## 9. 故障排查

常见错误码及解决方案

## 10. 参考实现索引 🆕

> **2026-05-26 新增**：Hello / Health 模块已升级为 SDK 接入参考实现。

详细接入规范请查看：[sample-module.md](./sample-module.md)

| 模块 | 演示的 SDK 能力 | 代码位置 |
|------|----------------|----------|
| Hello | configsdk.GetString/GetInt + i18nsdk.T (named 模板) | [hello/usecase.go](../../../go/modules/hello/usecase.go) |
| Health | i18nsdk.T (ICU plural) + health.Checker 接口 | [health/usecase.go](../../../go/modules/health/usecase.go) |

**新模块开发时，建议先阅读 [sample-module.md](./sample-module.md)，照 Hello / Health 抄一遍骨架，再填业务逻辑。**
