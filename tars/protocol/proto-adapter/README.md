# Proto Adapter

## 职责

Proto Adapter 负责 Protobuf message 与 Tars bytes 之间的转换适配。

## 负责

- Protobuf Request message 序列化为 Tars request bytes
- Tars response bytes 反序列化为 Protobuf Response message
- 协议版本兼容性处理

## 相关文档

- [docs/api/tars规范.md](../../../docs/api/tars规范.md)
