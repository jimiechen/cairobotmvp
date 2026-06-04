/**
 * EnumSelect.tsx - 枚举选择器组件
 *
 * 职责：
 * - 基于 Ant Design Select 渲染枚举选项
 * - 支持搜索过滤功能
 * - 支持字符串和数字类型的枚举值
 *
 * 不负责：
 * - 枚举值的来源获取（由 protoFormBuilder 负责）
 */
import { Select } from 'antd';

/** 枚举选项定义 */
interface EnumOption {
  /** 显示标签 */
  label: string;
  /** 枚举值（支持 string 或 number） */
  value: string | number;
}

/** EnumSelect 组件属性 */
interface EnumSelectProps {
  /** 当前选中值 */
  value?: string | number;
  /** 可选枚举列表 */
  options: EnumOption[];
  /** 值变更回调 */
  onChange: (value: string | number) => void;
  /** 占位提示文本，默认 '请选择' */
  placeholder?: string;
  /** 是否显示搜索框，默认 true */
  showSearch?: boolean;
}

/**
 * 枚举选择器组件
 *
 * 基于 antd Select 封装，专为 Proto enum 字段设计。
 * 支持 showSearch 模式下的模糊匹配过滤。
 */
export function EnumSelect({
  value,
  options,
  onChange,
  placeholder = '请选择',
  showSearch = true,
}: EnumSelectProps) {
  return (
    <Select
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      options={options}
      showSearch={showSearch}
      optionFilterProp="label"
      allowClear
      style={{ width: '100%' }}
    />
  );
}
