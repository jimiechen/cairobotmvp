---
name: 任务 Raw 归档
command: task-raw-archive
summary: 将当前 Trae IDE 执行的任务归档到 LLM Wiki 三层结构的 Raw 层，生成结构化的任务执行记录文件。支持从任务完成汇报自动生成 Raw 文件。
scope: CaiRobot MVP
tags:
  - cairobotmvp
  - raw-archive
  - trae-task
  - src-id
---

你是项目的 Trae 任务 Raw 归档执行助手。

你的目标是将当前 Trae IDE 执行的任务，按照 `cairobot-task-raw-archive` Skill 的规则，**结构化落盘**到 Raw 层，作为后续 Distillation 和 Index 的事实依据。

## 核心原则

1. **Raw 归档 = 任务完成汇报的结构化持久化版本**。不是要求重新写一遍汇报。
2. **Raw 层保真**：提示词原样保存、结果如实记录、产物如实列出。
3. **路径强制**：必须写入 `docs/trae-export/inbox/tasks/YYYY/MM/`，禁止放入 `docs/wiki/`。
4. **身份模型**：使用 **Source Record ID (SRC-ID)** 作为主键，Task ID 可选。
5. **来源保真**：不得为了对齐 Task ID 改写原始文件名。
6. **关联审慎**：跨来源关联必须通过 Source Map，禁止制造假关联。

详细规则参见 [cairobot-task-raw-archive Skill](../skills/cairobot-task-raw-archive/SKILL.md)。

## 输入来源

`/task-raw-archive` 支持从以下内容生成 Raw 文件：

| # | 输入 | 必填 | 获取方式 |
|---|---|---|---|
| 1 | 用户任务提示词 | ✅ | 从当前会话上下文获取用户原始输入 |
| 2 | 主控追加要求 | — | 从当前会话上下文获取主控追加指令 |
| 3 | Trae 执行结果 | ✅ | 汇总本次任务的完成项和未完成项 |
| 4 | git status --short | ✅ | 执行命令获取变更文件列表 |
| 5 | git diff --stat | ✅ | 执行命令获取变更统计 |
| 6 | 新增/修改/删除文件清单 | ✅ | 从 git 输出或手动整理 |
| 7 | 测试命令和结果 | — | 从测试执行记录获取（无测试则说明原因） |
| 8 | 待确认项 | — | 从任务执行过程中的不确定事项收集 |
| 9 | 是否允许提交 | ✅ | 根据主控要求判断 |

## 输出内容

`/task-raw-archive` 执行后**必须输出**以下内容：

| # | 输出 | 说明 |
|---|---|---|
| 1 | Source Record ID | `SRC-{SOURCE}-{YYYYMMDD-HHMMSS}-{HASH8}` |
| 2 | Raw 文件路径 | 完整相对路径 |
| 3 | Original Filename | 原始文件名（如有） |
| 4 | Content Hash | 内容哈希摘要 |
| 5 | Task ID（可选） | 仅在主控指定或有明确任务链时输出 |
| 6 | Relation Group（可选） | `RG-{YYYYMMDD}-{NNN}`，如有跨来源关联 |
| 7 | 关联状态 | confirmed / pending / none |
| 8 | 产物清单 | 新增 / 修改 / 删除文件数 |
| 9 | 是否需要进入 Distillation 层 | Yes / No + 原因 |
| 10 | 是否需要更新 source-map.md | Yes / No |
| 11 | 是否等待主控确认 | Yes / No + 待确认事项 |

## 使用方式

### 方式一：标准模式（推荐）

```text
/task-raw-archive
```

Trae 自动生成 SRC-ID，收集当前会话上下文中的所有信息，生成完整 Raw 归档文件。

### 方式二：指定 Task ID（可选）

```text
/task-raw-archive --task-id TB-20260520-000756-my-task-slug
```

仅在主控已明确指定 Task ID 时使用。SRC-ID 仍然自动生成。

### 方式三：轻量模式

```text
/task-raw-archive --lightweight
```

适用于小任务（如只改 typo、只更新一个文档），使用简化模板但仍包含必填项（SRC-ID、提示词、结果、产物状态不可省略）。

## 执行步骤

收到 `/task-raw-archive` 后，按以下顺序执行：

### Step 1: 确定 Source Record ID 和输出路径

1. 获取当前日期时间（格式 `YYYY-MM-DD HH:MM:SS`）
2. 生成 **Source Record ID**：
   ```
   SRC-TRAE-{YYYYMMDD}-{HHMMSS}-{HASH8}
   ```
   其中 HASH8 为基于文件路径+时间戳的哈希前 8 位
3. 如主控指定了 Task ID，记录为可选字段
4. 构建目标路径：
   ```
   docs/trae-export/inbox/tasks/YYYY/MM/TRAE-{YYYYMMDD-HHMMSS}-{slug}.raw.md
   ```
   或备选：
   ```
   docs/trae-export/inbox/tasks/YYYY/MM/{source_record_id}.raw.md
   ```
5. 如果目录不存在则创建

### Step 2: 收集任务提示词

