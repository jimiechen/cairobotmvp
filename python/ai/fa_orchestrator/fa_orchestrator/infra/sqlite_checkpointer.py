# -*- coding: utf-8 -*-
"""SQLiteCheckpointer：基于本地 SQLite 的 LangGraph checkpoint 持久化（PRD-12 F-19，ISS-FA4）。

职责边界：
- 负责：checkpoint / pending_writes 的 put/get_tuple/list 持久化，支撑断点续跑
  与人工裁决恢复（interrupt/Command(resume)），并为 replay/show 提供轨迹数据源。
- 不负责：图构建（graph_builder.py）、轨迹语义记录（graph_nodes.py）、报告产物。

为什么自实现：依赖白名单仅 langgraph + langchain-core（ADR-017 §9.1），
禁止引入 langgraph-checkpoint-sqlite；langgraph 自带 BaseCheckpointSaver 基类，
序列化复用其 JsonPlusSerializer（支持 set/tuple 等 channel 类型往返）。
SQLite 为本地单文件存储，非外部服务，单元/集成测试可直接使用。
"""
import sqlite3
import threading

from langgraph.checkpoint.base import BaseCheckpointSaver, CheckpointTuple

# checkpoint 表：thread+ns+checkpoint_id 唯一；checkpoint_id 为 langgraph 生成的
# 时间戳字符串，字典序与时间序一致，ORDER BY DESC 即取最新
_SCHEMA = """
CREATE TABLE IF NOT EXISTS checkpoints (
    thread_id TEXT NOT NULL,
    checkpoint_ns TEXT NOT NULL DEFAULT '',
    checkpoint_id TEXT NOT NULL,
    checkpoint_blob BLOB NOT NULL,
    checkpoint_type TEXT NOT NULL,
    metadata_blob BLOB NOT NULL,
    metadata_type TEXT NOT NULL,
    parent_checkpoint_id TEXT,
    PRIMARY KEY (thread_id, checkpoint_ns, checkpoint_id)
);
CREATE TABLE IF NOT EXISTS pending_writes (
    thread_id TEXT NOT NULL,
    checkpoint_ns TEXT NOT NULL DEFAULT '',
    checkpoint_id TEXT NOT NULL,
    task_id TEXT NOT NULL,
    idx INTEGER NOT NULL,
    channel TEXT NOT NULL,
    value_blob BLOB NOT NULL,
    value_type TEXT NOT NULL,
    PRIMARY KEY (thread_id, checkpoint_ns, checkpoint_id, task_id, idx)
);
"""


