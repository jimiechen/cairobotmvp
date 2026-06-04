/**
 * HistoryFilters.tsx - 历史记录筛选栏组件
 *
 * 职责：
 * - 协议号筛选输入
 * - 业务码筛选输入
 * - 时间范围选择
 * - 查询/导出/清空操作按钮
 *
 * 不负责：
 * - 数据加载（由父组件 HistoryPage 负责）
 * - 表格渲染（由 Ant Design Table 负责）
 */

import { Input, Button, Space, DatePicker } from 'antd';
import type { Dayjs } from 'dayjs';

const { RangePicker } = DatePicker;

/** 筛选栏属性 */
export interface HistoryFiltersProps {
  /** 协议号输入值 */
  protocolFilter: string;
  /** 业务码输入值 */
  businessCodeFilter: string;
  /** 时间范围 */
  timeRange: [Dayjs, Dayjs] | null;
  /** 是否有选中的行 */
  hasSelection: boolean;
  /** 协议号变更回调 */
  onProtocolChange: (value: string) => void;
  /** 业务码变更回调 */
  onBusinessCodeChange: (value: string) => void;
  /** 时间范围变更回调 */
  onTimeRangeChange: (range: [Dayjs, Dayjs] | null) => void;
  /** 查询按钮点击 */
  onSearch: () => void;
  /** 导出按钮点击 */
  onExport: () => void;
  /** 清空按钮点击 */
  onClearAll: () => void;
}

/**
 * 历史记录筛选栏
 *
 * 包含协议号、业务码、时间范围三个筛选项，
 * 以及查询、导出选中、清空全部三个操作按钮。
 */
export default function HistoryFilters({
  protocolFilter,
  businessCodeFilter,
  timeRange,
  hasSelection,
  onProtocolChange,
  onBusinessCodeChange,
  onTimeRangeChange,
  onSearch,
  onExport,
  onClearAll,
}: HistoryFiltersProps) {
  return (
    <Space style={{ marginBottom: 16 }} wrap>
      <Input
        placeholder="协议号（如 2100/2101）"
        value={protocolFilter}
        onChange={(e) => onProtocolChange(e.target.value)}
        allowClear
        style={{ width: 200 }}
      />
      <Input
        placeholder="业务码"
        value={businessCodeFilter}
        onChange={(e) => onBusinessCodeChange(e.target.value)}
        allowClear
        style={{ width: 120 }}
      />
      <RangePicker
        showTime
        value={timeRange}
        onChange={(dates) =>
          onTimeRangeChange(dates as [Dayjs, Dayjs] | null)
        }
      />
      <Button onClick={onSearch}>查询</Button>
      <Button onClick={onExport} disabled={!hasSelection}>
        导出选中
      </Button>
      <Button danger onClick={onClearAll}>
        清空全部
      </Button>
    </Space>
  );
}
