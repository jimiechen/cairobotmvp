/**
 * CLI send 命令实现
 *
 * 复用共享层：
 * - encodePacket() from @/lib/messagePacket
 * - sendRequest() from @/lib/apiClient
 *
 * 输出：终端 info 级别摘要 + Markdown 单次报告文件
 * 退出码：0=成功 / 1=业务失败 / 2=传输失败 / 3=参数错误/prod拦截
 */
import fs from 'fs';
import path from 'path';
import { encodePacket } from '../../lib/messagePacket.js';
import { sendRequest, ProtoTesterError } from '../../lib/apiClient.js';
import { getAllProtocols } from '../../lib/protoMetadata.js';
import type { ProtocolMeta } from '../../lib/protoMetadata.js';

/** send 命令选项 */
interface SendOptions {
  max?: string;
  min?: string;
  payload?: string;
  user?: string;
  env?: string;
  gateway?: string;
  token?: string;
  outputDir?: string;
}

export async function sendCommand(opts: SendOptions): Promise<void> {
  // 1. 参数校验：必填 max / min
  if (!opts.max || !opts.min) {
    console.error('错误: 必须指定 --max 和 --min');
    process.exit(3);
  }

  const maxType = parseInt(opts.max, 10);
  const minType = parseInt(opts.min, 10);

  if (isNaN(maxType) || isNaN(minType)) {
    console.error('错误: --max 和 --min 必须是数字');
    process.exit(3);
  }

  // 2. 查找协议元数据
  const protocol = getAllProtocols().find(
    (p: ProtocolMeta) => p.maxType === maxType && p.minType === minType,
  );

  if (!protocol) {
    console.error(`错误: 协议 ${maxType}/${minType} 未注册`);
    process.exit(3);
  }

  // 3. 解析 payload JSON
  let payloadObj: Record<string, unknown> = {};
  try {
    payloadObj = opts.payload ? JSON.parse(opts.payload) : {};
  } catch (_e) {
    console.error('错误: --payload 不是有效 JSON');
    process.exit(3);
  }

  // 4. 编码 MessagePacket（复用共享层，禁止在 cli/ 下重新实现）
  let packetBinary: Uint8Array;
  try {
    packetBinary = encodePacket({
      maxType,
      minType,
      payload: new Uint8Array(Buffer.from(JSON.stringify(payloadObj))),
      extend: { method: `${protocol.requestMessage}` },
    });
  } catch (e) {
    console.error(`错误: MessagePacket 编码失败: ${(e as Error).message}`);
    process.exit(3);
  }

  // 5. 发送请求（复用共享层 apiClient）
  const startTime = Date.now();
  try {
    const response = await sendRequest({
      maxType,
      minType,
      payload: packetBinary,
      gatewayUrl: opts.gateway,
      token: opts.token,
    });
    const durationMs = Date.now() - startTime;

    // 6. 输出终端摘要
    console.log(`► 协议: ${maxType}/${minType} ${protocol.name}`);
    console.log(`► 用户: ${opts.user ?? 'user_001'}`);
    console.log(`► Gateway: ${opts.gateway ?? 'http://localhost:8080'}`);
    console.log(`► 请求大小: ${packetBinary.length}B`);
    console.log(`► 响应大小: ${response.responseData.length}B`);
    console.log(`► 耗时: ${durationMs}ms`);
    console.log(
      `► 业务码: ${response.businessCode} (${response.businessCode === 10200 ? '成功' : '失败'})`,
    );

    // 7. 写入 Markdown 报告文件
    const reportDir = opts.outputDir ?? './proto-tester-reports';
    const reportPath = path.join(reportDir, `send-${maxType}-${minType}-${formatTimestamp()}.md`);

    fs.mkdirSync(reportDir, { recursive: true });

    const reportContent = generateSingleReport({
      protocol,
      maxType,
      minType,
      payload: payloadObj,
      response,
      durationMs,
    });

    fs.writeFileSync(reportPath, reportContent, 'utf-8');
    console.log(`► 详细报告: ${reportPath}`);

    // 退出码：0=成功 / 1=业务失败
    process.exit(response.businessCode === 10200 ? 0 : 1);
  } catch (e) {
    const durationMs = Date.now() - startTime;
    if (e instanceof ProtoTesterError) {
      console.error(`✗ 传输失败 [code=${e.code}]: ${e.message}`);
    } else {
      console.error(`✗ 传输失败: ${(e as Error).message}`);
    }
    console.log(`► 耗时: ${durationMs}ms`);
    process.exit(2);
  }
}

/**
 * 生成单次发送的 Markdown 报告内容
 */
function generateSingleReport(data: {
  protocol: ProtocolMeta;
  maxType: number;
  minType: number;
  payload: Record<string, unknown>;
  response: { status: number; businessCode: number; responseData: Uint8Array };
  durationMs: number;
}): string {
  return [
    '# Single Send Report',
    '',
    '## 元数据',
    `- 时间: ${new Date().toISOString()}`,
    `- 协议: ${data.maxType}/${data.minType} ${data.protocol.name}`,
    '',
    '## 请求',
    '```json',
    JSON.stringify(data.payload, null, 2),
    '```',
    '',
    '## 响应',
    `- HTTP Status: ${data.response.status}`,
    `- 业务码: ${data.response.businessCode}`,
    `- 耗时: ${data.durationMs}ms`,
    '',
  ].join('\n');
}

/** 格式化时间戳用于文件名 */
function formatTimestamp(): string {
  return new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
}
