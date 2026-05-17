# HTTP Gateway 规范

## 1. 概述

使用 grpc-gateway 将 gRPC 服务转换为 HTTP JSON 接口，供前端和开放平台使用。

## 2. HTTP 方法映射

| gRPC 方法名 | HTTP 方法 | 路径示例 |
|------------|----------|---------|
| GetXxx | GET | /v1/xxx/{id} |
| ListXxx | GET | /v1/xxx |
| CreateXxx | POST | /v1/xxx |
| UpdateXxx | PUT | /v1/xxx/{id} |
| DeleteXxx | DELETE | /v1/xxx/{id} |

## 3. 路径参数

- 使用 `{field_name}` 标记路径参数
- 路径参数必须对应请求消息中的字段
- 示例：
  ```proto
  rpc GetDevice(GetDeviceRequest) returns (GetDeviceResponse) {
    option (google.api.http) = {
      get: "/v1/devices/{device_id}"
    };
  }
  ```

## 4. 请求体

- POST 和 PUT 方法使用 `body: "*"` 将整个请求作为 body
- 示例：
  ```proto
  rpc CreateDevice(CreateDeviceRequest) returns (CreateDeviceResponse) {
    option (google.api.http) = {
      post: "/v1/devices"
      body: "*"
    };
  }
  ```

## 5. 查询参数

- GET 和 DELETE 方法的其他字段自动映射为查询参数
- 支持重复字段
- 示例：`GET /v1/devices?page_size=20&page_token=xxx`

## 6. 响应格式

- 成功响应：返回对应的响应消息
- 错误响应：返回标准错误格式
- 示例错误响应：
  ```json
  {
    "error": {
      "code": "NOT_FOUND",
      "message": "Device not found",
      "request_id": "xxx-xxx-xxx"
    }
  }
  ```

## 7. 状态码

- 200 OK：成功
- 400 Bad Request：请求参数错误
- 401 Unauthorized：未认证
- 403 Forbidden：无权限
- 404 Not Found：资源不存在
- 500 Internal Server Error：服务器内部错误

## 8. 相关文档

- [protobuf规范.md](protobuf规范.md)
- [OpenAPI规范.md](OpenAPI规范.md)
