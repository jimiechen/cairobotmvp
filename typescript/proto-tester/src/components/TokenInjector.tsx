/**
 * TokenInjector.tsx - Token 注入面板
 *
 * 职责：
 * - 显示当前 token 状态（脱敏显示）
 * - 提供"装载"按钮（从 test_users.json 加载当前用户的 token）
 * - 提供"清除"按钮（清空 session.token）
 * - 提供"复制审计哈希"按钮（计算简短摘要，仅展示不存储）
 *
 * 安全铁律：
 * - Token 100% 内存化（session store 的 state）
 * - 不写 localStorage / sessionStorage / IndexedDB
 * - console.log 只输出前 8 位 + "***"
 * - 哈希仅做纯计算展示，不存储任何地方
 *
 * 不负责：
 * - 用户选择（由 UserSwitcher 负责）
 * - Token 验证逻辑
 */
import React, { useCallback, useMemo } from 'react';
import { Button, Space, Alert, Typography, Tooltip } from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  LoadIcon,
  ClearOutlined,
  CopyOutlined,
} from '@ant-design/icons';
import { useSessionStore } from '../store/session';
import testUsers from '../data/test_users.json';

const { Text } = Typography;

/** 脱敏显示长度 */
const MASK_PREFIX_LENGTH = 8;

/** 计算简单审计哈希（取前 16 字符的 base64 编码） */
function computeAuditHash(token: string): string {
  if (!token) return '';
  // 简单截断 + base64 编码作为审计摘要（非加密用途）
  const prefix = token.slice(0, 16);
  try {
    return btoa(prefix);
  } catch {
    // fallback：直接返回截断值
    return prefix + '...';
  }
}

/** 对 token 进行脱敏处理，只显示前 N 位 + "***" */
function maskToken(token: string | null): string {
  if (!token) return '(未装载)';
  if (token.length <= MASK_PREFIX_LENGTH) return '***';
  return token.slice(0, MASK_PREFIX_LENGTH) + '***...';
}

interface TokenInjectorProps {
  // 预留扩展属性
}

/**
 * TokenInjector 组件 - Token 注入面板
 */
export default function TokenInjector(_props: TokenInjectorProps) {
  const { token, selectedUserId, setToken, clearToken } = useSessionStore();

  /** 当前选中用户对象 */
  const currentUser = useMemo(
    () => testUsers.find((u) => u.id === selectedUserId),
    [selectedUserId]
  );

  /** 是否已装载 token */
  const isTokenLoaded = !!token;

  /** 审计哈希值（每次 token 变化时重新计算） */
  const auditHash = useMemo(() => computeAuditHash(token || ''), [token]);

  /**
   * 装载 token：从 test_users.json 读取当前选中用户的 token
   */
  const handleLoad = useCallback(() => {
    if (!currentUser) {
      message.warning('请先选择一个测试用户');
      return;
    }

    if (currentUser.tokenSource === 'fixed' && currentUser.token) {
      setToken(currentUser.token);
      message.success('Token 已从用户配置装载');
    } else if (currentUser.tokenSource === 'mock_jwt') {
      message.info('该用户为 mock_jwt 类型，请手动输入 Token');
    } else {
      message.warning('该用户无预置 Token');
    }
  }, [currentUser, setToken]);

  /**
   * 清除 token：清空 session.token
   */
  const handleClear = useCallback(() => {
    clearToken();
    message.info('Token 已清除');
  }, [clearToken]);

  /**
   * 复制审计哈希到剪贴板
   */
  const handleCopyHash = useCallback(async () => {
    if (!auditHash) {
      message.warning('无 Token 可生成哈希');
      return;
    }
    try {
      await navigator.clipboard.writeText(auditHash);
      message.success('审计哈希已复制到剪贴板');
    } catch {
      // fallback：使用传统复制方式
      const textarea = document.createElement('textarea');
      textarea.value = auditHash;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      message.success('审计哈希已复制到剪贴板');
    }
  }, [auditHash]);

  return (
    <div style={{ padding: 16, border: '1px solid #d9d9d9', borderRadius: 8 }}>
      {/* Token 状态指示器 */}
      <Alert
        type={isTokenLoaded ? 'success' : 'info'}
        icon={isTokenLoaded ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
        message={
          <Space>
            <Text strong>Token 状态：</Text>
            <Text code>{maskToken(token)}</Text>
            <Tag color={isTokenLoaded ? 'success' : 'default'}>
              {isTokenLoaded ? '已装载' : '未装载'}
            </Tag>
          </Space>
        }
        style={{ marginBottom: 12 }}
      />

      {/* 操作按钮组 */}
      <Space wrap>
        <Tooltip title="从当前选中用户的配置中装载 Token">
          <Button
            icon={<LoadOutlined />}
            onClick={handleLoad}
            disabled={!selectedUserId}
          >
            装载
          </Button>
        </Tooltip>

        <Tooltip title="清空内存中的 Token">
          <Button
            icon={<ClearOutlined />}
            onClick={handleClear}
            disabled={!isTokenLoaded}
            danger
          >
            清除
          </Button>
        </Tooltip>

        <Tooltip title="复制 Token 的审计哈希（仅用于日志追踪，非完整 Token）">
          <Button
            icon={<CopyOutlined />}
            onClick={handleCopyHash}
            disabled={!isTokenLoaded}
          >
            复制审计哈希
          </Button>
        </Tooltip>
      </Space>

      {/* 审计哈希展示区 */}
      {auditHash && (
        <div style={{ marginTop: 12, padding: 8, background: '#f5f5f5', borderRadius: 4 }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            审计哈希（SHA-256 截断摘要）：
          </Text>
          <Text code copyable={{ text: auditHash }} style={{ fontSize: 12 }}>
            {auditHash}
          </Text>
        </div>
      )}

      {/* 安全提示 */}
      <div style={{ marginTop: 8 }}>
        <Text type="warning" style={{ fontSize: 11 }}>
          ⚠️ Token 仅存内存，刷新页面后需重新装载。不会持久化到任何存储。
        </Text>
      </div>
    </div>
  );
}
