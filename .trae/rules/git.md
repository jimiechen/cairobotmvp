# Git 工作流规则

## 1. 分支规则

推荐分支：

- `main`：稳定分支。
- `dev`：集成分支。
- `feature/*`：功能分支。
- `fix/*`：修复分支。
- `docs/*`：文档分支。
- `test/*`：测试相关分支。
- `hardware/*`：硬件相关分支。

所有功能开发应从 `dev` 创建分支。

## 2. 提交规则

提交必须小粒度。

提交信息格式：

```text
type(scope): 中文说明
```

示例：

```text
docs(prd): 添加MVP总纲
test(ai): 添加意图分类失败测试
feat(ai): 实现基础意图分类
fix(app): 修复设备断连状态判断
refactor(firmware): 简化锁仓状态机
```

允许的 type：

- `docs`
- `test`
- `feat`
- `fix`
- `refactor`
- `chore`
- `ci`
- `build`
- `hardware`

## 3. PR 规则

每个 PR 必须：

- 关联 Issue。
- 说明变更内容。
- 说明测试结果。
- 说明文档是否更新。
- 控制变更范围。
- 不混入无关格式化。

## 4. 合并规则

以下情况不得合并：

- 测试未通过。
- 没有关联 Issue。
- 没有说明测试结果。
- 涉及行为变化但未更新文档。
- PR 范围过大。
- 存在未解释的新增依赖。
