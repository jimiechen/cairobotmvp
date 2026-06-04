/**
 * HistoryTableConfig.tsx - 历史记录表格列定义与展开行
 *
 * 职责：
 * - 定义 Ant Design Table 的 columns 配置
 * - 定义展开行渲染（payload JSON + 响应摘要）
 *
 * 不负责：
 * - 数据加载（由 HistoryPage 负责）
 * - 筛选逻辑（由 HistoryFilters 负责）
 */

import { Button, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import type { HistoryRecord } from '../../store/history';
import { formatTraceDuration } from '../../utils/traceId';

/** 业务码到标签颜色的映射 */
function businessCodeTagColor(code: number): string {
  if (code === 0) return 'green';
  if (code >= 400 && code < 500) return 'orange';
  if (code >= 500) return 'red';
  return 'blue';
}

/**
 * 获取历史记录表格的列定义
 *
 * @param onDelete 删除按钮回调
 */
export function getHistoryColumns(
  onDelete: (id: number) => void,
): ColumnsType<HistoryRecord> {
  return [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (ts: number) => dayjs(ts).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '协议名',
      dataIndex: 'protocolName',
      key: 'protocolName',
      width: 200,
      ellipsis: true,
    },
    {
      title: '协议号',
      key: 'protocolNum',
      width: 120,
      render: (_: any, record: HistoryRecord) =>
        `${record.maxType}/${record.minType}`,
    },
    {
      title: 'traceId',
      dataIndex: 'traceId',
      key: 'traceId',
      width: 240,
      ellipsis: true,
      render: (id: string) => (
        <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{id}</span>
      ),
    },
    {
      title: '业务码',
      key: 'businessCode',
      width: 90,
      render: (_: any, record: HistoryRecord) => (
        <Tag color={businessCodeTagColor(record.responseSummary.businessCode)}>
          {record.responseSummary.businessCode}
        </Tag>
      ),
    },
    {
      title: '耗时',
      key: 'duration',
      width: 80,
      render: (_: any, record: HistoryRecord) =>
        formatTraceDuration(record.responseSummary.durationMs),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: any, record: HistoryRecord) => (
        <Button size="small" danger onClick={() => onDelete(record.id!)}>
          删除
        </Button>
      ),
    },
  ];
}

/**
 * 渲染历史记录的展开行内容（请求 payload + 响应摘要）
 */
export function renderHistoryExpandedRow(
  record: HistoryRecord,
): React.ReactNode {
  return (
    <div style={{ padding: '8px 16px' }}>
      <h4>请求 Payload</h4>
      <pre
        style={{
          maxHeight: 300,
          overflow: 'auto',
          background: '#f5f5f5',
          padding: 12,
        }}
      >
        {JSON.stringify(record.requestPayload, null, 2)}
      </pre>
      <h4>响应摘要</h4>
      <pre style={{ background: '#f5f5f5', padding: 12 }}>
        {JSON.stringify(record.responseSummary, null, 2)}
      </pre>
    </div>
  );
}
