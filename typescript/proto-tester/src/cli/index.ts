#!/usr/bin/env node
/**
 * proto-tester CLI 入口
 *
 * 注册 4 个子命令：send / run / capture / trace
 * 全局参数：--env / --gateway / --user / --token / --output-dir / --log-level / --no-color
 *
 * 退出码约定：
 * - 0 = 成功
 * - 1 = 业务失败（业务码非 10200）
 * - 2 = 传输失败（网络/超时）
 * - 3 = 参数错误 / prod 被拦截
 * - 4 = API 不可达（trace 等待实现）
 * - 5 = 未分类错误
 */
import { Command } from 'commander';
import { readFileSync } from 'fs';
import { resolve } from 'path';

const program = new Command();

program
  .name('proto-tester')
  .description('协议测试客户端 — MessagePacket + Protobuf 联调工具')
  .version('2.0.0');

// 全局参数定义
const globalOpts = program
  .option('-e, --env <name>', '环境名（dev/test/staging）', 'dev')
  .option('-g, --gateway <url>', '覆盖 Gateway URL')
  .option('-u, --user <id>', '测试用户 ID', 'user_001')
  .option('-t, --token <jwt>', '直接传入 JWT Token')
  .option('-o, --output-dir <path>', '报告输出目录', './proto-tester-reports')
  .option('--log-level <level>', '日志级别', 'info')
  .option('--no-color', '禁用彩色输出');

// send 子命令：发送单次协议请求
globalOpts
  .command('send')
  .description('发送单次协议请求')
  .option('--max <number>', '协议大类编号（maxType）')
  .option('--min <number>', '协议小类编号（minType）')
  .option('--payload <json>', '请求字段 JSON', '{}')
  .action(async (opts) => {
    // blocklist 校验：禁止 prod 环境
    const envName = program.opts().env;
    if (isBlocked(envName)) {
      console.error(`错误: 环境 "${envName}" 在拦截列表中，禁止使用`);
      process.exit(3);
    }
    const { sendCommand } = await import('./commands/send.js');
    await sendCommand({ ...opts, ...program.opts() });
  });

// trace 子命令：根据 traceId 聚合服务端日志
globalOpts
  .command('trace')
  .description('根据 traceId 聚合服务端日志')
  .requiredOption('-i, --id <traceId>', 'traceId')
  .option('--since <window>', '时间窗口（如 5m / 1h）', '5m')
  .action(async (opts) => {
    const { traceCommand } = await import('./commands/trace.js');
    await traceCommand({ ...opts, outputDir: program.opts().outputDir });
  });

// run 子命令 — 执行 YAML 测试用例集
globalOpts
  .command('run')
  .description('运行 YAML 测试用例集')
  .option('--suite <file>', 'YAML 用例文件路径')
  .option('--parallel <number>', '并行执行数量（默认 1，最大 4）', '1')
  .action(async (opts) => {
    // blocklist 校验：禁止 prod 环境
    const envName = program.opts().env;
    if (isBlocked(envName)) {
      console.error(`错误: 环境 "${envName}" 在拦截列表中，禁止使用`);
      process.exit(3);
    }
    const { runCommand } = await import('./commands/run.js');
    await runCommand({ ...opts, ...program.opts() });
  });

// capture 子命令 — 浏览器自动化场景执行
globalOpts
  .command('capture')
  .description('浏览器自动化场景执行')
  .option('--scenario <file>', 'YAML 场景文件路径')
  .option('--video <mode>', '录屏开关 on/off', 'on')
  .option('--screenshot <mode>', '截图开关 on/off', 'on')
  .action(async (opts) => {
    // blocklist 校验：禁止 prod 环境
    const envName = program.opts().env;
    if (isBlocked(envName)) {
      console.error(`错误: 环境 "${envName}" 在拦截列表中，禁止使用`);
      process.exit(3);
    }
    const { captureCommand } = await import('./commands/capture.js');
    await captureCommand({ ...opts });
  });

program.parse();

/**
 * 判断目标环境是否在 blocklist 中
 */
function isBlocked(envName: string): boolean {
  try {
    const endpointsPath = resolve(import.meta.dirname ?? '..', '../public/config/endpoints.json');
    const raw = readFileSync(endpointsPath, 'utf-8');
    const config = JSON.parse(raw);
    return Array.isArray(config.blocklist) && config.blocklist.includes(envName);
  } catch {
    return false;
  }
}
