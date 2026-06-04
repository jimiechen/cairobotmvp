/**
 * history.ts - IndexedDB 请求历史记录存储（CRUD 层）
 *
 * 职责：
 * - 管理请求/响应历史记录的持久化（IndexedDB via idb）
 * - 提供增删查接口
 *
 * 不负责：
 * - UI 展示（由 routes/history.tsx 负责）
 * - 容量策略清理（由 store/historyCleanup.ts 负责）
 * - traceId 生成（由 utils/traceId.ts 负责）
 *
 * Schema：
 * - 数据库名：proto-tester-db
 * - 版本：1
 * - Store 名：requestHistory
 * - KeyPath: id (autoIncrement)
 * - Indexes: by-traceId, by-timestamp, by-protocol [maxType, minType]
 */

import { openDB, type IDBPDatabase } from 'idb';
import { cleanupOldRecords } from './historyCleanup';

/** 数据库名称 */
const DB_NAME = 'proto-tester-db';

/** 数据库版本 */
const DB_VERSION = 1;

/** Store 名称 */
export const STORE_NAME = 'requestHistory';

/** 历史记录数据结构 */
export interface HistoryRecord {
  /** autoIncrement 主键，写入时不需要指定 */
  id?: number;
  /** 时间戳（Date.now()） */
  timestamp: number;
  /** UUID 格式的追踪 ID */
  traceId: string;
  /** 协议主类型号 */
  maxType: number;
  /** 协议子类型号 */
  minType: number;
  /** 协议名称，如 "HelloWorldRequest" */
  protocolName: string;
  /** 发送的表单值快照 */
  requestPayload: Record<string, any>;
  /** 响应摘要 */
  responseSummary: {
    /** HTTP 状态码 */
    status: number;
    /** 业务码 */
    businessCode: number;
    /** 请求耗时（毫秒） */
    durationMs: number;
    /** 错误信息（可选） */
    error?: string;
  };
}

/** 打开数据库时的 upgrade 回调类型 */
export interface HistoryDBSchema {
  [STORE_NAME]: HistoryRecord;
}

let dbInstance: IDBPDatabase<HistoryDBSchema> | null = null;

/**
 * 打开或创建 IndexedDB 数据库
 *
 * @returns 数据库实例 Promise
 */
export async function openDB(): Promise<IDBPDatabase<HistoryDBSchema>> {
  if (dbInstance) {
    return dbInstance;
  }

  dbInstance = await openDB<HistoryDBSchema>(DB_NAME, DB_VERSION, {
    upgrade(db) {
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        const store = db.createObjectStore(STORE_NAME, {
          keyPath: 'id',
          autoIncrement: true,
        });

        // 按 traceId 索引：支持按 traceId 查询同一请求的所有记录
        store.createIndex('by-traceId', 'traceId');

        // 按时间戳索引：支持按时间范围查询和排序
        store.createIndex('by-timestamp', 'timestamp');

        // 按协议号复合索引：支持按 maxType + minType 精确匹配
        store.createIndex('by-protocol', ['maxType', 'minType']);
      }
    },
  });

  return dbInstance;
}

/**
 * 新增一条历史记录
 *
 * 写入后自动异步触发容量清理。
 *
 * @param record 不含 id 的历史记录数据
 * @returns 新记录的自增主键 ID
 */
export async function addHistory(
  record: Omit<HistoryRecord, 'id'>,
): Promise<number> {
  const db = await openDB();
  const id = await db.add(STORE_NAME, record as HistoryRecord);

  // 异步触发清理，不阻塞返回
  cleanupOldRecords().catch((err) => {
    console.warn('[history] cleanupOldRecords 异步执行失败:', err);
  });

  return id as number;
}

/**
 * 分页获取历史记录列表（按时间倒序）
 *
 * @param opts 分页参数
 * @returns 历史记录数组
 */
export async function getHistoryList(opts?: {
  limit?: number;
  offset?: number;
}): Promise<HistoryRecord[]> {
  const db = await openDB();
  const limit = opts?.limit ?? 50;
  const offset = opts?.offset ?? 0;

  const tx = db.transaction(STORE_NAME, 'readonly');
  const store = tx.objectStore(STORE_NAME);
  const index = store.index('by-timestamp');

  // 按 timestamp 倒序游标遍历
  const records: HistoryRecord[] = [];
  let cursor = await index.openCursor(null, 'prev');
  let skipped = 0;

  while (cursor && records.length < limit) {
    if (skipped < offset) {
      skipped++;
      cursor = await cursor.continue();
      continue;
    }
    records.push(cursor.value);
    cursor = await cursor.continue();
  }

  return records;
}

/**
 * 按 traceId 查询关联的所有历史记录
 *
 * @param traceId 目标 traceId
 * @returns 匹配的历史记录数组
 */
export async function getHistoryByTraceId(
  traceId: string,
): Promise<HistoryRecord[]> {
  const db = await openDB();
  const records = await db.getAllFromIndex(
    STORE_NAME,
    'by-traceId',
    traceId,
  );
  return records;
}

/**
 * 按协议号查询历史记录
 *
 * @param maxType 协议主类型号
 * @param minType 协议子类型号
 * @returns 匹配的历史记录数组
 */
export async function getHistoryByProtocol(
  maxType: number,
  minType: number,
): Promise<HistoryRecord[]> {
  const db = await openDB();
  const records = await db.getAllFromIndex(
    STORE_NAME,
    'by-protocol',
    [maxType, minType],
  );
  return records;
}

/**
 * 删除单条历史记录
 *
 * @param id 记录主键
 */
export async function deleteHistory(id: number): Promise<void> {
  const db = await openDB();
  await db.delete(STORE_NAME, id);
}

/**
 * 清空全部历史记录
 *
 * 注意：二次确认逻辑应在 UI 层实现，本函数仅执行删除操作。
 */
export async function clearAllHistory(): Promise<void> {
  const db = await openDB();
  await db.clear(STORE_NAME);
}
