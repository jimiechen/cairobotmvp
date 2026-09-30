# ISS-FA1 测试报告：编排骨架与数据层（TDD）

| 字段 | 值 |
|---|---|
| 日期 | 2026-09-30 |
| 分支 | feature/fa-orchestrator-skeleton（自 main 创建） |
| Issue | ISS-FA1（docs/drafts/PRD-12-M1-M5-Issue拆分草案.md） |
| 相关 PRD | PRD-12（F-01/F-02/F-03/F-05骨架/F-06/F-07/F-08/F-09） |
| 相关 ADR | ADR-017 |
| 执行人 | Trae |

## 1. 测试环境

- macOS（arm64）/ Python 3.11.0 / pytest 7.4.3
- 依赖（精确锁定，ADR-017 白名单内）：langgraph==1.2.12、langchain-core==1.6.6、pyyaml>=6.0.1
- 外部服务：无。全部测试 Mock LLM 与外部服务；MySQL（192.168.1.6:3306）与 HTTP（192.168.1.6:18000）为仓库外环境，禁止真实调用

## 2. 测试对象

`python/ai/fa_orchestrator/`（新增包，实现 11 文件 936 行 + 测试 7 文件 717 行）：

| 模块 | 职责 | PRD 功能项 |
|---|---|---|
| sql_contracts.py | Q1/Q2b/候选行 SQL 契约常量与六模块表名映射 | F-06 |
| id_resolver.py | fid→match.id 四路径兜底解析（纯函数） | F-07 |
| integrity.py | 六模块完整性核查与日级就绪率报告 | F-08 |
| config.py | YAML 编排配置加载校验 | F-01 |
| state.py / graph_builder.py | TypedState 与 StateGraph 装配（Mermaid 可导出） | F-01 |
| llm_gateway.py | OpenAI 兼容统一网关（超时/重试/密钥脱敏） | F-03 |
| planner_agent.py | Plan-and-Execute 结构化计划产出 | F-02 |
| data_source.py / data_agent.py | fixture 数据源、取数编排、溯源快照 | F-06/F-09 |
| cli.py / __main__.py | `run --date` 骨架 | F-05 |

## 3. 测试用例与结果

命令：

```bash
cd python && make unit
# 等价：cd python/ai && python3 -m pytest fa_orchestrator/tests/ -v
```

结果：**61 passed**（fa_orchestrator 60 + hello 既有 1），0 failed，8.06s。

分组覆盖：

| 测试文件 | 用例数 | 覆盖点 |
|---|---|---|
| test_sql_contracts.py | 10 | Q1 契约逐字断言（500com/is_active/参数化/排序/fid 别名）、Q2b 六模块表名、占位符生成、表名白名单、`match` 反引号、play_type='jingcai' |
| test_id_resolver.py | 13 | 四路径参数化（cc+窗口/跨日±2边界/仅cc/队名归一/别名）、多候选取最近、非竞彩玩法排除、全未命中抛错含场次编号 |
| test_integrity.py | 8 | 六模块就绪/部分缺口/全缺 blocked/未知模块拒绝/就绪率 6/7/空日报告 |
| test_config.py | 6 | YAML 加载、缺文件/缺 entry_point/边引用未定义节点/入口未定义均报错 |
| test_graph_builder.py | 4 | 编译产物类型、注册缺失报错、Mermaid 导出含三节点、planner→data_agent→ready_report 流转顺序（trace 断言） |
| test_llm_gateway.py | 7 | Mock Server 契约（path/Authorization Bearer/model/temperature/messages）、AIMessage 解析、5xx 重试成功/耗尽报错、from_env、缺环境变量报错、密钥不落 stdout/stderr |
| test_planner_agent.py | 5 | 结构化计划、必填字段校验、markdown 包裹 JSON 容错、非法 JSON 报错、提示词含竞彩日 |
| test_cli.py | 7 | fixture 数据源 Q1/Q2b/候选行、DataAgent 端到端就绪率报告、解析失败不静默跳过、CLI run --date 输出与落盘 |

## 4. 端到端验证（CLI）

```bash
cd python/ai/fa_orchestrator && python3 -m fa_orchestrator run --date 2026-09-29 --fixture tests/fixtures/sample_day_20260929.json
```

输出（exit=0）：竞彩日 2026-09-29 数据就绪率 6/7；七场明细含 fid、四路径解析落点 match_id、状态与缺口（周二013 缺 vote → incomplete）。就绪率报告 JSON 可落盘（--output）。

## 5. TDD 过程记录

1. 测试红：实现前 8 个测试模块收集失败（ModuleNotFoundError: fa_orchestrator.xxx），失败原因符合预期。
2. 实现绿：首轮 51/60 通过；修复 3 处——实现 bug 1 个（FixtureDataSource.__init__ 属性赋值顺序）、测试构造笔误 2 处（dict 键覆盖导致样本数错、缺口断言顺序），结构化调整 1 处（每场明细内嵌 integrity 子对象，同步 CLI 打印）。
3. 回归：make unit 61 passed（含 hello 既有测试，无破坏）。

## 6. Bug 列表

| Bug | 等级 | 状态 | 说明 |
|---|---|---|---|
| FIXTURE-INIT-ORDER | P2 | 已修复 | FixtureDataSource.__init__ 先引用后赋值 self.fixture，测试立即暴露 |
| PY-MK-PYTHON-VAR | P2 | 已修复 | python/Makefile 缺 PYTHON 变量定义，make unit 一直实际未执行 pytest（`-m: command not found` 被静默吞掉） |

## 7. 工程规范核查

- 依赖白名单：pyproject 中无白名单外 AI 框架（grep langchain(?!-core)/crewai/autogen 为空）
- 文件行数：最大 148 行（test_id_resolver.py），全部 ≤300；方法均 ≤30 行
- 中文注释：核心类/方法均已注释并说明职责边界与失败行为
- SQL 安全：全参数化 %s，`match` 反引号，Q2b 表名白名单校验
- 密钥安全：api_key 仅进请求头，日志打印 base_url/model/字节数；测试断言密钥不落 stdout/stderr

## 8. 风险与遗留

- R1：dev 分支不存在，本分支自 main 创建（违反 git.md §1 默认值，待主控确认是否补建 dev）
- R2：MySqlDataSource 仅为骨架（R-04 联调环境未就绪），ISS-FA2 前需实现真实 MySQL 通道
- R3：`make rules` 测试用例注册表检查失败为存量问题（typescript/go/node_modules 项，与本 Issue 无关，需另立修复任务）
- R4：pyyaml 为白名单说明外的工程依赖（F-01 YAML 解析必需，非 AI 框架，不违反 ADR-017 §9.1）

## 9. 结论

ISS-FA1 验收标准达成：StateGraph 骨架跑通 planner→data_agent→ready_report 链路，样本日 fixture 就绪率报告可产出，全部测试通过。可提交评审。
