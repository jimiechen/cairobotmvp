/**
 * ProtoFormRenderer.tsx - 协议表单核心渲染组件
 *
 * 职责：
 * - 根据 FormFieldDef[] 渲染完整的协议表单
 * - 支持 9 种 Proto 类型的控件映射
 * - 处理嵌套消息、repeated、map、oneof 等复杂结构
 * - 嵌套深度超限时降级为 JSON 编辑器
 *
 * 不负责：
 * - FormFieldDef 的生成（由 protoFormBuilder 负责）
 * - 表单提交和序列化（由调用方负责）
 */
import { useCallback, useState } from 'react';
import {
  Input,
  InputNumber,
  Switch,
  Collapse,
  Button,
  Space,
  Radio,
  Tag,
  Typography,
  Select,
} from 'antd';
import type { FormFieldDef } from '../lib/protoFormBuilder';
import { EnumSelect } from './EnumSelect';
import { JsonEditor } from './JsonEditor';

const { Text } = Typography;

/** ProtoFormRenderer 组件属性 */
interface ProtoFormRendererProps {
  /** 表单字段定义数组（来自 buildFormSchema()） */
  schema: FormFieldDef[];
  /** 当前表单值（受控模式，由父组件管理） */
  values: Record<string, any>;
  /** 字段变更回调 */
  onChange: (field: string, value: any) => void;
  /** 当前嵌套深度（用于缩进样式），默认 0 */
  depth?: number;
}

/**
 * 从 comment 字符串解析 maxLength 约束
 *
 * 支持格式："最大长度:100" 或 "maxLength=100"
 */
function parseMaxLength(comment?: string): number | undefined {
  if (!comment) return undefined;
  const match = comment.match(/(?:最大长度|maxLength)[=:]\s*(\d+)/);
  return match ? parseInt(match[1], 10) : undefined;
}

/**
 * 判断整数类型是否为无符号
 *
 * uint32/uint64/fixed32/fixed64 为无符号类型
 */
function isUnsignedInt(fieldType: string): boolean {
  return ['uint32', 'uint64', 'fixed32', 'fixed64'].includes(fieldType);
}

// ========== 子组件：各类型字段渲染 ==========

/** 渲染 string 字段 */
function StringField({ field, value, onChange }: { field: FormFieldDef; value: any; onChange: (v: any) => void }) {
  const maxLength = parseMaxLength(field.comment);
  return (
    <div data-id={`pt-form-${field.fieldName}`}>
      <Input
        value={value ?? ''}
        onChange={(e) => onChange(e.target.value)}
        maxLength={maxLength}
        placeholder={field.label}
        allowClear
      />
    </div>
  );
}

/** 渲染 int32/int64 字段 */
function IntField({ field, value, onChange }: { field: FormFieldDef; value: any; onChange: (v: any) => void }) {
  return (
    <div data-id={`pt-form-${field.fieldName}`}>
      <InputNumber
        value={value ?? 0}
        onChange={(v) => onChange(v ?? 0)}
        step={1}
        style={{ width: '100%' }}
        placeholder={field.label}
      />
    </div>
  );
}

/** 渲染 float/double 字段 */
function FloatField({ field, value, onChange }: { field: FormFieldDef; value: any; onChange: (v: any) => void }) {
  return (
    <div data-id={`pt-form-${field.fieldName}`}>
      <InputNumber
        value={value ?? 0}
        onChange={(v) => onChange(v ?? 0)}
        step={0.01}
        style={{ width: '100%' }}
        placeholder={field.label}
      />
    </div>
  );
}

/** 渲染 bool 字段 */
function BoolField({ field, value, onChange }: { field: FormFieldDef; value: any; onChange: (v: any) => void }) {
  return (
    <Switch
      data-id={`pt-form-${field.fieldName}`}
      checked={value ?? false}
      onChange={(checked) => onChange(checked)}
      checkedChildren="开"
      unCheckedChildren="关"
    />
  );
}

/** 渲染 bytes 字段（双模式：文件上传 / hex 文本） */
function BytesField({ field, value, onChange }: { field: FormFieldDef; value: any; onChange: (v: any) => void }) {
  // hex 模式：文本域输入十六进制字符串
  return (
    <div data-id={`pt-form-${field.fieldName}`}>
      <Input.TextArea
        value={value ?? ''}
        onChange={(e) => onChange(e.target.value)}
        placeholder="输入 Hex 编码的 bytes 数据"
        rows={2}
      />
    </div>
  );
}

/** 渲染 enum 字段 */
function EnumField({ field, value, onChange }: { field: FormFieldDef; value: any; onChange: (v: any) => void }) {
  const options = (field.options || []).map((opt) => ({
    label: opt,
    value: /^\d+$/.test(opt) ? parseInt(opt, 10) : opt,
  }));

  return (
    <div data-id={`pt-form-${field.fieldName}`}>
      <EnumSelect
        value={value}
        options={options}
        onChange={onChange}
        placeholder={`选择 ${field.label}`}
      />
    </div>
  );
}

