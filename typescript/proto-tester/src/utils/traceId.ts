/**
 * traceId.ts - traceId 工具函数
 *
 * 职责：
 * - 生成 UUID v4 格式的 traceId
 * - 格式化请求耗时（毫秒/秒）
 * - 从响应 packet 中提取 traceId
 *
 * 不负责：
 * - traceId 持久化存储（由 store/history.ts 负责）
 * - traceId 与后端日志的关联（由 routes/trace.tsx 负责）
 */

import { v4 as uuidv4 } from 'uuid';

/** 1 秒对应的毫秒数 */
const SECOND_MS = 1000;

/** UUID v4 正则表达式 */
const UUID_V4_REGEX =
  /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

/**
 * 生成 UUID v4 格式的 traceId
 *
 * @returns UUID v4 字符串，如 "550e8400-e29b-41d4-a716-446655440000"
 */
export function generateTraceId(): string {
  return uuidv4();
}

/**
 * 格式化请求耗时
 *
 * 规则：
 * - < 1000ms：显示 "Xms"（如 245ms）
 * - >= 1000ms：显示 "X.Xs"（如 1.2s），保留一位小数
 *
 * @param ms 耗时毫秒数
 * @returns 格式化后的字符串
 */
export function formatTraceDuration(ms: number): string {
  if (ms < SECOND_MS) {
    return `${ms}ms`;
  }
  // 保留一位小数
  const seconds = (ms / SECOND_MS).toFixed(1);
  return `${seconds}s`;
}

/**
 * 从响应 packet 数据中提取 traceId
 *
 * 尝试从 JSON 序列化的 packet extend 字段中解析 traceId。
 * 如果无法解析或不存在 traceId 字段，返回 null。
 *
 * @param packetData 响应二进制数据（通常是 JSON 编码的 packet）
 * @returns 提取到的 traceId 字符串，或 null
 */
export function extractTraceIdFromPacket(
  packetData: Uint8Array,
): string | null {
  if (!packetData || packetData.length === 0) {
    return null;
  }

  try {
    const text = new TextDecoder().decode(packetData);
    const parsed = JSON.parse(text);

    // 兼容多种可能的 traceId 存储位置
    if (typeof parsed.traceId === 'string' && parsed.traceId) {
      return parsed.traceId;
    }

    // 从 extend 对象中查找
    if (parsed.extend && typeof parsed.extend.traceId === 'string') {
      return parsed.extend.traceId;
    }

    return null;
  } catch {
    // 非文本数据或非法 JSON，无法提取
    return null;
  }
}
