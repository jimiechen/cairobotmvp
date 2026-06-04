/**
 * browser.ts - Playwright 浏览器控制模块
 *
 * 职责：
 * - 延迟加载 Playwright（optionalDependencies）
 * - 提供 launchBrowser() 启动 Chromium 实例
 * - Playwright 缺失时给出明确的安装指引
 *
 * 仅被 capture 命令使用。send/run/trace 不引入此模块。
 *
 * 不负责：
 * - 场景步骤执行（由 commands/capture.ts 负责）
 * - 浏览器交互逻辑
 */

let playwright: typeof import('playwright') | null = null;

/**
 * 延迟加载 Playwright 模块
 *
 * 首次调用时尝试 import，失败则抛出含安装指引的错误。
 * 后续调用直接返回缓存的实例。
 *
 * @returns Playwright 模块命名空间
 * @throws Error 含安装指引（退出码提示为 5）
 */
async function getPlaywright(): Promise<typeof import('playwright')> {
  if (!playwright) {
    try {
      playwright = await import('playwright');
    } catch {
      throw new Error(
        'Playwright 未安装。请执行: pnpm add -D playwright && npx playwright install chromium\n' +
          '退出码: 5',
      );
    }
  }
  return playwright;
}

/**
 * 启动 Chromium 浏览器实例
 *
 * @param headless - 是否无头模式运行，默认 true
 * @returns Browser 实例，调用方需负责关闭
 */
export async function launchBrowser(headless = true) {
  const pw = await getPlaywright();
  return await pw.chromium.launch({ headless });
}
