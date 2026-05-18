# 文档规则

## 1. 文档优先级

本项目文档优先级：

1. PRD：定义做什么、不做什么。
2. ADR：定义为什么这样设计。
3. API 文档：定义模块如何通信。
4. 测试文档：定义如何验证。
5. README：定义目录用途和快速开始。

## 2. 目录与文件存放规范

### 2.1 固定目录约束

#### 必备目录

| 内容 | 必须存放位置 |
| --- | --- |
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

#### 可选目录（MVP 阶段按需创建）

| 内容 | 必须存放位置 |
| --- | --- |
| AI 安全策略文档 | `docs/ai-safety/` |
| 硬件说明 | `docs/hardware/` |
| 固件说明 | `docs/firmware/` |
| App 说明 | `docs/app/` |
| 固件代码 | `firmware/` |
| App 代码 | `app/` |
| EDA 文件 | `hardware/eda/` |
| BOM | `hardware/bom/` |
| Datasheet | `hardware/datasheets/` |
| 手板/结构说明 | `mechanical/` |
| 测试数据 | `tests/fixtures/` |
| 安全边界测试用例 | `tests/safety-cases/` |
| 脚本 | `scripts/` |
| 草稿文件 | `docs/drafts/` 或 `tmp/` |
| Tabbit / TabAI 导出收件箱 | `docs/tabbit/` |
| TRAE 任务执行导出收件箱 | `docs/trae-export/` |

### 2.2 禁止行为

- 不允许在仓库根目录随意新增 md、tmp、draft、new、final2 之类文件
- 不允许临时脚本散落在根目录
- 不允许把 PRD 放在聊天导出文件夹
- 不允许把测试数据塞到 src 目录
- 不允许随意新增顶层目录

### 2.3 临时文件处理

如果必须放草稿，统一放在：

```text
docs/drafts/
```

或：

```text
tmp/
```

并在 `.gitignore` 中说明是否忽略。

## 3. PRD 规则

PRD 放在：

```text
docs/prd/
```

PRD 必须包含：

- 背景
- 目标
- 非目标
- 用户或使用角色
- 核心流程
- 功能需求
- 验收标准
- 风险

## 4. ADR 规则

ADR 放在：

```text
docs/adr/
```

ADR 必须包含：

- 状态
- 背景
- 决策
- 后果
- 替代方案

## 5. API 文档规则

API 和通信协议放在：

```text
docs/api/
```

协议文档必须包含：

- 消息方向
- 字段定义
- 示例 JSON
- 错误码
- 兼容性说明

## 6. 测试文档规则

测试策略放在：

```text
docs/testing/
```

必须说明：

- 测试范围
- 测试类型
- 测试命令
- 验收标准
- 不通过时如何处理

## 7. 协议文档同步要求

代码行为变化时，必须同步更新文档。

如果无需更新文档，PR 中必须说明原因。

### 7.1 Protobuf 协议同步要求

新增或修改 Protobuf 协议时，必须：

1. 更新 `docs/api/协议编号注册表.md`
2. 更新 `docs/api/openapi-protobuf映射规范.md`
3. 确保 max + min 唯一性检查
4. 更新 LLM Wiki
5. 更新相关 PRD/ADR
