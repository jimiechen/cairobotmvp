/**
 * App.tsx - proto-tester 应用主入口
 *
 * 职责：
 * - 路由配置（首页协议发送 / 历史记录 / 链路追踪）
 * - 全局布局（Header + Content + 侧栏）
 * - 协议发送页面的核心编排（选择协议 → 填写表单 → 发送 → 查看响应）
 *
 * 不负责：
 * - 单个组件的内部逻辑（由各 components/ 负责）
 * - HTTP 请求细节（由 apiClient 负责）
 * - 状态持久化（由各 store 负责）
 */
import { useEffect, useState, useCallback, memo } from 'react';
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import { Layout, Menu, Typography, Card, Button, Space, Tag, Alert, Collapse, Input, message, Divider, Empty } from 'antd';
import {
  SendOutlined,
  HistoryOutlined,
  SearchOutlined,
  ApiOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  LoadingOutlined,
} from '@ant-design/icons';
import styled from 'styled-components';

// 组件
import { ProtoFormRenderer } from './components/ProtoFormRenderer';
import UserSwitcher from './components/UserSwitcher';
import TokenInjector from './components/TokenInjector';
import { ExtendParamsPanel } from './components/ExtendParamsPanel';
import { JsonEditor } from './components/JsonEditor';

// Store
import { useSessionStore } from './store/session';
import { useProtocolsStore } from './store/protocols';
import { addHistory } from './store/history';

// Lib
import { getAllProtocols, getMessageSchema, type ProtocolMeta } from './lib/protoMetadata';
import { buildFormSchema, type FormFieldDef } from './lib/protoFormBuilder';
import { sendRequest } from './lib/apiClient';
import { encodePayload, decodePayload } from './lib/protoPayload';
import type { ProtoTesterError } from './lib/errors';

// 路由页面
import HistoryPage from './routes/history';
import TracePage from './routes/trace';

// 样式组件
import { theme } from './styles/theme';
import {
  MainLayout,
  ProtocolSider,
  RightSider,
  ProtocolCard,
  ProtocolName,
  ProtocolMetaInfo,
  ResponsePre,
} from './styles/common';

const { Header, Content, Footer } = Layout;
const { Title, Text, Paragraph } = Typography;

// ==================== styled-components (App.tsx 私有) ====================

/** 搜索框容器 */
const SearchContainer = styled.div`
  padding: ${theme.spacing.md} ${theme.spacing.lg} ${theme.spacing.sm};
`;

/** 协议列表滚动区域 */
const ProtocolListScroll = styled.div`
  overflow-y: auto;
  height: calc(100% - 48px);
`;

/** 中间内容区 Layout */
const ContentLayout = styled(Layout)`
  flex: 1;
  background: ${theme.colors.background};
`;

/** 内容区域 */
const ContentArea = styled(Content)`
  padding: ${theme.spacing.lg};
  overflow-y: auto;
`;

/** Header 样式化 */
const StyledHeader = styled(Header)`
  background: ${theme.colors.headerBg};
  padding: 0 ${theme.spacing.xxl};
  display: flex;
  align-items: center;
  justify-content: space-between;
`;

/** Header 标题 */
const HeaderTitle = styled(Title).withConfig({
  shouldForwardProp: (prop) => prop !== '$as',
})`
  && {
    color: #fff;
    margin: 0;
  }
`;

/** Header 导航菜单 */
const HeaderMenu = styled(Menu)`
  && {
    background: transparent;
    flex: 1;
    justify-content: flex-end;
    min-width: 200px;
  }
`;

/** Footer 样式化 */
const StyledFooter = styled(Footer)`
  text-align: center;
  padding: ${theme.spacing.lg} ${theme.spacing.xxl};
`;

/** Footer 文字 */
const FooterText = styled(Text).withConfig({
  shouldForwardProp: (prop) => prop !== '$as',
})`
  && {
    font-size: ${theme.fontSize.sm};
  }
`;

/** Token 截断显示 */
const TokenCodeText = styled(Text).withConfig({
  shouldForwardProp: (prop) => prop !== '$as',
})`
  && {
    font-size: ${theme.fontSize.xs};
  }
`;

// ==================== 子页面组件 ====================

