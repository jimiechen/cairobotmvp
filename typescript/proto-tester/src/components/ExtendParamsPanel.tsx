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

const { Text } = Typography;

/** ExtendParamsPanel 组件属性 */
interface ExtendParamsPanelProps {
  /** 当前 extend 键值对 */
  values: Record<string, string>;
  /** 值变更回调：key 变更时传旧 key 和新 key，value 变更时传 key 和新值 */
  onChange: (key: string, value: string) => void;
  /** 全局默认 extend 参数（只读参考） */
  globalValues?: Record<string, string>;
}

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
    <div data-id="pt-extend-panel" style={{ padding: '8px 0' }}>
      {/* 可编辑的键值对列表 */}
      {entries.map(([key, val]) => (
        <div key={key} style={{ display: 'flex', gap: 8, marginBottom: 4, alignItems: 'center' }}>
          <Input
            value={key}
            disabled
            style={{ width: 140 }}
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
        </div>
      ))}

      {/* 新增键值对输入区 */}
      <div style={{ display: 'flex', gap: 8, marginTop: 8, alignItems: 'center' }}>
        <Input
          placeholder="新 Key"
          value={newKey}
          onChange={(e) => setNewKey(e.target.value)}
          size="small"
          style={{ width: 140 }}
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
      </div>

      {/* 全局默认参数参考区 */}
      {globalValues && Object.keys(globalValues).length > 0 && (
        <>
          <Divider style={{ margin: '12px 0 8px' }} />
          <Text type="secondary" style={{ fontSize: 12 }}>全局默认参数（只读参考）</Text>
          <div style={{ marginTop: 4 }}>
            {Object.entries(globalValues).map(([key, val]) => (
              <Tag key={key} color="blue" style={{ marginBottom: 4 }}>
                {key}: {val}
              </Tag>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
