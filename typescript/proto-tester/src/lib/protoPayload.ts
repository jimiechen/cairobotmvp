/**
 * protoPayload.ts - 表单值 ↔ Protobuf bytes 编解码模块
 *
 * 职责：
 * - 将前端表单 formValues 编码为对应 Protobuf 消息的二进制字节
 * - 将服务端返回的 Protobuf 二进制反序列化为可读 JSON 对象
 * - 通过消息名称动态查找生成的 Protobuf 类并序列化/反序列化
 *
 * 不负责：
 * - MessagePacket 封装（由 messagePacket.ts 负责）
 * - HTTP 发送（由 apiClient.ts 负责）
 */

import { com as healthCom } from '@proto/base/health';
import { com as helloCom } from '@proto/base/hello';
import { com as configCom } from '@proto/base/app_config';
import { com as i18nCom } from '@proto/base/i18n';
// 使用与 protoc-gen-ts 生成代码一致的导入方式（避免触发类型声明缺失错误）
import * as pb from 'google-protobuf';

// 从各命名空间提取请求消息类
// 注意：protoc-gen-ts 生成的命名空间路径取决于 .proto 的 package 声明
const HealthMessages = healthCom.mineplanet.pojo.health;
const HelloMessages = helloCom.mineplanet.pojo.hello;
// app_config 和 i18n 的 package 都是 com.mineplanet.pojo（非子命名空间）
const ConfigMessages = configCom.mineplanet.pojo;
const I18nMessages = i18nCom.mineplanet.pojo;

/** 消息名称 → Protobuf 类构造函数的注册表（请求消息） */
const messageRegistry: Record<string, new (data?: any) => pb.Message> = {
  // System 模块
  ServiceHealthCheckRequest: HealthMessages.ServiceHealthCheckRequest,
  HelloWorldRequest: HelloMessages.HelloWorldRequest,

  // Config 模块
  AppConfigsReq: ConfigMessages.AppConfigsReq,
  AppConfigVersionReq: ConfigMessages.AppConfigVersionReq,

  // I18n 模块
  AppFetchLanguageReq: I18nMessages.AppFetchLanguageReq,
  AppFetchLangPackReq: I18nMessages.AppFetchLangPackReq,
  AppFetchLangDifferenceReq: I18nMessages.AppFetchLangDifferenceReq,
};

/** 响应消息名称 → Protobuf 类（用于反序列化） */
const responseRegistry: Record<string, { deserialize(bytes: Uint8Array): pb.Message }> = {
  // System 模块
  ServiceHealthCheckResponse: HealthMessages.ServiceHealthCheckResponse,
  HelloWorldResponse: HelloMessages.HelloWorldResponse,

  // Config 模块
  AppConfigsRsp: ConfigMessages.AppConfigsRsp,
  AppConfigVersionRsp: ConfigMessages.AppConfigVersionRsp,

  // I18n 模块
  AppFetchLanguageRsp: I18nMessages.AppFetchLanguageRsp,
  AppFetchLangPackRsp: I18nMessages.AppFetchLangPackRsp,
  AppFetchLangDifferenceRsp: I18nMessages.AppFetchLangDifferenceRsp,
};

/**
 * 将表单值编码为 Protobuf 二进制字节
 *
 * @param messageName - Protobuf 消息类名（如 "ServiceHealthCheckRequest"）
 * @param formValues - 前端表单收集的键值对（字段名 → 值）
 * @returns 序列化后的 Uint8Array；消息未注册时返回空数组
 */
export function encodePayload(messageName: string, formValues: Record<string, any>): Uint8Array {
  const MsgClass = messageRegistry[messageName];
  if (!MsgClass) {
    console.warn(`[protoPayload] 未注册的消息类型: ${messageName}，发送空 payload`);
    return new Uint8Array(0);
  }

  try {
    const instance = new MsgClass(formValues);
    return new Uint8Array(instance.serialize());
  } catch (err) {
    console.error(`[protoPayload] 编码 ${messageName} 失败:`, err);
    return new Uint8Array(0);
  }
}

/**
 * 检查指定消息类型是否已注册（可用于 UI 提示）
 */
export function isMessageRegistered(messageName: string): boolean {
  return messageName in messageRegistry;
}

/**
 * 检查指定响应消息类型是否已注册
 */
export function isResponseRegistered(responseName: string): boolean {
  return responseName in responseRegistry;
}

/**
 * 将 Protobuf 二进制字节反序列化为可读对象（使用 toJSON）
 *
 * @param responseName - 响应 Protobuf 消息类名（如 "ServiceHealthCheckResponse"）
 * @param bytes - 服务端返回的 MessagePacket.data 字段（Protobuf 二进制）
 * @returns 反序列化后的 JSON 对象；未注册或失败时返回原始 hex/raw 信息
 */
export function decodePayload(responseName: string, bytes: Uint8Array): any {
  console.log(`[protoPayload] decodePayload called: name=${responseName}, bytesLen=${bytes?.length}`);
  const MsgClass = responseRegistry[responseName];
  console.log(`[protoPayload] MsgClass found: ${!!MsgClass}, keys=${Object.keys(responseRegistry)}`);
  if (!MsgClass || !bytes || bytes.length === 0) {
    // 未注册时降级为 hex/raw 展示
    const raw = new TextDecoder().decode(bytes);
    console.warn(`[protoPayload] 降级为 hex/raw: name=${responseName}, registered=${responseName in responseRegistry}`);
    return { _raw: raw, _hex: Array.from(bytes || []).map(b => b.toString(16).padStart(2, '0')).join(' ') };
  }

  try {
    const instance = MsgClass.deserialize(bytes);
    console.log(`[protoPayload] 反序列化成功: ${responseName}`, instance);
    const obj = instance.toObject();
    console.log(`[protoPayload] toObject 结果:`, obj);
    return obj;
  } catch (err) {
    console.error(`[protoPayload] 解码 ${responseName} 失败:`, err);
    const raw = new TextDecoder().decode(bytes);
    return { _raw: raw, _hex: Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join(' '), _error: String(err) };
  }
}
