/**
 * T1 协议元数据与表单构建器单元测试
 *
 * 测试范围：
 * 1. protoMetadata.ts 对 protocols.json 的解析正确性
 * 2. protoFormBuilder.ts 的表单 Schema 构建逻辑
 */
import { describe, it, expect } from 'vitest';
import {
  getAllProtocols,
  getProtocolByMaxMin,
  getMessageSchema,
  searchProtocols,
} from '../../src/lib/protoMetadata';
import { buildFormSchema } from '../../src/lib/protoFormBuilder';

describe('T1 protoMetadata 协议元数据查询', () => {
  it('getAllProtocols 应返回非空协议列表', () => {
    const protocols = getAllProtocols();
    expect(protocols.length).toBeGreaterThan(0);
    // 协议编号注册表有 14 条记录
    expect(protocols.length).toBe(14);
  });

  it('getProtocolByMaxMin 按 maxType+minType 精确查找', () => {
    const protocol = getProtocolByMaxMin(2100, 2101);
    expect(protocol).toBeDefined();
    expect(protocol!.name).toBe('HelloWorldRequest');
    expect(protocol!.direction).toBe('C->S');
  });

  it('getProtocolByMaxMin 未找到时返回 undefined', () => {
    const protocol = getProtocolByMaxMin(9999, 9999);
    expect(protocol).toBeUndefined();
  });

  it('getMessageSchema 返回 HelloWorldRequest 的完整字段', () => {
    const schema = getMessageSchema('HelloWorldRequest');
    expect(schema).toBeDefined();
    expect(schema!.fields.length).toBe(2);
    expect(schema!.fields[0].name).toBe('name');
    expect(schema!.fields[0].type).toBe('string');
    expect(schema!.fields[1].name).toBe('lang_code');
    expect(schema!.fields[1].type).toBe('string');
  });

  it('getMessageSchema 返回 Result 的字段', () => {
    const schema = getMessageSchema('Result');
    expect(schema).toBeDefined();
    expect(schema!.fields.length).toBe(2);
    expect(schema!.fields.map((f) => f.name)).toEqual(['code', 'message']);
  });

  it('searchProtocols 按名称模糊搜索', () => {
    const results = searchProtocols('hello');
    expect(results.length).toBeGreaterThan(0);
    expect(results.some((p) => p.name.includes('Hello'))).toBe(true);
  });

  it('searchProtocols 按描述模糊搜索', () => {
    const results = searchProtocols('健康');
    expect(results.length).toBeGreaterThan(0);
  });

  it('searchProtocols 无匹配时返回空数组', () => {
    const results = searchProtocols('zzz_not_exist_xxx');
    expect(results).toEqual([]);
  });
});

