---
name: cairobot-coding-standard
description: CaiRobot MVP 编码规范强制执行 Skill。当用户要求实现功能、编写代码，或涉及 services/、ai/、web/、app/、firmware/ 目录时激活。
trigger:
  - "实现"
  - "编写代码"
  - "实现功能"
  - 涉及 services/、ai/、web/、app/、firmware/ 目录
priority: medium
blocking: false
---

# CaiRobot MVP 编码规范强制执行 Skill

## 1. Skill 职责

本 Skill 强制执行 CaiRobot MVP 项目的编码规范。

**负责**：
- 命名规范检查
- 文件规模检查
- 注释规范检查
- 类职责检查

**不负责**：
- 具体业务逻辑
- 测试编写（由 cairobot-tdd-loop 处理）

详细规则参见：[.trae/rules/coding.md](../../.trae/rules/coding.md)

## 2. 强制执行检查

### 2.1 命名规范

| 类型 | 规范 | 示例 |
|---|---|---|
| 变量/参数 | lowerCamelCase | `doorStatus`、`hasPermission` |
| 常量 | UPPER_SNAKE_CASE | `MAX_FILE_LINES` |
| 类/接口 | UpperCamelCase | `IntentClassifier`、`DeviceService` |
| 布尔变量 | is/has/can/should 开头 | `isLocked`、`hasParentPermission` |
| 枚举值 | UPPER_SNAKE_CASE | `HOMEWORK_EXPLANATION` |

### 2.2 文件规模

| 类型 | 推荐上限 | 绝对上限 |
|---|---|---|
| 单文件 | 300 行 | 500 行 |
| 单类 | 200 行 | - |
| 单方法 | 30 行 | 50 行 |
| PR 改动 | 10 个文件 | - |
| PR 代码行数 | 300 行 | - |

> 注：自动生成代码（Protobuf、ORM、路由表）不计入上限。

### 2.3 注释规范

- 默认使用中文注释
- 核心类、核心方法必须有注释
- 注释解释"为什么"和"约束"
- 禁止无意义注释
- TODO/FIXME 必须带说明

### 2.4 类职责

- 单一职责原则
- 禁止上帝类
- 状态机类只做状态迁移
- 工具类只放纯函数

## 3. 完成前校验清单

代码编写完成后，必须确认：

- [ ] 命名符合规范
- [ ] 文件规模未超限
- [ ] 核心类/方法有注释
- [ ] 类职责单一
- [ ] 无魔法数字
- [ ] 错误处理规范

## 4. 违规警告

以下行为需警告并要求修复：

- 命名不符合规范
- 文件超过 300 行未拆分
- 方法超过 30 行未拆分
- 核心代码无注释
- 存在上帝类

## 5. 联动 Skill

- 任务启动时激活 cairobot-active-gap-filling
- 涉及协议时激活 cairobot-proto-registry-guard
- 涉及测试时激活 cairobot-tdd-loop
