---
name: cairobot-html-distillation
description: CaiRobot MVP 每日 22:00 HTML 蒸馏 Skill。将测试报告、日报蒸馏为 Markdown 并更新 LLM Wiki。
trigger:
  - "蒸馏"
  - "生成报告"
  - "每天 22:00"
  - PR 创建后
priority: medium
blocking: false
---

# CaiRobot MVP HTML 蒸馏 Skill

## 1. Skill 职责

本 Skill 强制执行日报和测试报告的 Markdown 蒸馏流程。

**负责**：
- 测试报告蒸馏
- 日报蒸馏
- LLM Wiki 更新

详细规则参见：
- [.trae/rules/reporting.md](../../.trae/rules/reporting.md)

## 2. 强制执行步骤

### 2.1 每日固定工作流

每天必须完成以下工作：

1. **21:30** - 停止扩大开发范围
2. **21:30-21:45** - 整理今日修改文件
3. **21:45-22:00** - 生成日报
4. **22:00 前** - 提交日报
5. **22:00 后** - 将日报蒸馏为 Markdown
6. **同步更新** - LLM Wiki

### 2.2 蒸馏流程

```
日报/测试报告 → Markdown 蒸馏 → docs/reports/distilled/
                              ↓
                        LLM Wiki 更新
```

## 3. 报告位置

| 报告类型 | 位置 |
|---|---|
| 测试报告 | `docs/reports/testing/` |
| 日报 | `docs/reports/daily/` |
| Markdown 蒸馏 | `docs/reports/distilled/` |
| Standalone HTML | `docs/reports/html/` |

## 4. LLM Wiki 更新要求

每次蒸馏后必须更新以下索引：

- `docs/wiki/PRD索引.md`
- `docs/wiki/ADR索引.md`
- `docs/wiki/测试索引.md`
- `docs/wiki/Bug索引.md`
- `docs/wiki/LLM-WIKI.md`

## 5. 完成前校验清单

蒸馏完成后，必须确认：

- [ ] 测试报告已蒸馏到 docs/reports/distilled/
- [ ] 日报已蒸馏到 docs/reports/distilled/
- [ ] Standalone HTML 报告已生成（如有测试）
- [ ] LLM Wiki 索引已更新

## 6. 联动 Skill

- 任务完成时激活 cairobot-daily-report
- 测试完成时激活 cairobot-tdd-loop