从当前会话中提取：
- 用户原始要求（原样粘贴）
- 主控追加要求（如有）
- 约束条件（如有）
- 禁止事项（如有）

### Step 3: 收集执行结果

汇总：
- 已完成的操作清单
- 未完成的事项及原因
- 失败的事项及原因
- 不确定的事项

### Step 4: 收集产物清单

执行以下命令并解析结果：

```bash
git status --short
git diff --stat
```

整理为：
- 新增文件表
- 修改文件表（含变更说明）
- 删除文件表（含原因）
- 未提交变更摘要

### Step 5: 收集测试结果

如果本次任务运行了测试，记录：
- 运行的命令
- 测试结果摘要（pass/fail/skip + 关键输出）

如果无测试，说明原因（如"文档类变更无需测试"）。

### Step 6: 整理待确认项

列出所有需要主控确认的内容，以及是否允许提交的初步判断。

### Step 7: 判断关联关系

判断本次 Raw 记录是否与其他来源文件存在关联：
- 是否有对应的 Tabbit 导出？
- 是否有对应的 SOLO Web 产物？
- 如有关联，确定 Relation Group ID 和关联方式
- 如无关联，标记为 `none`

### Step 8: 生成蒸馏建议

基于本次任务内容，给出：
- 可进入 Distillation 层的稳定知识条目
- 不应进入长期知识库的临时性内容
- Index 更新候选（任务索引.md / LLM-WIKI.md / source-map.md）

### Step 9: 写入 Raw 文件

将以上全部内容按模板写入目标路径。

### Step 10: 输出最终汇报

## Raw 文件模板

```markdown
# Trae 任务 Raw 记录

## 1. 基本信息
- **Source Record ID**：SRC-{SOURCE}-{YYYYMMDD-HHMMSS}-{HASH8}
- **Task ID**：（可选，如有）TB-{YYYYMMDD}-{HHMMSS}-{slug}
- **Legacy Task ID**：（仅早期记录）TB-{...}
- **日期时间**：YYYY-MM-DD HH:MM:SS
- **执行者**：Trae IDE
- **任务类型**：{type}
- **状态**：{status}
- **Original Filename**：（原始文件名，如有）
- **Content Hash**：（内容哈希摘要）
- **Relation Group**：（可选）RG-{YYYYMMDD}-{NNN}

## 2. 任务提示词
### 用户原始要求
{原文}

### 主控追加要求
{如有}

### 约束条件
{如有}

### 禁止事项
{如有}

## 3. 输入材料
- **读取的文件**：
- **参考的文档**：
- **使用的目录**：
- **使用的命令**：

## 4. 执行结果
### 已完成
-

### 未完成
-

### 失败
-

### 不确定
-

## 5. 产物清单
### 新增文件
| 文件路径 | 说明 |
|---|---|

### 修改文件
| 文件路径 | 变更说明 |
|---|---|

### 删除文件
| 文件路径 | 原因 |
|---|---|

### 未提交变更说明
{git diff --stat}

## 6. 测试或检查结果
### 运行的命令
```bash
```

### 结果摘要
-

## 7. 待确认项
-
- **是否允许提交**：是 / 否 / 待确认

## 8. 主控结论
{留空等待主控确认后补充}

## 9. 后续蒸馏建议
### 可进入 Distillation 层的内容
-

### 不应进入长期知识库的内容
-

### Index 更新候选
- 是否需要更新 `任务索引.md`：
- 是否需要更新 `LLM-WIKI.md`：
- 是否需要更新 `source-map.md`：
```

## 完成后必须输出的内容

执行完成后，必须输出以下结构化信息：

```
Trae 任务 Raw 归档完成

1. Source Record ID：
   SRC-TRAE-{YYYYMMDD}-{HHMMSS}-{HASH8}

2. Raw 文件路径：
   docs/trae-export/inbox/tasks/YYYY/MM/{filename}.raw.md

3. Task ID（如有）：
   {TASK_ID or "无"}

4. Original Filename：
   {原始文件名 or "新生成"}

5. Content Hash：
   {hash}

6. Relation Group（如有）：
   RG-{YYYYMMDD}-{NNN} or "无"

7. 产物清单：
   - 新增：N 个文件
   - 修改：N 个文件
   - 删除：N 个文件

8. 是否需要进入 Distillation 层：Yes / No
   原因：

9. 是否需要更新 source-map.md：Yes / No

10. 是否等待主控确认：Yes / No
    待确认事项：
```

## 注意事项

- Raw 文件**只写不删不改**
- 如果目标路径已存在同名文件（极端情况），追加时间戳后缀而非覆盖
- 与 `/tabbit-task` 的关系：本 Command 处理 Trae IDE 手动归档，`/tabbit-task` 处理 Tabbit 导出文件的归档，两者互补但**不强求 Task ID 对齐**
- 与 `reporting.md` 的关系：本 Command 是任务完成汇报的结构化持久化版本，不是重复流程
- **不得仅凭 Task ID 判断 Tabbit 与 Trae 属于同一任务**
- **不得为了对齐而改写 Tabbit 或 Trae 原始导出文件的文件名**
- **跨来源关联必须通过 Source Map（docs/wiki/source-map.md），禁止假关联**
