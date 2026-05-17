# 测试规则

## 1. 测试目标

测试用于验证需求，不是为了追求形式覆盖率。

测试必须回答：

- 功能是否符合 PRD。
- 异常情况是否正确处理。
- 状态转换是否正确。
- 协议输入输出是否稳定。
- 关键安全边界是否没有被破坏。

## 2. 测试分层

本项目测试分为：

- 单元测试：验证单个函数、类、模块。
- 参数化测试：验证多组输入输出。
- 状态机测试：验证状态转换。
- 协议测试：验证 JSON、BLE、HTTP 等协议。
- 集成测试：验证模块之间协作。
- 端到端测试：验证完整用户流程。
- 回归测试：验证历史问题不复发。

## 3. 测试文件放置规范

| 测试类型 | 位置 |
| --- | --- |
| AI 模块单元测试 | `ai/.../tests/` 或 `tests/unit/ai/` |
| 固件单元测试 | `firmware/.../test/` |
| App 单元测试 | `app/.../test/` |
| App UI 测试 | `app/.../androidTest/` 或 `app/.../ui-test/` |
| 集成测试 | `tests/integration/` |
| 端到端测试 | `tests/e2e/` |
| 测试数据 fixtures | `tests/fixtures/` |
| 安全边界测试用例 | `tests/safety-cases/` |

## 4. 测试文件命名规范

### 4.1 文件命名

测试文件命名建议：

```text
intent_classifier_test.py
device_state_machine_test.cpp
learning_mode_viewmodel_test.kt
prompt_rewrite_service_test.py
```

### 4.2 测试函数命名

测试名称应表达业务含义。

推荐：

```text
test_当舱门未关闭时_锁定命令应失败
test_当请求为拍题讲解时_应进入扫描流程
test_当设备断连时_App应进入错误状态
```

如果测试框架不适合中文函数名，可以使用英文函数名，但测试描述必须中文。

## 5. 测试数据

测试数据放在：

```text
tests/fixtures/
```

安全或边界用例放在：

```text
tests/safety-cases/
```

## 6. 外部服务

单元测试不得直接调用真实外部服务。

包括：

- 真实大模型 API
- 真实 OCR 服务
- 真实云端接口
- 真实硬件设备

应使用 Mock、Fake 或本地 fixture。

## 7. 单个测试文件规模限制

- 单个测试文件：推荐 ≤ 300 行
- 超过推荐限制应按场景拆分测试文件
