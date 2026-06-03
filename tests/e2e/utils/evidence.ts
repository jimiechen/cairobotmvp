import { Page } from '@playwright/test'
import path from 'path'
import fs from 'fs/promises'

export interface EvidenceOptions {
  dataId: string
  description: string
  fullPage?: boolean
}

const EVIDENCE_DIR = path.resolve(__dirname, '../evidence')

export async function takeEvidence(
  page: Page,
  options: EvidenceOptions
): Promise<string> {
  const { dataId, description, fullPage = true } = options
  const ts = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19)
  const safeDesc = description.replace(/[/\\?%*:|"<>]/g, '-')
  const filename = `${dataId}-${safeDesc}-${ts}.png`
  const filepath = path.join(EVIDENCE_DIR, filename)

  await fs.mkdir(EVIDENCE_DIR, { recursive: true })
  await page.screenshot({ path: filepath, fullPage })
  return filepath
}

export async function clickAndScreenshot(
  page: Page,
  dataId: string,
  options: EvidenceOptions
): Promise<void> {
  await takeEvidence(page, {
    ...options,
    description: options.description + '-操作前',
  })
  await page.locator(`[data-id="${dataId}"]`).click()
  await page.waitForTimeout(500)
  await takeEvidence(page, {
    ...options,
    description: options.description + '-操作后',
  })
}
