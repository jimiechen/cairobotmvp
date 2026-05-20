# Skill 与 Command 命名规范

## 1. 问题背景

早期 Skill 名称以 `cairobotmvp-` 或 `cairobot-` 作为前缀，导致：

- 模糊匹配时搜索结果前缀高度重复，真正有用的语义被挤到后面
- AI 和人都难以快速判断 Skill 的用途
- Skill 的中文说明不够靠前，选择时只能看到英文 slug

## 2. 核心原则

**Skill 的可发现性优先于项目名前缀统一。**

展示名必须告诉 AI 和人"这个 Skill 何时使用"。项目归属通过 tags/scope 表达。

## 3. Skill 命名规范

### 3.1 两层命名

| 类型 | 规则 | 示例 |
|---|---|---|
| 展示名（name） | 中文用途在前，必要时括号补英文 | `任务 Raw 归档` |
| 目录名/Slug | 不以项目名前缀开头，使用功能 slug | `task-raw-archive` |

### 3.2 Front Matter 格式

每个 SKILL.md 必须使用以下 front matter 格式：

```yaml
---
name: 中文展示名
slug: 英文功能标识
summary: 中文一句话说明用途
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - ...相关标签
trigger:
  - 触发词1
  - 触发词2
priority: high | medium | highest
blocking: true | false
---
```

### 3.3 name 规则

1. **不得**以 `cairobotmvp` 或 `cairobot` 开头
2. **必须**中文用途优先
3. 应清楚表达"这个 Skill 什么时候被使用"

### 3.4 summary 规则

1. **必须**中文优先
2. **必须**说明"这个 Skill 解决什么问题"
3. 第一句话应包含适用场景

### 3.5 slug 规则

1. 使用短英文功能名，如 `task-raw-archive`、`llm-wiki-distillation`
2. 不带项目名前缀
3. 使用 kebab-case 格式

### 3.6 tags / scope 规则

1. `scope` 固定为 `CaiRobot MVP`
2. `tags` 中**必须**包含 `cairobotmvp`
3. 额外 tags 用于表达模块、场景等分类信息

### 3.7 目录名规则

MVP 阶段策略：**先优化展示名和内容，暂不强制改目录名**。

原因：目录名可能被 AGENTS.md、Command 文件、Skill 内部引用、历史 Raw 记录等多处引用，大规模重命名风险高。

如需后续迁移目录名，应单独做迁移任务并全局替换引用。

## 4. Command 命名规范

### 4.1 Command 名规则

- 保持短英文，动词明确
- 不以 `cairobotmvp` 或 `cairobot` 开头
- 使用 kebab-case 格式

推荐示例：

```text
/task-raw-archive
/daily-distill
/llm-wiki-distill
/tabbit-task
/source-map-update
```

### 4.2 Command 文件头格式

每个 Command 文件必须包含中文用途说明：

```markdown
---
name: 每日知识蒸馏
command: daily-distill
summary: 按日期读取 Raw 材料，生成日报、知识蒸馏、主控汇报、Index 候选和 Source Map 候选。
tags:
  - cairobotmvp
  - daily
  - distillation
  - solo-web
---

# /daily-distill：每日知识蒸馏

用于本地 Trae 手动补跑或回放 SOLO Web 的每日定时蒸馏流程。
```

### 4.3 Command 说明要求

1. 顶部**必须有**中文用途说明（summary）
2. command 名保持短英文
3. 正文第一段必须说明"什么时候用这个 Command"
4. 如果 Command 引用 Skill，应引用中文展示名 + slug

## 5. AGENTS.md 索引规则

Skill 索引表必须使用以下格式（**必须包含实际路径列**）：

```markdown
| 展示名 | Slug | 实际路径 | 适用场景 | Priority | Blocking | 关键词 |
|---|---|---|---|---|---|---|
| 任务 Raw 归档 | task-raw-archive | `.trae/skills/cairobot-task-raw-archive/SKILL.md` | Trae 每次任务完成后归档 | high | true | raw, trae, 归档 |
```

要求：

1. 第一列**必须**是中文展示名
2. 第二列是英文 slug
3. 第三列**必须**是实际文件路径（确保可反查到目录）
4. 第四列说明适用场景
5. 包含关键词列方便模糊匹配

