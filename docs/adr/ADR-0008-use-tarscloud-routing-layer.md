# ADR-0008：使用 TarsCloud/TarsGo 作为单网关后的内部 RPC 与服务治理层

## 1. 基本信息

| 字段 | 值 |
|---|---|
| ID | ADR-0008 |
| 名称 | 使用 TarsCloud/TarsGo 作为单网关后的内部 RPC 与服务治理层 |
| 状态 | 已确认 |
| 创建日期 | 2026-05-18 |
| 最后更新 | 2026-05-18 |
| 创建人 | 项目团队 |

## 2. 背景

CaiRobot MVP 项目前期在 ADR-0001 和 ADR-0003 中规划了基于 gRPC + Protobuf 的内部服务通信方案，并计划使用 grpc-gateway 将 gRPC 服务转换为多 REST path 的 HTTP JSON 接口供外部调用。

随着架构演进，项目需要：
1. 统一外部入口，避免为每个业务命令暴露独立 REST path。
2. 强化内部服务治理能力（服务注册、发现、负载均衡、灰度发布、监控）。
3. 保持 Protobuf 作为业务消息结构和协议编号的唯一契约来源。
4. 降低网关层复杂度，Gateway 只做协议转换和路由，不做业务逻辑。

经过评估，TarsCloud/TarsGo 在服务治理、运维成熟度、字节接口灵活性方面更符合当前阶段需求。

## 3. 决策

CaiRobot MVP 对外采用单网关协议入口，只暴露 `POST /api/hello`。请求体使用仓库已有的 `MessagePacket`，其中 `maxType` 和 `minType` 对应业务 Protobuf Request message 内部的 `Type.max` 和 `Type.min`，`data` 字段承载业务 Protobuf Request 序列化后的 bytes。Gateway 以 `maxType:minType` 作为主路由键查询 routes.yaml，并转发到内部 TarsGo servant。

内部核心服务通信从原先规划的 gRPC + Protobuf 主链路调整为 TarsCloud/TarsGo 主链路。TarsGo 服务采用统一 Tars 标准接口：`int Xxx(vector<byte> request, map<string,string> extend, out vector<byte> response)`。Protobuf message 仍然是业务字段和协议编号的唯一契约，Tars IDL 不重复定义业务字段结构。

TarsCloud 负责内部服务注册、服务发现、负载均衡、超时控制、监控、灰度发布和运维治理。Gateway 负责 MessagePacket 解析、Protobuf marshal/unmarshal、extend 构造、Tars 调用和响应 MessagePacket 封装。

## 4. 为什么从 gRPC 主链路调整为 TarsCloud/TarsGo

| 维度 | gRPC 方案（原规划） | TarsCloud/TarsGo 方案（当前决策） |
|---|---|---|
| 外部入口 | grpc-gateway 生成多 REST path | 单网关 `POST /api/hello` + MessagePacket |
| 服务治理 | 需额外引入 Consul/etcd + 自研治理 | TarsCloud 内置注册、发现、负载均衡、灰度 |
| 运维工具 | 相对简单，需自行搭建 | TarsCloud 提供成熟 Web 管理台 |
| 字节接口 | gRPC 强依赖 proto service/method | Tars 标准 bytes 接口更灵活，与 MessagePacket 解耦 |
| 协议契约 | Protobuf 定义业务字段 + gRPC 定义 RPC | Protobuf 定义业务字段，Tars IDL 只定义方法签名 |
| 学习成本 | gRPC 生态广，团队熟悉 | TarsGo 在国内有成熟案例，文档完善 |

核心原因：
1. **统一入口需求**：项目明确要求外部只有一个业务入口 `/api/hello`，grpc-gateway 的多 path 生成模式与此冲突。
2. **服务治理需求**：MVP 阶段需要快速搭建可运维、可灰度的服务网格，TarsCloud 提供开箱即用的治理能力。
3. **协议解耦需求**：希望 Gateway 路由键是 `maxType:minType`（协议编号）而非 `proto package/service/method`，Tars bytes 接口更契合此设计。

## 5. 与 Protobuf 的关系

Protobuf 仍然是 CaiRobot MVP 的业务消息契约与协议编号来源：
- 所有业务 Request / Response 结构由 Protobuf message 定义。
- 协议身份由 Protobuf message 内部 `Type.max + Type.min` 定义。
- `MessagePacket.data` 是 Protobuf Request 序列化后的 bytes。
- Tars 方法中的 `vector<byte> request` 和 `vector<byte> response` 分别承载 Protobuf Request / Response 序列化后的 bytes。
- Tars IDL 不定义业务字段结构，只定义统一方法签名。

