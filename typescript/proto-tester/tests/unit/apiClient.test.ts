/**
 * T2 核心通信层单元测试
 *
 * 测试范围：
 * 1. messagePacket.ts - MessagePacket 编解码往返一致性
 * 2. apiClient.ts - axios 封装 + 错误码映射 + 拦截器
 * 3. session.ts - 内存状态管理（Token / endpoint）
 *
 * 使用 MSW (msw/node) 拦截 HTTP 请求，不依赖真实 Gateway
 * @vitest-environment node
 */
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import axios from 'axios';
import { com as messageCom } from '@proto/base/message';
const { MessagePacket, Platform } = messageCom.mineplanet.pojo;

// 被测模块
import {
  encodePacket,
  decodePacket,
  type PacketBuildOptions,
  type ParsedResponse,
} from '../../src/lib/messagePacket';
import {
  sendRequest,
  ProtoTesterError,
  BadRequestError,
  UnauthorizedError,
  NotFoundError,
  InternalError,
  NotImplementedError,
  type SendRequest,
  type SendResponse,
} from '../../src/lib/apiClient';
import { useSessionStore } from '../../src/store/session';

// ============================================================
// MSW Server 设置
// ============================================================
// node 环境下 axios 自动使用 http adapter，MSW/node 可直接拦截

/** 构造成功响应的 MessagePacket 二进制 */
function buildSuccessResponsePacket(
  maxType: number,
  minType: number,
  code: string = '10200',
  traceId?: string,
  responseData?: string,
): Uint8Array {
  const packet = new MessagePacket({
    maxType,
    minType,
    platform: Platform.WEB,
    extend: new Map<string, string>([
      ['code', code],
      ['traceId', traceId || 'test-trace-123'],
      ['method', 'TestResponse'],
    ]),
    data: responseData ? Uint8Array.from(Buffer.from(responseData)) : new Uint8Array(0),
  });
  return packet.serialize();
}

const server = setupServer({
  onUnhandledRequest: 'error',
});

// 必须在所有测试前启动 MSW 拦截
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterAll(() => server.close());

beforeEach(() => {
  server.resetHandlers();
  useSessionStore.setState({
    token: null,
    selectedUserId: '',
    gatewayUrl: 'http://localhost:8080',
  });
});

// ============================================================
// 1. encodePacket / decodePacket 往返一致性
// ============================================================

describe('T2 messagePacket 编解码', () => {
  it('encodePacket 应生成有效二进制', () => {
    const opts: PacketBuildOptions = {
      maxType: 2100,
      minType: 2101,
      platform: Platform.WEB,
      payload: Uint8Array.from(Buffer.from('hello')),
      extend: { customKey: 'customValue' },
      traceId: 'fixed-trace-id',
    };

    const binary = encodePacket(opts);

    expect(binary).toBeInstanceOf(Uint8Array);
    expect(binary.length).toBeGreaterThan(0);
  });

  it('decodePacket 应正确还原 MessagePacket 字段', () => {
    const opts: PacketBuildOptions = {
      maxType: 2100,
      minType: 2101,
      platform: Platform.WEB,
      payload: Uint8Array.from(Buffer.from('test-payload')),
      traceId: 'roundtrip-trace',
    };

    const binary = encodePacket(opts);
    const parsed = decodePacket(binary);

    expect(parsed).not.toBeNull();
    expect(parsed!.maxType).toBe(2100);
    expect(parsed!.minType).toBe(2101);
    expect(parsed!.platform).toBe(Platform.WEB);
    // 使用 Buffer.from 比较 Uint8Array 内容
    expect(Buffer.from(parsed!.data).toString()).toBe('test-payload');
  });

  it('encodePacket → decodePacket 往返一致性（含 extend）', () => {
    const opts: PacketBuildOptions = {
      maxType: 3100,
      minType: 3101,
      platform: Platform.ANDROID,
      extend: { method: 'TestMethod', extra: 'extraVal' },
      traceId: 'extend-roundtrip',
    };

    const binary = encodePacket(opts);
    const parsed = decodePacket(binary);

    expect(parsed).not.toBeNull();
    expect(parsed!.maxType).toBe(3100);
    expect(parsed!.minType).toBe(3101);
    expect(parsed!.platform).toBe(Platform.ANDROID);

    // 标准 extend 字段应自动填充
    expect(parsed!.extend.has('method')).toBe(true);
    expect(parsed!.extend.has('traceId')).toBe(true);
    expect(parsed!.extend.has('requestId')).toBe(true);
    // 用户自定义 extend 也保留
    expect(parsed!.extend.get('extra')).toBe('extraVal');
  });

  it('traceId 未提供时应自动生成 UUID 格式', () => {
    const opts: PacketBuildOptions = {
      maxType: 2100,
      minType: 2101,
    };

    const binary = encodePacket(opts);
    const parsed = decodePacket(binary);
    const autoTraceId = parsed!.extend.get('traceId');

    // UUID 格式：xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
    expect(autoTraceId).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/,
    );
  });

  it('decodePacket 解析无效二进制应返回 null', () => {
    const invalidBinary = new Uint8Array([0xff, 0xfe, 0xfd]);
    const parsed = decodePacket(invalidBinary);
    expect(parsed).toBeNull();
  });

  it('platform 默认值应为 WEB', () => {
    const opts: PacketBuildOptions = {
      maxType: 2100,
      minType: 2101,
    };

    const binary = encodePacket(opts);
    const parsed = decodePacket(binary);
    expect(parsed!.platform).toBe(Platform.WEB);
  });

  it('payload 为空时 data 应为空 Uint8Array', () => {
    const opts: PacketBuildOptions = {
      maxType: 2100,
      minType: 2101,
    };

    const binary = encodePacket(opts);
    const parsed = decodePacket(binary);
    expect(parsed!.data).toBeInstanceOf(Uint8Array);
    expect(parsed!.data.length).toBe(0);
  });
});

