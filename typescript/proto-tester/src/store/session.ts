/**
 * session.ts - 内存状态管理（Token + endpoint）
 *
 * 职责：
 * - 管理 Token（100% 内存化，不持久化到任何存储）
 * - 管理当前选中的测试用户 ID
 * - 管理 Gateway URL
 *
 * 铁律：Token 仅存内存，不写 localStorage / sessionStorage / IndexedDB
 *
 * 不负责：
 * - Token 刷新逻辑（由调用方决定）
 * - HTTP 请求发送（由 apiClient.ts 负责）
 */

import { create } from 'zustand';

/** Session Store 状态定义 */
interface SessionState {
  /** 当前认证 Token（仅内存，刷新即丢失） */
  token: string | null;
  /** 当前选中的测试用户 ID */
  selectedUserId: string;
  /** 当前 Gateway URL */
  gatewayUrl: string;

  // Actions
  setToken: (token: string | null) => void;
  setSelectedUserId: (id: string) => void;
  setGatewayUrl: (url: string) => void;
  clearToken: () => void;
}

export const useSessionStore = create<SessionState>((set) => ({
  token: null,
  selectedUserId: '',
  gatewayUrl: 'http://localhost:8080',

  setToken: (token) => set({ token }),

  setSelectedUserId: (id) => set({ selectedUserId: id }),

  setGatewayUrl: (url) => set({ gatewayUrl: url }),

  clearToken: () => set({ token: null }),
}));
