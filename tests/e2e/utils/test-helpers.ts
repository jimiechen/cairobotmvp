import { Page, BrowserContext, expect } from '@playwright/test'
import * as fs from 'fs'
import * as path from 'path'

const FIXTURES_DIR = path.resolve(__dirname, '../fixtures')
const MOCK_TOKEN = 'e2e-mock-jwt-token-for-testing'

export function loadTestData<T>(filename: string): T {
  const filePath = path.join(FIXTURES_DIR, filename)
  const raw = fs.readFileSync(filePath, 'utf-8')
  return JSON.parse(raw) as T
}

export function buildSuccessResponse(data: unknown): object {
  return { code: 200, msg: '操作成功', data }
}

export function buildErrorResponse(code: number, msg: string, errors?: unknown[]): object {
  const result: Record<string, unknown> = { code, msg }
  if (errors && errors.length > 0) {
    result.errors = errors
  }
  return result
}

export async function setupMockAPI(page: Page, testData: Record<string, unknown>): Promise<void> {
  await page.route('**/api/v1/login', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ code: 200, data: { token: MOCK_TOKEN } }),
    })
  })

  await page.route('**/api/v1/getinfo', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 200,
        data: {
          id: 1,
          nickName: 'E2E测试管理员',
          avatar: '',
          name: 'admin',
          userId: 1,
          roles: ['admin'],
          permissions: ['*:*:*'],
        },
      }),
    })
  })

  await page.route('**/api/v1/getmenu**', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 200,
        data: [
          { id: 1, parentId: 0, path: '/config/schema', name: 'Schema管理', component: 'config/schema-list' },
          { id: 2, parentId: 0, path: '/config/value', name: '配置发布', component: 'config/value-publish' },
          { id: 3, parentId: 0, path: '/i18n/string', name: '字符串管理', component: 'i18n/string-list' },
          { id: 4, parentId: 0, path: '/i18n/pack', name: '语言包管理', component: 'i18n/pack-manage' },
          { id: 5, parentId: 0, path: '/i18n/import-export', name: '导入导出', component: 'i18n/import-export' },
        ],
      }),
    })
  })

  await page.route('**/api/v1/menurole', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 200,
        data: [
          {
            path: '/config',
            component: 'Layout',
            redirect: '/config/schema',
            menuName: 'ConfigAdmin',
            title: '配置管理',
            icon: 'setting',
            visible: '1',
            noCache: false,
            children: [
              {
                path: 'schema',
                component: '/config/schema-list',
                menuName: 'SchemaList',
                title: 'Schema列表',
                icon: 'list',
                visible: '1',
                noCache: false,
              },
              {
                path: 'value',
                component: '/config/value-publish',
                menuName: 'ValuePublish',
                title: '配置发布',
                icon: 'edit',
                visible: '1',
                noCache: false,
              },
            ],
          },
          {
            path: '/i18n',
            component: 'Layout',
            redirect: '/i18n/string',
            menuName: 'I18nAdmin',
            title: '国际化管理',
            icon: 'language',
            visible: '1',
            noCache: false,
            children: [
              {
                path: 'string',
                component: '/i18n/string-list',
                menuName: 'StringList',
                title: '字符串列表',
                icon: 'document',
                visible: '1',
                noCache: false,
              },
              {
                path: 'pack',
                component: '/i18n/pack-manage',
                menuName: 'PackManage',
                title: '语言包管理',
                icon: 'package',
                visible: '1',
                noCache: false,
              },
              {
                path: 'import-export',
                component: '/i18n/import-export',
                menuName: 'ImportExport',
                title: '导入导出',
                icon: 'download',
                visible: '1',
                noCache: false,
              },
            ],
          },
        ],
      }),
    })
  })

  await page.route('**/api/v1/app-config', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 200,
        data: {
          sys_app_name: 'CaiRobot Admin',
          sys_app_logo: '',
          sys_copyright: '© 2026 CaiRobot MVP',
          sys_icp: '',
          sys_version: 'v1.0.0',
        },
      }),
    })
  })

  await page.route('**/api/admin/v1/config/schema**', (route) => {
    const method = route.request().method()
    if (method === 'GET') {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(buildSuccessResponse(testData.config_schemas)),
      })
    } else if (method === 'POST') {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(buildSuccessResponse({ id: 99 })),
      })
    } else if (method === 'PUT') {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(buildSuccessResponse(null)),
      })
    } else if (method === 'DELETE') {
      const url = new URL(route.request().url())
      const idParam = url.searchParams.get('id')
      if (idParam && isNaN(Number(idParam))) {
        route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({ code: 400, msg: '参数错误：ID 必须为数字' }),
        })
        return
      }
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(buildSuccessResponse(null)),
      })
    }
  })

  await page.route('**/api/admin/v1/config/value/publish**', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(buildSuccessResponse(testData.i18n_pack_result)),
    })
  })

  await page.route('**/api/admin/v1/config/value/versions**', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(buildSuccessResponse({ versions: testData.config_versions })),
    })
  })

  await page.route('**/api/admin/v1/i18n/string**', (route) => {
    const method = route.request().method()
    if (method === 'GET') {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(buildSuccessResponse(testData.i18n_strings)),
      })
    } else if (method === 'POST') {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(buildSuccessResponse({ id: 99 })),
      })
    } else if (method === 'PUT') {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(buildSuccessResponse(null)),
      })
    } else if (method === 'DELETE') {
      const url = new URL(route.request().url())
      const idParam = url.searchParams.get('id')
      if (idParam && isNaN(Number(idParam))) {
        route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({ code: 400, msg: '参数错误：ID 必须为数字' }),
        })
        return
      }
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(buildSuccessResponse(null)),
      })
    }
  })

  await page.route('**/api/admin/v1/i18n/pack/publish**', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(buildSuccessResponse(testData.i18n_pack_result)),
    })
  })

  await page.route('**/api/admin/v1/i18n/pack/rollback**', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(buildSuccessResponse(null)),
    })
  })

  await page.route('**/api/admin/v1/i18n/import/csv**', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(testData.import_result_success),
    })
  })

  await page.route('**/api/admin/v1/i18n/export/csv**', (route) => {
    const csvContent = 'string_key,string_value,group_name,template_type\ngreeting.hello,你好，世界！,common,plain\nerror.notFound,页面未找到,error,plain\n'
    route.fulfill({
      status: 200,
      contentType: 'text/csv; charset=utf-8',
      headers: { 'Content-Disposition': 'attachment; filename="strings_1.csv"' },
      body: csvContent,
    })
  })
}

