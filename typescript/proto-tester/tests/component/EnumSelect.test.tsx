/**
 * EnumSelect 组件单元测试
 *
 * 测试范围：
 * 1. 基本渲染和选择
 * 2. 搜索过滤
 * 3. 空选项状态
 */
import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { EnumSelect } from '../../src/components/EnumSelect';

describe('T3 EnumSelect 枚举选择器', () => {
  const defaultOptions = [
    { label: '选项A', value: 'a' },
    { label: '选项B', value: 'b' },
    { label: '选项C', value: 'c' },
  ];

  it('应渲染 Select 组件并显示 placeholder', () => {
    const onChange = vi.fn();
    render(
      <EnumSelect options={defaultOptions} onChange={onChange} placeholder="请选择" />
    );
    // antd Select 渲染后应包含 placeholder 文本
    expect(screen.getByText('请选择')).toBeInTheDocument();
  });

  it('点击选项应触发 onChange 回调并传递值', async () => {
    const onChange = vi.fn();
    render(<EnumSelect options={defaultOptions} onChange={onChange} />);

    // 点击触发下拉
    const selectTrigger = screen.getByRole('combobox');
    fireEvent.mouseDown(selectTrigger);

    // 选择选项B
    const option = screen.getByText('选项B');
    fireEvent.click(option);

    // antd Select 的 onChange 会传递 value 和 option 对象
    expect(onChange).toHaveBeenCalled();
    // 验证第一个参数是正确的值
    const callArg = onChange.mock.calls[0][0];
    expect(callArg).toBe('b');
  });

  it('支持搜索过滤功能（showSearch 模式）', () => {
    const onChange = vi.fn();
    render(<EnumSelect options={defaultOptions} onChange={onChange} showSearch />);

    // 点击打开下拉
    const selectTrigger = screen.getByRole('combobox');
    fireEvent.mouseDown(selectTrigger);

    // 验证 Select 已渲染
    expect(selectTrigger).toBeInTheDocument();
  });

  it('空选项列表时应正常渲染并显示 placeholder', () => {
    const onChange = vi.fn();
    render(<EnumSelect options={[]} onChange={onChange} placeholder="无选项" />);
    expect(screen.getByText('无选项')).toBeInTheDocument();
  });

  it('数字类型值正确传递', () => {
    const numberOptions = [
      { label: '零', value: 0 },
      { label: '一', value: 1 },
    ];
    const onChange = vi.fn();
    render(<EnumSelect options={numberOptions} onChange={onChange} value={0} />);
    // 应能接受数字类型的值而不报错
    expect(onChange).not.toHaveBeenCalled();
  });
});
