/**
 * UserSwitcher.tsx - 测试用户切换组件
 *
 * 职责：
 * - 从 session store 读取当前用户
 * - 提供下拉框切换测试用户
 * - 切换时自动注入 fixed token 到内存
 * - mock_jwt 用户提示手动输入
 *
 * 不负责：
 * - Token 持久化（由 session store 保证 100% 内存）
 * - Token 刷新逻辑
 */
import React, { useCallback } from 'react';
import { Select, Tag, message } from 'antd';
import { useSessionStore } from '../store/session';
import testUsers from '../data/test_users.json';

/** 角色到中文标签颜色的映射 */
const ROLE_COLOR_MAP: Record<string, string> = {
  admin: 'red',
  operator: 'blue',
  viewer: 'default',
  user: 'green',
  attacker: 'orange',
};

interface UserSwitcherProps {
  /** 用户切换后的回调 */
  onUserChange?: (userId: string) => void;
}

/**
 * UserSwitcher 组件
 *
 * @param props.onUserChange - 用户 ID 变更回调
 */
export default function UserSwitcher({ onUserChange }: UserSwitcherProps) {
  const { selectedUserId, setSelectedUserId, setToken } = useSessionStore();

  /**
   * 处理用户切换
   * @param userId - 选中的用户 ID
   */
  const handleUserChange = useCallback(
    (userId: string) => {
      // 更新选中的用户 ID
      setSelectedUserId(userId);

      // 查找目标用户
      const targetUser = testUsers.find((u) => u.id === userId);

      if (!targetUser) {
        message.warning('未找到该用户配置');
        return;
      }

      // 根据 tokenSource 类型处理
      if (targetUser.tokenSource === 'fixed' && targetUser.token) {
        // fixed token：自动注入到内存
        setToken(targetUser.token);
        message.success(`已切换至 ${targetUser.name}，token 已自动装载`);
      } else if (targetUser.tokenSource === 'mock_jwt') {
        // mock_jwt：清空 token 并提示手动输入
        setToken(null);
        message.info(
          `${targetUser.name} 需要手动输入 Token，请在下方注入面板操作`
        );
      }

      // 触发外部回调
      onUserChange?.(userId);
    },
    [setSelectedUserId, setToken, onUserChange]
  );

  /** 渲染下拉选项的标签（用户名 + 角色 badge） */
  const renderOptionLabel = (user: (typeof testUsers)[number]) => (
    <span>
      {user.name}
      <Tag color={ROLE_COLOR_MAP[user.role] || 'default'} style={{ marginLeft: 8 }}>
        {user.role}
      </Tag>
    </span>
  );

  return (
    <Select
      data-id="pt-user-switcher"
      style={{ width: 240 }}
      placeholder="选择测试用户"
      value={selectedUserId || undefined}
      onChange={handleUserChange}
      options={testUsers.map((user) => ({
        value: user.id,
        label: renderOptionLabel(user),
      }))}
      optionLabelProp="value"
      optionRender={(option) => {
        const user = testUsers.find((u) => u.id === option.value);
        return user ? renderOptionLabel(user) : option.label;
      }}
    />
  );
}
