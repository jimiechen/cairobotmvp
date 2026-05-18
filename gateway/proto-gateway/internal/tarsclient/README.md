# Tars Client Module

## 概述

Tars 客户端模块封装了 TarsGo 客户端的调用逻辑，提供服务发现、负载均衡、连接池管理等功能。

## 职责

- Tars 服务客户端封装
- 服务发现与集成
- 负载均衡策略实现
- 连接池管理
- 超时控制
- 重试机制

## 文件说明

| 文件 | 说明 |
|------|------|
| client.go | Tars 客户端封装 |
| pool.go | 连接池管理 |

## 负载均衡策略

- Round Robin（轮询）
- Weighted Round Robin（加权轮询）
- Consistent Hash（一致性哈希）
- Least Connection（最小连接数）

## 当前状态

- ⏳ 等待实现 Tars 客户端封装
- ⏳ 等待实现服务发现集成
- ⏳ 等待实现负载均衡策略
- ⏳ 等待实现连接池管理

## 相关文档

- [Code Wiki](../../../docs/wiki/CODE-WIKI.md)
- [ADR-0008](../../../docs/adr/ADR-0008-use-tarscloud-routing-layer.md)
