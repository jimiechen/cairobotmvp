/**
 * trace.tsx - traceId 检索页面
 *
 * 职责：
 * - traceId 输入 + 查询
 * - 时间线展示（Ant Design Timeline）
 * - 从 IndexedDB 本地查询匹配记录（stub 模式）
 * - 一键复制 traceId
 * - 服务端日志跳转（stub：显示"待接入"提示）
 *
 * 不负责：
 * - 后端 API 日志查询（当前为 stub，待接入）
 * - 历史数据存储（由 store/history.ts 负责）
 */

import { useState } from 'react';
import {
  Input,
  Button,
  Timeline,
  Card,
  Space,
  message,
  Tag,
  Empty,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import { getHistoryByTraceId, type HistoryRecord } from '../store/history';
import { formatTraceDuration } from '../utils/traceId';

const { Text, Paragraph } = Typography;

/** 时间线节点颜色映射（根据 HTTP 状态码） */
function timelineColor(status: number): string {
  if (status >= 200 && status < 300) return 'green';
  if (status >= 400 && status < 500) return 'orange';
  if (status >= 500) return 'red';
  return 'blue';
}

export default function TracePage() {
  const [traceIdInput, setTraceIdInput] = useState('');
  const [records, setRecords] = useState<HistoryRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchedId, setSearchedId] = useState('');

  /** 执行 traceId 查询 */
  const handleSearch = async () => {
    const trimmed = traceIdInput.trim();
    if (!trimmed) {
      message.warning('请输入 traceId');
      return;
    }

    setLoading(true);
    setSearchedId(trimmed);

    try {
      // 从本地 IndexedDB 查询匹配的记录
      // TODO(ai, MVP2): 接入后端日志查询 API
      const result = await getHistoryByTraceId(trimmed);
      setRecords(result);

      if (result.length === 0) {
        message.info('未找到匹配的记录');
      }
    } catch (err) {
      message.error('查询失败');
      console.error('[trace]', err);
    } finally {
      setLoading(false);
    }
  };

  /** 一键复制 traceId 到剪贴板 */
  const handleCopyTraceId = async () => {
    const text = searchedId || traceIdInput.trim();
    if (!text) return;

    try {
      await navigator.clipboard.writeText(text);
      message.success('已复制到剪贴板');
    } catch {
      // fallback: 使用传统方式
      const textarea = document.createElement('textarea');
      textarea.value = text;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      message.success('已复制到剪贴板');
    }
  };

  /** 服务端日志跳转（stub） */
  const handleServerLogJump = () => {
    // TODO(ai, MVP2): 接入真实的服务端日志系统 URL
    message.info('服务端日志跳转功能待接入');
  };

  /** 渲染单条时间线节点 */
  function renderTimelineItem(record: HistoryRecord) {
    const hasError = !!record.responseSummary.error;

    return (
      <Timeline.Item
        key={record.id}
        color={timelineColor(record.responseSummary.status)}
        label={dayjs(record.timestamp).format('MM-DD HH:mm:ss')}
      >
        <Card size="small" style={{ marginBottom: 8 }}>
          <Space direction="vertical" size={4} style={{ width: '100%' }}>
            <div>
              <Text strong>{record.protocolName}</Text>
              <Tag style={{ marginLeft: 8 }}>
                {record.maxType}/{record.minType}
              </Tag>
              <Tag color={timelineColor(record.responseSummary.status)}>
                HTTP {record.responseSummary.status}
              </Tag>
              <Tag>业务码: {record.responseSummary.businessCode}</Tag>
              <Text type="secondary">
                耗时: {formatTraceDuration(record.responseSummary.durationMs)}
              </Text>
            </div>

            {/* 请求 Payload */}
            <Paragraph
              ellipsis={{ rows: 2, expandable: true }}
              style={{ marginBottom: 4 }}
            >
              <pre style={{ fontSize: 12 }}>
                {JSON.stringify(record.requestPayload, null, 2)}
              </pre>
            </Paragraph>

            {/* 错误信息 */}
            {hasError && (
              <Text type="danger">{record.responseSummary.error}</Text>
            )}
          </Space>
        </Card>
      </Timeline.Item>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      <h2>traceId 检索</h2>

      {/* 输入区 */}
      <Space style={{ marginBottom: 16 }} wrap>
        <Input.Search
          placeholder="请输入 traceId"
          value={traceIdInput}
          onChange={(e) => setTraceIdInput(e.target.value)}
          onSearch={handleSearch}
          enterButton="查询"
          loading={loading}
          style={{ width: 400 }}
          allowClear
        />
        {(searchedId || traceIdInput.trim()) && (
          <Button onClick={handleCopyTraceId}>复制 traceId</Button>
        )}
        {records.length > 0 && (
          <Button onClick={handleServerLogJump}>服务端日志</Button>
        )}
      </Space>

      {/* 结果展示区 */}
      {records.length > 0 ? (
        <Timeline mode="left">
          {records.map(renderTimelineItem)}
        </Timeline>
      ) : searchedId && !loading ? (
        <Empty description={`未找到 traceId "${searchedId}" 的相关记录`} />
      ) : null}
    </div>
  );
}
