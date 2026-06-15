/**
 * ExtendParamsPanel.tsx - MessagePacket extend 参数编辑面板
 *
 * 职责：
 * - 显示和编辑 MessagePacket 的 extend map（键值对）
 * - 支持增删键值对操作
 * - 显示全局默认参数作为只读参考
 *
 * 不负责：
 * - extend 参数的业务校验（由调用方负责）
 */
import { useState } from 'react';
import { Input, Button, Space, Tag, Typography, Divider } from 'antd';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import styled from 'styled-components';
import { theme } from '../styles/theme';

const { Text } = Typography;

// ==================== styled-components ====================

/** 面板根容器 */
const PanelRoot = styled.div`
  padding: ${theme.spacing.sm} 0;
`;

/** 键值对行 */
const KvRow = styled.div`
  display: flex;
  gap: ${theme.spacing.sm};
  margin-bottom: 4px;
  align-items: center;
`;

/** 新增输入行 */
const AddRow = styled(KvRow)`
  margin-top: ${theme.spacing.sm};
`;

/** Key 输入框固定宽度 */
const KeyInputWidth = '140px';

/** 全局默认标签容器 */
const GlobalTagsContainer = styled.div`
  margin-top: 4px;
`;

/** 分割线间距 */
const StyledDivider = styled(Divider)`
  && {
    margin: ${theme.spacing.md} 0 ${theme.spacing.sm};
  }
`;

/** 全局默认说明文字 */
const GlobalLabel = styled(Text)`
  && {
    font-size: ${theme.fontSize.sm};
  }
`;

/** 标签底部间距 */
const GlobalTag = styled(Tag)`
  && {
    margin-bottom: 4px;
  }
`;

// ==================== 组件属性 ====================

/** ExtendParamsPanel 组件属性 */
interface ExtendParamsPanelProps {
  /** 当前 extend 键值对 */
  values: Record<string, string>;
  /** 值变更回调：key 变更时传旧 key 和新 key，value 变更时传 key 和新值 */
  onChange: (key: string, value: string) => void;
  /** 全局默认 extend 参数（只读参考） */
  globalValues?: Record<string, string>;
}

// ==================== 组件实现 ====================

/**
 * MessagePacket extend 参数编辑面板
 *
 * UI 结构：
 * 1. 可编辑区域：当前 values 的键值对列表
 * 2. 添加按钮：新增空键值对
 * 3. 全局默认区域：显示 globalValues 作为参考（只读）
 */
export function ExtendParamsPanel({
  values,
  onChange,
  globalValues,
}: ExtendParamsPanelProps) {
  const [newKey, setNewKey] = useState('');
  const [newValue, setNewValue] = useState('');

  /** 处理添加新的键值对 */
  const handleAdd = () => {
    const trimmedKey = newKey.trim();
    if (trimmedKey && !(trimmedKey in values)) {
      onChange(trimmedKey, newValue);
      setNewKey('');
      setNewValue('');
    }
  };

  /** 处理删除键值对 */
  const handleDelete = (key: string) => {
    // 删除时传空字符串表示移除
    onChange(key, '');
  };

  const entries = Object.entries(values).filter(([, v]) => v !== '');

  return (
    <PanelRoot data-id="pt-extend-panel">
      {/* 可编辑的键值对列表 */}
      {entries.map(([key, val]) => (
        <KvRow key={key}>
          <Input
            value={key}
            disabled
            style={{ width: KeyInputWidth }}
            size="small"
          />
          <Input
            value={val}
            onChange={(e) => onChange(key, e.target.value)}
            size="small"
            style={{ flex: 1 }}
          />
          <Button
            type="text"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(key)}
            size="small"
          />
        </KvRow>
      ))}

      {/* 新增键值对输入区 */}
      <AddRow>
        <Input
          placeholder="新 Key"
          value={newKey}
          onChange={(e) => setNewKey(e.target.value)}
          size="small"
          style={{ width: KeyInputWidth }}
          onPressEnter={handleAdd}
        />
        <Input
          placeholder="Value"
          value={newValue}
          onChange={(e) => setNewValue(e.target.value)}
          size="small"
          style={{ flex: 1 }}
          onPressEnter={handleAdd}
        />
        <Button
          type="dashed"
          icon={<PlusOutlined />}
          onClick={handleAdd}
          size="small"
        >
          添加
        </Button>
      </AddRow>

      {/* 全局默认参数参考区 */}
      {globalValues && Object.keys(globalValues).length > 0 && (
        <>
          <StyledDivider />
          <GlobalLabel type="secondary">全局默认参数（只读参考）</GlobalLabel>
          <GlobalTagsContainer>
            {Object.entries(globalValues).map(([key, val]) => (
              <GlobalTag key={key} color="blue">
                {key}: {val}
              </GlobalTag>
            ))}
          </GlobalTagsContainer>
        </>
      )}
    </PanelRoot>
  );
}
