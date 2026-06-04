/**
 * CLI run 命令 — 调用 runner 执行 YAML 用例集
 *
 * 职责：
 * - 解析命令行参数（--suite, --parallel, --env, --gateway）
 * - 参数校验与前置检查
 * - 调用 runSuite() 执行用例
 * - 输出执行摘要和退出码
 *
 * 不负责：
 * - 用例执行细节（由 runner.ts 负责）
 * - Schema 校验（由 schema.ts 负责）
 */
import { runSuite } from '../runner.js';

/** run 命令选项接口 */
interface RunOptions {
  suite?: string;
  parallel?: string;
  env?: string;
  gateway?: string;
}

/**
 * 执行 run 命令
 *
 * @param opts - 从 commander 解析的选项
 * @throws 通过 process.exit 退出，不抛异常
 */
export async function runCommand(opts: RunOptions): Promise<void> {
  // 1. 必填参数校验
  if (!opts.suite) {
    console.error('错误: 必须指定 --suite <file>');
    process.exit(3);
  }

  // 2. 文件格式校验
  if (!opts.suite.endsWith('.yaml') && !opts.suite.endsWith('.yml')) {
    console.error('错误: --suite 文件必须是 .yaml 或 .yml 格式');
    process.exit(3);
  }

  const parallel = parseInt(opts.parallel ?? '1', 10);

  console.log(`► 用例集: ${opts.suite}`);
  console.log(`► 并行度: ${parallel}`);
  console.log('');

  // 3. 执行用例集
  const results = await runSuite(opts.suite, {
    parallel,
    env: opts.env,
    gateway: opts.gateway,
  });

  // 4. 输出摘要
  const passed = results.filter((r) => r.passed).length;
  const failed = results.filter((r) => !r.passed).length;
  console.log('');
  console.log('═══════════════════════════');
  console.log(`总计: ${results.length} | 通过: ${passed} | 失败: ${failed}`);
  console.log('═══════════════════════════');

  // 5. 退出码：0=全通过 / 1=有失败
  process.exit(failed > 0 ? 1 : 0);
}