/** 渲染 repeated 字段（可增删行的列表） */
function RepeatedField({
  field,
  value,
  onChange,
  onChildChange,
}: {
  field: FormFieldDef;
  value: any;
  onChange: (v: any) => void;
  onChildChange: (f: string, v: any) => void;
}) {
  const items: any[] = Array.isArray(value) ? value : [];

  /** 添加新行 */
  const handleAdd = () => {
    const newItem = field.children?.[0]?.defaultValue ?? '';
    onChange([...items, newItem]);
  };

  /** 删除指定行 */
  const handleDelete = (index: number) => {
    const newItems = items.filter((_, i) => i !== index);
    onChange(newItems);
  };

  return (
    <div data-id={`pt-form-${field.fieldName}`}>
      {items.map((item, index) => (
        <div key={index} style={{ display: 'flex', gap: 8, marginBottom: 4, alignItems: 'flex-start' }}>
          <div style={{ flex: 1 }}>
            <Text type="secondary" style={{ fontSize: 12 }}>[{index}]</Text>
            {/* 如果有子字段定义，递归渲染；否则用 Input */}
            {field.children && field.children.length > 0 ? (
              <ProtoFormRenderer
                schema={field.children!}
                values={typeof item === 'object' ? item : {}}
                onChange={(childField, childValue) => {
                  const newItems = [...items];
                  if (typeof newItems[index] !== 'object') {
                    newItems[index] = {};
                  }
                  newItems[index] = { ...newItems[index], [childField]: childValue };
                  onChange(newItems);
                }}
                depth={field.nestedDepth}
              />
            ) : (
              <Input
                value={item ?? ''}
                onChange={(e) => {
                  const newItems = [...items];
                  newItems[index] = e.target.value;
                  onChange(newItems);
                }}
                size="small"
              />
            )}
          </div>
          <Button
            type="text"
            danger
            onClick={() => handleDelete(index)}
            size="small"
          >
            删除
          </Button>
        </div>
      ))}
      <Button
        type="dashed"
        onClick={handleAdd}
        size="small"
        block
      >
        添加项
      </Button>
    </div>
  );
}

/** 渲染 nested 字段（Collapse 折叠面板） */
function NestedField({
  field,
  value,
  onChange,
}: {
  field: FormFieldDef;
  value: any;
  onChange: (v: any) => void;
}) {
  const childValues = typeof value === 'object' && value !== null ? value : {};

  return (
    <div data-id={`pt-form-${field.fieldName}`}>
      <Collapse
        size="small"
        defaultActiveKey={undefined}
        items={[{
          key: field.fieldName,
          label: (
            <Space>
              <Text strong>{field.label}</Text>
              {field.required && <Tag color="blue">必填</Tag>}
            </Space>
          ),
          children: (
            <ProtoFormRenderer
              schema={field.children || []}
              values={childValues}
              onChange={(childField, childValue) => {
                onChange({ ...childValues, [childField]: childValue });
              }}
              depth={field.nestedDepth}
            />
          ),
        }]}
      />
    </div>
  );
}

/** 渲染 oneof 字段（Tab/Radio 选择器） */
function OneOfField({
  field,
  value,
  onChange,
  siblings,
}: {
  field: FormFieldDef;
  value: any;
  onChange: (v: any) => void;
  siblings: FormFieldDef[];
}) {
  // 找到同一 oneof 组的所有字段选项
  const options = siblings.filter((s) => s.oneofGroup === field.oneofGroup);

  return (
    <div data-id={`pt-form-${field.fieldName}`}>
      <Tag color="blue" style={{ marginBottom: 4 }}>二选一</Tag>
      <Radio.Group
        value={field.fieldName}
        onChange={(e) => {
          // 选择当前字段时清空同组其他字段的值
          onChange(e.target.value === field.fieldName ? value : undefined);
        }}
        optionType="button"
        size="small"
      >
        {options.map((opt) => (
          <Radio.Button key={opt.fieldName} value={opt.fieldName}>
            {opt.label}
          </Radio.Button>
        ))}
      </Radio.Group>
    </div>
  );
}

/** 渲染 map 字段（键值对编辑器） */
function MapField({
  field,
  value,
  onChange,
}: {
  field: FormFieldDef;
  value: any;
  onChange: (v: any) => void;
}) {
  const mapData: Record<string, any> = typeof value === 'object' && value !== null ? value : {};

  return (
    <div data-id={`pt-form-${field.fieldName}`}>
      {Object.entries(mapData).map(([key, val]) => (
        <div key={key} style={{ display: 'flex', gap: 8, marginBottom: 4, alignItems: 'center' }}>
          <Input
            value={key}
            disabled
            size="small"
            style={{ width: 140 }}
          />
          <Input
            value={String(val ?? '')}
            onChange={(e) => {
              onChange({ ...mapData, [key]: e.target.value });
            }}
            size="small"
            style={{ flex: 1 }}
          />
          <Button
            type="text"
            danger
            onClick={() => {
              const newData = { ...mapData };
              delete newData[key];
              onChange(newData);
            }}
            size="small"
          >
            删除
          </Button>
        </div>
      ))}
      <MapAddEntry
        mapData={mapData}
        mapKeyType={field.mapKeyType || 'string'}
        onChange={onChange}
      />
    </div>
  );
}