/** 首页：协议发送器 */
function SenderPage() {
  const { token, gatewayUrl } = useSessionStore();
  const {
    protocols,
    setProtocols,
    selectedProtocol,
    selectProtocol,
    searchQuery,
    setSearchQuery,
    getFilteredProtocols,
  } = useProtocolsStore();

  // 表单状态
  const [formValues, setFormValues] = useState<Record<string, any>>({});
  const [formSchema, setFormSchema] = useState<FormFieldDef[]>([]);
  const [extendValues, setExtendValues] = useState<Record<string, string>>({
    traceId: '',
    token: '',
  });
  // 发送状态
  const [sending, setSending] = useState(false);
  const [responseData, setResponseData] = useState<string>('');
  const [responseError, setResponseError] = useState<string>('');

  // 初始化协议列表
  useEffect(() => {
    const list = getAllProtocols();
    setProtocols(list);
    // 默认选中第一个 Request 类型协议
    const firstReq = list.find((p) => p.messageType === 'Request');
    if (firstReq) {
      handleSelectProtocol(firstReq);
    }
  }, []);

  // 选择协议 → 构建表单 Schema
  const handleSelectProtocol = useCallback((protocol: ProtocolMeta) => {
    selectProtocol(protocol);
    setResponseData('');
    setResponseError('');
    setFormValues({});

    // 构建请求表单
    if (protocol.requestMessage) {
      const schema = getMessageSchema(protocol.requestMessage);
      if (schema) {
        setFormSchema(buildFormSchema(schema));
        return;
      }
    }
    // 无 schema 时降级为空
    setFormSchema([]);
  }, [selectProtocol]);

  // 表单字段变更
  const handleFormChange = useCallback((field: string, value: any) => {
    setFormValues((prev) => ({ ...prev, [field]: value }));
  }, []);

  // Extend 参数变更
  const handleExtendChange = useCallback((key: string, value: string) => {
    if (value === '') {
      // 删除操作
      setExtendValues((prev) => {
        const next = { ...prev };
        delete next[key];
        return next;
      });
    } else {
      setExtendValues((prev) => ({ ...prev, [key]: value }));
    }
  }, []);

  // 发送请求
  const handleSend = useCallback(async () => {
    if (!selectedProtocol) {
      message.warning('请先选择一个协议');
      return;
    }

    setSending(true);
    setResponseData('');
    setResponseError('');

    try {
      // 将 session 中的 token 注入 extend
      const finalExtend = { ...extendValues };
      if (token && !finalExtend.token) {
        finalExtend.token = token;
      }

      // 将表单值编码为 Protobuf payload
      const payload = selectedProtocol.requestMessage
        ? encodePayload(selectedProtocol.requestMessage, formValues)
        : undefined;

      const response = await sendRequest({
        maxType: selectedProtocol.maxType,
        minType: selectedProtocol.minType,
        payload,
        extend: finalExtend,
        gatewayUrl,
      });

      // 显示响应数据（按 MessagePacket 协议扁平化展示）
      const packet = response.responsePacket;
      // 查找响应消息名：当前选中的是 Request 协议，需匹配同 maxType 的 Response 协议
      const respProto = protocols.find(
        p => p.maxType === selectedProtocol.maxType && p.direction === 'S->C' && p.responseMessage,
      );
      const respName = respProto?.responseMessage || selectedProtocol.responseMessage;
      console.log('[App] selectedProtocol:', selectedProtocol.name, 'respName:', respName, 'respProto:', respProto?.name);
      let dataDisplay: any = null;
      if (packet?.data && packet.data.length > 0 && respName) {
        dataDisplay = decodePayload(respName, packet.data);
      } else if (packet?.data && packet.data.length > 0) {
        const raw = new TextDecoder().decode(packet.data);
        dataDisplay = { _raw: raw, _hex: Array.from(packet.data).map(b => b.toString(16).padStart(2, '0')).join(' ') };
      }
      setResponseData(
        JSON.stringify(
          {
            _meta: {
              code: response.businessCode,
              traceId: response.traceId,
              durationMs: response.durationMs,
              httpStatus: response.status,
            },
            // MessagePacket 平级字段（与协议定义一致）
            maxType: packet?.maxType,
            minType: packet?.minType,
            platform: packet?.platform,
            extend: packet?.extend ? Object.fromEntries(packet.extend) : null,
            data: dataDisplay,
          },
          null,
          2,
        ) || '(空响应)',
      );
      message.success(`请求成功 (${response.durationMs}ms)`);

      // 保存到历史记录
      try {
        await addHistory({
          timestamp: Date.now(),
          traceId: response.traceId,
          maxType: selectedProtocol.maxType,
          minType: selectedProtocol.minType,
          protocolName: selectedProtocol.name,
          requestPayload: formValues,
          responseSummary: {
            status: response.status,
            businessCode: response.businessCode,
            durationMs: response.durationMs,
          },
        });
      } catch {
        // IndexedDB 写入失败不阻塞主流程
      }
    } catch (err) {
      const error = err as ProtoTesterError;
      setResponseError(error.message || '请求失败');
      message.error(error.message || '请求失败');

      // 失败也记录历史
      try {
        await addHistory({
          timestamp: Date.now(),
          traceId: error.traceId || '',
          maxType: selectedProtocol!.maxType,
          minType: selectedProtocol!.minType,
          protocolName: selectedProtocol!.name,
          requestPayload: formValues,
          responseSummary: {
            status: 0,
            businessCode: error.businessCode || 0,
            durationMs: 0,
            error: error.message,
          },
        });
      } catch {
        // ignore
      }
    } finally {
      setSending(false);
    }
  }, [selectedProtocol, formValues, extendValues, token, gatewayUrl, protocols]);

  // 过滤后的协议列表
  const filteredProtocols = getFilteredProtocols();
  // 仅显示 Request 类型的协议用于发送
  const requestProtocols = filteredProtocols.filter((p) => p.messageType === 'Request');

  return (
    <MainLayout>
      {/* 左侧：协议列表 */}
      <ProtocolSider width={280} theme="light">
        <SearchContainer>
          <Input
            placeholder="搜索协议..."
            prefix={<SearchOutlined />}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            allowClear
            size="small"
          />
        </SearchContainer>
        <ProtocolListScroll>
          {requestProtocols.map((p) => (
            <ProtocolCardItem
              key={`${p.maxType}-${p.minType}`}
              protocol={p}
              isSelected={selectedProtocol?.maxType === p.maxType && selectedProtocol?.minType === p.minType}
              onSelect={handleSelectProtocol}
            />
          ))}
          {requestProtocols.length === 0 && (
            <Empty description="无匹配协议" image={Empty.PRESENTED_IMAGE_SIMPLE} style={{ marginTop: 40 }} />
          )}
        </ProtocolListScroll>
      </ProtocolSider>

      {/* 中间：表单 + 响应 */}
      <ContentLayout>
        <ContentArea>
          {/* 协议信息头部 */}
          {selectedProtocol && (
            <Card size="small" style={{ marginBottom: 16 }} title={
              <Space>
                <ApiOutlined />
                <span>{selectedProtocol.name}</span>
                <Tag color="blue">{selectedProtocol.maxType}:{selectedProtocol.minType}</Tag>
                <Tag>{selectedProtocol.direction}</Tag>
              </Space>
            }>
              <Paragraph type="secondary" style={{ marginBottom: 0 }}>{selectedProtocol.description}</Paragraph>
            </Card>
          )}

          {/* 请求表单 */}
          {formSchema.length > 0 ? (
            <Card
              size="small"
              title="请求参数"
              style={{ marginBottom: 16 }}
              extra={
                <Button
                  type="primary"
                  icon={sending ? <LoadingOutlined /> : <SendOutlined />}
                  onClick={handleSend}
                  loading={sending}
                  data-id="pt-btn-send"
                >
                  发送请求
                </Button>
              }
            >
              <ProtoFormRenderer
                schema={formSchema}
                values={formValues}
                onChange={handleFormChange}
              />
            </Card>
          ) : selectedProtocol ? (
            <Card size="small" title="请求参数" style={{ marginBottom: 16 }}>
              <Alert type="info" message="该协议无结构化 Schema，请通过 Extend 参数传递原始数据。" showIcon />
              <div style={{ marginTop: 8 }}>
                <JsonEditor
                  value={formValues._raw || '{}'}
                  onChange={(v) => handleFormChange('_raw', v)}
                  height="200px"
                />
              </div>
              <Button
                type="primary"
                icon={sending ? <LoadingOutlined /> : <SendOutlined />}
                onClick={handleSend}
                loading={sending}
                style={{ marginTop: 8 }}
                data-id="pt-btn-send"
              >
                发送请求
              </Button>
            </Card>
          ) : (
            <Empty description="请从左侧选择一个协议" style={{ marginTop: 60 }} />
          )}

          {/* 响应区域 */}
          {(responseData || responseError) && (
            <Card
              size="small"
              title={
                <Space>
                  {responseError ? (
                    <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
                  ) : (
                    <CheckCircleOutlined style={{ color: '#52c41a' }} />
                  )}
                  <span>响应结果</span>
                </Space>
              }
              style={{ marginBottom: 16 }}
            >
              {responseError ? (
                <Alert type="error" message={responseError} showIcon />
              ) : (
                <ResponsePre>{responseData}</ResponsePre>
              )}
            </Card>
          )}
        </ContentArea>
      </ContentLayout>

      {/* 右侧：Extend 参数 + 用户/TOKEN */}
      <RightSider width={300} theme="light">
        <Collapse defaultActiveKey={['user', 'extend']} size="small">
          <Collapse.Panel header="测试用户 & Token" key="user">
            <div style={{ marginBottom: 12 }}>
              <UserSwitcher />
            </div>
            <TokenInjector />
          </Collapse.Panel>

          <Collapse.Panel header="Extend 参数" key="extend">
            <ExtendParamsPanel
              values={extendValues}
              onChange={handleExtendChange}
              globalValues={{
                method: 'UnknownMethod',
                platform: 'WEB',
              }}
            />
          </Collapse.Panel>

          <Collapse.Panel header="网关配置" key="gateway">
            <Text type="secondary" style={{ fontSize: 12 }}>Gateway URL:</Text>
            <br />
            <Text code style={{ fontSize: 12 }}>{gatewayUrl}</Text>
            <Divider style={{ margin: '8px 0' }} />
            <Text type="secondary" style={{ fontSize: 12 }}>当前 Token:</Text>
            <br />
            <TokenCodeText code>
              {token ? `${token.slice(0, 8)}***...` : '(未设置)'}
            </TokenCodeText>
          </Collapse.Panel>
        </Collapse>
      </RightSider>
    </MainLayout>
  );
}

