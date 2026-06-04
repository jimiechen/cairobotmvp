/**
 * CLI 报告生成器
 *
 * 支持三种格式：
 * - generateSingleReport(): 单次发送 Markdown 报告
 * - generateSuiteReport(): 用例集总报告（Markdown + JUnit XML）
 * - generateCaptureReport(): 场景录屏报告（T8 使用）
 */

// --- 类型定义 ---

/** 单条测试用例结果 */
interface TestCaseResult {
  caseId: string;
  name: string;
  passed: boolean;
  error?: string;
  traceId?: string;
  durationMs: number;
  reportPath?: string;
}

/** 用例集报告输出 */
interface SuiteReportOutput {
  summaryMd: string;
  junitXml: string;
}

/** 证据项（截图/视频/请求/响应） */
interface EvidenceItem {
  type: 'screenshot' | 'video' | 'request' | 'response';
  path: string;
}

// --- 导出函数 ---

/**
 * 生成单次发送的 Markdown 报告
 *
 * @param req 请求数据（含 payload）
 * @param rsp 响应数据（含 status / businessCode / durationMs）
 * @param meta 协议元数据（含 maxType / minType / protocolName）
 * @returns 完整 Markdown 字符串
 */
export function generateSingleReport(
  req: { payload: Record<string, unknown> },
  rsp: { status: number; businessCode: number; durationMs: number },
  meta: { maxType: number; minType: number; protocolName: string },
): string {
  const lines = [
    '# Single Send Report',
    '',
    '## 元数据',
    `- 时间: ${new Date().toISOString()}`,
    `- 协议: ${meta.maxType}/${meta.minType} ${meta.protocolName}`,
    '',
    '## 请求',
    '```json',
    JSON.stringify(req.payload, null, 2),
    '```',
    '',
    '## 响应',
    `- HTTP Status: ${rsp.status}`,
    `- 业务码: ${rsp.businessCode}`,
    `- 耗时: ${rsp.durationMs}ms`,
    '',
  ];
  return lines.join('\n');
}

/**
 * 生成用例集总报告（Markdown 摘要 + JUnit XML）
 *
 * @param cases 测试用例结果数组
 * @returns 含 summaryMd 和 junitXml 的输出对象
 */
export function generateSuiteReport(cases: TestCaseResult[]): SuiteReportOutput {
  const passed = cases.filter((c) => c.passed).length;
  const failed = cases.filter((c) => !c.passed).length;

  const summaryMd = [
    '# Suite Run Report',
    '',
    '## Summary',
    `- Total: ${cases.length}`,
    `- Passed: ${passed}`,
    `- Failed: ${failed}`,
    `- Rate: ${cases.length > 0 ? ((passed / cases.length) * 100).toFixed(1) : '0'}%`,
    '',
    '## Failed Cases',
    ...cases
      .filter((c) => !c.passed)
      .map((c) => `- [${c.caseId}] ${c.name}: ${c.error ?? 'unknown'}`),
    '',
  ].join('\n');

  const junitXml = generateJUnitXml(cases);

  return { summaryMd, junitXml };
}

/**
 * 生成场景录屏报告
 *
 * @param scenario 场景名称
 * @param evidences 证据列表（截图/视频等）
 * @returns Markdown 报告字符串
 */
export function generateCaptureReport(scenario: string, evidences: EvidenceItem[]): string {
  return [
    '# Capture Report',
    '',
    `Scenario: ${scenario}`,
    '',
    '## Evidences',
    ...evidences.map((e) => `- ${e.type}: ${e.path}`),
    '',
  ].join('\n');
}

// --- 内部辅助函数 ---

/**
 * 生成 JUnit XML 格式的测试报告
 *
 * 可被标准 CI 工具（Jenkins / GitHub Actions 等）解析
 */
function generateJUnitXml(cases: TestCaseResult[]): string {
  const failures = cases.filter((c) => !c.passed);
  const xml = [
    '<?xml version="1.0" encoding="UTF-8"?>',
    `<testsuites tests="${cases.length}" failures="${failures.length}">`,
    `  <testsuite name="proto-tester" tests="${cases.length}" failures="${failures.length}">`,
    ...cases.map((c) =>
      c.passed
        ? `    <testcase name="${c.caseId}" classname="proto-tester" time="${(c.durationMs / 1000).toFixed(3)}" />`
        : `    <testcase name="${c.caseId}" classname="proto-tester" time="${(c.durationMs / 1000).toFixed(3)}">
      <failure message="${escapeXmlAttribute(c.error ?? 'unknown')}">${escapeXmlAttribute(c.error ?? '')}</failure>
    </testcase>`,
    ),
    '  </testsuite>',
    '</testsuites>',
  ];
  return xml.join('\n');
}

/** 转义 XML 属性中的特殊字符 */
function escapeXmlAttribute(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}