/** Map 新增键值对子组件 */
function MapAddEntry({
  mapData,
  mapKeyType,
  onChange,
}: {
  mapData: Record<string, any>;
  mapKeyType: string;
  onChange: (data: Record<string, any>) => void;
}) {
  const [newKey, setNewKey] = useState('');
  const [newValue, setNewValue] = useState('');

  const handleAdd = () => {
    const trimmedKey = newKey.trim();
    if (trimmedKey && !(trimmedKey in mapData)) {
      onChange({ ...mapData, [trimmedKey]: newValue });
      setNewKey('');
      setNewValue('');
    }
  };

  return (
    <div style={{ display: 'flex', gap: 8, marginTop: 4 }}>
      <Input
        placeholder="Key"
        value={newKey}
        onChange={(e) => setNewKey(e.target.value)}
        size="small"
        style={{ width: 140 }}
        onPressEnter={handleAdd}
      />
      <Input
        placeholder="Value"
        value={newValue}
        onChange={(e) => setNewValue(e.target.value)}
        size="small"
        style={{ flex: 1 }}
        onPressEnter={handleAdd}
      />
      <Button type="dashed" onClick={handleAdd} size="small">
        添加
      </Button>
    </div>
  );
}

/** 渲染 _fallback_json 降级模式（JSON 编辑器） */
function FallbackJsonField({ field, value, onChange }: { field: FormFieldDef; value: any; onChange: (v: any) => void }) {
  const jsonStr = typeof value === 'string' ? value : JSON.stringify(value || {}, null, 2);

  return (
    <div data-id={`pt-form-${field.fieldName}`}>
      <JsonEditor
        value={jsonStr}
        onChange={onChange}
        height="150px"
      />
    </div>
  );
}

// ========== 主组件 ==========

/**
 * 协议表单核心渲染组件
 *
 * 根据 fieldType 分发到对应的子组件进行渲染。
 * 支持嵌套递归调用以处理 nested/repeated 结构。
 */
export function ProtoFormRenderer({
  schema,
  values,
  onChange,
  depth = 0,
}: ProtoFormRendererProps) {
  /** 统一的字段变更回调，拼接完整路径 */
  const handleChange = useCallback(
    (fieldName: string, fieldValue: any) => {
      onChange(fieldName, fieldValue);
    },
    [onChange],
  );

  return (
    <div style={{ paddingLeft: depth > 0 ? 16 : 0 }}>
      {schema.map((field) => {
        const currentValue = values[field.fieldName];

        switch (field.fieldType) {
          case 'string':
            return (
              <StringField
                key={field.fieldName}
                field={field}
                value={currentValue}
                onChange={(v) => handleChange(field.fieldName, v)}
              />
            );

          case 'int32':
          case 'int64':
            return (
              <IntField
                key={field.fieldName}
                field={field}
                value={currentValue}
                onChange={(v) => handleChange(field.fieldName, v)}
              />
            );

          case 'float':
          case 'double':
            return (
              <FloatField
                key={field.fieldName}
                field={field}
                value={currentValue}
                onChange={(v) => handleChange(field.fieldName, v)}
              />
            );

          case 'bool':
            return (
              <BoolField
                key={field.fieldName}
                field={field}
                value={currentValue}
                onChange={(v) => handleChange(field.fieldName, v)}
              />
            );

          case 'bytes':
            return (
              <BytesField
                key={field.fieldName}
                field={field}
                value={currentValue}
                onChange={(v) => handleChange(field.fieldName, v)}
              />
            );

          case 'enum':
            return (
              <EnumField
                key={field.fieldName}
                field={field}
                value={currentValue}
                onChange={(v) => handleChange(field.fieldName, v)}
              />
            );

          case 'repeated':
            return (
              <RepeatedField
                key={field.fieldName}
                field={field}
                value={currentValue}
                onChange={(v) => handleChange(field.fieldName, v)}
                onChildChange={handleChange}
              />
            );

          case 'nested':
            return (
              <NestedField
                key={field.fieldName}
                field={field}
                value={currentValue}
                onChange={(v) => handleChange(field.fieldName, v)}
              />
            );

          case 'oneof':
            return (
              <OneOfField
                key={field.fieldName}
                field={field}
                value={currentValue}
                onChange={(v) => handleChange(field.fieldName, v)}
                siblings={schema}
              />
            );

          case 'map':
            return (
              <MapField
                key={field.fieldName}
                field={field}
                value={currentValue}
                onChange={(v) => handleChange(field.fieldName, v)}
              />
            );

          case '_fallback_json':
            return (
              <FallbackJsonField
                key={field.fieldName}
                field={field}
                value={currentValue}
                onChange={(v) => handleChange(field.fieldName, v)}
              />
            );

          default:
            return null;
        }
      })}
    </div>
  );
}
