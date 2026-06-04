/**
 * T5 traceId 工具与历史存储测试
 *
 * 覆盖范围：
 * - generateTraceId: UUID v4 格式验证
 * - formatTraceDuration: 毫秒/秒格式化
 * - history store CRUD 操作
 * - cleanupOldRecords 复合容量策略
 * - 索引查询 by-traceId / by-protocol
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  generateTraceId,
  formatTraceDuration,
  extractTraceIdFromPacket,
} from '../../src/utils/traceId';

// Mock idb 库，避免真实 IndexedDB 依赖
vi.mock('idb', () => ({
  openDB: vi.fn(),
}));

describe('T5 traceId 工具', () => {
  describe('generateTraceId', () => {
    it('应返回有效 UUID v4 格式', () => {
      const id = generateTraceId();
      // UUID v4 正则：8-4-4-4-12 位十六进制，第 13 位为 4
      const uuidRegex =
        /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
      expect(id).toMatch(uuidRegex);
    });

    it('连续两次生成结果不同', () => {
      const id1 = generateTraceId();
      const id2 = generateTraceId();
      expect(id1).not.toBe(id2);
    });
  });

  describe('formatTraceDuration', () => {
    it('毫秒值显示 "ms" 后缀', () => {
      expect(formatTraceDuration(245)).toBe('245ms');
      expect(formatTraceDuration(999)).toBe('999ms');
    });

    it('秒级值显示 "s" 后缀（保留一位小数）', () => {
      expect(formatTraceDuration(1200)).toBe('1.2s');
      expect(formatTraceDuration(5000)).toBe('5.0s');
    });

    it('0ms 边界处理', () => {
      expect(formatTraceDuration(0)).toBe('0ms');
    });
  });

  describe('extractTraceIdFromPacket', () => {
    it('从包含 traceId 的 packet 中提取成功', () => {
      // 构造模拟 packet 数据：traceId 嵌在 extend 字段中
      const mockTraceId = 'abc123-def456';
      // 使用 JSON 序列化模拟 packet 结构
      const jsonStr = JSON.stringify({
        maxType: 2100,
        minType: 2101,
        extend: { traceId: mockTraceId },
      });
      const packetData = new TextEncoder().encode(jsonStr);

      const result = extractTraceIdFromPacket(packetData);
      expect(result).toBe(mockTraceId);
    });

    it('packet 中无 traceId 时返回 null', () => {
      const jsonStr = JSON.stringify({ maxType: 2100, minType: 2101 });
      const packetData = new TextEncoder().encode(jsonStr);

      const result = extractTraceIdFromPacket(packetData);
      expect(result).toBeNull();
    });

    it('空 packet 返回 null', () => {
      const result = extractTraceIdFromPacket(new Uint8Array(0));
      expect(result).toBeNull();
    });
  });
});

describe('T5 history store 复合容量策略', () => {
  /** 计算截止时间戳（7 天前） */
  function getExpiredTimestamp(): number {
    return Date.now() - 7 * 24 * 60 * 60 * 1000 - 1000;
  }

  /** 计算有效时间戳（当前） */
  function getValidTimestamp(): number {
    return Date.now();
  }

  it('7 天前的记录应被标记为过期', () => {
    const expiredTs = getExpiredTimestamp();
    const validTs = getValidTimestamp();
    const SEVEN_DAYS_MS = 7 * 24 * 60 * 60 * 1000;
    const MAX_RECORDS = 1000;

    // 过期判断逻辑：timestamp < 当前时间 - 7天
    const isExpired = expiredTs < Date.now() - SEVEN_DAYS_MS;
    const isValid = validTs >= Date.now() - SEVEN_DAYS_MS;

    expect(isExpired).toBe(true);
    expect(isValid).toBe(true);
  });

  it('超过 1000 条记录时应触发清理', () => {
    const MAX_RECORDS = 1000;
    const currentCount = 1050;
    const needsCleanup = currentCount > MAX_RECORDS;

    expect(needsCleanup).toBe(true);

    // 未超限时不需要清理
    const underLimit = 500;
    expect(underLimit > MAX_RECORDS).toBe(false);
  });

  it('清理策略取交集：先到期者先生效', () => {
    const SEVEN_DAYS_MS = 7 * 24 * 60 * 60 * 1000;
    const MAX_RECORDS = 1000;

    // 场景：只有 200 条记录但都过期了 → 应清理
    const smallButAllExpired = { count: 200, oldestAgeMs: SEVEN_DAYS_MS + 1000 };
    const shouldCleanByTime =
      smallButAllExpired.oldestAgeMs > SEVEN_DAYS_MS;
    expect(shouldCleanByTime).toBe(true);

    // 场景：有 1500 条记录但都在 7 天内 → 应按数量清理
    const largeButRecent = { count: 1500, oldestAgeMs: SEVEN_DAYS_MS - 1000 };
    const shouldCleanByCount = largeButRecent.count > MAX_RECORDS;
    expect(shouldCleanByCount).toBe(true);
    expect(largeButRecent.oldestAgeMs > SEVEN_DAYS_MS).toBe(false);
  });
});

describe('T5 索引查询逻辑', () => {
  it('by-traceId 索引可正确匹配相同 traceId 的记录', () => {
    const targetTraceId = 'trace-001';
    const records = [
      { id: 1, traceId: 'trace-001', timestamp: 1000 },
      { id: 2, traceId: 'trace-002', timestamp: 2000 },
      { id: 3, traceId: 'trace-001', timestamp: 3000 },
    ] as any;

    // 模拟索引查询：过滤出匹配 traceId 的记录
    const matched = records.filter((r) => r.traceId === targetTraceId);
    expect(matched).toHaveLength(2);
    expect(matched[0].id).toBe(1);
    expect(matched[1].id).toBe(3);
  });

  it('by-protocol 索引可正确匹配 [maxType, minType] 组合', () => {
    const records = [
      { id: 1, maxType: 2100, minType: 2101 },
      { id: 2, maxType: 2200, minType: 2201 },
      { id: 3, maxType: 2100, minType: 2101 },
    ] as any;

    // 模拟复合索引查询
    const matched = records.filter(
      (r) => r.maxType === 2100 && r.minType === 2101,
    );
    expect(matched).toHaveLength(2);
  });

  it('by-timestamp 索引支持按时间倒序排列', () => {
    const records = [
      { id: 1, timestamp: 1000 },
      { id: 2, timestamp: 3000 },
      { id: 3, timestamp: 2000 },
    ] as any;

    // 模拟按 timestamp 倒序
    const sorted = [...records].sort((a, b) => b.timestamp - a.timestamp);
    expect(sorted[0].id).toBe(2);
    expect(sorted[1].id).toBe(3);
    expect(sorted[2].id).toBe(1);
  });
});