describe('T1 protoFormBuilder 表单 Schema 构建', () => {
  it('HelloWorldRequest → 2 个 string 字段', () => {
    const schema = getMessageSchema('HelloWorldRequest')!;
    const formFields = buildFormSchema(schema);
    expect(formFields.length).toBe(2);
    expect(formFields[0].fieldName).toBe('name');
    expect(formFields[0].fieldType).toBe('string');
    expect(formFields[1].fieldName).toBe('lang_code');
    expect(formFields[1].fieldType).toBe('string');
  });

  it('AppConfigsReq → env/client_scope/client_version/request_modules', () => {
    const schema = getMessageSchema('AppConfigsReq')!;
    const formFields = buildFormSchema(schema);
    expect(formFields.length).toBe(4);

    const fieldNames = formFields.map((f) => f.fieldName);
    expect(fieldNames).toContain('env');
    expect(fieldNames).toContain('client_scope');
    expect(fieldNames).toContain('client_version');
    expect(fieldNames).toContain('requested_modules');

    // requested_modules 是 repeated string
    const repeatedField = formFields.find((f) => f.fieldName === 'requested_modules')!;
    expect(repeatedField.fieldType).toBe('repeated');
  });

  it('AppConfigsRsp → result(嵌套) + 8 个强类型模块 + dynamic_modules(repeated)', () => {
    const schema = getMessageSchema('AppConfigsRsp')!;
    const formFields = buildFormSchema(schema);
    // result + 8 强类型模块 + dynamic_modules = 10
    expect(formFields.length).toBe(10);

    const fieldNames = formFields.map((f) => f.fieldName);
    expect(fieldNames).toContain('result');
    expect(fieldNames).toContain('base_cfg');
    expect(fieldNames).toContain('dynamic_modules');

    // result 是嵌套消息
    const resultField = formFields.find((f) => f.fieldName === 'result')!;
    expect(resultField.fieldType).toBe('nested');
    expect(resultField.nestedSchema).toBeDefined();
    expect(resultField.nestedDepth).toBe(1);

    // dynamic_modules 是 repeated nested
    const dynField = formFields.find((f) => f.fieldName === 'dynamic_modules')!;
    expect(dynField.fieldType).toBe('repeated');
  });

  it('DynamicConfigModule → module_key + version + fields(map) + descriptors(repeated)', () => {
    const schema = getMessageSchema('DynamicConfigModule')!;
    const formFields = buildFormSchema(schema);
    expect(formFields.length).toBe(4);

    const fieldNames = formFields.map((f) => f.fieldName);
    expect(fieldNames).toEqual(['module_key', 'version', 'fields', 'descriptors']);

    // fields 是 map
    const mapField = formFields.find((f) => f.fieldName === 'fields')!;
    expect(mapField.fieldType).toBe('map');
    expect(mapField.mapKeyType).toBe('string');
    expect(mapField.mapValueType).toBe('string');

    // descriptors 是 repeated nested
    const descField = formFields.find((f) => f.fieldName === 'descriptors')!;
    expect(descField.fieldType).toBe('repeated');
  });

  it('FieldDescriptor → 3 个 string + 1 个 bool', () => {
    const schema = getMessageSchema('FieldDescriptor')!;
    const formFields = buildFormSchema(schema);
    expect(formFields.length).toBe(4);

    const boolField = formFields.find((f) => f.fieldName === 'is_required')!;
    expect(boolField.fieldType).toBe('bool');
  });

  it('嵌套深度 > 5 时返回 _fallback_json', () => {
    // 构造一个深度 > 5 的嵌套 schema 来测试降级逻辑
    const deepSchema = {
      fullName: 'test.DeepL5',
      protoFile: 'test.proto',
      fields: [
        { name: 'child', number: 1, type: 'DeepL6', label: 'optional', oneof: null },
      ],
      nestedTypes: ['DeepL6'],
      enums: [],
    };
    // 用 depth=5 调用，内部再嵌套一层就超过 5
    const formFields = buildFormSchema(deepSchema as any, 5);
    // 深度为 5 时，子字段应为 fallback
    expect(formFields.length).toBe(1);
    expect(formFields[0].fieldType).toBe('nested');
    // 验证 children 中包含 _fallback_json
    if (formFields[0].children) {
      const fallback = formFields[0].children.find((c) => c.fieldType === '_fallback_json');
      expect(fallback).toBeDefined();
      expect(fallback!.label).toContain('超限降级');
    }
  });

  it('PayMethod 包含 double 类型字段', () => {
    const schema = getMessageSchema('PayMethod')!;
    const formFields = buildFormSchema(schema);
    expect(formFields.length).toBe(4);

    const rangMin = formFields.find((f) => f.fieldName === 'rang_min')!;
    expect(rangMin.fieldType).toBe('double');
  });

  it('LanguageMeta 包含 bool 字段 is_default', () => {
    const schema = getMessageSchema('LanguageMeta')!;
    const formFields = buildFormSchema(schema);
    const isDefault = formFields.find((f) => f.fieldName === 'is_default')!;
    expect(isDefault.fieldType).toBe('bool');
  });

  it('MuteDuration 包含 int32 字段 seconds', () => {
    const schema = getMessageSchema('MuteDuration')!;
    const formFields = buildFormSchema(schema);
    const seconds = formFields.find((f) => f.fieldName === 'seconds')!;
    expect(seconds.fieldType).toBe('int32');
  });

  it('AppConfigVersionReq 包含 map 字段', () => {
    const schema = getMessageSchema('AppConfigVersionReq')!;
    const formFields = buildFormSchema(schema);
    expect(formFields.length).toBe(3);

    const knownVersions = formFields.find((f) => f.fieldName === 'known_versions')!;
    expect(knownVersions.fieldType).toBe('map');
    expect(knownVersions.mapKeyType).toBe('string');
    expect(knownVersions.mapValueType).toBe('int64');
  });

  it('AppConfigVersionRsp 包含 has_changes bool 和 map 字段', () => {
    const schema = getMessageSchema('AppConfigVersionRsp')!;
    const formFields = buildFormSchema(schema);
    expect(formFields.length).toBe(4);

    const hasChanges = formFields.find((f) => f.fieldName === 'has_changes')!;
    expect(hasChanges.fieldType).toBe('bool');
  });
});
