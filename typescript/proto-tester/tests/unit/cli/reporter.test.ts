/**
 * CLI reporter 单元测试
 *
 * 覆盖三种报告格式：
 * - generateSingleReport: Markdown 含关键字段
 * - generateSuiteReport: Markdown + JUnit XML 结构
 * - generateCaptureReport: 含证据列表
 * - JUnit XML 格式可被标准工具解析
 */
import { describe, it, expect } from 'vitest';
import {
  generateSingleReport,
  generateSuiteReport,
  generateCaptureReport,
} from '@/cli/reporter';

describe('reporter.generateSingleReport', () => {
  it('输出包含协议元数据关键字段', () => {
    const md = generateSingleReport(
      { payload: { name: 'test' } },
      { status: 200, businessCode: 10200, durationMs: 42 },
      { maxType: 2100, minType: 2101, protocolName: 'HelloWorld' },
    );

    expect(md).toContain('# Single Send Report');
    expect(md).toContain('2100/2101');
    expect(md).toContain('HelloWorld');
    expect(md).toContain('200');
    expect(md).toContain('10200');
    expect(md).toContain('42ms');
  });

  it('输出包含 JSON 格式的请求体', () => {
    const md = generateSingleReport(
      { payload: { key: 'value', num: 123 } },
      { status: 200, businessCode: 10200, durationMs: 10 },
      { maxType: 1, minType: 1, protocolName: 'Test' },
    );

    expect(md).toContain('"key": "value"');
    expect(md).toContain('"num": 123');
  });

  it('业务码非 10200 时正常输出', () => {
    const md = generateSingleReport(
      { payload: {} },
      { status: 400, businessCode: 10400, durationMs: 5 },
      { maxType: 1, minType: 1, protocolName: 'FailCase' },
    );

    expect(md).toContain('10400');
    expect(md).toContain('5ms');
  });
});

describe('reporter.generateSuiteReport', () => {
  it('全通过时 summaryMd 统计正确', () => {
    const cases = [
      { caseId: 'TC001', name: 'HelloWorld', passed: true, durationMs: 100 },
      { caseId: 'TC002', name: 'HealthCheck', passed: true, durationMs: 50 },
    ];

    const result = generateSuiteReport(cases);

    expect(result.summaryMd).toContain('# Suite Run Report');
    expect(result.summaryMd).toContain('Total: 2');
    expect(result.summaryMd).toContain('Passed: 2');
    expect(result.summaryMd).toContain('Failed: 0');
    expect(result.summaryMd).toContain('Rate: 100.0%');
  });

  it('有失败时 summaryMd 列出失败用例', () => {
    const cases = [
      { caseId: 'TC001', name: 'HelloWorld', passed: true, durationMs: 100 },
      {
        caseId: 'TC002',
        name: 'AuthFail',
        passed: false,
        error: 'Token expired',
        durationMs: 200,
      },
    ];

    const result = generateSuiteReport(cases);

    expect(result.summaryMd).toContain('Passed: 1');
    expect(result.summaryMd).toContain('Failed: 1');
    expect(result.summaryMd).toContain('[TC002] AuthFail: Token expired');
    expect(result.summaryMd).toContain('Rate: 50.0%');
  });

  it('JUnit XML 包含标准的 testsuite 根元素', () => {
    const cases = [
      { caseId: 'TC001', name: 'Ok', passed: true, durationMs: 10 },
    ];

    const result = generateSuiteReport(cases);

    expect(result.junitXml).toContain('<?xml version="1.0" encoding="UTF-8"?>');
    expect(result.junitXml).toContain('<testsuites ');
    expect(result.junitXml).toContain('<testsuite name="proto-tester"');
    expect(result.junitXml).toContain('tests="1"');
    expect(result.junitXml).toContain('failures="0"');
  });

  it('JUnit XML 失败用例包含 failure 元素', () => {
    const cases = [
      {
        caseId: 'TC_FAIL',
        name: 'BrokenTest',
        passed: false,
        error: 'assertion failed',
        durationMs: 999,
      },
    ];

    const result = generateSuiteReport(cases);

    expect(result.junitXml).toContain('<testcase name="TC_FAIL"');
    expect(result.junitXml).toContain('<failure message="assertion failed">');
    expect(result.junitXml).toContain('</failure>');
    expect(result.junitXml).toContain('failures="1"');
  });

  it('JUnit XML 通过用例无 failure 元素', () => {
    const cases = [
      { caseId: 'TC_OK', name: 'Passing', passed: true, durationMs: 1 },
    ];

    const result = generateSuiteReport(cases);

    expect(result.junitXml).toContain('<testcase name="TC_OK"');
    expect(result.junitXml).not.toContain('<failure');
  });

  it('空用例集返回零统计', () => {
    const result = generateSuiteReport([]);

    expect(result.summaryMd).toContain('Total: 0');
    expect(result.summaryMd).toContain('Rate: 0%');
    expect(result.junitXml).toContain('tests="0"');
  });

  it('JUnit XML 特殊字符被正确转义', () => {
    const cases = [
      {
        caseId: 'TC_XML',
        name: 'XmlTest',
        passed: false,
        error: 'value < "bad" & \'worse\' > end',
        durationMs: 10,
      },
    ];

    const result = generateSuiteReport(cases);

    // XML 特殊字符应被转义
    expect(result.junitXml).toContain('&lt;');
    expect(result.junitXml).toContain('&gt;');
    expect(result.junitXml).toContain('&amp;');
    expect(result.junitXml).toContain('&quot;');
    expect(result.junitXml).not.toContain('< "bad"');
  });
});

describe('reporter.generateCaptureReport', () => {
  it('输出包含场景名和证据列表', () => {
    const report = generateCaptureReport('LoginFlow', [
      { type: 'screenshot', path: '/tmp/login-1.png' },
      { type: 'video', path: '/tmp/login.mp4' },
      { type: 'request', path: '/tmp/req.json' },
      { type: 'response', path: '/tmp/rsp.json' },
    ]);

    expect(report).toContain('# Capture Report');
    expect(report).toContain('Scenario: LoginFlow');
    expect(report).toContain('- screenshot: /tmp/login-1.png');
    expect(report).toContain('- video: /tmp/login.mp4');
    expect(report).toContain('- request: /tmp/req.json');
    expect(report).toContain('- response: /tmp/rsp.json');
  });

  it('无证据时仅输出场景标题', () => {
    const report = generateCaptureReport('EmptyScenario', []);

    expect(report).toContain('Scenario: EmptyScenario');
    expect(report).not.toContain('- screenshot:');
  });
});
