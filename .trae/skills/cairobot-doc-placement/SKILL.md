---
name: 文档归档位置规范
slug: doc-placement
summary: 目录与文件存放守卫，确保文件放在正确位置。新增文件、修改目录结构或涉及 docs/ 变更时激活。
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - doc-placement
  - directory-layout
  - file-organization
trigger:
  - "新增文件"
  - "修改目录"
  - "创建目录"
  - 涉及 docs/ 目录
  - 涉及根目录文件
priority: medium
blocking: false
---

# CaiRobot MVP 目录与文件存放守卫 Skill

## 1. Skill 职责

本 Skill 强制执行 CaiRobot MVP 项目的目录与文件存放规范。

**负责**：
- 目录存放位置检查
- 文件命名规范
- 禁止行为检查

详细规则参见：
- [.trae/rules/docs.md](../../.trae/rules/docs.md)

## 2. 目录存放规范

### 2.1 必备目录

| 内容 | 必须存放位置 |
|---|---|
| PRD | `docs/prd/` |
| ADR | `docs/adr/` |
| API/协议 | `docs/api/` |
| 测试策略 | `docs/testing/` |
| 报告目录 | `docs/reports/` |
| Wiki | `docs/wiki/` |
| 协议定义 | `proto/` |
| 后端服务代码 | `services/` |
| AI 模块代码 | `ai/` |
| 前端代码 | `web/` |

### 2.2 可选目录（MVP 阶段按需创建）

| 内容 | 必须存放位置 |
|---|---|
| AI 安全策略文档 | `docs/ai-safety/` |
| 硬件说明 | `docs/hardware/` |
| 固件说明 | `docs/firmware/` |
| App 说明 | `docs/app/` |
| 固件代码 | `firmware/` |
| App 代码 | `app/` |
| EDA 文件 | `hardware/eda/` |
| BOM | `hardware/bom/` |
| 测试数据 | `tests/fixtures/` |
| 安全边界测试用例 | `tests/safety-cases/` |
| 脚本 | `scripts/` |
| 草稿文件 | `docs/drafts/` 或 `tmp/` |

## 3. PRD 文档结构

每个 PRD 必须包含：

- 背景
- 目标
- 非目标
- 使用角色
- 核心流程
- 功能需求
- 非功能需求
- 权限要求
- 接口要求
- 数据要求
- 验收标准
- TDD 测试要求
- 截图/视频证据要求
- Standalone HTML 报告要求
- 风险
- 相关 Issue
- 相关 ADR
- 相关 Proto
- 相关测试报告

## 4. ADR 文档结构

每个 ADR 必须包含：

- 状态
- 背景
- 决策
- 理由
- 替代方案
- 后果
- 约束
- 后续待确认事项

## 5. 禁止行为

以下行为**绝对禁止**：

- 在仓库根目录随意新增 md、tmp、draft、new、final2 之类文件
- 临时脚本散落在根目录
- 把 PRD 放在聊天导出文件夹
- 把测试数据塞到 src 目录
- 随意新增顶层目录

## 6. 完成前校验清单

文件/目录变更前，必须确认：

- [ ] 文件存放位置符合规范
- [ ] 不在禁止位置新增文件
- [ ] PRD/ADR 结构完整
- [ ] 相关索引已更新

## 7. 违规阻断

以下行为必须**立即停止**：

- 在根目录新增 md 文件
- 把 PRD 放在错误位置
- 随意新增顶层目录

## 8. 联动 Skill

- 文档变更时激活 cairobot-daily-report
- 协议变更时激活 cairobot-proto-registry-guard