// ==================== 性能优化：memo 化协议卡片 ====================

interface ProtocolCardItemProps {
  protocol: ProtocolMeta;
  isSelected: boolean;
  onSelect: (protocol: ProtocolMeta) => void;
}

const ProtocolCardItem = memo(function ProtocolCardItem({ protocol, isSelected, onSelect }: ProtocolCardItemProps) {
  return (
    <ProtocolCard $selected={isSelected} onClick={() => onSelect(protocol)}>
      <ProtocolName strong>{protocol.name}</ProtocolName>
      <ProtocolMetaInfo type="secondary">
        {protocol.maxType}:{protocol.minType} | {protocol.description}
      </ProtocolMetaInfo>
    </ProtocolCard>
  );
});

// ==================== 主布局组件 ====================

function AppLayout() {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <StyledHeader>
        <HeaderTitle level={4}>
          <ApiOutlined style={{ marginRight: 8 }} />
          proto-tester
        </HeaderTitle>
        <HeaderMenu
          theme="dark"
          mode="horizontal"
          selectable={false}
          items={[
            { key: 'sender', icon: <SendOutlined />, label: <Link to="/">协议发送</Link> },
            { key: 'history', icon: <HistoryOutlined />, label: <Link to="/history">历史记录</Link> },
            { key: 'trace', icon: <ClockCircleOutlined />, label: <Link to="/trace">链路追踪</Link> },
          ]}
        />
      </StyledHeader>
      <Layout>
        <Content style={{ margin: 0, height: 'calc(100vh - 64px)' }}>
          <Routes>
            <Route path="/" element={<SenderPage />} />
            <Route path="/history" element={<HistoryPage />} />
            <Route path="/trace" element={<TracePage />} />
          </Routes>
        </Content>
      </Layout>
      <StyledFooter>
        <FooterText type="secondary">
          proto-tester v2.0 &copy; CaiRobot MVP &mdash; 内网研发工具，禁止公网部署
        </FooterText>
      </StyledFooter>
    </Layout>
  );
}

// ==================== 根组件 ====================

export default function App() {
  return (
    <BrowserRouter>
      <AppLayout />
    </BrowserRouter>
  );
}
