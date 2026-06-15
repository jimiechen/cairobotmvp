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
import styled from 'styled-components';
import dayjs from 'dayjs';
import { getHistoryByTraceId, type HistoryRecord } from '../store/history';
import { formatTraceDuration } from '../utils/traceId';
import { theme } from '../styles/theme';
import { PageContainer, SearchBarWrapper } from '../styles/common';

const { Text, Paragraph } = Typography;

// ==================== styled-components ====================

/** 时间线卡片 */
const TraceCard = styled(Card).withConfig({
  shouldForwardProp: (prop) => prop !== '$as',
})`
  && {
    margin-bottom: ${theme.spacing.sm};
  }
`;

/** 时间线内容区全宽容器 */
const TimelineContent = styled(Space)`
  && {
    width: 100%;
  }
`;

/** Payload 预览区域 */
const PayloadParagraph = styled(Paragraph).withConfig({
  shouldForwardProp: (prop) => prop !== '$as',
})`
  && {
    margin-bottom: 4px;
  }
`;

/** JSON pre 标签 */
const JsonPre = styled.pre`
  font-size: ${theme.fontSize.sm};
`;

/** 搜索输入框宽度 */
const SearchInputWidth = '400px';

// ==================== 页面逻辑 ====================

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
        <TraceCard size="small">
          <TimelineContent direction="vertical" size={4}>
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
            <PayloadParagraph
              ellipsis={{ rows: 2, expandable: true }}
            >
              <JsonPre>
                {JSON.stringify(record.requestPayload, null, 2)}
              </JsonPre>
            </PayloadParagraph>

            {/* 错误信息 */}
            {hasError && (
              <Text type="danger">{record.responseSummary.error}</Text>
            )}
          </TimelineContent>
        </TraceCard>
      </Timeline.Item>
    );
  }

  return (
    <PageContainer>
      <h2>traceId 检索</h2>

      {/* 输入区 */}
      <SearchBarWrapper>
        <Space wrap>
          <Input.Search
            placeholder="请输入 traceId"
            value={traceIdInput}
            onChange={(e) => setTraceIdInput(e.target.value)}
            onSearch={handleSearch}
            enterButton="查询"
            loading={loading}
            style={{ width: SearchInputWidth }}
            allowClear
          />
          {(searchedId || traceIdInput.trim()) && (
            <Button onClick={handleCopyTraceId}>复制 traceId</Button>
          )}
          {records.length > 0 && (
            <Button onClick={handleServerLogJump}>服务端日志</Button>
          )}
        </Space>
      </SearchBarWrapper>

      {/* 结果展示区 */}
      {records.length > 0 ? (
        <Timeline mode="left">
          {records.map(renderTimelineItem)}
        </Timeline>
      ) : searchedId && !loading ? (
        <Empty description={`未找到 traceId "${searchedId}" 的相关记录`} />
      ) : null}
    </PageContainer>
  );
}