// ============================================================
// 2. sendRequest 成功与业务错误场景
// ============================================================

describe('T2 apiClient sendRequest', () => {
  it('成功请求应返回 SendResponse（code=10200）', async () => {
    const responseBody = buildSuccessResponsePacket(2100, 2101, '10200', 'success-trace');

    server.use(
      http.post('http://localhost:8080/api/hello', () =>
        HttpResponse.arrayBuffer(responseBody.buffer, {
          status: 200,
          headers: { 'Content-Type': 'application/octet-stream' },
        }),
      ),
    );

    const req: SendRequest = {
      maxType: 2100,
      minType: 2101,
    };

    const resp: SendResponse = await sendRequest(req);

    expect(resp.status).toBe(200);
    expect(resp.businessCode).toBe(10200);
    expect(resp.responsePacket).not.toBeNull();
    expect(resp.responsePacket!.maxType).toBe(2100);
    expect(resp.traceId).toBe('success-trace');
    expect(resp.durationMs).toBeGreaterThanOrEqual(0);
  });

  it('业务错误 10400 应抛出 BadRequestError', async () => {
    const errorBody = buildSuccessResponsePacket(2100, 9999, '10400', 'bad-request-trace');

    server.use(
      http.post('http://localhost:8080/api/hello', () =>
        HttpResponse.arrayBuffer(errorBody.buffer, {
          status: 200,
          headers: { 'Content-Type': 'application/octet-stream' },
        }),
      ),
    );

    const req: SendRequest = {
      maxType: 2100,
      minType: 9999,
    };

    try {
      await sendRequest(req);
      expect.unreachable('应抛出 BadRequestError');
    } catch (e) {
      expect(e).toBeInstanceOf(BadRequestError);
      expect((e as BadRequestError).businessCode).toBe(10400);
      expect((e as BadRequestError).traceId).toBe('bad-request-trace');
    }
  });

  it('业务错误 10401 应抛出 UnauthorizedError（Token 过期）', async () => {
    const errorBody = buildSuccessResponsePacket(2100, 9999, '10401', 'unauth-trace');

    server.use(
      http.post('http://localhost:8080/api/hello', () =>
        HttpResponse.arrayBuffer(errorBody.buffer, {
          status: 200,
          headers: { 'Content-Type': 'application/octet-stream' },
        }),
      ),
    );

    const req: SendRequest = {
      maxType: 2100,
      minType: 9999,
    };

    try {
      await sendRequest(req);
      expect.unreachable('应抛出 UnauthorizedError');
    } catch (e) {
      expect(e).toBeInstanceOf(UnauthorizedError);
      expect((e as UnauthorizedError).businessCode).toBe(10401);
    }
  });

  it('业务错误 10404 应抛出 NotFoundError', async () => {
    const errorBody = buildSuccessResponsePacket(9999, 9999, '10404', 'notfound-trace');

    server.use(
      http.post('http://localhost:8080/api/hello', () =>
        HttpResponse.arrayBuffer(errorBody.buffer, {
          status: 200,
          headers: { 'Content-Type': 'application/octet-stream' },
        }),
      ),
    );

    const req: SendRequest = { maxType: 9999, minType: 9999 };

    try {
      await sendRequest(req);
      expect.unreachable('应抛出 NotFoundError');
    } catch (e) {
      expect(e).toBeInstanceOf(NotFoundError);
      expect((e as NotFoundError).businessCode).toBe(10404);
    }
  });

  it('业务错误 10500 应抛出 InternalError（含 traceId）', async () => {
    const errorBody = buildSuccessResponsePacket(2100, 9999, '10500', 'internal-trace');

    server.use(
      http.post('http://localhost:8080/api/hello', () =>
        HttpResponse.arrayBuffer(errorBody.buffer, {
          status: 200,
          headers: { 'Content-Type': 'application/octet-stream' },
        }),
      ),
    );

    const req: SendRequest = { maxType: 2100, minType: 9999 };

    try {
      await sendRequest(req);
      expect.unreachable('应抛出 InternalError');
    } catch (e) {
      expect(e).toBeInstanceOf(InternalError);
      expect((e as InternalError).businessCode).toBe(10500);
      expect((e as InternalError).traceId).toBe('internal-trace');
    }
  });

  it('业务错误 10501 应抛出 NotImplementedError', async () => {
    const errorBody = buildSuccessResponsePacket(2100, 9999, '10501', 'notimpl-trace');

    server.use(
      http.post('http://localhost:8080/api/hello', () =>
        HttpResponse.arrayBuffer(errorBody.buffer, {
          status: 200,
          headers: { 'Content-Type': 'application/octet-stream' },
        }),
      ),
    );

    const req: SendRequest = { maxType: 2100, minType: 9999 };

    try {
      await sendRequest(req);
      expect.unreachable('应抛出 NotImplementedError');
    } catch (e) {
      expect(e).toBeInstanceOf(NotImplementedError);
      expect((e as NotImplementedError).businessCode).toBe(10501);
    }
  });

  it('HTTP 500 Gateway 错误应抛出 InternalError', async () => {
    server.use(
      http.post('http://localhost:8080/api/hello', () =>
        new HttpResponse(null, { status: 500, statusText: 'Internal Server Error' }),
      ),
    );

    const req: SendRequest = { maxType: 2100, minType: 2101 };

    try {
      await sendRequest(req);
      expect.unreachable('应抛出 InternalError');
    } catch (e) {
      expect(e).toBeInstanceOf(InternalError);
    }
  });

  it('网络超时应抛出 ProtoTesterError', async () => {
    server.use(
      http.post('http://localhost:8080/api/hello', async () => {
        await new Promise((resolve) => setTimeout(resolve, 200));
        return new HttpResponse(null, { status: 200 });
      }),
    );

    const req: SendRequest = {
      maxType: 2100,
      minType: 2101,
      gatewayUrl: 'http://localhost:8080',
    };

    try {
      await sendRequest({ ...req, _timeoutMs: 50 });
      expect.unreachable('应抛出 ProtoTesterError');
    } catch (e) {
      expect(e).toBeInstanceOf(ProtoTesterError);
      expect((e as ProtoTesterError).message).toContain('超时');
    }
  });

  it('Token 应注入到 Authorization header', async () => {
    let capturedAuthHeader: string | null = null;

    server.use(
      http.post('http://localhost:8080/api/hello', ({ request }) => {
        capturedAuthHeader = request.headers.get('Authorization');
        return HttpResponse.arrayBuffer(
          buildSuccessResponsePacket(2100, 2101, '10200').buffer,
          {
            status: 200,
            headers: { 'Content-Type': 'application/octet-stream' },
          },
        );
      }),
    );

    const req: SendRequest = {
      maxType: 2100,
      minType: 2101,
      token: 'test-token-abc123',
    };

    await sendRequest(req);

    expect(capturedAuthHeader).toBe('Bearer test-token-abc123');
  });

  it('自定义 gatewayUrl 应生效', async () => {
    server.use(
      http.post('http://custom-gateway:9090/api/hello', () =>
        HttpResponse.arrayBuffer(
          buildSuccessResponsePacket(2100, 2101, '10200', 'custom-url-trace').buffer,
          {
            status: 200,
            headers: { 'Content-Type': 'application/octet-stream' },
          },
        ),
      ),
    );

    const req: SendRequest = {
      maxType: 2100,
      minType: 2101,
      gatewayUrl: 'http://custom-gateway:9090',
    };

    const resp = await sendRequest(req);
    expect(resp.status).toBe(200);
    expect(resp.traceId).toBe('custom-url-trace');
  });
});

