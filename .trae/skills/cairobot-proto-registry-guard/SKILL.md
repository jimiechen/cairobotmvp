---
name: 协议编号注册表守卫
slug: proto-registry-guard
summary: Protobuf 协议编号唯一性守卫，确保 max+min 编号无重复且已登记到注册表。定义协议、新增接口或涉及 proto/ 时激活。
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - protobuf
  - protocol
  - registry
trigger:
  - "定义协议"
  - "新增接口"
  - "修改 proto"
  - 涉及 proto/ 目录
  - 新增 Request/Response/Event/Command/Callback
priority: high
blocking: true
---

# CaiRobot MVP Protobuf 协议编号唯一性守卫 Skill

## 1. Skill 职责

本 Skill 强制执行 Protobuf 协议编号的唯一性和注册规范。

**负责**：
- max + min 编号唯一性检查
- 编号注册规范
- 协议变更同步

**不负责**：
- 具体业务协议设计（由 PRD/ADR 处理）

详细规则参见：
- [.trae/rules/docs.md](../../.trae/rules/docs.md)
- [docs/api/协议编号注册表.md](../../docs/api/协议编号注册表.md)
- [docs/api/protobuf规范.md](../../docs/api/protobuf规范.md)

## 2. 强制执行步骤

### 2.1 新增协议检查
新增任何 Request/Response/Event/Command/Callback 时：

1. 确定 max 编号（按模块分配区间）
2. 确定 min 编号（递增）
3. 检查 max + min 组合是否唯一
4. 在 proto 文件中声明 enum Type
5. 在协议编号注册表登记

### 2.2 编号分配区间

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

### 2.3 Type 枚举规范

每个业务 message 必须包含 enum Type：

```proto
message ServiceHealthCheckRequest {
  enum Type {
    none = 0;
    max = 2100;  // 协议大类
    min = 2097;  // 协议小类
  }
  // 业务字段...
}
```

## 3. 完成前硬校验清单

协议定义完成前，必须确认：

- [ ] proto 文件已创建/修改
- [ ] enum Type 已声明
- [ ] max + min 已确定
- [ ] 编号唯一性已验证（无冲突）
- [ ] 已登记到 docs/api/协议编号注册表.md
- [ ] OpenAPI 映射已更新（如适用）

## 4. 违规阻断

以下行为**必须失败 CI**：

- max + min 编号重复
- 编号未登记到注册表
- 删除已发布编号未标记 reserved
- Request 和 Response 使用相同编号

## 5. CI 检查

协议变更必须通过 scripts/ci/check_proto_registry.py：

```bash
python3 scripts/ci/check_proto_registry.py
```

## 6. 联动 Skill

- 协议变更时激活 cairobot-tdd-loop（写契约测试）
- 协议变更时激活 cairobot-doc-placement（更新映射文档）
- 任务完成时激活 cairobot-daily-report
