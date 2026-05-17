# Protobuf 规范

## 1. 核心原则

在 CaiRobot MVP 中，Protobuf 协议编号 `max + min` 是接口报文的唯一身份。任何新增、修改、删除接口报文，都必须同步更新协议编号注册表、OpenAPI 映射、测试用例、测试报告和 LLM Wiki。

## 2. 命名规范

### 2.1 包名
- 使用小写字母、数字和点
- 格式：`com.mineplanet.pojo.[service]`
- 示例：`com.mineplanet.pojo.health`、`com.mineplanet.pojo.user`

### 2.2 消息名
- 使用 UpperCamelCase
- 示例：`User`、`CreateUserRequest`、`ListUsersResponse`

### 2.3 服务名
- 使用 UpperCamelCase
- 以 `Service` 结尾
- 示例：`UserService`、`DeviceService`

### 2.4 RPC 方法名
- 使用 UpperCamelCase
- 动作在前，资源在后
- 示例：`CreateUser`、`ListUsers`、`GetDevice`

### 2.5 字段名
- 使用 lower_snake_case（注意：历史原因，maxType、minType 保留 camelCase）
- 示例：`user_id`、`created_at`、`request_id`

### 2.6 枚举名
- 使用 UpperCamelCase
- 示例：`UserStatus`、`DeviceType`

### 2.7 枚举值
- 使用 UPPER_SNAKE_CASE
- 以枚举名前缀开头
- 示例：`USER_STATUS_ACTIVE`、`DEVICE_TYPE_ROBOT`

### 2.8 Type 枚举
- 每个业务 Request/Response message 内部必须声明 enum Type
- Type.max 表示协议大类
- Type.min 表示协议小类
- 格式：
  ```proto
  message ServiceHealthCheckRequest {
    enum Type {
      none = 0;
      max = 2100;
      min = 2097;
    }
  }
  ```

## 3. 协议编号规范

### 3.1 核心规则
1. `max + min` 组合在全仓库内必须唯一
2. 每个业务报文的 Type.max + Type.min 必须登记到 [协议编号注册表.md](./协议编号注册表.md)
3. 已发布的 `max + min` 不得复用
4. 删除接口后，原编号必须标记为 `reserved`，不得分配给新接口
5. Request 和 Response 应分别拥有独立编号
6. Event、Command、Callback 也必须拥有独立编号

### 3.2 编号分配建议

| max 范围 | 模块 |
|---:|---|
| 1000-1999 | 通用基础协议 |
| 2000-2999 | 系统、健康检查、网关基础能力 |
| 3000-3999 | 认证与权限 |
| 4000-4999 | 服务商后台系统 |
| 5000-5999 | 终端用户中台系统 |
| 6000-6999 | 开放平台 API |
| 7000-7999 | AI 服务系统 |
| 8000-8999 | 设备通信与设备网关 |
| 9000-9999 | App、Web、前端交互协议 |

## 4. 通用返回规范

### 4.1 Result 通用返回
所有响应报文应优先使用 `com.mineplanet.pojo.pb.Result` 表达接口处理结果。

Result 字段含义：
- `code`：响应码，成功建议使用 `10200`。
- `message`：响应消息。

### 4.2 PageInfo 分页规范
分页响应应使用 `PageInfo`。
- `pageSize`：每页大小
- `pageToken`：分页令牌
- `totalCount`：总数
- `nextPageToken`：下一页令牌

### 4.3 ErrorDetail 错误详情
参数校验响应应使用 `ValidationResult` 和 `ErrorDetail`。
- `field`：字段名
- `message`：错误消息
- `code`：错误码

## 5. MessagePacket 网关统一入口

网关统一入口报文使用 `MessagePacket`，位于 `proto/base/message.proto`。

```proto
message MessagePacket {
  int32 maxType = 1;             // 必填，协议大类
  int32 minType = 2;             // 必填，协议小类
  map<string, string> extend = 3; // 非必填，通用透传上下文
  Platform platform = 4;         // 非必填，平台类型
  bytes data = 5;                // 必填，业务协议包
}
```

路由规则：
1. 解析 MessagePacket
2. 提取 maxType 和 minType
3. 根据 max + min 找到对应的业务协议
4. 解析 data 字段为对应的业务 Protobuf Message

## 6. 字段编号规范

- 字段编号 1-15 占用 1 字节，优先用于高频字段
- 字段编号 16-2047 占用 2 字节
- 字段编号一旦发布，不得复用
- 删除字段必须使用 `reserved` 标记

## 7. 向后兼容规则

- 不得删除现有字段
- 不得修改现有字段编号
- 不得修改现有字段类型
- 不得修改现有字段标签（optional/repeated）
- 新增字段必须标记为 optional
- 删除字段必须使用 reserved
- 不得修改已发布的 max + min

## 8. 选项规范

### 8.1 go_package
- 格式：`github.com/jimiechen/mineplanet/protocols/generated/go/proto/[directory]`
- 示例：`github.com/jimiechen/mineplanet/protocols/generated/go/proto/base`

### 8.2 java_package
- 格式：`com.mineplanet.pojo` 或 `com.mineplanet.pojo.[service]`

### 8.3 java_outer_classname
- 格式：`[Service]Proto`
- 示例：`HealthProto`、`MessageProto`

### 8.4 optimize_for
- 使用 `LITE_RUNTIME`

## 9. 文件结构规范

- 每个 proto 文件只能有一个 `syntax = "proto3";` 声明
- import 路径必须与目录一致
- 基础协议放在 `proto/base/` 目录下
- 业务协议按模块分目录

```
proto/
├── base/
│   ├── message.proto  # 网关统一入口
│   ├── result.proto   # 通用返回
│   └── health.proto   # 健康检查
└── [module]/
    └── [service].proto
```

## 10. 协议变更评审规则

协议变更必须按以下流程进行：
1. 更新 Protobuf 文件
2. 更新协议编号注册表
3. 更新 OpenAPI 映射
4. 更新测试用例
5. 更新相关文档（PRD、ADR、LLM Wiki）
6. 提交 PR 评审
7. 必须有项目主控确认后才能合并

## 11. 相关文档

- [ADR-0003-服务协议使用Protobuf.md](../adr/ADR-0003-服务协议使用Protobuf.md)
- [协议编号注册表.md](./协议编号注册表.md)
- [openapi-protobuf映射规范.md](./openapi-protobuf映射规范.md)
- [gRPC接口规范.md](./gRPC接口规范.md)
- [HTTP-gateway规范.md](./HTTP-gateway规范.md)
- [OpenAPI规范.md](./OpenAPI规范.md)
