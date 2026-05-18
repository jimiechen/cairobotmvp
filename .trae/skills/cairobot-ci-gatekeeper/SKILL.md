---
name: cairobot-ci-gatekeeper
description: CaiRobot MVP CI 阻断合并守卫 Skill。PR 创建或合并前必须激活，确保所有 CI 检查通过。
trigger:
  - "创建 PR"
  - "合并 PR"
  - "请求合并"
  - PR 状态变更
priority: high
blocking: true
---

# CaiRobot MVP CI 阻断合并守卫 Skill

## 1. Skill 职责

本 Skill 强制执行 CI 阻断合并规则，确保 PR 必须通过所有检查才能合并。

**负责**：
- CI 检查确认
- 跳过原因验证
- 本地等价命令要求

详细规则参见：
- [.trae/rules/tdd.md](../../.trae/rules/tdd.md)
- [.trae/rules/testing.md](../../.trae/rules/testing.md)

## 2. CI 检查清单

### 2.1 必须通过的检查

| Job | 作用 | 失败后果 |
|---|---|---|
| `docs-check` | 关键文档存在性检查 | **CI 失败** |
| `proto-check` | 协议编号唯一性检查 | **CI 失败** |
| `report-check` | 报告存在性检查 | **CI 失败** |

### 2.2 通过或说明跳过的检查

| Job | 作用 | 跳过条件 |
|---|---|---|
| `go-test` | Golang 单元/集成测试 | 服务骨架未实现 |
| `python-test` | Python 单元/集成测试 | AI 服务骨架未实现 |
| `web-test` | ReactJS App 测试 | App 骨架未实现 |
| `admin-web-test` | AdminWeb 测试 | Admin 骨架未实现 |

### 2.3 跳过说明要求

跳过检查时，必须在 PR 中明确输出：

```
跳过 go-test：services/hello-go/go.mod 不存在，当前尚未实现 Golang HelloWorld 骨架。
```

不得静默跳过。

## 3. 本地等价命令

如果 CI 无法运行（如 GitHub Actions 未配置），必须提供本地等价命令：

```bash
# 文档检查
python3 scripts/ci/check_required_docs.py

# 协议检查
python3 scripts/ci/check_proto_registry.py

# Golang 测试
cd services/[service-name] && go test ./...

# Python 测试
cd ai/service && python -m pytest

# ReactJS 测试
cd web/app && npm test -- --runInBand

# 报告检查
python3 scripts/ci/check_reports.py
```

## 4. 完成前硬校验清单

PR 合并前，必须确认：

- [ ] `docs-check` 已通过
- [ ] `proto-check` 已通过
- [ ] `report-check` 已通过
- [ ] `go-test` 已通过或有跳过说明
- [ ] `python-test` 已通过或有跳过说明
- [ ] `web-test` 已通过或有跳过说明
- [ ] `admin-web-test` 已通过或有跳过说明
- [ ] 本地等价命令已提供（如 CI 未运行）

**新增硬校验**：
- [ ] 无 GitHub Actions run URL 时，自动判定为 CI 未通过。

## 5. 违规阻断

以下情况**必须阻止合并**：

- 任何"必须通过"的检查失败
- 跳过检查但无跳过说明
- 测试失败但未说明原因

## 6. 联动 Skill

- PR 创建时激活 cairobot-daily-report
- 协议变更时激活 cairobot-proto-registry-guard
- Git 规范时激活 cairobot-git-discipline