## 6. 与 gRPC 的关系

gRPC 不再作为 CaiRobot 内部核心服务调用主链路。内部核心服务调用主链路改为 TarsCloud/TarsGo。Protobuf 仍然保留为业务消息结构、协议编号和跨语言契约来源。gRPC 规范暂时保留为历史兼容或特定场景参考，但必须标记为非主链路；后续如彻底废弃，需另行 ADR。

具体调整：
- `docs/api/gRPC接口规范.md` 标记为历史兼容/非主链路。
- 不再新增 gRPC service 定义。
- 现有 proto 文件中的 `service` 定义如需保留，仅作为文档参考，不用于实际 gRPC 调用。

## 7. 与 HTTP Gateway 的关系

HTTP Gateway 层从 grpc-gateway 多 REST path 模式收敛为单网关模式：
- 唯一业务入口：`POST /api/hello`
- Content-Type：`application/octet-stream`
- 请求体：`MessagePacket` 序列化后的二进制内容
- 路由键：`MessagePacket.maxType:MessagePacket.minType`
- Gateway 查询 routes.yaml 后调用内部 TarsGo servant

## 8. 与 OpenAPI 的关系

OpenAPI 在当前架构中仅描述单网关入口 `/api/hello` 的外层 MessagePacket 结构和调试方式，不再作为每个业务命令的独立 REST path 契约。具体业务请求和响应结构由 `maxType/minType` 对应的 Protobuf message 决定。

OpenAPI 可用于：
- 文档展示和调试说明
- SDK 生成时的外层包装说明
- 第三方接入时的入口协议文档

## 9. 正面影响

1. 外部入口统一，降低客户端集成复杂度。
2. 内部服务治理能力增强（注册、发现、负载均衡、灰度、监控）。
3. Gateway 职责清晰，只做协议转换和路由。
4. Protobuf 作为唯一业务契约，避免 Tars 与 Protobuf 字段重复定义。
5. 协议编号（maxType/minType）成为全链路统一路由键，可追溯、可审计。

## 10. 负面影响

1. 引入 TarsCloud 新依赖，团队需要学习 Tars 运维工具。
2. gRPC 方案的前期设计工作部分废弃，存在沉没成本。
3. Python 侧 AI 服务如需直接调用 Tars，需要 Tars Python 客户端支持。
4. 本地开发环境需要搭建 Tars 框架或 Mock 方案。

## 11. 风险与缓解措施

| 风险 | 等级 | 缓解措施 |
|---|---|---|
| TarsCloud 运维复杂度 | 中 | MVP 阶段使用单机或 Docker 部署，逐步熟悉运维工具 |
| Python 侧 Tars 调用 | 中 | AI 服务优先通过 Gateway 调用，不直接调用 Tars；如需直接调用，评估 TarsPy 方案 |
| 团队学习成本 | 低 | 文档先行，S0 阶段只搭骨架不实现复杂业务 |
| gRPC 历史兼容 | 低 | 保留 gRPC 规范文档并标记状态，不删除历史文件 |

## 12. 约束

1. 外部业务入口只有 `POST /api/hello`。
2. Tars IDL 不定义业务 struct，只定义统一 bytes 方法签名。
3. Protobuf 是业务字段、协议编号和跨语言契约的唯一来源。
4. routes.yaml 中只能配置协议编号注册表中已登记的编号。
5. 每个 Tars interface 必须包含 Health 和 HealthCheck。
6. Tars 方法 return 与 Protobuf Result.code 使用同一项目状态码体系，但处于不同层级。

## 13. 后续演进计划

1. S0 阶段（当前）：完成文档、ADR、目录骨架、Tars IDL 定义、routes.yaml 最小闭环。
2. S1 阶段：搭建 TarsCloud 本地环境，实现 Gateway 到 Tars 的 HelloWorld 调用。
3. S2 阶段：实现核心服务（Auth、UserCenter）的 TarsGo 骨架和 HealthCheck。
4. S3 阶段：完善服务治理、监控、灰度发布能力。
5. 未来评估：如 gRPC 生态在特定场景（如流式通信）有显著优势，可另行 ADR 评估共存或回退方案。

## 14. 相关文档

- [ADR-0001-总体系统架构.md](ADR-0001-总体系统架构.md)
- [ADR-0003-服务协议使用Protobuf.md](ADR-0003-服务协议使用Protobuf.md)
- [docs/api/tars规范.md](../api/tars规范.md)
- [docs/api/protobuf规范.md](../api/protobuf规范.md)
- [docs/api/协议编号注册表.md](../api/协议编号注册表.md)
- [docs/wiki/CODE-WIKI.md](../wiki/CODE-WIKI.md)
