/**
 * CLI trace 命令实现
 *
 * 后端 API 未就绪时降级为本地报告 + 提示手动检索。
 * 退出码 4（API 不可达），T8 接入后端 API 后改为正常退出码。
 */
import fs from 'fs';
import path from 'path';

/** trace 命令选项 */
interface TraceOptions {
  id: string;
  since?: string;
  outputDir?: string;
}

export async function traceCommand(opts: TraceOptions): Promise<void> {
  console.log(`► TraceID: ${opts.id}`);
  console.log(`► 时间窗口: ${opts.since ?? '5m'}`);

  // CLI 无浏览器环境，IndexedDB 不可用
  console.log('► 本地历史: CLI 模式暂不支持 IndexedDB 查询');
  console.log('► 服务端日志: 待接入 /api/dev/trace API');
  console.log('');
  console.log('提示: 请手动在 Web UI 中搜索此 traceId，或等待后端 API 就绪');

  // 写入简短 Markdown 报告
  const reportDir = opts.outputDir ?? './proto-tester-reports';
  fs.mkdirSync(reportDir, { recursive: true });

  const reportPath = path.join(reportDir, `trace-${opts.id}.md`);
  fs.writeFileSync(
    reportPath,
    [
      '# Trace Report',
      '',
      `TraceID: ${opts.id}`,
      `Since: ${opts.since ?? '5m'}`,
      'Status: pending (API not available)',
      '',
    ].join('\n'),
    'utf-8',
  );
  console.log(`► 报告: ${reportPath}`);

  process.exit(4); // API 不可达，T8 实现后移除此退出码
}
