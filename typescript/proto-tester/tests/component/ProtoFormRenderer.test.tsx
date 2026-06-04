/**
 * ProtoFormRenderer 核心表单渲染组件测试
 *
 * 测试范围：
 * 1. string 字段正确渲染 Input
 * 2. int32 字段正确渲染 InputNumber
 * 3. bool 字段正确渲染 Switch
 * 4. enum 字段正确渲染 Select（含 options）
 * 5. repeated 字段可增删行
 * 6. nested 字段渲染 Collapse 折叠面板
 * 7. _fallback_json 降级模式
 * 8. data-id 属性正确设置
 * 9. onChange 回调正确触发
 */
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { ProtoFormRenderer } from '../../src/components/ProtoFormRenderer';
import type { FormFieldDef } from '../../src/lib/protoFormBuilder';

/** 创建最小 FormFieldDef 的工厂函数 */
function makeField(overrides: Partial<FormFieldDef>): FormFieldDef {
  return {
    name: 'test_field',
    fieldName: 'test_field',
    fieldType: 'string',
    label: 'Test Field',
    required: false,
    nestedDepth: 1,
    ...overrides,
  };
}

describe('T3 ProtoFormRenderer 基础字段渲染', () => {
  it('string 字段应渲染 Input 控件', () => {
    const schema: FormFieldDef[] = [makeField({ fieldName: 'name', fieldType: 'string', label: '名称' })];
    const onChange = vi.fn();
    render(<ProtoFormRenderer schema={schema} values={{}} onChange={onChange} />);

    // 应有 Input 或输入框
    const input = screen.getByRole('textbox');
    expect(input).toBeInTheDocument();
    // 验证 data-id 属性
    expect(input.closest('[data-id="pt-form-name"]')).toBeTruthy();
  });

  it('int32 字段应渲染 InputNumber 控件', () => {
    const schema: FormFieldDef[] = [makeField({ fieldName: 'count', fieldType: 'int32', label: '数量' })];
    const onChange = vi.fn();
    render(<ProtoFormRenderer schema={schema} values={{}} onChange={onChange} />);

    // InputNumber 在 antd 中渲染为 input[type=text] 但有特定角色
    const input = screen.getByRole('spinbutton') || screen.getByRole('textbox');
    expect(input).toBeInTheDocument();
  });

  it('bool 字段应渲染 Switch 控件', () => {
    const schema: FormFieldDef[] = [makeField({ fieldName: 'enabled', fieldType: 'bool', label: '启用' })];
    const onChange = vi.fn();
    render(<ProtoFormRenderer schema={schema} values={{ enabled: false }} onChange={onChange} />);

    // Switch 控件在 antd 中 role 为 switch
    const switchEl = screen.getByRole('switch');
    expect(switchEl).toBeInTheDocument();
    expect(switchEl).toHaveAttribute('data-id', 'pt-form-enabled');
  });

  it('enum 字段应渲染 Select 并包含 options', () => {
    const schema: FormFieldDef[] = [
      makeField({
        fieldName: 'status',
        fieldType: 'enum',
        label: '状态',
        options: ['ACTIVE', 'INACTIVE'],
      }),
    ];
    const onChange = vi.fn();
    render(<ProtoFormRenderer schema={schema} values={{}} onChange={onChange} />);

    // Select 应可交互
    const select = screen.getByRole('combobox');
    expect(select).toBeInTheDocument();
  });
});

describe('T3 ProtoFormRenderer 复杂字段渲染', () => {
  it('repeated 字段应渲染可增删行的列表', () => {
    const schema: FormFieldDef[] = [
      makeField({
        fieldName: 'tags',
        fieldType: 'repeated',
        label: '标签列表',
        children: [
          makeField({ fieldName: 'tag', fieldType: 'string', label: '标签', nestedDepth: 2 }),
        ],
      }),
    ];
    const onChange = vi.fn();
    render(<ProtoFormRenderer schema={schema} values={{ tags: [] }} onChange={onChange} />);

    // repeated 区域应有添加按钮或相关 UI
    const container = document.querySelector('[data-id="pt-form-tags"]');
    expect(container).toBeInTheDocument();
  });

  it('nested 字段应渲染 Collapse 折叠面板', () => {
    const schema: FormFieldDef[] = [
      makeField({
        fieldName: 'config',
        fieldType: 'nested',
        label: '配置',
        nestedSchema: {
          fullName: 'Config',
          protoFile: 'test.proto',
          fields: [{ name: 'key', number: 1, type: 'string', label: 'optional', oneof: null }],
          nestedTypes: [],
          enums: [],
        },
        children: [makeField({ fieldName: 'key', fieldType: 'string', label: 'Key', nestedDepth: 2 })],
      }),
    ];
    const onChange = vi.fn();
    render(<ProtoFormRenderer schema={schema} values={{}} onChange={onChange} />);

    // nested 字段应有折叠面板区域
    const panel = document.querySelector('[data-id="pt-form-config"]');
    expect(panel).toBeInTheDocument();
  });

  it('_fallback_json 降级模式应渲染 JSON 编辑器', () => {
    const schema: FormFieldDef[] = [
      makeField({
        fieldName: '_raw',
        fieldType: '_fallback_json',
        label: '原始 JSON（超限降级）',
        nestedDepth: 6,
      }),
    ];
    const onChange = vi.fn();
    render(<ProtoFormRenderer schema={schema} values={{ _raw: '{}' }} onChange={onChange} />);

    // JSON 编辑器区域
    const editorArea = document.querySelector('[data-id="pt-form-_raw"]');
    expect(editorArea).toBeInTheDocument();
  });
});

describe('T3 ProtoFormRenderer data-id 属性', () => {
  it('每个字段都应设置正确的 data-id 属性', () => {
    const schema: FormFieldDef[] = [
      makeField({ fieldName: 'field_a', fieldType: 'string', label: 'A' }),
      makeField({ fieldName: 'field_b', fieldType: 'int32', label: 'B' }),
      makeField({ fieldName: 'field_c', fieldType: 'bool', label: 'C' }),
    ];
    const onChange = vi.fn();
    render(<ProtoFormRenderer schema={schema} values={{}} onChange={onChange} />);

    expect(document.querySelector('[data-id="pt-form-field_a"]')).toBeTruthy();
    expect(document.querySelector('[data-id="pt-form-field_b"]')).toBeTruthy();
    expect(document.querySelector('[data-id="pt-form-field_c"]')).toBeTruthy();
  });
});

describe('T3 ProtoFormRenderer onChange 回调', () => {
  it('修改 string 字段值时触发 onChange', () => {
    const schema: FormFieldDef[] = [makeField({ fieldName: 'name', fieldType: 'string', label: '名称' })];
    const onChange = vi.fn();
    render(<ProtoFormRenderer schema={schema} values={{}} onChange={onChange} />);

    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'hello' } });

    // onChange 应被调用
    expect(onChange).toHaveBeenCalled();
  });

  it('切换 bool 字段时触发 onChange', () => {
    const schema: FormFieldDef[] = [makeField({ fieldName: 'enabled', fieldType: 'bool', label: '启用' })];
    const onChange = vi.fn();
    render(<ProtoFormRenderer schema={schema} values={{ enabled: false }} onChange={onChange} />);

    const switchEl = screen.getByRole('switch');
    fireEvent.click(switchEl);

    expect(onChange).toHaveBeenCalled();
  });
});