## 6. 新增 Skill 检查清单

新增 Skill 时必须逐项确认：

- [ ] name 不以 cairobotmvp 或 cairobot 开头
- [ ] name 是中文用途名称
- [ ] summary 第一行是中文说明
- [ ] slug 是短英文功能名
- [ ] scope 为 CaiRobot MVP
- [ ] tags 包含 cairobotmvp
- [ ] 目录名不带项目名前缀
- [ ] AGENTS.md 索引已登记中文展示名、slug、适用场景、关键词
- [ ] 正文第一段说明"什么时候使用"
- [ ] 联动 Skill 引用时使用中文展示名 + slug

## 7. 新增 Command 检查清单

新增 Command 时必须逐项确认：

- [ ] command 名不以 cairobotmvp 或 cairobot 开头
- [ ] command 名是短英文动词短语
- [ ] 顶部有中文 summary
- [ ] 正文第一段说明用途
- [ ] 引用 Skill 时用中文展示名 + slug
- [ ] tags 包含 cairobotmvp

## 8. 已有 Skill 映射表

| 旧目录名（当前） | 展示名 | Slug | 当前状态 |
|---|---|---|---|
| `cairobot-active-gap-filling` | 主动补齐工程缺口 | active-gap-filling | ✅ 已优化 front matter |
| `cairobot-engineering-workflow` | 工程闭环流程 | engineering-workflow | ✅ 已优化 front matter |
| `cairobot-tdd-loop` | TDD 红绿重构循环 | tdd-loop | ✅ 已优化 front matter |
| `cairobot-proto-registry-guard` | 协议编号注册表守卫 | proto-registry-guard | ✅ 已优化 front matter |
| `cairobot-ci-gatekeeper` | CI 阻断合并守卫 | ci-gatekeeper | ✅ 已优化 front matter |
| `cairobot-coding-standard` | 编码规范强制执行 | coding-standard | ✅ 已优化 front matter |
| `cairobot-daily-report` | 每日汇报与事故上报 | daily-report | ✅ 已优化 front matter |
| `cairobot-git-discipline` | Git 提交纪律 | git-discipline | ✅ 已优化 front matter |
| `cairobot-doc-placement` | 文档归档位置规范 | doc-placement | ✅ 已优化 front matter |
| `cairobot-html-distillation` | HTML 蒸馏与报告生成 | html-distillation | ✅ 已优化 front matter |
| `cairobot-llm-wiki-distillation` | LLM Wiki 知识蒸馏 | llm-wiki-distillation | ✅ 已优化 front matter |
| `cairobot-scheduled-knowledge-distillation` | 定时知识蒸馏 | scheduled-knowledge-distillation | ✅ 已优化 front matter |
| `cairobot-task-raw-archive` | 任务 Raw 归档 | task-raw-archive | ✅ 已优化 front matter |
| `tabbit-task-distillation` | Tabbit 任务蒸馏 | tabbit-task-distillation | ✅ 已优化 front matter（保留原名） |

> 注：目录名暂未修改，待后续专门迁移任务处理。

## 9. 禁止事项（硬规则，强制执行）

以下行为**绝对禁止**，违反者必须立即停止并上报：

1. **🚫 硬规则：新增 Skill 的 `name` 字段不得以 `cairobot` 或 `cairobotmvp` 开头。** 这是不可协商的命名红线。项目归属只能通过 `scope` 和 `tags` 表达，不能通过 name 前缀表达。
2. 🚫 新增 Command 的 `name` 字段不得以 `cairobot` 或 `cairobotmvp` 开头
3. 🚫 让 summary 只有英文（summary 第一行必须是中文）
4. 🚫 把项目归属从 tags 中删除（tags 必须包含 cairobotmvp）
5. 🚫 为了美化名称破坏已有引用
6. 🃏 大规模重命名目录而不同步更新所有引用
7. 🚫 删除已有 Skill
8. 🚫 改业务代码
9. 🚫 git push（除非主控明确要求）

## 10. 相关文档

- [AGENTS.md](../../AGENTS.md) §19 Skill 索引
- [.trae/rules/docs.md](../../.trae/rules/docs.md) 文档规则
- [.trae/rules/reporting.md](../../.trae/rules/reporting.md) 汇报规则
