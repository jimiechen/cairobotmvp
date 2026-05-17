# PRD-09：HelloWorld 与 HealthCheck 验收规范

## 1. 基本信息

| 字段 | 值 |
|---|---|
| ID | PRD-09 |
| 名称 | HelloWorld 与 HealthCheck 验收规范 |
| 状态 | 草稿 |
| 优先级 | P0 |
| 创建日期 | 2026-05-17 |
| 最后更新 | 2026-05-17 |
| 创建人 | 项目团队 |

## 2. 背景

本 PRD 定义 HelloWorld 和 HealthCheck 的验收标准，用于验证工程规范、TDD 流程、CI/CD 和协议设计是否正确落地。

## 3. 目标

- 验证工程规范是否完整
- 验证 TDD 流程是否正确执行
- 验证 CI/CD 是否正常工作
- 验证协议设计是否正确
- 验证报告和文档流程是否完整

## 4. 非目标

- 实现真实业务功能
- 实现登录、数据库、支付等复杂功能
- 实现真实 AI 能力

## 5. 验收标准

### 5.1 协议验收

- [ ] HelloWorldRequest 和 HelloWorldResponse 已定义
- [ ] HealthCheckRequest 和 HealthCheckResponse 已定义
- [ ] 所有协议已注册 max + min 编号
- [ ] max + min 编号唯一，无冲突
- [ ] 协议编号已登记到 `docs/api/协议编号注册表.md`

### 5.2 测试验收

- [ ] 单元测试已编写并通过
- [ ] 集成测试已编写并通过
- [ ] 协议序列化/反序列化测试已编写并通过
- [ ] 测试覆盖率 ≥ 80%

### 5.3 CI 验收

- [ ] GitHub Actions 已配置
- [ ] docs-check 已通过
- [ ] proto-check 已通过
- [ ] go-test 已通过或说明跳过原因
- [ ] python-test 已通过或说明跳过原因
- [ ] web-test 已通过或说明跳过原因
- [ ] admin-web-test 已通过或说明跳过原因
- [ ] report-check 已通过

### 5.4 文档验收

- [ ] PRD 已更新
- [ ] ADR 已更新（如有架构决策）
- [ ] LLM Wiki 已更新
- [ ] 测试报告已生成

### 5.5 报告验收

- [ ] 测试报告已生成（`docs/reports/testing/`）
- [ ] Standalone HTML 报告已生成（`docs/reports/html/`）
- [ ] 日报已提交（`docs/reports/daily/`）
- [ ] Markdown 蒸馏已完成（`docs/reports/distilled/`）

## 6. 协议编号分配

| max | min | Message | 说明 |
|---:|---:|---|---|
| 2100 | 2097 | ServiceHealthCheckRequest | 服务健康检查请求（已存在） |
| 2100 | 2098 | ServiceHealthCheckResponse | 服务健康检查响应（已存在） |
| 2100 | 2101 | HelloWorldRequest | HelloWorld 请求（待实现） |
| 2100 | 2102 | HelloWorldResponse | HelloWorld 响应（待实现） |
| 2100 | 2103 | HealthCheckRequest | 健康检查请求（待实现） |
| 2100 | 2104 | HealthCheckResponse | 健康检查响应（待实现） |

## 7. CI 检查要求

### 7.1 必须通过的检查

- `docs-check`：关键文档存在
- `proto-check`：协议编号唯一且已注册
- `report-check`：报告目录和文件存在

### 7.2 可跳过的检查（需说明原因）

- `go-test`：如果 Golang 骨架尚未实现
- `python-test`：如果 Python 骨架尚未实现
- `web-test`：如果 ReactJS App 骨架尚未实现
- `admin-web-test`：如果 AdminWeb 骨架尚未实现

### 7.3 跳过原因示例

```text
跳过 go-test：services/hello-go/go.mod 不存在，当前尚未实现 Golang HelloWorld 骨架。
跳过 python-test：ai/service 暂无 pyproject.toml 或 requirements.txt。
跳过 web-test：web/app/package.json 不存在，当前尚未实现 ReactJS App 骨架。
跳过 admin-web-test：web/provider-admin/package.json 不存在，当前尚未实现 AdminWeb 骨架。
```

## 8. 本地等价命令

如果 CI 无法运行，必须提供本地等价命令：

```bash
# 文档检查
python3 scripts/ci/check_required_docs.py

# 协议检查
python3 scripts/ci/check_proto_registry.py

# 报告检查
python3 scripts/ci/check_reports.py

# Golang 测试（如果已实现）
cd services/hello-go && go test ./...

# Python 测试（如果已实现）
cd ai/service && python -m pytest

# ReactJS 测试（如果已实现）
cd web/app && npm test -- --runInBand
```

## 9. 风险

- CI 配置可能需要根据实际环境调整
- 协议编号可能需要重新规划
- 测试覆盖率可能难以达到 80%

## 10. 相关文档

### 相关 Issue

（待创建）

### 相关 ADR

- [ADR-0001-总体系统架构.md](../adr/ADR-0001-总体系统架构.md)
- [ADR-0003-服务协议使用Protobuf.md](../adr/ADR-0003-服务协议使用Protobuf.md)

### 相关 Proto

- [proto/base/health.proto](../../proto/base/health.proto)

### 相关测试报告

（待生成）