class SQLiteCheckpointer(BaseCheckpointSaver):
    """本地 SQLite checkpoint 存储器（同步接口）。

    状态约束：单连接（check_same_thread=False）+ 实例级互斥锁；
    langgraph 会从 executor 线程池并发调用 put/put_writes，多线程交错
    execute/commit 会触发 sqlite3 连接内部死锁，因此所有连接操作必须
    持锁串行（与官方 SqliteSaver 同模式）；连接随实例生命周期关闭。
    """

    def __init__(self, db_path: str):
        """输入：SQLite 数据库文件路径（不存在则创建）。"""
        super().__init__()
        self._conn = sqlite3.connect(db_path, check_same_thread=False)
        self._lock = threading.Lock()
        self._conn.executescript(_SCHEMA)
        self._conn.commit()

    def put(self, config, checkpoint, metadata, new_versions):
        """持久化一个 checkpoint。

        输入：config（含 thread_id）、checkpoint（langgraph Checkpoint dict）、
        metadata、new_versions（channel 版本，仅透传不落库）。
        输出：携带 checkpoint_id 的 RunnableConfig。
        失败行为：thread_id 缺失抛 ValueError；SQL 异常向上抛出。
        """
        configurable = dict(config.get("configurable") or {})
        thread_id = configurable.get("thread_id")
        if not thread_id:
            raise ValueError("checkpoint 写入缺少 configurable.thread_id")
        ns = configurable.get("checkpoint_ns", "")
        ckpt_type, ckpt_blob = self.serde.dumps_typed(checkpoint)
        meta_type, meta_blob = self.serde.dumps_typed(metadata)
        with self._lock:
            self._conn.execute(
                "INSERT OR REPLACE INTO checkpoints VALUES (?,?,?,?,?,?,?,?)",
                (thread_id, ns, checkpoint["id"], ckpt_blob, ckpt_type,
                 meta_blob, meta_type, configurable.get("checkpoint_id")),
            )
            self._conn.commit()
        return {"configurable": {**configurable, "checkpoint_id": checkpoint["id"]}}

    def get_tuple(self, config):
        """读取 checkpoint。

        输入：config（thread_id 必需；checkpoint_id 可选，缺省取最新）。
        输出：CheckpointTuple（checkpoint/metadata/pending_writes）或 None（无记录）。
        """
        with self._lock:
            row = self._query_one(config)
            if row is None:
                return None
            configurable = dict(config.get("configurable") or {})
            (checkpoint_id, ckpt_blob, ckpt_type, meta_blob, meta_type,
             parent_id) = row
            checkpoint = self.serde.loads_typed((ckpt_type, ckpt_blob))
            metadata = self.serde.loads_typed((meta_type, meta_blob))
            thread_id = configurable["thread_id"]
            ns = configurable.get("checkpoint_ns", "")
            parent_config = None
            if parent_id:
                parent_config = {"configurable": {
                    "thread_id": thread_id, "checkpoint_ns": ns,
                    "checkpoint_id": parent_id,
                }}
            return CheckpointTuple(
                config={"configurable": {
                    **configurable, "checkpoint_id": checkpoint_id}},
                checkpoint=checkpoint,
                metadata=metadata,
                parent_config=parent_config,
                pending_writes=self._load_writes(thread_id, ns, checkpoint_id),
            )

    def list(self, config, *, filter=None, before=None, limit=None):
        """按 checkpoint_id 降序迭代 thread 历史（replay 逐超步回放数据源）。

        锁约束：生成器不能跨 yield 持锁，先在锁内物化全部结果再迭代。
        """
        configurable = (config or {}).get("configurable") or {}
        thread_id = configurable.get("thread_id")
        if not thread_id:
            return
        ns = configurable.get("checkpoint_ns", "")
        sql = ("SELECT checkpoint_id, checkpoint_blob, checkpoint_type, metadata_blob,"
               " metadata_type, parent_checkpoint_id FROM checkpoints"
               " WHERE thread_id=? AND checkpoint_ns=? ORDER BY checkpoint_id DESC")
        params: list = [thread_id, ns]
        if limit is not None:
            sql += " LIMIT ?"
            params.append(int(limit))
        with self._lock:
            rows = self._conn.execute(sql, params).fetchall()
            tuples = [self._tuple_from_row(config, row) for row in rows]
        yield from tuples

    def put_writes(self, config, writes, task_id, task_path=""):
        """持久化 pending writes（interrupt resume 值经此恢复）。

        输入：writes 为 (channel, value) 序列；task_id 标识写入任务。
        失败行为：thread 下无任何 checkpoint 时抛 ValueError。
        """
        configurable = dict(config.get("configurable") or {})
        thread_id = configurable["thread_id"]
        ns = configurable.get("checkpoint_ns", "")
        with self._lock:
            checkpoint_id = configurable.get("checkpoint_id") or self._latest_id(
                thread_id, ns
            )
            if not checkpoint_id:
                raise ValueError("pending writes 写入缺少目标 checkpoint")
            for idx, (channel, value) in enumerate(writes):
                value_type, value_blob = self.serde.dumps_typed(value)
                self._conn.execute(
                    "INSERT OR REPLACE INTO pending_writes VALUES (?,?,?,?,?,?,?,?)",
                    (thread_id, ns, checkpoint_id, task_id, idx,
                     channel, value_blob, value_type),
                )
            self._conn.commit()

    def _query_one(self, config):
        """按 config 查询单条 checkpoint 行（指定 id 精确查，否则最新）。

        锁约束：仅允许在持有 self._lock 时调用。
        """
        configurable = (config or {}).get("configurable") or {}
        thread_id = configurable.get("thread_id")
        if not thread_id:
            return None
        ns = configurable.get("checkpoint_ns", "")
        checkpoint_id = configurable.get("checkpoint_id")
        if checkpoint_id:
            sql = ("SELECT checkpoint_id, checkpoint_blob, checkpoint_type,"
                   " metadata_blob, metadata_type, parent_checkpoint_id"
                   " FROM checkpoints WHERE thread_id=? AND checkpoint_ns=?"
                   " AND checkpoint_id=?")
            params: list = [thread_id, ns, checkpoint_id]
        else:
            sql = ("SELECT checkpoint_id, checkpoint_blob, checkpoint_type,"
                   " metadata_blob, metadata_type, parent_checkpoint_id"
                   " FROM checkpoints WHERE thread_id=? AND checkpoint_ns=?"
                   " ORDER BY checkpoint_id DESC LIMIT 1")
            params = [thread_id, ns]
        return self._conn.execute(sql, params).fetchone()

    def _tuple_from_row(self, config, row) -> CheckpointTuple:
        """把 checkpoint 行组装为 CheckpointTuple（含 pending_writes）。

        锁约束：仅允许在持有 self._lock 时调用。
        """
        configurable = dict((config or {}).get("configurable") or {})
        (checkpoint_id, ckpt_blob, ckpt_type, meta_blob, meta_type,
         parent_id) = row
        thread_id = configurable["thread_id"]
        ns = configurable.get("checkpoint_ns", "")
        parent_config = None
        if parent_id:
            parent_config = {"configurable": {
                "thread_id": thread_id, "checkpoint_ns": ns,
                "checkpoint_id": parent_id,
            }}
        return CheckpointTuple(
            config={"configurable": {**configurable, "checkpoint_id": checkpoint_id}},
            checkpoint=self.serde.loads_typed((ckpt_type, ckpt_blob)),
            metadata=self.serde.loads_typed((meta_type, meta_blob)),
            parent_config=parent_config,
            pending_writes=self._load_writes(thread_id, ns, checkpoint_id),
        )

    def _load_writes(self, thread_id: str, ns: str, checkpoint_id: str) -> list:
        """读取指定 checkpoint 的 pending writes（按 idx 排序）。

        锁约束：仅允许在持有 self._lock 时调用。
        """
        rows = self._conn.execute(
            "SELECT task_id, channel, value_blob, value_type FROM pending_writes"
            " WHERE thread_id=? AND checkpoint_ns=? AND checkpoint_id=? ORDER BY idx",
            (thread_id, ns, checkpoint_id),
        ).fetchall()
        return [
            (task_id, channel, self.serde.loads_typed((value_type, value_blob)))
            for task_id, channel, value_blob, value_type in rows
        ]

    def _latest_id(self, thread_id: str, ns: str) -> str | None:
        """查询 thread 最新 checkpoint_id（put_writes 未显式指定时兜底）。

        锁约束：仅允许在持有 self._lock 时调用。
        """
        row = self._conn.execute(
            "SELECT checkpoint_id FROM checkpoints WHERE thread_id=? AND"
            " checkpoint_ns=? ORDER BY checkpoint_id DESC LIMIT 1",
            (thread_id, ns),
        ).fetchone()
        return row[0] if row else None
