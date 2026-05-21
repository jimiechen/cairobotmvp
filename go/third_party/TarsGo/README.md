# TarsGo 第三方依赖基线

## 基线版本

TarsCloud/TarsGo **v1.4.6**

## 用途

作为 `proto-gateway` 和 `tars/system` 后续 S1 二次开发的 TarsGo 框架依赖基线。

- **不用于**：local 模式运行时依赖（local 模式不连接真实 TarsCloud）
- **用于**：S1 阶段 TarsGoInvoker 实现、Tars 协议编解码、TarsCloud 客户端集成
- **来源**：[TarsCloud/TarsGo](https://github.com/TarsCloud/TarsGo) 官方 Release 1.4.6

## 位置

```
go/third_party/TarsGo/TarsGo-1.4.6/
```

## 目录约定依据

- [CODE-WIKI.md §13.2](../../docs/wiki/CODE-WIKI.md) — Go 语言资产目录结构中 `third_party/TarsGo/` 条目
- 本目录为 CODE-WIKI 预留的第三方 Go 依赖落盘位置

## 使用约束

1. 本基线为只读参考，二次开发应基于此目录进行
2. 不得将真实 TarsCloud 服务实例配置放入仓库
3. 不得修改本基线内的上游源码（如需 patch 应通过 replace 或 fork 方式）
4. local 模式（`GATEWAY_INVOKER_MODE=local`）不依赖此基线运行
5. 微服务模式（`GATEWAY_INVOKER_MODE=tars`）S1 实现阶段将引用此基线

## 相关文档

- [tars规范.md](../../docs/api/tars规范.md)
- [CODE-WIKI.md](../../docs/wiki/CODE-WIKI.md)
- [ADR-0008-use-tarscloud-routing-layer.md](../../docs/adr/ADR-0008-use-tarscloud-routing-layer.md)
- [proto-gateway/README.md](../gateway/proto-gateway/README.md)
