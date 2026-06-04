/**
 * apiClient.ts - axios 封装 + 错误码映射
 *
 * 职责：
 * - 封装 HTTP POST 请求到 Gateway
 * - 将业务错误码映射为自定义异常类
 * - 注入 Authorization header（Token）
 * - 记录请求耗时
 *
 * 不负责：
 * - MessagePacket 编解码（由 messagePacket.ts 负责）
 * - 错误类定义（由 errors.ts 负责）
 * - Token 持久化（由 session store 负责，且仅存内存）
 */

import axios, { type AxiosRequestConfig, type AxiosResponse } from 'axios';
import { encodePacket, decodePacket } from './messagePacket';
import {
  CODE_SUCCESS,
  CODE_BAD_REQUEST,
  CODE_UNAUTHORIZED,
  CODE_NOT_FOUND,
  CODE_INTERNAL_ERROR,
  CODE_NOT_IMPLEMENTED,
  ProtoTesterError,
  BadRequestError,
  UnauthorizedError,
  NotFoundError,
  InternalError,
  NotImplementedError,
} from './errors';
import type { ParsedResponse } from './messagePacket';

// 重新导出错误类，方便调用方单点导入
export {
  ProtoTesterError,
  BadRequestError,
  UnauthorizedError,
  NotFoundError,
  InternalError,
  NotImplementedError,
} from './errors';

/** 发送请求的参数 */
export interface SendRequest {
  maxType: number;
  minType: number;
  payload?: Uint8Array;
  extend?: Record<string, string>;
  token?: string;
  gatewayUrl?: string;
  /** 内部测试用：覆盖默认超时（ms） */
  _timeoutMs?: number;
}

/** 发送请求的响应 */
export interface SendResponse {
  status: number;
  businessCode: number;
  responsePacket: ParsedResponse | null;
  responseData: Uint8Array;
  traceId: string;
  durationMs: number;
}

/**
 * 根据业务码抛出对应的自定义异常
 */
function throwByBusinessCode(businessCode: number, traceId?: string): never {
  switch (businessCode) {
    case CODE_BAD_REQUEST:
      throw new BadRequestError(businessCode, traceId);
    case CODE_UNAUTHORIZED:
      throw new UnauthorizedError(businessCode, traceId);
    case CODE_NOT_FOUND:
      throw new NotFoundError(businessCode, traceId);
    case CODE_INTERNAL_ERROR:
      throw new InternalError(businessCode, traceId);
    case CODE_NOT_IMPLEMENTED:
      throw new NotImplementedError(businessCode, traceId);
    default:
      throw new ProtoTesterError(0, businessCode, traceId, `未知业务码: ${businessCode}`);
  }
}

const DEFAULT_GATEWAY_URL = 'http://localhost:8080';
const DEFAULT_TIMEOUT_MS = 10000;

/**
 * 发送 Protobuf 请求到 Gateway 并返回结构化响应
 *
 * 流程：
 * 1. 使用 encodePacket 编码请求体
 * 2. axios POST 到 gatewayUrl + '/api/hello'
 * 3. 解码响应二进制
 * 4. 映射业务错误码到自定义异常
 * 5. 返回 SendResponse
 */
export async function sendRequest(req: SendRequest): Promise<SendResponse> {
  const startTime = Date.now();
  const gatewayUrl = req.gatewayUrl || DEFAULT_GATEWAY_URL;
  const timeoutMs = req._timeoutMs || DEFAULT_TIMEOUT_MS;

  // 编码请求包
  const requestBody = encodePacket({
    maxType: req.maxType,
    minType: req.minType,
    payload: req.payload,
    extend: req.extend,
  });

  // 构建请求配置
  const config: AxiosRequestConfig<Uint8Array> = {
    method: 'POST',
    url: `${gatewayUrl}/api/hello`,
    data: requestBody,
    headers: {
      'Content-Type': 'application/octet-stream',
      ...(req.token ? { Authorization: `Bearer ${req.token}` } : {}),
    },
    responseType: 'arraybuffer',
    timeout: timeoutMs,
  };

  let response: AxiosResponse<ArrayBuffer>;
  try {
    response = await axios.request(config);
  } catch (error) {
    if (axios.isAxiosError(error)) {
      if (error.code === 'ECONNABORTED') {
        throw new ProtoTesterError(0, undefined, undefined, `请求超时 (${timeoutMs}ms)`);
      }
      if (error.response) {
        throw new InternalError(undefined, undefined);
      }
      throw new ProtoTesterError(0, undefined, undefined, `网络不可达: ${error.message}`);
    }
    throw error;
  }

  const durationMs = Date.now() - startTime;
  const responseData = new Uint8Array(response.data);

  // 解码响应包
  const parsed = decodePacket(responseData);

  // 无有效响应包时，根据 HTTP 状态码判断
  if (!parsed) {
    if (response.status >= 500) {
      throw new InternalError(undefined, undefined);
    }
    return {
      status: response.status,
      businessCode: CODE_SUCCESS,
      responsePacket: null,
      responseData,
      traceId: '',
      durationMs,
    };
  }

  // 提取业务码和 traceId
  const codeStr = parsed.extend.get('code');
  const businessCode = codeStr ? parseInt(codeStr, 10) : CODE_SUCCESS;
  const traceId = parsed.extend.get('traceId') || '';

  // 非成功业务码 → 抛出对应异常
  if (businessCode !== CODE_SUCCESS) {
    throwByBusinessCode(businessCode, traceId);
  }

  return {
    status: response.status,
    businessCode,
    responsePacket: parsed,
    responseData,
    traceId,
    durationMs,
  };
}
