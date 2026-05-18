# ADR-0010 每日知识蒸馏定时任务

## 状态

已采纳

## 背景

根据项目规范（AGENTS.md §15），要求每天 22:00 前完成：
1. 提交日报
2. Markdown 蒸馏
3. 更新 LLM Wiki

为了自动化这个流程，减少人工操作，需要创建一个定时任务来自动执行这些步骤。

## 决策

使用 GitHub Actions Schedule 功能创建每日定时任务：

### 技术选型

| 选项 | 说明 | 选择原因 |
|------|------|----------|
| GitHub Actions Schedule | GitHub 原生支持的定时触发功能 | 与现有 CI 系统集成，无需额外基础设施 |
| Python 脚本 | 实现具体的蒸馏逻辑 | 项目已有 Python 脚本基础，便于维护 |

### 任务流程

1. **触发时间**：每天 UTC 14:00（对应北京时间 22:00）
2. **检查变更**：判断当天是否有 Git 提交
3. **执行流程**：
   - 如有变更：执行完整知识蒸馏流程
   - 如无变更：生成无代码变更日报
4. **自动提交**：将生成的文件提交回仓库

### 文件结构

```
.github/workflows/
└── daily-knowledge-distillation.yml  # GitHub Actions 定时任务定义

scripts/
└── daily-knowledge-distillation.py    # 每日知识蒸馏执行脚本

docs/reports/daily/
└── distillation-YYYY-MM-DD.log        # 每日执行日志
```

## 影响

- 自动化了每日知识蒸馏流程
- 减少了人工操作的繁琐性
- 确保了流程的一致性和及时性
- 当前为占位实现，后续需要完善具体逻辑

## 替代方案

曾考虑以下方案但未采纳：

| 方案 | 内容 | 未采纳原因 |
|------|------|------------|
| 使用 Cron + 脚本 | 本地服务器定时执行 | 需要维护额外服务器 |
| 使用第三方 CI 服务 | 如 Travis CI、CircleCI | 增加额外依赖 |

## 后续动作

1. ✅ 创建 GitHub Actions workflow
2. ✅ 创建 Python 辅助脚本
3. ⬜ 完善日报生成逻辑
4. ⬜ 实现 Tabbit 任务蒸馏
5. ⬜ 实现 LLM Wiki 更新
6. ⬜ 测试完整流程
