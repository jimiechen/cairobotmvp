/**
 * protocols.ts - 协议列表 Zustand Store
 *
 * 职责：
 * - 管理协议列表状态（筛选、收藏、选中）
 * - 提供搜索过滤能力
 * - 协调 protoMetadata 和 UI 层的数据流
 *
 * 不负责：
 * - Proto 文件解析（由 gen-protocols.mjs 完成）
 * - 表单 Schema 构建（由 protoFormBuilder.ts 负责）
 */

import { create } from 'zustand';
import type { ProtocolMeta } from '../lib/protoMetadata';
import {
  getAllProtocols as fetchAllProtocols,
  searchProtocols as searchFromMetadata,
} from '../lib/protoMetadata';

/** Protocols Store 状态定义 */
interface ProtocolsState {
  /** 全部协议列表 */
  protocols: ProtocolMeta[];
  /** 收藏的协议 key 集合（格式："maxType-minType"） */
  favorites: Set<string>;
  /** 当前搜索关键词 */
  searchQuery: string;
  /** 当前选中的协议 */
  selectedProtocol: ProtocolMeta | null;

  // Actions
  /** 设置协议列表（通常在 App 初始化时调用一次） */
  setProtocols: (protocols: ProtocolMeta[]) => void;
  /** 切换收藏状态 */
  toggleFavorite: (key: string) => void;
  /** 设置搜索关键词 */
  setSearchQuery: (query: string) => void;
  /** 选中/取消选中协议 */
  selectProtocol: (protocol: ProtocolMeta | null) => void;
  /** 获取经过搜索+收藏过滤后的协议列表 */
  getFilteredProtocols: () => ProtocolMeta[];
}

export const useProtocolsStore = create<ProtocolsState>((set, get) => ({
  protocols: [],
  favorites: new Set<string>(),
  searchQuery: '',
  selectedProtocol: null,

  setProtocols: (protocols) => set({ protocols }),

  toggleFavorite: (key) =>
    set((state) => {
      const nextFavorites = new Set(state.favorites);
      if (nextFavorites.has(key)) {
        nextFavorites.delete(key);
      } else {
        nextFavorites.add(key);
      }
      return { favorites: nextFavorites };
    }),

  setSearchQuery: (query) => set({ searchQuery: query }),

  selectProtocol: (protocol) => set({ selectedProtocol: protocol }),

  getFilteredProtocols: () => {
    const state = get();
    let result = state.protocols;

    // 搜索过滤
    if (state.searchQuery.trim()) {
      result = searchFromMetadata(state.searchQuery);
    }

    // 收藏排序：收藏的排前面
    return [...result].sort((a, b) => {
      const aKey = `${a.maxType}-${a.minType}`;
      const bKey = `${b.maxType}-${b.minType}`;
      const aFav = state.favorites.has(aKey) ? 1 : 0;
      const bFav = state.favorites.has(bKey) ? 1 : 0;
      return bFav - aFav;
    });
  },
}));