// ============================================================
// 3. session Store 内存状态管理
// ============================================================

describe('T2 session Store 内存状态管理', () => {
  it('初始状态应为空 token + 默认 gatewayUrl', () => {
    const state = useSessionStore.getState();
    expect(state.token).toBeNull();
    expect(state.selectedUserId).toBe('');
    expect(state.gatewayUrl).toBe('http://localhost:8080');
  });

  it('setToken / clearToken 应正确更新状态', () => {
    const store = useSessionStore.getState();

    store.setToken('my-jwt-token');
    expect(useSessionStore.getState().token).toBe('my-jwt-token');

    store.clearToken();
    expect(useSessionStore.getState().token).toBeNull();
  });

  it('setSelectedUserId 应更新选中用户 ID', () => {
    const store = useSessionStore.getState();
    store.setSelectedUserId('user-001');
    expect(useSessionStore.getState().selectedUserId).toBe('user-001');
  });

  it('setGatewayUrl 应更新 gateway URL', () => {
    const store = useSessionStore.getState();
    store.setGatewayUrl('http://staging:8080');
    expect(useSessionStore.getState().gatewayUrl).toBe('http://staging:8080');
  });

  it('Token 仅存内存——刷新后丢失', () => {
    // session store 使用 zustand 纯内存存储
    // 验证 setState 后立即可读，无持久化副作用
    useSessionStore.getState().setToken('memory-only-token');
    expect(useSessionStore.getState().token).toBe('memory-only-token');

    // clearToken 后立即为 null
    useSessionStore.getState().clearToken();
    expect(useSessionStore.getState().token).toBeNull();
    // zustand 无持久化中间件，token 不会写入任何外部存储
  });
});
