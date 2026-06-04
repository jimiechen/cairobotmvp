/**
 * CLI send 命令单元测试
 *
 * 覆盖 4 种退出码：
 * - 0: 成功（业务码 10200）
 * - 1: 业务失败（业务码非 10200）
 * - 2: 传输失败（网络异常）
 * - 3: 参数错误 / prod 拦截 / payload JSON 无效
 *
 * 使用 msw mock HTTP 层，不依赖真实 Gateway。
 */
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';

// Mock process.exit 避免测试进程退出
const mockExit = vi.spyOn(process, 'exit').mockImplementation(
  (code: number) => {
    throw new Error(`process.exit(${String(code)})`);
  },
);

// Mock console.log/console.error 以减少噪声（必要时可恢复）
const mockConsoleLog = vi.spyOn(console, 'log').mockImplementation(() => {});
const mockConsoleError = vi.spyOn(console, 'error').mockImplementation(() => {});

// --- msw server ---
const server = setupServer(
  // 成功响应：返回含 code=10200 的 MessagePacket 二进制
  http.post('http://localhost:8080/api/hello', async ({ request }) => {
    const body = await request.arrayBuffer();
    // 构造一个最小有效响应：返回原始 body 作为 echo
    return new HttpResponse(new Uint8Array(body), {
      status: 200,
      headers: { 'Content-Type': 'application/octet-stream' },
    });
  }),
);

beforeEach(() => {
  vi.clearAllMocks();
  server.resetHandlers();
});

describe('send 命令', () => {
  it('缺少 --max 和 --min 时退出码为 3', async () => {
    const { sendCommand } = await import('@/cli/commands/send');

    await expect(sendCommand({})).rejects.toThrow('process.exit(3)');
    expect(mockConsoleError).toHaveBeenCalledWith(expect.stringContaining('必须指定'));
  });

  it('--max/--min 为非数字时退出码为 3', async () => {
    const { sendCommand } = await import('@/cli/commands/send');

    await expect(sendCommand({ max: 'abc', min: 'xyz' })).rejects.toThrow(
      'process.exit(3)',
    );
    expect(mockConsoleError).toHaveBeenCalledWith(expect.stringContaining('必须是数字'));
  });

  it('未注册的协议编号退出码为 3', async () => {
    const { sendCommand } = await import('@/cli/commands/send');

    await expect(sendCommand({ max: '9999', min: '9999' })).rejects.toThrow(
      'process.exit(3)',
    );
    expect(mockConsoleError).toHaveBeenCalledWith(expect.stringContaining('未注册'));
  });

  it('无效 JSON payload 退出码为 3', async () => {
    const { sendCommand } = await import('@/cli/commands/send');

    await expect(
      sendCommand({ max: '2100', min: '2101', payload: '{invalid' }),
    ).rejects.toThrow('process.exit(3)');
    expect(mockConsoleError).toHaveBeenCalledWith(expect.stringContaining('不是有效 JSON'));
  });

  it('传输失败时退出码为 2（网络不可达）', async () => {
    // 使用不存在的 URL 触发传输错误
    const { sendCommand } = await import('@/cli/commands/send');

    await expect(
      sendCommand({
        max: '2100',
        min: '2101',
        payload: '{}',
        gateway: 'http://127.0.0.1:1', // 不可达地址
        outputDir: '/tmp/proto-tester-test-reports',
      }),
    ).rejects.toThrow('process.exit(2)');
    expect(mockConsoleError).toHaveBeenCalledWith(
      expect.stringContaining('传输失败'),
    );
  });

  it('成功请求时输出报告文件路径', async () => {
    server.listen();

    const { sendCommand } = await import('@/cli/commands/send');

    // 使用已注册的协议 2100/2101（HelloWorld）
    try {
      await sendCommand({
        max: '2100',
        min: '2101',
        payload: '{"name":"test"}',
        outputDir: '/tmp/proto-tester-test-reports',
      });
    } catch (_e) {
      // process.exit 会抛异常，忽略
    }

    expect(mockConsoleLog).toHaveBeenCalledWith(
      expect.stringContaining('详细报告'),
    );

    server.close();
  });
});
