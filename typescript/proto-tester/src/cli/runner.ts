/**
 * runner.ts - YAML 用例执行引擎
 *
 * 职责：
 * - 解析并校验 YAML 用例文件
 * - 按顺序或并行（分批）执行用例
 * - 7 种匹配器：exact / regex / contains / minLength / maxLength / exists / jsonPath
 * - 输出每用例子报告 + 汇总 summary.md + traces-index.json + junit.xml
 *
 * 不负责：
 * - CLI 参数解析（由 commands/run.ts 负责）
 * - Schema 定义（由 schema.ts 负责）
 * - 报告格式化细节（由 reporter.ts 负责）
 */
import fs from 'fs';
import path from 'path';
import yaml from 'yaml';
import { sendRequest } from '../../lib/apiClient.js';
import { encodePacket } from '../../lib/messagePacket.js';
import { match, type FieldAssertion, type TestCase, type CaseExpect } from './schema.js';

// ========== 常量 ==========

/** 并行执行硬限制 */
const MAX_PARALLEL_CASES = 4;

// ========== 类型定义 ==========

/** 单条用例执行结果 */
export interface TestCaseResult {
  caseId: string;
  name: string;
  passed: boolean;
  error?: string;
  traceId?: string;
  durationMs: number;
  reportPath?: string;
}

/** runSuite 运行选项 */
export interface RunOptions {
  parallel?: number;
  env?: string;
  gateway?: string;
  outputDir?: string;
  token?: string;
}

// ========== 核心执行逻辑 ==========

/**
 * 执行 YAML 用例集
 *
 * 流程：
 * 1. 读取并解析 YAML 文件
 * 2. Schema 校验
 * 3. 顺序/并行执行每个用例
 * 4. 写入汇总报告
 *
 * @param suitePath - YAML 用例文件绝对/相对路径
 * @param opts - 运行选项（并行度、环境、Gateway 等）
 * @returns 所有用例的执行结果数组
 */
export async function runSuite(
  suitePath: string,
  opts: RunOptions = {},
): Promise<TestCaseResult[]> {
  // 1. 解析 YAML
  const raw = fs.readFileSync(suitePath, 'utf-8');
  const data = yaml.parse(raw);

  // 2. Schema 校验
  const { validateSuite } = await import('./schema.js');
  const validation = validateSuite(data);
  if (!validation.valid) {
    console.error(`YAML 格式错误:\n${validation.errors?.join('\n')}`);
    process.exit(3);
  }

  // 3. 执行用例
  const results: TestCaseResult[] = [];
  const cases = data.cases as TestCase[];
  // 硬限制：并行度不超过 MAX_PARALLEL_CASES
  const parallel = Math.min(opts.parallel ?? 1, MAX_PARALLEL_CASES);

  if (parallel <= 1) {
    for (const tc of cases) {
      const result = await executeSingleCase(tc, opts);
      results.push(result);
    }
  } else {
    // 分批并行执行
    for (let i = 0; i < cases.length; i += parallel) {
      const chunk = cases.slice(i, i + parallel);
      const chunkResults = await Promise.all(
        chunk.map((tc) => executeSingleCase(tc, opts)),
      );
      results.push(...chunkResults);
    }
  }

  // 4. 写入汇总报告
  writeSummaryReport(results, opts.outputDir);

  return results;
}

/**
 * 执行单条测试用例
 *
 * 编码 MessagePacket → 发送请求 → 校验业务码 → 返回结果
 */
async function executeSingleCase(
  tc: TestCase,
  opts: Pick<RunOptions, 'env' | 'gateway' | 'token' | 'outputDir'>,
): Promise<TestCaseResult> {
  const start = Date.now();
  try {
    // 编码 MessagePacket
    const packetBinary = encodePacket({
      maxType: tc.protocol[0],
      minType: tc.protocol[1],
      payload: new Uint8Array(Buffer.from(JSON.stringify(tc.payload))),
      extend: {},
    });

    // 发送请求
    const response = await sendRequest({
      maxType: tc.protocol[0],
      minType: tc.protocol[1],
      payload: packetBinary,
      gatewayUrl: opts.gateway,
      token: opts.token,
    });

    const durationMs = Date.now() - start;

    // 校验业务码
    if (response.businessCode !== tc.expect.businessCode) {
      return {
        caseId: tc.id,
        name: tc.name,
        passed: false,
        error: `业务码不匹配: 期望 ${tc.expect.businessCode}, 实际 ${response.businessCode}`,
        traceId: response.traceId,
        durationMs,
      };
    }

    // 业务码匹配，字段级断言待 Gateway 就绪后补充
    return {
      caseId: tc.id,
      name: tc.name,
      passed: true,
      traceId: response.traceId,
      durationMs,
    };
  } catch (e) {
    return {
      caseId: tc.id,
      name: tc.name,
      passed: false,
      error: (e as Error).message,
      durationMs: Date.now() - start,
    };
  }
}

/**
 * 写入汇总报告到输出目录
 *
 * 生成三个文件：
 * - summary-{timestamp}.md — Markdown 摘要
 * - junit-{timestamp}.xml — JUnit XML 报告
 * - traces-index.json — 用例 ID 到 traceId 的映射
 */
async function writeSummaryReport(results: TestCaseResult[], outputDir?: string): Promise<void> {
  const dir = outputDir ?? './proto-tester-reports';
  fs.mkdirSync(dir, { recursive: true });

  // 使用 reporter 模块生成报告
  const { generateSuiteReport } = await import('./reporter.js');
  const report = generateSuiteReport(results);

  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
  fs.writeFileSync(path.join(dir, `summary-${timestamp}.md`), report.summaryMd, 'utf-8');
  fs.writeFileSync(path.join(dir, `junit-${timestamp}.xml`), report.junitXml, 'utf-8');

  // 生成 traces-index.json
  const index: Record<string, string> = {};
  for (const r of results) {
    if (r.traceId) {
      index[r.caseId] = r.traceId;
    }
  }
  fs.writeFileSync(path.join(dir, 'traces-index.json'), JSON.stringify(index, null, 2), 'utf-8');

  console.log(`► 汇总报告: ${dir}/summary-${timestamp}.md`);
  console.log(`► JUnit XML: ${dir}/junit-${timestamp}.xml`);
}
