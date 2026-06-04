/**
 * protoFormBuilder.ts - messageSchemas 驱动的表单 Schema 构建器
 *
 * 职责：
 * - 将 MessageSchema 转换为 FormFieldDef[]，驱动前端表单渲染
 * - 处理 11 种 Proto 类型到表单类型的映射
 * - 处理嵌套消息、repeated、map、oneof 等复杂结构
 * - 嵌套深度超限时降级为 JSON 原始输入
 *
 * 数据来源约束：
 * - 唯一数据来源是 protocols.json.messageSchemas（构建期产物）
 * - 禁止运行时 import .proto 或用 toObject() 推断字段
 */

import type { MessageSchema, FieldSchema } from './protoMetadata';
import protocolsRaw from '@/data/protocols.json';

/** 表单字段定义，用于驱动前端表单组件渲染 */
export interface FormFieldDef {
  name: string;
  fieldName: string;
  fieldType:
    | 'string'
    | 'int32'
    | 'int64'
    | 'float'
    | 'double'
    | 'bool'
    | 'bytes'
    | 'enum'
    | 'repeated'
    | 'nested'
    | 'map'
    | 'oneof'
    | '_fallback_json';
  label: string;
  required: boolean;
  defaultValue?: any;
  options?: string[];
  oneofGroup?: string;
  nestedSchema?: MessageSchema;
  nestedDepth: number;
  children?: FormFieldDef[];
  mapKeyType?: string;
  mapValueType?: string;
  comment?: string;
}

/** 最大嵌套深度限制，超过后降级为原始 JSON 输入 */
const MAX_NESTED_DEPTH = 5;

/** 整数类型集合（映射为 int32 或 int64） */
const INTEGER_TYPES = new Set([
  'int32', 'int64', 'uint32', 'uint64',
  'sint32', 'sint64', 'fixed32', 'fixed64',
  'sfixed32', 'sfixed64',
]);

/** 浮点类型集合 */
const FLOAT_TYPES = new Set(['float', 'double']);

/**
 * 判断 Proto 类型是否为整数类型
 * int64 系列映射为 'int64'，其余整数映射为 'int32'
 */
function mapIntegerType(type: string): 'int32' | 'int64' {
  if (type === 'int64' || type === 'uint64' || type === 'sint64' || type === 'fixed64' || type === 'sfixed64') {
    return 'int64';
  }
  return 'int32';
}

/**
 * 将 Proto 字段 label + type 映射为 FormFieldDef.fieldType
 */
function mapFieldType(field: FieldSchema): FormFieldDef['fieldType'] {
  if (field.label === 'repeated') return 'repeated';
  if (field.label === 'map') return 'map';
  if (INTEGER_TYPES.has(field.type)) return mapIntegerType(field.type);
  if (FLOAT_TYPES.has(field.type)) return field.type as 'float' | 'double';
  if (field.type === 'bool') return 'bool';
  if (field.type === 'string') return 'string';
  if (field.type === 'bytes') return 'bytes';
  // 剩余视为嵌套消息引用
  return 'nested';
}

/**
 * 判断字段是否必填
 *
 * 三层优先级策略：
 * 1. field_meta.json 标注（预留）
 * 2. oneof 组内字段视为条件必选
 * 3. 默认 false（proto3 无 required 关键字）
 */
function isFieldRequired(field: FieldSchema): boolean {
  if (field.oneof) return true;
  return false;
}

/**
 * 构建 fallback 降级字段
 * 当嵌套深度超过 MAX_NESTED_DEPTH 时返回
 */
function buildFallbackField(): FormFieldDef {
  return {
    name: '_raw',
    fieldName: '_raw',
    fieldType: '_fallback_json',
    label: '原始 JSON（超限降级）',
    required: false,
    nestedDepth: MAX_NESTED_DEPTH + 1,
    comment: '嵌套深度超过限制，已降级为 JSON 编辑器',
  };
}

// ========== Schema 查找缓存 ==========

/** 全局 schema 查找缓存，避免重复遍历 */
const schemaCache = new Map<string, MessageSchema>();

/** 全部 messageSchemas 的扁平引用 */
const allSchemas: Record<string, MessageSchema> = protocolsRaw.messageSchemas;

/**
 * 按名称查找 message schema（带缓存）
 */
function findMessageSchema(typeName: string, parentSchema: MessageSchema): MessageSchema | undefined {
  // 先在父 schema 的 nestedTypes 中按短名匹配
  const nestedMatch = parentSchema.nestedTypes.find(
    (n) => n === typeName || n.endsWith(`.${typeName}`),
  );
  if (nestedMatch && allSchemas[nestedMatch]) {
    const cached = schemaCache.get(nestedMatch);
    if (cached) return cached;
    schemaCache.set(nestedMatch, allSchemas[nestedMatch]);
    return allSchemas[nestedMatch];
  }

  // 再在全局 schemas 中搜索
  for (const [fullName, schema] of Object.entries(allSchemas)) {
    if (fullName === typeName || fullName.endsWith(`.${typeName}`)) {
      const cached = schemaCache.get(fullName);
      if (cached) return cached;
      schemaCache.set(fullName, schema);
      return schema;
    }
  }
  return undefined;
}

/**
 * 将 MessageSchema 转换为表单字段定义数组
 *
 * @param schema - 来自 protocols.json 的 MessageSchema
 * @param depth - 当前嵌套深度（从 0 开始）
 * @returns 表单字段定义数组
 */
export function buildFormSchema(schema: MessageSchema, depth = 0): FormFieldDef[] {
  if (depth > MAX_NESTED_DEPTH) {
    return [buildFallbackField()];
  }

  const fields: FormFieldDef[] = [];

  for (const field of schema.fields) {
    const fieldType = mapFieldType(field);
    const baseField: FormFieldDef = {
      name: field.name,
      fieldName: field.name,
      fieldType,
      label: field.name,
      required: isFieldRequired(field),
      nestedDepth: depth + 1,
      comment: field.comment || undefined,
    };

    // oneof 分组标记
    if (field.oneof) {
      baseField.oneofGroup = field.oneof;
    }

    // 根据 fieldType 补充额外信息
    switch (fieldType) {
      case 'nested': {
        const nestedSchema = findMessageSchema(field.type, schema);
        if (nestedSchema && depth < MAX_NESTED_DEPTH) {
          baseField.nestedSchema = nestedSchema;
          baseField.children = buildFormSchema(nestedSchema, depth + 1);
        } else if (depth >= MAX_NESTED_DEPTH) {
          baseField.children = [buildFallbackField()];
        }
        break;
      }

      case 'repeated': {
        // repeated 元素可能是嵌套消息，尝试展开
        const elementSchema = findMessageSchema(field.type, schema);
        if (elementSchema && depth < MAX_NESTED_DEPTH) {
          baseField.nestedSchema = elementSchema;
          baseField.children = buildFormSchema(elementSchema, depth + 1);
        }
        break;
      }

      case 'map': {
        // protobufjs 对 map 字段的 type 存储的是 value 类型
        // key 默认为 string（大多数场景），value 取自 field.type
        baseField.mapKeyType = 'string';
        baseField.mapValueType = field.type;
        break;
      }

      case 'enum': {
        // 从当前 schema 的 enums 中查找对应枚举值列表
        const enumDef = schema.enums.find((e) => e.name === field.type);
        if (enumDef) {
          baseField.options = Object.keys(enumDef.values);
        }
        break;
      }
    }

    fields.push(baseField);
  }

  return fields;
}
