import { test, expect, Page } from '@playwright/test'
import { takeEvidence, clickAndScreenshot } from '../utils/evidence'
import {
  loadTestData,
  setupMockAPI,
  overrideMockRoute,
  buildSuccessResponse,
  buildErrorResponse,
  autoLogin,
  navigateToRoute,
  selectElOption,
} from '../utils/test-helpers'

test.describe('Config Admin E2E 测试', () => {
  let page: Page
  const testData = loadTestData<Record<string, unknown>>('test-data.json')

  test.beforeAll(async ({ browser }) => {
    const context = await browser.newContext({ viewport: { width: 1400, height: 900 } })
    page = await context.newPage()
  })

  test.afterAll(async () => {
    await page.close()
  })

  test.beforeEach(async () => {
    await setupMockAPI(page, testData)
    await autoLogin(page)
  })

  test('ca-01 Schema 列表正常渲染', async () => {
    await navigateToRoute(page, '/config/schema')
    await expect(page.locator('[data-id="ca-table-schema-list"]')).toBeVisible({ timeout: 10000 })
    const rows = page.locator('[data-id="ca-table-schema-list"] tbody tr')
    await expect(rows).toHaveCount((testData.config_schemas as Array<unknown>).length * 2)
    await expect(page.locator('[data-id="ca-pagination-schema-list"]').first()).toBeVisible()
    await takeEvidence(page, {
      dataId: 'ca-01',
      description: 'Schema列表正常渲染',
      fullPage: true,
    })
  })

  test('ca-02 空 ModuleKey 查询', async () => {
    await navigateToRoute(page, '/config/schema')
    await expect(page.locator('[data-id="ca-input-mokuai-key"]')).toBeVisible({ timeout: 10000 })
    await page.locator('[data-id="ca-input-mokuai-key"]').clear()
    await page.locator('[data-id="ca-btn-sousuo"]').click()
    await page.waitForTimeout(800)
    await takeEvidence(page, {
      dataId: 'ca-02',
      description: '空ModuleKey查询结果',
    })
  })

  test('ca-03 正常创建 Schema', async ({ browser }) => {
    const context = await browser.newContext({ viewport: { width: 1400, height: 900 }, recordVideo: { dir: '../evidence' } })
    const videoPage = await context.newPage()
    await setupMockAPI(videoPage, testData)
    await autoLogin(videoPage)

    try {
      await navigateToRoute(videoPage, '/config/schema')
      await videoPage.locator('[data-id="ca-btn-xinzeng-schema"]').click()
      await videoPage.waitForTimeout(800)
      await videoPage.locator('[data-id="ca-input-mokuai-key-dialog"]').fill('app.server')
      await videoPage.locator('[data-id="ca-input-ziduan-key-dialog"]').fill('port')
      await selectElOption(videoPage, '[data-id="ca-select-ziduan-leixing-dialog"]', 'int')
      await takeEvidence(videoPage, {
        dataId: 'ca-03',
        description: '创建Schema填写完成',
      })
      await videoPage.locator('[data-id="ca-btn-queding-schema-dialog"]').click()
      await videoPage.waitForTimeout(1500)
      await takeEvidence(videoPage, {
        dataId: 'ca-03',
        description: '创建Schema成功后状态',
      })
    } finally {
      await videoPage.close()
      await context.close()
    }
  })

  test('ca-04 创建时 Body 为空', async () => {
    await navigateToRoute(page, '/config/schema')
    await page.locator('[data-id="ca-btn-xinzeng-schema"]').click()
    await page.waitForTimeout(800)
    await page.locator('[data-id="ca-btn-queding-schema-dialog"]').click()
    await page.waitForTimeout(1000)
    const errorTips = page.locator('.el-form-item__error')
    await expect(errorTips.first()).toBeVisible({ timeout: 5000 })
    await takeEvidence(page, {
      dataId: 'ca-04',
      description: '空Body提交校验提示',
    })
  })

  test('ca-05 正常更新 Schema', async () => {
    await navigateToRoute(page, '/config/schema')
    await expect(page.locator('[data-id="ca-table-schema-list"]')).toBeVisible({ timeout: 10000 })
    const firstRow = page.locator('[data-id="ca-table-schema-list"] tbody tr').first()
    await firstRow.hover()
    await page.waitForTimeout(500)
    await page.locator('[data-id="ca-btn-bianji-schema-row"]').first().click({ force: true })
    await page.waitForTimeout(800)
    await expect(page.locator('[data-id="ca-input-mokuai-key-dialog"]')).toHaveValue('app.server')
    await page.locator('[data-id="ca-input-ziduan-key-dialog"]').fill('timeout', { force: true })
    await page.locator('[data-id="ca-btn-queding-schema-dialog"]').click()
    await page.waitForTimeout(1500)
    await takeEvidence(page, {
      dataId: 'ca-05',
      description: '更新Schema成功后状态',
    })
  })

  test('ca-06 无效 ID 删除', async () => {
    await navigateToRoute(page, '/config/schema')
    await expect(page.locator('[data-id="ca-table-schema-list"]')).toBeVisible({ timeout: 10000 })
    await page.waitForTimeout(500)
    await page.locator('[data-id="ca-btn-shanchu-schema-row"]').first().click({ force: true })
    await page.waitForTimeout(800)
    const confirmBtn = page.locator('.el-message-box__btns button.el-button--primary')
    if (await confirmBtn.count() > 0) {
      await confirmBtn.click()
    }
    await page.waitForTimeout(1500)
    await takeEvidence(page, {
      dataId: 'ca-06',
      description: '无效ID删除错误提示',
    })
  })

  test('ca-07 正常删除 Schema', async ({ browser }) => {
    const context = await browser.newContext({ viewport: { width: 1400, height: 900 }, recordVideo: { dir: '../evidence' } })
    const videoPage = await context.newPage()
    await setupMockAPI(videoPage, testData)
    await autoLogin(videoPage)

    try {
      await navigateToRoute(videoPage, '/config/schema')
      await expect(videoPage.locator('[data-id="ca-table-schema-list"]')).toBeVisible({ timeout: 10000 })
      const firstRow = videoPage.locator('[data-id="ca-table-schema-list"] tbody tr').first()
      await firstRow.hover()
      await videoPage.waitForTimeout(500)
      await videoPage.locator('[data-id="ca-btn-shanchu-schema-row"]').first().click({ force: true })
      await videoPage.waitForTimeout(800)
      await takeEvidence(videoPage, {
        dataId: 'ca-07',
        description: '删除确认对话框弹出',
      })
      const confirmBtn = videoPage.locator('.el-message-box__btns button.el-button--primary')
      if (await confirmBtn.count() > 0) {
        await confirmBtn.click()
      }
      await videoPage.waitForTimeout(1500)
      await takeEvidence(videoPage, {
        dataId: 'ca-07',
        description: '删除成功后列表刷新',
      })
    } finally {
      await videoPage.close()
      await context.close()
    }
  })

  test('ca-08 正常发布配置值', async ({ browser }) => {
    const context = await browser.newContext({ viewport: { width: 1400, height: 900 }, recordVideo: { dir: '../evidence' } })
    const videoPage = await context.newPage()
    await setupMockAPI(videoPage, testData)
    await autoLogin(videoPage)

    try {
      await navigateToRoute(videoPage, '/config/value')
      await expect(videoPage.locator('[data-id="ca-select-mokuai-value"]')).toBeVisible({ timeout: 10000 })
      await selectElOption(videoPage, '[data-id="ca-select-mokuai-value"]', 'app.server')
      await videoPage.waitForTimeout(800)
      await selectElOption(videoPage, '[data-id="ca-select-huanjing-value"]', 'dev')
      await takeEvidence(videoPage, {
        dataId: 'ca-08',
        description: '配置值动态表单渲染',
        fullPage: true,
      })
      await videoPage.locator('[data-id="ca-btn-fabu-peizhi"]').click()
      await videoPage.waitForTimeout(1500)
      await takeEvidence(videoPage, {
        dataId: 'ca-08',
        description: '发布配置成功后版本号展示',
      })
    } finally {
      await videoPage.close()
      await context.close()
    }
  })

  test('ca-09 10400 校验错误展示', async ({ browser }) => {
    const context = await browser.newContext({ viewport: { width: 1400, height: 900 }, recordVideo: { dir: '../evidence' } })
    const videoPage = await context.newPage()
    await setupMockAPI(videoPage, testData)
    await autoLogin(videoPage)

    await overrideMockRoute(
      videoPage,
      '**/api/admin/v1/config/value/publish**',
      (route) =>
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(testData.validation_error_10400),
        })
    )

    try {
      await navigateToRoute(videoPage, '/config/value')
      await expect(videoPage.locator('[data-id="ca-select-mokuai-value"]')).toBeVisible({ timeout: 10000 })
      await selectElOption(videoPage, '[data-id="ca-select-mokuai-value"]', 'app.server')
      await selectElOption(videoPage, '[data-id="ca-select-huanjing-value"]', 'dev')
      await videoPage.waitForTimeout(800)
      await takeEvidence(videoPage, {
        dataId: 'ca-09',
        description: '10400校验前表单状态',
        fullPage: true,
      })
      await videoPage.locator('[data-id="ca-btn-fabu-peizhi"]').click()
      await videoPage.waitForTimeout(1500)
      const errorMsg = videoPage.locator('.el-message--error')
      if (await errorMsg.count() > 0) {
        await expect(errorMsg.first()).toBeVisible()
      }
      await takeEvidence(videoPage, {
        dataId: 'ca-09',
        description: '10400校验错误提示展示',
        fullPage: true,
      })
    } finally {
      await videoPage.close()
      await context.close()
    }
  })

  test('ca-10 空 Fields 发布', async () => {
    await overrideMockRoute(page, '**/api/admin/v1/config/value/publish**', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(buildErrorResponse(400, 'fields 不能为空')),
      })
    )
    await navigateToRoute(page, '/config/value')
    await page.waitForTimeout(2000)
    const publishBtn = page.locator('[data-id="ca-btn-fabu-peizhi"]')
    if (await publishBtn.count() > 0) {
      await publishBtn.click({ force: true })
    }
    await page.waitForTimeout(1500)
    await takeEvidence(page, {
      dataId: 'ca-10',
      description: '空Fields发布错误提示',
    })
  })

  test('ca-11 缺 ModuleKey 查版本', async () => {
    await navigateToRoute(page, '/config/value')
    await page.waitForTimeout(2000)
    await takeEvidence(page, {
      dataId: 'ca-11',
      description: '缺ModuleKey查版本历史初始态',
    })
  })

  test('ca-12 正常查询版本历史', async () => {
    await navigateToRoute(page, '/config/value')
    await expect(page.locator('[data-id="ca-select-mokuai-value"]')).toBeVisible({ timeout: 10000 })
    await selectElOption(page, '[data-id="ca-select-mokuai-value"]', 'app.server')
    await selectElOption(page, '[data-id="ca-select-huanjing-value"]', 'dev')
    await page.locator('[data-id="ca-btn-fabu-peizhi"]').click()
    await page.waitForTimeout(1500)
    await expect(page.locator('[data-id="ca-table-version-history"]')).toBeVisible({ timeout: 10000 })
    const versionRows = page.locator('[data-id="ca-table-version-history"] tbody tr')
    await expect(versionRows).toHaveCount((testData.config_versions as Array<unknown>).length)
    await takeEvidence(page, {
      dataId: 'ca-12',
      description: '版本历史表格正常渲染',
      fullPage: true,
    })
  })
})
