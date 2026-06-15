/**
 * common.ts - 通用样式组件
 *
 * 从各组件中提取的可复用 styled-components，
 * 消除内联样式的重复定义，统一引用 theme 设计令牌。
 */
import styled from 'styled-components';
import { Layout, Typography } from 'antd';
import { theme } from './theme';

const { Sider } = Layout;
const { Text, Paragraph } = Typography;

// ==================== 布局组件 ====================

/** SenderPage 主布局：左右分栏 + 中间内容区 */
export const MainLayout = styled.div`
  display: flex;
  gap: ${theme.spacing.lg};
  height: 100%;
`;

/** 左侧协议列表 Sider */
export const ProtocolSider = styled(Sider)`
  background: ${theme.colors.background};
  border-right: 1px solid ${theme.colors.border};
`;

/** 右侧参数面板 Sider */
export const RightSider = styled(Sider)`
  background: ${theme.colors.sidebarBg};
  border-left: 1px solid ${theme.colors.border};
  padding: ${theme.spacing.lg};
  overflow-y: auto;
`;

// ==================== 协议列表组件 ====================

/** 单个协议卡片（可选中态） */
export const ProtocolCard = styled.div<{ $selected?: boolean }>`
  padding: ${theme.spacing.md} ${theme.spacing.lg};
  cursor: pointer;
  border-bottom: 1px solid ${theme.colors.border};
  background: ${(props) => (props.$selected ? theme.colors.selectedBg : 'transparent')};

  &:hover {
    background: ${(props) => (props.$selected ? theme.colors.selectedBg : theme.colors.hoverBg)};
  }
`;

/** 协议名称（加粗） */
export const ProtocolName = styled(Text).withConfig({
  shouldForwardProp: (prop) => prop !== '$as',
})`
  && {
    font-weight: 600;
    font-size: ${theme.fontSize.base};
  }
`;

/** 协议元信息（maxType:minType | 描述） */
export const ProtocolMetaInfo = styled(Text).withConfig({
  shouldForwardProp: (prop) => prop !== '$as',
})`
  && {
    font-size: ${theme.fontSize.xs};
    color: ${theme.colors.text.secondary};
    margin-top: 4px;
    display: block;
  }
`;

// ==================== 响应区域组件 ====================

/** JSON 响应预览区域 */
export const ResponsePre = styled.pre`
  background: ${theme.colors.codeBg};
  padding: ${theme.spacing.lg};
  border-radius: 6px;
  font-size: ${theme.fontSize.sm};
  max-height: 400px;
  overflow: auto;
`;

// ==================== 页面容器组件 ====================

/** 标准页面内容容器（用于 history / trace 等子页面） */
export const PageContainer = styled.div`
  padding: ${theme.spacing.xxl};
`;

/** 搜索栏间距容器 */
export const SearchBarWrapper = styled.div`
  margin-bottom: ${theme.spacing.lg};
`;
