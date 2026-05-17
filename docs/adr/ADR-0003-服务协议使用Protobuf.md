# ADR-0003：服务协议使用 Protobuf

## 1. 基本信息

| 字段 | 值 |
|---|---|
| ID | ADR-0003 |
| 名称 | 服务协议使用 Protobuf |
| 状态 | 草稿 |
| 创建日期 | 2026-05-17 |
| 最后更新 | 2026-05-17 |
| 创建人 | 项目团队 |

## 2. 背景

（待补充）

## 3. 决策

1. **统一协议定义**：使用 Protocol Buffers 作为所有服务间接口和开放平台契约的定义语言
2. **内部通信**：使用 gRPC + Protobuf
3. **开放平台对外**：使用 HTTPS JSON，由 Protobuf 通过 grpc-gateway 转换生成
4. **Protobuf 存放位置**：根目录 proto/

## 4. 理由

（待补充）

## 5. 替代方案

- OpenAPI / Swagger
- Thrift
- JSON Schema

## 6. 后果

### 6.1 正面
- 强类型，编译时检查
- 二进制序列化，性能好
- 跨语言支持好
- 版本兼容设计完善

### 6.2 负面
- 有学习成本
- 需要额外的编译步骤
- 可读性不如纯 JSON

## 7. 约束

- Protobuf 文件统一放在 proto/ 目录下
- 按服务和功能分目录组织
- 遵循 Protobuf 编码规范
- 所有请求必须包含 request_id
- 所有响应必须包含错误码或状态字段
- 开放平台 API 必须包含 version 字段

## 8. 后续待确认事项

- Protobuf 版本（proto2 vs proto3）
- 代码生成工具链
- 版本管理策略
- 向后兼容保证策略

## 9. 相关文档

- [ADR-0001-总体系统架构.md](ADR-0001-总体系统架构.md)
- [PRD-03-开放平台API.md](../prd/PRD-03-开放平台API.md)
