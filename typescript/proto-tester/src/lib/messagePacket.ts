/**
 * messagePacket.ts - MessagePacket 编解码模块
 *
 * 职责：
 * - 将业务参数编码为 MessagePacket Protobuf 二进制
 * - 将响应二进制解码为结构化 ParsedResponse 对象
 * - 自动填充标准 extend 字段（method / traceId / requestId）
 *
 * 不负责：
 * - HTTP 请求发送（由 apiClient.ts 负责）
 * - 错误码映射（由 apiClient.ts 负责）
 */

import { v4 as uuidv4 } from 'uuid';
import { com as messageCom } from '@proto/base/message';
const { MessagePacket, Platform } = messageCom.mineplanet.pojo;

/** 构建 MessagePacket 的选项 */
export interface PacketBuildOptions {
  maxType: number;
  minType: number;
  payload?: Uint8Array;
  platform?: typeof Platform[keyof typeof Platform];
  extend?: Record<string, string>;
  traceId?: string;
}

/** 解码后的响应结构 */
export interface ParsedResponse {
  maxType: number;
  minType: number;
  platform: number;
  extend: Map<string, string>;
  data: Uint8Array;
  rawPacket: InstanceType<typeof MessagePacket>;
}

/**
 * 将选项编码为 MessagePacket 二进制
 *
 * 自动填充标准 extend 字段：
 * - method：从 extend.method 或默认值
 * - traceId：从参数或自动生成 UUID
 * - requestId：自动生成 UUID
 */
export function encodePacket(opts: PacketBuildOptions): Uint8Array {
  const traceId = opts.traceId || uuidv4();
  const requestId = uuidv4();

  // 合并标准 extend + 用户自定义 extend
  const mergedExtend = new Map<string, string>([
    ['method', opts.extend?.method || 'UnknownMethod'],
    ['traceId', traceId],
    ['requestId', requestId],
    ...Object.entries(opts.extend || {}),
  ]);

  // google-protobuf 3.21.2 writeBytes 要求 Buffer 兼容的 Uint8Array
  const safePayload = opts.payload
    ? Uint8Array.from(opts.payload)
    : new Uint8Array(0);

  const packet = new MessagePacket({
    maxType: opts.maxType,
    minType: opts.minType,
    platform: opts.platform ?? Platform.WEB,
    extend: mergedExtend,
    data: safePayload,
  });

  return packet.serialize();
}

/**
 * 将二进制数据解码为结构化响应对象
 *
 * 解析失败时返回 null（不抛异常，由调用方决定处理方式）
 */
export function decodePacket(binary: Uint8Array): ParsedResponse | null {
  try {
    const rawPacket = MessagePacket.deserialize(binary);
    return {
      maxType: rawPacket.maxType,
      minType: rawPacket.minType,
      platform: rawPacket.platform,
      extend: rawPacket.extend || new Map(),
      data: rawPacket.data || new Uint8Array(0),
      rawPacket,
    };
  } catch {
    return null;
  }
}
