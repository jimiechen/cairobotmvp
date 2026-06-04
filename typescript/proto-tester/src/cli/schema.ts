/**
 * schema.ts - YAML 用例集 / 场景 Schema 定义 + 匹配器引擎
 *
 * 职责：
 * - 定义 Suite（用例集）和 Scenario（场景）的 JSON Schema
 * - 提供 validateSuite() / validateScenario() 校验函数
 * - 提供 7 种匹配器：exact / regex / contains / minLength / maxLength / exists / jsonPath
 * - 校验失败时返回详细错误信息列表
 *
 * 不负责：
 * - YAML 解析（由 yaml 库负责）
 * - 用例执行（由 runner.ts 负责）
 */
import Ajv from 'ajv';
import type { JSONSchemaType } from 'ajv/dist/types';

// ========== 类型定义 ==========

/** 匹配器类型枚举 */
export type MatcherType = 'exact' | 'regex' | 'contains' | 'minLength' | 'maxLength' | 'exists' | 'jsonPath';

/** 字段级断言 */
export interface FieldAssertion {
  field: string;
  matcher: MatcherType;
  value?: unknown;
}

/** 用例期望结果 */
export interface CaseExpect {
  businessCode: number;
  responseFields?: FieldAssertion[];
  errorMessage?: {
    matcher: string;
    value?: unknown;
  };
}

/** 单条测试用例 */
export interface TestCase {
  id: string;
  name: string;
  protocol: [number, number];
  payload: Record<string, unknown>;
  expect: CaseExpect;
}

/** 用例集默认配置 */
export interface SuiteDefaults {
  env: string;
  user?: string;
  gatewayOverride?: string | null;
}

/** YAML 用例集（Suite）结构 */
export interface Suite {
  name: string;
  description?: string;
  defaults?: SuiteDefaults;
  cases: TestCase[];
}

/** 场景步骤定义 */
export interface ScenarioStep {
  action: string;
  [key: string]: unknown;
}

/** YAML 场景（Scenario）结构 */
export interface Scenario {
  name: string;
  description?: string;
  steps: ScenarioStep[];
}

// ========== Schema 定义 ==========

/** Suite JSON Schema — 用于校验 YAML 用例集文件 */
export const SuiteSchema: JSONSchemaType<Suite> = {
  type: 'object',
  properties: {
    name: { type: 'string' },
    description: { type: 'string' },
    defaults: {
      type: 'object',
      properties: {
        env: { type: 'string' },
        user: { type: 'string' },
        gatewayOverride: { type: ['string', 'null'] },
      },
      required: ['env'],
      additionalProperties: false,
    },
    cases: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          id: { type: 'string' },
          name: { type: 'string' },
          protocol: {
            type: 'array',
            items: { type: 'number' },
            minItems: 2,
            maxItems: 2,
          },
          payload: { type: 'object' },
          expect: {
            type: 'object',
            properties: {
              businessCode: { type: 'number' },
              responseFields: {
                type: 'array',
                items: {
                  type: 'object',
                  properties: {
                    field: { type: 'string' },
                    matcher: {
                      type: 'string',
                      enum: ['exact', 'regex', 'contains', 'minLength', 'maxLength', 'exists', 'jsonPath'],
                    },
                    value: {},
                  },
                  required: ['field', 'matcher'],
                  additionalProperties: false,
                },
              },
              errorMessage: {
                type: 'object',
                properties: {
                  matcher: { type: 'string' },
                  value: {},
                },
                additionalProperties: false,
              },
            },
            required: ['businessCode'],
            additionalProperties: false,
          },
        },
        required: ['id', 'name', 'protocol', 'payload', 'expect'],
        additionalProperties: false,
      },
    },
  },
  required: ['name', 'cases'],
  additionalProperties: false,
};

