/**
 * protoMetadata.ts - 运行时协议元数据查询模块
 *
 * 职责：
 * - 从 protocols.json 加载协议元数据（构建期产物）
 * - 提供 protocol / messageSchema 的查询接口
 * - 纯函数，无副作用
 *
 * 不负责：
 * - 解析 .proto 文件（由 gen-protocols.mjs 在构建期完成）
 * - 表单 UI 渲染（由 protoFormBuilder.ts 负责）
 */

import protocolsJSON from '@/data/protocols.json';

/** 单条协议的元数据（来自协议编号注册表） */
export interface ProtocolMeta {
  maxType: number;
  minType: number;
  name: string;
  direction: string;
  messageType: string;
  protoFile: string;
  openapi: string;
  status: string;
  description: string;
  requestMessage?: string;
  responseMessage?: string;
}

/** Proto 字段的 Schema 定义 */
export interface FieldSchema {
  name: string;
  number: number;
  type: string;
  label: string;
  oneof: string | null;
  comment?: string | null;
}

/** Enum 类型定义 */
export interface EnumSchema {
  name: string;
  fullName: string;
  values: Record<string, number>;
}

/** 完整 Message Schema（含嵌套类型引用） */
export interface MessageSchema {
  fullName: string;
  protoFile: string;
  fields: FieldSchema[];
  nestedTypes: string[];
  enums: EnumSchema[];
}

/** 获取全部协议列表 */
export function getAllProtocols(): ProtocolMeta[] {
  return protocolsJSON.protocols;
}

/** 按 maxType + minType 精确查找单条协议 */
export function getProtocolByMaxMin(maxType: number, minType: number): ProtocolMeta | undefined {
  return protocolsJSON.protocols.find(
    (p) => p.maxType === maxType && p.minType === minType,
  );
}

/** 按 message 名称获取其 Schema（含字段、嵌套类型、枚举）
 *  支持全限定名（com.mineplanet.pojo.hello.HelloWorldRequest）和短名（HelloWorldRequest）查找
 */
export function getMessageSchema(messageName: string): MessageSchema | undefined {
  // 精确匹配全限定名
  if (protocolsJSON.messageSchemas[messageName]) {
    return protocolsJSON.messageSchemas[messageName];
  }
  // 按短名后缀匹配（如 "HelloWorldRequest" 匹配 "com.xxx.HelloWorldRequest"）
  const matchedKey = Object.keys(protocolsJSON.messageSchemas).find(
    (key) => key === messageName || key.endsWith(`.${messageName}`),
  );
  return matchedKey ? protocolsJSON.messageSchemas[matchedKey] : undefined;
}

/** 按 name 或 description 模糊搜索协议 */
export function searchProtocols(query: string): ProtocolMeta[] {
  const lowerQuery = query.toLowerCase();
  return protocolsJSON.protocols.filter(
    (p) =>
      p.name.toLowerCase().includes(lowerQuery) ||
      (p.description && p.description.toLowerCase().includes(lowerQuery)),
  );
}
