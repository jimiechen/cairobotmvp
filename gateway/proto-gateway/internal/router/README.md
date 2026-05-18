# Router Module

## 概述

路由模块负责加载和解析路由配置，并根据 Protobuf 请求的 package、service、method 查找对应的 TarsGo 服务路由信息。

## 职责

- 加载和解析 routes.yaml 配置文件
- 提供路由查找功能
- 路由规则验证
- 路由缓存

## 文件说明

| 文件 | 说明 |
|------|------|
| config.go | 配置加载和解析 |
| router.go | 路由查找逻辑 |

## 核心功能

### 配置加载

从 configs/routes.yaml 加载路由配置。

### 路由查找

支持多种路由查找策略：
1. 精确匹配：package + service + method
2. 服务级匹配：package + service
3. 包级匹配：package

### 路由验证

- 验证 Tars 服务信息完整性
- 验证超时时间配置
- 验证鉴权和审计配置

## 当前状态

- ⏳ 等待实现配置加载逻辑
- ⏳ 等待实现路由查找逻辑
- ⏳ 等待实现路由验证逻辑

## 相关文档

- [routes.yaml](../configs/routes.yaml)
- [Code Wiki](../../../docs/wiki/CODE-WIKI.md)