/** Scenario JSON Schema — 用于校验 YAML 场景文件 */
export const ScenarioSchema: JSONSchemaType<Scenario> = {
  type: 'object',
  properties: {
    name: { type: 'string' },
    description: { type: 'string' },
    steps: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          action: { type: 'string' },
        },
        required: ['action'],
        additionalProperties: true,
      },
    },
  },
  required: ['name', 'steps'],
  additionalProperties: false,
};

// ========== Ajv 实例与校验函数 ==========

const suiteAjv = new Ajv({ allErrors: true });
const scenarioAjv = new Ajv({ allErrors: true });

const validateSuiteFn = suiteAjv.compile(SuiteSchema);
const validateScenarioFn = scenarioAjv.compile(ScenarioSchema);

/**
 * 校验 YAML 用例集数据是否符合 Suite Schema
 *
 * @param data - 从 YAML 解析的未知数据
 * @returns 校验结果，valid=true 时 errors 为 undefined
 */
export function validateSuite(data: unknown): { valid: boolean; errors?: string[] } {
  if (validateSuiteFn(data)) {
    return { valid: true };
  }
  return {
    valid: false,
    errors: validateSuiteFn.errors?.map(
      (e) => `${e.instancePath ? e.instancePath + ' ' : ''}${e.message ?? '未知错误'}`,
    ),
  };
}

/**
 * 校验 YAML 场景数据是否符合 Scenario Schema
 *
 * @param data - 从 YAML 解析的未知数据
 * @returns 校验结果，valid=true 时 errors 为 undefined
 */
export function validateScenario(data: unknown): { valid: boolean; errors?: string[] } {
  if (validateScenarioFn(data)) {
    return { valid: true };
  }
  return {
    valid: false,
    errors: validateScenarioFn.errors?.map(
      (e) => `${e.instancePath ? e.instancePath + ' ' : ''}${e.message ?? '未知错误'}`,
    ),
  };
}

// ========== 匹配器引擎 ==========

/**
 * 根据匹配器类型对实际值进行断言
 *
 * 支持 7 种匹配器：
 * - exact: 精确相等
 * - regex: 正则表达式匹配
 * - contains: 子串包含
 * - minLength: 最小长度
 * - maxLength: 最大长度
 * - exists: 存在性（非 null/undefined）
 * - jsonPath: 点号路径存在性（简化版）
 *
 * @param actual - 实际响应值
 * @param assertion - 字段级断言定义
 * @returns 是否匹配
 */
export function match(actual: unknown, assertion: FieldAssertion): boolean {
  switch (assertion.matcher) {
    case 'exact':
      return actual === assertion.value;
    case 'regex':
      return new RegExp(String(assertion.value)).test(String(actual));
    case 'contains':
      return String(actual).includes(String(assertion.value));
    case 'minLength':
      return (actual != null ? String(actual).length : 0) >= Number(assertion.value);
    case 'maxLength':
      return (actual != null ? String(actual).length : 0) <= Number(assertion.value);
    case 'exists':
      return actual !== undefined && actual !== null;
    case 'jsonPath':
      return matchJsonPath(actual, String(assertion.value));
    default:
      throw new Error(`未知匹配器: ${assertion.matcher}`);
  }
}

/**
 * 简化版 JSONPath 存在性检查
 *
 * 仅支持 a.b.c 和 a[0].b 格式，不支持通配符和递归下降。
 *
 * @param obj - 待查询对象
 * @param pathStr - 点号分隔的路径字符串（支持 [n] 数组索引）
 * @returns 路径指向的值是否存在且非 null/undefined
 */
function matchJsonPath(obj: unknown, pathStr: string): boolean {
  // 将 a[0].b 格式转为 a.0.b
  const normalized = pathStr.replace(/\[(\d+)\]/g, '.$1');
  const parts = normalized.split('.').filter((p) => p !== '');
  let current: unknown = obj;

  for (const part of parts) {
    if (current == null) return false;
    if (typeof current === 'object') {
      current = (current as Record<string, unknown>)[part];
    } else {
      return false;
    }
  }

  return current !== undefined && current !== null;
}
