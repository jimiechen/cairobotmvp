# Protobuf 规范

## 1. 命名规范

### 1.1 包名
- 使用小写字母、数字和点
- 格式：`cairobot.[service].[version]`
- 示例：`cairobot.common.v1`、`cairobot.user_center.v1`

### 1.2 消息名
- 使用 UpperCamelCase
- 示例：`User`、`CreateUserRequest`、`ListUsersResponse`

### 1.3 服务名
- 使用 UpperCamelCase
- 以 `Service` 结尾
- 示例：`UserService`、`DeviceService`

### 1.4 RPC 方法名
- 使用 UpperCamelCase
- 动作在前，资源在后
- 示例：`CreateUser`、`ListUsers`、`GetDevice`

### 1.5 字段名
- 使用 lower_snake_case
- 示例：`user_id`、`created_at`、`request_id`

### 1.6 枚举名
- 使用 UpperCamelCase
- 示例：`UserStatus`、`DeviceType`

### 1.7 枚举值
- 使用 UPPER_SNAKE_CASE
- 以枚举名前缀开头
- 示例：`USER_STATUS_ACTIVE`、`DEVICE_TYPE_ROBOT`

## 2. 字段编号

- 字段编号 1-15 占用 1 字节，优先用于高频字段
- 字段编号 16-2047 占用 2 字节
- 字段编号一旦发布，不得复用
- 删除字段必须使用 `reserved` 标记

## 3. 通用字段规范

### 3.1 请求 ID
所有请求消息必须包含：
```proto
string request_id = 1; // 请求唯一标识
```

### 3.2 响应状态
所有响应消息必须包含：
```proto
Error error = 1; // 错误信息，成功时为空
```

其中 `Error` 定义在 `common/v1/error.proto`

### 3.3 开放平台版本
开放平台 API 请求必须包含：
```proto
string version = 2; // API 版本，如 "v1"
```

### 3.4 时间戳
统一使用：
```proto
int64 created_at = N; // 创建时间戳（秒）
int64 updated_at = N; // 更新时间戳（秒）
```

或使用 Google 标准类型：
```proto
google.protobuf.Timestamp created_at = N;
google.protobuf.Timestamp updated_at = N;
```

## 4. 分页规范

参考 `common/v1/pagination.proto`

### 请求
```proto
message PageRequest {
  int32 page_size = 1; // 每页数量，默认 20，最大 100
  string page_token = 2; // 分页令牌，首次为空
}
```

### 响应
```proto
message PageResponse {
  string next_page_token = 1; // 下一页令牌，无更多数据时为空
  int64 total_count = 2; // 总数
}
```

## 5. 向后兼容规则

- 不得删除现有字段
- 不得修改现有字段编号
- 不得修改现有字段类型
- 不得修改现有字段标签（optional/repeated）
- 新增字段必须标记为 optional
- 删除字段必须使用 reserved

## 6. 文件结构

每个服务的 proto 文件建议按以下结构组织：
```
proto/[service]/v1/
├── [service].proto      # 服务定义
├── [service]_model.proto  # 数据模型
└── [service]_request.proto  # 请求响应定义
```

## 7. 相关文档

- [ADR-0003-服务协议使用Protobuf.md](../adr/ADR-0003-服务协议使用Protobuf.md)
- [gRPC接口规范.md](gRPC接口规范.md)
- [HTTP-gateway规范.md](HTTP-gateway规范.md)
