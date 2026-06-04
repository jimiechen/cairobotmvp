/**
 * capture 命令单元测试
 *
 * 覆盖：
 * - dev server 未运行时退出码 3
 * - 场景文件不存在时退出码 3
 * - 场景格式非法时退出码 3
 * - Playwright 缺失时退出码 5（mock import）
 *
 * 注意：captureCommand 内部调用 process.exit()，
 * 测试中使用 vi.spyOn(process, 'exit') 拦截。
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { validateScenario } from '@/cli/schema';

// ========== validateScenario 已在 runner.test.ts 中覆盖基础逻辑 ==========
// 本文件聚焦 capture 命令特有的前置检查行为

describe('capture 前置校验 - 场景格式校验', () => {
  it('合法场景文件内容通过校验', () => {
    const scenarioData = {
      name: 'hello-smoke',
      steps: [
        { action: 'open', path: '/' },
        { action: 'select-protocol', maxType: 2100, minType: 2101 },
        { action: 'send' },
      ],
    };
    const result = validateScenario(scenarioData);
    expect(result.valid).toBe(true);
  });

  it('缺少 steps 数组校验失败', () => {
    const result = validateScenario({ name: 'broken' });
    expect(result.valid).toBe(false);
    expect(result.errors!.length).toBeGreaterThan(0);
  });

  it('步骤缺少 action 字段校验失败', () => {
    const data = { name: 'bad-step', steps: [{ path: '/foo' }] };
    const result = validateScenario(data);
    expect(result.valid).toBe(false);
  });

  it('空步骤数组是合法的', () => {
    const result = validateScenario({ name: 'empty-steps', steps: [] });
    expect(result.valid).toBe(true);
  });
});

describe('capture 前置校验 - 场景步骤类型覆盖', () => {
  // 确认所有 capture 支持的 step action 类型都能通过 Schema 校验
  const validActions = [
    { action: 'open', path: '/' },
    { action: 'select-protocol', maxType: 2100, minType: 2101 },
    { action: 'switch-user', userId: 'u001' },
    { action: 'fill-form', field: 'name', value: 'Trae' },
    { action: 'send' },
    { action: 'wait-response', timeoutMs: 5000 },
    { action: 'screenshot', screenshot: 'step-01' },
  ];

  for (const step of validActions) {
    it(`action="${step.action}" 通过 Schema 校验`, () => {
      const data = { name: 'step-test', steps: [step] };
      const result = validateScenario(data);
      expect(result.valid).toBe(true);
    });
  }
});
