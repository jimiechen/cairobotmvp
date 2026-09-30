# R-04 MySQL 联调测试报告：MySqlDataSource 真实数据通道（TDD）

| 字段 | 值 |
|---|---|
| 日期 | 2026-09-30 |
| 分支 | feature/fa-orchestrator-skeleton |
| 任务 | R-04 联调环境补齐（ISS-FA3 前置，主控确认清单第 3 项） |
| 相关 PRD | PRD-12（F-06/F-07/F-08/F-09；§13 R-04；§17 接口清单） |
| 相关 ADR | ADR-017 |
| Issue | 无独立 Issue（Issue 拆分草案主控确认清单项；范围经用户确认：分析类查询不纳入） |
| 执行人 | Trae |

## 1. 测试环境

- macOS（arm64）/ Python 3.11.0 / pytest 7.4.3 / pymysql 1.4.6（本任务新增锁定）
- 真实联调环境（R-04 探测确认，只读操作）：
  - MySQL：192.168.1.6:3306，库 **workbench**，九张契约表全部存在且有真实数据
  - 连接凭据走 `FA_MYSQL_*` 环境变量（与仓库 docs 既有记录一致，不入代码不入日志）
- 外部依赖边界：HTTP API（192.168.1.6:18000）可达但本任务不涉及；单元测试全程 Fake 连接，禁止真实连库

## 2. 联调环境探测结论（只读）

| 探测项 | 结果 |
|---|---|
| Q1 契约 SQL 原样执行 | ✅ 样本日 2026-09-29 返回 13 场（fixture 7 场为其子集） |
| 四路径解析真实落点 | ✅ 周一003（fid=1398605）→ `match`.id=2635 |
| 六模块数据 | ✅ 2635 六模块行数 231/272/32/20/6/4（全就绪） |
| 模块表 match_id 域 | ✅ = `match`.id（数字域）；matches.id（4164）全零反证 |
| `source_id` | 第三方站点编号（4517966），非 500com fid，不能替代解析 |
| `match` 表数据质量 | ⚠️ 8 组同场次重复行（数据集中于小 id；ORDER BY id + min 语义确定性安全） |
| 契约列名差异 | ⚠️ SQL_JC_MATCH_CANDIDATES 的 cc/home_team/away_team 与实际列 cc_id/home/away 不符 |

## 3. 测试用例与结果

命令：

```bash
cd python && make unit          # 单元测试（Fake 连接，禁真实连库）
cd python && FA_MYSQL_INTEGRATION=1 FA_MYSQL_HOST=… FA_MYSQL_PORT=3306 \
  FA_MYSQL_USER=… FA_MYSQL_PASSWORD=… FA_MYSQL_DB=workbench make integration
```

结果：**单元 151 passed, 4 skipped**（fa_orchestrator 150 = FA1/FA2 131 + 本任务 20；集成 4 例默认门控 skip 并输出原因）；**真实联调 4 passed（0.76s）**。

| 测试文件 | 用例数 | 覆盖点 |
|---|---|---|
| test_mysql_data_source.py | 18 | 配置优先级与环境变量兜底、缺配置抛 MySqlConfigError、repr 密码脱敏、懒连接、Q1 契约 SQL 原样执行与参数化、候选行列名适配（cc_id→cc 等）、适配 SQL 常量/确定性排序/反引号、契约文本保留审计、Q2b 按 match_id 六模块核查与占位符断言、无映射返回空不臆测、缺模块计 0、分析类查询恒空（契约未归档）、DataAgent match_ids 传参契约、溯源 channel=mysql、数据库异常不吞掉 |
| integration/test_mysql_integration.py | 4 | 真实 Q1 清单（≥7 场含锚点 fid）、四路径解析真实落点、Q2b 六模块真实行数全非零、DataAgent 真实通道端到端（13 场 + channel=mysql） |
| test_cli.py（扩展 2 例） | 2 | --source mysql 缺配置报错退出码 2；mysql 通道拒绝 --analysis-output（分析类契约未归档） |

## 4. 端到端验证（CLI 真实通道）

```bash
cd python/ai && PYTHONPATH=fa_orchestrator \
  FA_MYSQL_HOST=… FA_MYSQL_PORT=3306 FA_MYSQL_USER=… FA_MYSQL_PASSWORD=… FA_MYSQL_DB=workbench \
  python3 -m fa_orchestrator run --date 2026-09-29 --source mysql --output /tmp/fa_r04_ready_report.json
```

