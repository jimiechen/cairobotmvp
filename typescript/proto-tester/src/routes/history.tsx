/**
 * history.tsx - 请求历史列表页面
 *
 * 职责：
 * - 管理历史记录的加载状态和筛选状态
 * - 协调筛选栏、表格、导出、清空等子功能
 *
 * 不负责：
 * - 历史数据存储（由 store/history.ts 负责）
 * - 筛选栏 UI（由 HistoryFilters 组件负责）
 * - 表格列定义（由 HistoryTableConfig 组件负责）
 */

import { useState, useEffect, useCallback } from 'react';
import { Table, Modal, message } from 'antd';
import type { TablePaginationConfig } from 'antd/es/table';
import dayjs from 'dayjs';
import {
  getHistoryList,
  deleteHistory,
  clearAllHistory,
  type HistoryRecord,
} from '../store/history';
import HistoryFilters from './components/HistoryFilters';
import { getHistoryColumns, renderHistoryExpandedRow } from './components/HistoryTableConfig';

/** 每页默认条数 */
const PAGE_SIZE = 20;

export default function HistoryPage() {
  const [records, setRecords] = useState<HistoryRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [current, setCurrent] = useState(1);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  // 筛选状态
  const [protocolFilter, setProtocolFilter] = useState('');
  const [businessCodeFilter, setBusinessCodeFilter] = useState('');
  const [timeRange, setTimeRange] =
    useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);

  /** 加载历史列表（含客户端侧筛选） */
  const fetchRecords = useCallback(async () => {
    setLoading(true);
    try {
      let list = await getHistoryList({
        limit: PAGE_SIZE,
        offset: (current - 1) * PAGE_SIZE,
      });

      // 客户端侧按协议号筛选（格式：2100/2101）
      if (protocolFilter.trim()) {
        const parts = protocolFilter.split('/').map((s) => s.trim());
        if (parts.length === 2) {
          const maxT = parseInt(parts[0], 10);
          const minT = parseInt(parts[1], 10);
          if (!isNaN(maxT) && !isNaN(minT)) {
            list = list.filter(
              (r) => r.maxType === maxT && r.minType === minT,
            );
          }
        }
      }

      // 客户端侧按业务码筛选
      if (businessCodeFilter.trim()) {
        const code = parseInt(businessCodeFilter, 10);
        if (!isNaN(code)) {
          list = list.filter(
            (r) => r.responseSummary.businessCode === code,
          );
        }
      }

      // 客户端侧按时间范围筛选
      if (timeRange?.[0] && timeRange?.[1]) {
        const startMs = timeRange[0].valueOf();
        const endMs = timeRange[1].valueOf();
        list = list.filter(
          (r) => r.timestamp >= startMs && r.timestamp <= endMs,
        );
      }

      setRecords(list);
    } catch (err) {
      message.error('加载历史记录失败');
      console.error('[history]', err);
    } finally {
      setLoading(false);
    }
  }, [current, protocolFilter, businessCodeFilter, timeRange]);

  useEffect(() => {
    fetchRecords();
  }, [fetchRecords]);

  /** 删除单条记录 */
  const handleDelete = async (id: number) => {
    try {
      await deleteHistory(id);
      message.success('已删除');
      fetchRecords();
    } catch (err) {
      message.error('删除失败');
    }
  };

  /** 清空全部记录（二次确认） */
  const handleClearAll = () => {
    Modal.confirm({
      title: '确认清空',
      content: '确定要清空全部请求历史吗？此操作不可撤销。',
      okText: '确认清空',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await clearAllHistory();
          message.success('已清空全部记录');
          setRecords([]);
          setSelectedRowKeys([]);
        } catch (err) {
          message.error('清空失败');
        }
      },
    });
  };

  /** 导出选中记录为 JSON 文件下载 */
  const handleExport = () => {
    const selected = records.filter((r) =>
      selectedRowKeys.includes(r.id!),
    );
    if (selected.length === 0) {
      message.warning('请先选择要导出的记录');
      return;
    }

    const jsonStr = JSON.stringify(selected, null, 2);
    const blob = new Blob([jsonStr], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `history-export-${Date.now()}.json`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
    message.success(`已导出 ${selected.length} 条记录`);
  };

  /** 分页配置 */
  const pagination: TablePaginationConfig = {
    current,
    pageSize: PAGE_SIZE,
    total,
    showSizeChanger: false,
    onChange: (page) => setCurrent(page),
  };

  return (
    <div style={{ padding: 24 }}>
      <h2>请求历史</h2>

      <HistoryFilters
        protocolFilter={protocolFilter}
        businessCodeFilter={businessCodeFilter}
        timeRange={timeRange}
        hasSelection={selectedRowKeys.length > 0}
        onProtocolChange={setProtocolFilter}
        onBusinessCodeChange={setBusinessCodeFilter}
        onTimeRangeChange={setTimeRange}
        onSearch={fetchRecords}
        onExport={handleExport}
        onClearAll={handleClearAll}
      />

      <Table<HistoryRecord>
        rowKey="id"
        columns={getHistoryColumns(handleDelete)}
        dataSource={records}
        loading={loading}
        pagination={pagination}
        rowSelection={{
          selectedRowKeys,
          onChange: setSelectedRowKeys,
        }}
        expandable={{
          expandedRowRender: renderHistoryExpandedRow,
        }}
        size="middle"
        scroll={{ x: 1000 }}
      />
    </div>
  );
}
