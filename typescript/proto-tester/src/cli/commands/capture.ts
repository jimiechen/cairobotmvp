/**
 * CLI capture 命令 — 浏览器自动化场景执行
 *
 * 职责：
 * - 检查 vite dev server 是否运行（未启动时优雅失败）
 * - 加载 YAML 场景文件并校验格式
 * - 启动 Playwright 浏览器按步骤执行场景
 * - 支持录屏和截图输出
 * - 生成 Capture Markdown 报告
 *
 * 前提条件：vite dev server 必须已在 http://127.0.0.1:3001 运行
 *
 * 退出码约定：
 * - 0 = 成功
 * - 3 = 参数错误 / dev server 未运行
 * - 5 = Playwright 缺失
 */
import fs from 'fs';
import path from 'path';
import yaml from 'yaml';
import { launchBrowser } from '../browser.js';
import { validateScenario } from '../schema.js';

/** capture 命令选项接口 */
interface CaptureOptions {
  scenario?: string;
  video?: string;
  screenshot?: string;
}

/**
 * 执行 capture 命令
 *
 * @param opts - 从 commander 解析的选项
 * @throws 通过 process.exit 退出，不抛异常
 */
export async function captureCommand(opts: CaptureOptions): Promise<void> {
  // 1. 必填参数校验
  if (!opts.scenario) {
    console.error('错误: 必须指定 --scenario <file>');
    process.exit(3);
  }

  // 2. 检查 vite dev server 是否运行
  const isDevServerRunning = await checkDevServer();
  if (!isDevServerRunning) {
    console.error('错误: vite dev server 未在 http://127.0.0.1:3001 运行');
    console.error('请先执行: pnpm --filter proto-tester dev');
    process.exit(3);
  }

  // 3. 加载并校验场景文件
  let scenarioData: unknown;
  try {
    const raw = fs.readFileSync(opts.scenario, 'utf-8');
    scenarioData = yaml.parse(raw);
  } catch (e) {
    console.error(`错误: 场景文件解析失败: ${(e as Error).message}`);
    process.exit(3);
  }

  const validation = validateScenario(scenarioData);
  if (!validation.valid) {
    console.error(`错误: 场景格式不合法:\n${validation.errors?.join('\n')}`);
    process.exit(3);
  }

  // 4. 启动浏览器
  const videoOn = (opts.video ?? 'on') === 'on';
  let browser;
  try {
    browser = await launchBrowser(/* headless */ true);
  } catch (e) {
    console.error((e as Error).message);
    process.exit(5); // Playwright 缺失
  }

  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
    video: videoOn ? { dir: './proto-tester-reports/capture-videos' } : undefined,
  });

  const page = await context.newPage();

  try {
    // 导航到 proto-tester Web UI
    await page.goto('http://127.0.0.1:3001/', { waitUntil: 'networkidle' });

    // 执行场景步骤
    const steps = (scenarioData as { steps: unknown[] }).steps ?? [];
    for (const step of steps) {
      await executeStep(page, step, opts);
    }

    console.log('✓ 场景执行完成');
  } finally {
    await context.close();
    await browser.close();
  }

  // 5. 生成报告
  const { generateCaptureReport } = await import('../reporter.js');
  const report = generateCaptureReport(opts.scenario!, []);
  const dir = './proto-tester-reports';
  fs.mkdirSync(dir, { recursive: true });
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
  fs.writeFileSync(path.join(dir, `capture-${timestamp}.md`), report, 'utf-8');
  console.log(`► 报告: ${dir}/capture-${timestamp}.md`);

  process.exit(0);
}

/**
 * 检查 vite dev server 是否可达
 *
 * 使用原生 fetch 探测 http://127.0.0.1:3001/
 *
 * @returns dev server 是否在运行
 */
async function checkDevServer(): Promise<boolean> {
  try {
    const res = await fetch('http://127.0.0.1:3001/');
    return res.ok || res.status < 500;
  } catch {
    return false;
  }
}

/**
 * 执行单个场景步骤
 *
 * 支持 action 类型：
 * - open: 导航到指定路径
 * - select-protocol: 选择协议（UI 操作预留）
 * - switch-user: 切换测试用户
 * - fill-form: 填写表单字段
 * - send: 点击发送按钮
 * - wait-response: 等待响应
 * - screenshot: 截图保存
 *
 * @param page - Playwright Page 实例
 * @param step - 步骤定义对象
 * @param opts - capture 命令选项（截图开关等）
 */
async function executeStep(
  page: import('playwright').Page,
  step: Record<string, unknown>,
  opts: CaptureOptions,
): Promise<void> {
  const action = step.action as string;

  switch (action) {
    case 'open':
      await page.goto(
        step.path ? `http://127.0.0.1:3001${step.path}` : 'http://127.0.0.1:3001/',
        { waitUntil: 'networkidle' },
      );
      break;

    case 'select-protocol': {
      // 通过 data-id 锚点操作协议树区域
      await page.click('[data-id="pt-protocol-tree"]');
      console.log(`  → 选择协议: ${step.maxType}/${step.minType}`);
      break;
    }

    case 'switch-user':
      await page.click('[data-id="pt-user-switcher"]');
      console.log(`  → 切换用户: ${step.userId}`);
      break;

    case 'fill-form': {
      const field = String(step.field ?? '');
      const input = page.locator(
        `[data-id="pt-form-${field}"] input, [data-id="pt-form-${field}"] textarea`,
      );
      await input.fill(String(step.value ?? ''));
      console.log(`  → 填写字段: ${field} = ${step.value}`);
      break;
    }

    case 'send':
      await page.click('[data-id="pt-btn-send"]');
      console.log('  → 点击发送');
      break;

    case 'wait-response':
      await page.waitForTimeout(Number(step.timeoutMs ?? 5000));
      break;

    case 'screenshot':
      if (opts.screenshot !== 'off') {
        const name = String(step.screenshot ?? `step-${Date.now()}`);
        const screenshotDir = './proto-tester-reports/screenshots';
        fs.mkdirSync(screenshotDir, { recursive: true });
        await page.screenshot({
          path: `${screenshotDir}/${name}.png`,
          fullPage: false,
        });
        console.log(`  → 截图: ${name}.png`);
      }
      break;

    default:
      console.warn(`  ⚠ 未知动作: ${action}`);
  }
}