输出（exit=0）：**13/13 全就绪**，全部场次解析出 match_id（2635~2655），溯源快照 `channel=mysql`、`ready_rate=1.0`。

```text
竞彩日 2026-09-29 数据就绪率: 13/13
  周一003 亚美尼亚 vs 黑山 fid=1398605 match_id=2635 状态=ready 缺口=-
  周一004 土耳其 vs 意大利 fid=1398610 match_id=2636 状态=ready 缺口=-
  …（共 13 场，全部 ready）
```

门控关闭验证：`make integration`（未设 FA_MYSQL_INTEGRATION）输出显式跳过原因（testing.md §8.2）。

## 5. TDD 过程记录

1. 测试红：test_mysql_data_source 收集失败（MySqlConfigError 未定义）；test_cli 新增 2 例失败（argparse 不识别 --source），失败原因均符合预期。
2. 实现绿：mysql_data_source.py（新文件，228 行，职责拆分自 data_source.py 以满足单文件 ≤300 行）+ data_agent 传参 + cli 双通道 + pyproject 锁定 pymysql==1.4.6。
3. 联调绿：FA_MYSQL_INTEGRATION=1 门控 4/4 通过；CLI 真实通道 13/13 就绪。

## 6. 联调发现与适配决策（核心产出）

| # | 发现 | 决策 |
|---|---|---|
| D1 | 契约 SQL_JC_MATCH_CANDIDATES 列名与 workbench 实际不符（cc→cc_id、home_team→home、away_team→away） | 契约文本保留原样供审计对照；数据源内以 SQL_JC_CANDIDATES_MYSQL 适配执行，get_sql_contracts 双文本并列暴露差异 |
| D2 | 模块表 match_id 域 = `match`.id，非 500com fid | DataAgent 将四路径解析结果 match_ids 透传 query_q2b_counts（可选参数，fixture 行为不变） |
| D3 | `match` 表 8 组重复场次，数据集中于小 id | 候选查询 ORDER BY id + 解析器 min 语义 → 确定性取小 id；重复行数据质量问题登记待上游清理 |
| D4 | 分析类查询（Q3/Q5/Q9~Q12）SQL 契约在《需求文档-局域网数据获取》§四，外部文档未归档 | 经主控确认本阶段不实现：一律返回 None/空并注释说明；待需求文档归档后另行任务 |
| D5 | `match` 表无别名列 | 解析路径 4（别名学习）在 MySQL 通道自然不触发，1-3 路径已覆盖样本日 |

## 7. 工程规范核查

- 文件行数：mysql_data_source 228 / data_source 117 / data_agent 95 / cli 127 / 测试文件 267、85，全部 ≤300
- 中文注释：核心类/方法均说明职责边界、输入输出与失败行为；联调适配均有"为什么"说明
- 安全边界：凭据仅走环境变量/构造参数，repr 密码脱敏；只读查询，无任何写入语句
- SQL 硬性规则：Q1 契约原文执行、全参数化 %s、`match` 保留字反引号、六模块表名白名单（build_q2b_sql 校验）
- 依赖：pymysql==1.4.6 精确锁定，pyproject 注明"DB 驱动非 AI 框架，不违反 ADR-017 白名单"

## 8. 风险与遗留

- R1（承接 FA1/FA2）：dev 分支不存在；`make rules` 308 项存量失败与本任务无关
- R2：`match` 表重复行（8 组）与 D1 契约差异需上游（需求文档维护方）确认：以实际库为准修契约，还是修库对齐契约
- R3：分析类查询（Q3/Q5/Q9~Q12）与 value_ratio 数据块待《需求文档》归档后实现（ISS-FA3 前确认）
- R4：联调测试含真实环境锚点（样本日 13 场/2635 行数），若联调库数据清理需同步更新锚点
- R5：HTTP API 通道（python-server，18000）未联调，属后续数据通道补抓范围

## 9. 结论

R-04 联调环境补齐达成：MySQL 真实通道（Q1/候选行/Q2b + 四路径解析）端到端可运行，样本日 13/13 全就绪，溯源声明 channel=mysql；三条仓库内契约联调通过，三条环境差异以适配层消化并登记。ISS-FA3 数据层前置阻塞解除，可提交评审。
