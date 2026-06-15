/**
 * theme.ts - 统一设计令牌（Design Tokens）
 *
 * 集中管理颜色、间距、字号等设计变量，
 * 供所有 styled-components 引用，确保样式一致性。
 */

export const theme = {
  colors: {
    primary: '#1890ff',
    success: '#52c41a',
    error: '#ff4d4f',
    warning: '#faad14',

    // 布局与边框
    border: '#f0f0f0',
    background: '#fff',
    sidebarBg: '#fafafa',
    selectedBg: '#e6f4ff',
    hoverBg: '#fafafa',
    codeBg: '#f6f8fa',

    // 文字
    text: {
      primary: 'rgba(0, 0, 0, 0.85)',
      secondary: 'rgba(0, 0, 0, 0.45)',
      code: 'rgba(0, 0, 0, 0.85)',
    },

    // Ant Design Header 深色背景
    headerBg: '#001529',
  },

  spacing: {
    xs: '4px',
    sm: '8px',
    md: '12px',
    lg: '16px',
    xl: '24px',
    xxl: '32px',
  },

  fontSize: {
    xs: '11px',
    sm: '12px',
    base: '13px',
    md: '14px',
    lg: '16px',
  },

  layout: {
    siderWidth: 280,
    rightSiderWidth: 300,
    headerHeight: 64,
  },
} as const;
