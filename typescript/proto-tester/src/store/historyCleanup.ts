/**
 * historyCleanup.ts - 历史记录复合容量策略清理
 *
 * 职责：
 * - 执行 7 天 OR 1000 条取交集的容量清理
 * - 每次 addHistory 后异步触发
 * - 应用启动时执行一次全量清理
 *
 * 不负责：
 * - UI 展示（由 routes/history.tsx 负责）
 * - CRUD 操作（由 store/history.ts 负责）
 */

import { openDB } from './history';

/** Store 名称，与 history.ts 保持一致 */
const STORE_NAME = 'requestHistory';

/** 复合容量策略：最大保留天数 */
export const MAX_AGE_DAYS = 7;

/** 复合容量策略：最大记录数 */
export const MAX_RECORDS = 1000;

/**
 * 复合容量策略清理过期记录
 *
 * 规则：7 天 OR 1000 条取交集（先到期者先生效）。
 *
 * 清理步骤：
 * 1. 删除所有超过 7 天的记录（按 timestamp 正序游标）
 * 2. 如果仍超过 1000 条上限，继续从最旧的开始删除超出部分
 */
export async function cleanupOldRecords(): Promise<void> {
  const db = await openDB();
  const tx = db.transaction(STORE_NAME, 'readwrite');
  const store = tx.objectStore(STORE_NAME);

  // 计算截止时间戳（7 天前）
  const cutoffTime = Date.now() - MAX_AGE_DAYS * 24 * 60 * 60 * 1000;

  // 1. 删除超期记录（按 timestamp 正序游标，最早的先删）
  const timestampIdx = store.index('by-timestamp');
  let expiredCursor = await timestampIdx.openCursor();

  while (expiredCursor) {
    if (expiredCursor.value.timestamp >= cutoffTime) {
      // 已到未过期区域，停止扫描
      break;
    }
    await expiredCursor.delete();
    expiredCursor = await expiredCursor.continue();
  }

  // 2. 如果仍超过上限，继续从最旧的开始删除
  const count = await store.count();
  if (count > MAX_RECORDS) {
    const excessCount = count - MAX_RECORDS;
    let overflowCursor = await timestampIdx.openCursor();
    let deleted = 0;

    while (overflowCursor && deleted < excessCount) {
      await overflowCursor.delete();
      deleted++;
      overflowCursor = await overflowCursor.continue();
    }
  }

  await tx.done;
}
