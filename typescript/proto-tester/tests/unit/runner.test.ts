/**
 * runner.match 单元测试
 *
 * 覆盖 7 种匹配器：
 * - exact: 精确匹配
 * - regex: 正则表达式匹配
 * - contains: 包含匹配
 * - minLength: 最小长度
 * - maxLength: 最大长度
 * - exists: 存在性检查
 * - jsonPath: JSON 路径存在性
 *
 * 同时覆盖：
 * - Schema 校验失败场景
 * - 并行执行参数传递
 */
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { match } from '@/cli/schema';
import { validateSuite, validateScenario } from '@/cli/schema';

// ========== 匹配器测试 ==========

describe('runner.match - exact 匹配器', () => {
  it('值完全相等时返回 true', () => {
    expect(match('hello', { field: 'msg', matcher: 'exact', value: 'hello' })).toBe(true);
  });

  it('值不相等时返回 false', () => {
    expect(match('hello', { field: 'msg', matcher: 'exact', value: 'world' })).toBe(false);
  });

  it('数字精确匹配', () => {
    expect(match(10200, { field: 'code', matcher: 'exact', value: 10200 })).toBe(true);
  });
});

describe('runner.match - regex 匹配器', () => {
  it('正则匹配成功时返回 true', () => {
    expect(match('hello-world-123', { field: 'text', matcher: 'regex', value: '^hello.*\\d+$' })).toBe(true);
  });

  it('正则不匹配时返回 false', () => {
    expect(match('hello-world', { field: 'text', matcher: 'regex', value: '^\\d+$' })).toBe(false);
  });
});

describe('runner.match - contains 匹配器', () => {
  it('包含子串时返回 true', () => {
    expect(match('Hello World', { field: 'text', matcher: 'contains', value: 'World' })).toBe(true);
  });

  it('不包含子串时返回 false', () => {
    expect(match('Hello World', { field: 'text', matcher: 'contains', value: 'Foo' })).toBe(false);
  });
});

describe('runner.match - minLength 匹配器', () => {
  it('字符串长度 >= 阈值时返回 true', () => {
    expect(match('hello', { field: 'name', matcher: 'minLength', value: 3 })).toBe(true);
  });

  it('字符串长度 < 阈值时返回 false', () => {
    expect(match('hi', { field: 'name', matcher: 'minLength', value: 5 })).toBe(false);
  });

  it('null 值视为长度 0，不满足 minLength > 0', () => {
    expect(match(null, { field: 'val', matcher: 'minLength', value: 1 })).toBe(false);
  });
});

describe('runner.match - maxLength 匹配器', () => {
  it('字符串长度 <= 阈值时返回 true', () => {
    expect(match('hi', { field: 'name', matcher: 'maxLength', value: 5 })).toBe(true);
  });

  it('字符串长度 > 阈值时返回 false', () => {
    expect(match('hello world', { field: 'name', matcher: 'maxLength', value: 3 })).toBe(false);
  });
});

describe('runner.match - exists 匹配器', () => {
  it('非 null/undefined 值返回 true', () => {
    expect(match('', { field: 'val', matcher: 'exists' })).toBe(true);
    expect(match(0, { field: 'val', matcher: 'exists' })).toBe(true);
    expect(match(false, { field: 'val', matcher: 'exists' })).toBe(true);
  });

  it('null 返回 false', () => {
    expect(match(null, { field: 'val', matcher: 'exists' })).toBe(false);
  });

  it('undefined 返回 false', () => {
    expect(match(undefined, { field: 'val', matcher: 'exists' })).toBe(false);
  });
});

describe('runner.match - jsonPath 匹配器', () => {
  it('点号路径命中时返回 true', () => {
    const obj = { user: { name: 'Trae', age: 1 } };
    expect(match(obj, { field: 'data', matcher: 'jsonPath', value: 'user.name' })).toBe(true);
  });

  it('数组索引路径命中时返回 true', () => {
    const obj = { items: [{ id: 1 }, { id: 2 }] };
    expect(match(obj, { field: 'data', matcher: 'jsonPath', value: 'items[0].id' })).toBe(true);
  });

  it('路径不存在时返回 false', () => {
    const obj = { user: { name: 'Trae' } };
    expect(match(obj, { field: 'data', matcher: 'jsonPath', value: 'user.email' })).toBe(false);
  });

  it('中间路径为 null 时返回 false', () => {
    const obj = { user: null };
    expect(match(obj, { field: 'data', matcher: 'jsonPath', value: 'user.name' })).toBe(false);
  });
});

