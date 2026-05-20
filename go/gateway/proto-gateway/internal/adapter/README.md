# Adapter

## 职责

Adapter 负责 MessagePacket.data 与业务 Protobuf message 的转换。

## 负责

- MessagePacket.data 与业务 Protobuf message 的转换
- Protobuf object 到 Tars request bytes 的转换
- Tars response bytes 到业务 Protobuf response 的转换

## 不负责

- Tars struct 字段映射（Tars IDL 不定义业务 struct）
- 业务语义修改

## 错误处理

- 必须保留校验错误并转换为统一错误码

## 相关文档

- [docs/api/tars规范.md](../../../docs/api/tars规范.md)
