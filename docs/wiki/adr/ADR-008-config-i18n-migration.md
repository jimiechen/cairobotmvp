# ADR-008: 全局配置与多语言迁移

## Status: Accepted
## Date: 2026-05-24
## Decision Makers: 项目主控

### Context (背景)

旧项目存在三大问题：

1. **双重存储**：config_query.go 直接 SQL + sys_config_version 并存，业务读路径混乱
2. **单文件超限**：历史遗留 compose.go 340 行，含硬编码 switch/case
3. **缓存硬编码版本号**：旧代码使用 "v3" 字符串作为缓存 key 后缀

### Decision (决策)

采用三层架构：

1. **Schema Registry**（sys_config_schema 表驱动配置解析）
2. **SDK 层**（configsdk + i18nsdk，业务只通过 SDK 访问）
3. **admin 边界**（admin-server 仅运维内网可达，业务服务 0 引用）

关键技术决策：

- compose.go 用 map 注册表替代 switch/case（消除 module_key 硬编码）
- Protobuf dynamic_modules 字段承载新增动态配置
- 三层缓存：L1(LRU) → L2(Redis) → L3(Service)

### Consequences (后果)

**Positive:**

- 单一数据源，消除双重读路径
- Schema-driven，新增配置零代码变更
- admin 边界铁律防止历史事故重演

**Negative:**

- 强类型字段兼容层增加了复杂度
- 迁移期需维护镜像写入
- Tars IDL 需同步维护

### 协议号决策

6002/6004/6006/6008/6010 **保留**，
请求与响应协议号一一对应（6001↔6002、6003↔6004...），
便于追踪调试，注册表规则第 5 条保持不变。