describe('runner.match - 未知匹配器', () => {
  it('抛出错误提示未知匹配器名称', () => {
    expect(() => match('x', { field: 'f', matcher: 'unknown' as any })).toThrow('未知匹配器: unknown');
  });
});

// ========== Schema 校验测试 ==========

describe('validateSuite - 合法用例集', () => {
  it('最小合法用例集通过校验', () => {
    const data = {
      name: 'test-suite',
      cases: [
        {
          id: 'tc-001',
          name: 'Test Case',
          protocol: [2100, 2101],
          payload: {},
          expect: { businessCode: 10200 },
        },
      ],
    };
    const result = validateSuite(data);
    expect(result.valid).toBe(true);
    expect(result.errors).toBeUndefined();
  });

  it('完整用例集（含 defaults + responseFields）通过校验', () => {
    const data = {
      name: 'full-suite',
      description: 'Full test suite',
      defaults: { env: 'test', user: 'u1' },
      cases: [
        {
          id: 'tc-001',
          name: 'HelloWorld',
          protocol: [2100, 2101],
          payload: { name: 'Trae' },
          expect: {
            businessCode: 10200,
            responseFields: [{ field: 'reply', matcher: 'exists' }],
          },
        },
      ],
    };
    const result = validateSuite(data);
    expect(result.valid).toBe(true);
  });
});

describe('validateSuite - 非法用例集', () => {
  it('缺少 name 字段校验失败', () => {
    const result = validateSuite({ cases: [] });
    expect(result.valid).toBe(false);
    expect(result.errors).toBeDefined();
    expect(result.errors!.some((e) => e.includes('name'))).toBe(true);
  });

  it('protocol 不是两元素数组校验失败', () => {
    const data = {
      name: 'bad-proto',
      cases: [
        {
          id: 'tc-001',
          name: 'BadProto',
          protocol: [2100],
          payload: {},
          expect: { businessCode: 10200 },
        },
      ],
    };
    const result = validateSuite(data);
    expect(result.valid).toBe(false);
  });

  it('无效 matcher 值校验失败', () => {
    const data = {
      name: 'bad-matcher',
      cases: [
        {
          id: 'tc-001',
          name: 'BadMatcher',
          protocol: [2100, 2101],
          payload: {},
          expect: {
            businessCode: 10200,
            responseFields: [{ field: 'f', matcher: 'invalid_matcher' }],
          },
        },
      ],
    };
    const result = validateSuite(data);
    expect(result.valid).toBe(false);
  });

  it('空用例集也通过校验（cases 可以为空数组）', () => {
    const data = { name: 'empty', cases: [] };
    const result = validateSuite(data);
    expect(result.valid).toBe(true);
  });
});

// ========== Scenario Schema 校验测试 ==========

describe('validateScenario - 合法场景', () => {
  it('最小合法场景通过校验', () => {
    const data = {
      name: 'smoke-test',
      steps: [{ action: 'open' }],
    };
    const result = validateScenario(data);
    expect(result.valid).toBe(true);
  });

  it('含多个步骤的场景通过校验', () => {
    const data = {
      name: 'multi-step',
      steps: [
        { action: 'open', path: '/' },
        { action: 'fill-form', field: 'name', value: 'Trae' },
        { action: 'send' },
      ],
    };
    const result = validateScenario(data);
    expect(result.valid).toBe(true);
  });
});

describe('validateScenario - 非法场景', () => {
  it('缺少 steps 字段校验失败', () => {
    const result = validateScenario({ name: 'no-steps' });
    expect(result.valid).toBe(false);
  });

  it('步骤缺少 action 校验失败', () => {
    const data = { name: 'no-action', steps: [{ path: '/' }] };
    const result = validateScenario(data);
    expect(result.valid).toBe(false);
  });
});