export async function overrideMockRoute(
  page: Page,
  urlPattern: string | RegExp,
  handler: (route: import('@playwright/test').Route) => void
): Promise<void> {
  await page.route(urlPattern, handler)
}

export async function waitForElementVisible(page: Page, dataId: string, timeout = 5000): Promise<void> {
  await page.locator(`[data-id="${dataId}"]`).waitFor({ state: 'visible', timeout })
}

export async function fillInputByDataId(page: Page, dataId: string, value: string): Promise<void> {
  const el = page.locator(`[data-id="${dataId}"]`)
  await el.fill(value)
}

export async function selectOptionByDataId(page: Page, dataId: string, value: string): Promise<void> {
  const el = page.locator(`[data-id="${dataId}"]`)
  await el.selectOption(value)
}

export async function checkDevServerRunning(baseURL: string): Promise<boolean> {
  try {
    const response = await fetch(baseURL, { signal: AbortSignal.timeout(3000) })
    return response.ok || response.status < 500
  } catch {
    return false
  }
}

export async function autoLogin(page: Page): Promise<void> {
  await page.context().addCookies([
    {
      name: 'Admin-Token',
      value: MOCK_TOKEN,
      domain: 'localhost',
      path: '/',
    },
  ])
}

export async function navigateToRoute(page: Page, routePath: string): Promise<void> {
  await page.goto(routePath, { waitUntil: 'domcontentloaded' })
  await page.waitForTimeout(2000)
}

export async function selectElOption(page: Page, dataId: string, value: string): Promise<void> {
  const select = page.locator(dataId)
  await select.click()
  await page.waitForTimeout(500)
  const dropdownItems = page.locator('.el-select-dropdown:visible .el-select-dropdown__item')
  const count = await dropdownItems.count()
  for (let i = 0; i < count; i++) {
    const text = await dropdownItems.nth(i).textContent()
    if (text && text.trim() === value) {
      await dropdownItems.nth(i).click()
      await page.waitForTimeout(300)
      return
    }
  }
  const fallback = page.locator('.el-select-dropdown:visible .el-select-dropdown__item:has-text("' + value + '")').first()
  if (await fallback.count() > 0) {
    await fallback.click()
  }
  await page.waitForTimeout(300)
}

export async function fillElInputNumber(page: Page, dataId: string, value: string): Promise<void> {
  const input = page.locator(`[data-id="${dataId}"] input`)
  await input.click({ force: true })
  await input.fill(value, { force: true })
  await input.press('Tab')
  await page.waitForTimeout(500)
}
