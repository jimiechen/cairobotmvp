# -*- coding: utf-8 -*-
"""SQLiteCheckpointer 单元测试（PRD-12 F-19，ISS-FA4）。

验证 checkpoint 持久化契约：put/get_tuple 往返、list 降序迭代、
pending_writes 携带、thread 隔离、serde 非 JSON 类型往返。
SQLite 为本地文件存储（非外部服务），全程不连真实 MySQL/HTTP。
"""
import pytest

from fa_orchestrator.sqlite_checkpointer import SQLiteCheckpointer


def _config(thread_id: str = "t1", checkpoint_id: str | None = None) -> dict:
    """构造 RunnableConfig 形态的寻址键。"""
    configurable: dict = {"thread_id": thread_id}
    if checkpoint_id:
        configurable["checkpoint_id"] = checkpoint_id
    return {"configurable": configurable}


def _checkpoint(cp_id: str, trace: list) -> dict:
    """构造最小 Checkpoint 结构（字段名与 langgraph 1.x TypedDict 对齐）。"""
    return {
        "v": 4,
        "id": cp_id,
        "ts": "2026-09-30T10:00:00",
        "channel_values": {"trace": trace},
        "channel_versions": {"trace": len(trace)},
        "versions_seen": {},
    }


@pytest.fixture()
def checkpointer(tmp_path):
    return SQLiteCheckpointer(str(tmp_path / "ckpt.db"))


class TestSQLiteCheckpointer:
    def test_put后get_tuple返回最新checkpoint(self, checkpointer):
        cfg = _config()
        checkpointer.put(cfg, _checkpoint("cp1", ["n1"]), {"source": "loop"}, {})
        checkpointer.put(cfg, _checkpoint("cp2", ["n1", "n2"]), {"source": "loop"}, {})
        tp = checkpointer.get_tuple(cfg)
        assert tp.checkpoint["id"] == "cp2"
        assert tp.checkpoint["channel_values"]["trace"] == ["n1", "n2"]
        assert tp.metadata["source"] == "loop"

    def test_空库get_tuple返回None(self, checkpointer):
        assert checkpointer.get_tuple(_config()) is None

    def test_list按checkpoint_id降序迭代(self, checkpointer):
        cfg = _config()
        for cp_id, trace in (("cp1", ["a"]), ("cp2", ["a", "b"]), ("cp3", ["a", "b", "c"])):
            checkpointer.put(cfg, _checkpoint(cp_id, trace), {}, {})
        ids = [tp.checkpoint["id"] for tp in checkpointer.list(cfg)]
        assert ids == ["cp3", "cp2", "cp1"]

    def test_list支持limit截取(self, checkpointer):
        cfg = _config()
        for cp_id in ("cp1", "cp2", "cp3"):
            checkpointer.put(cfg, _checkpoint(cp_id, []), {}, {})
        ids = [tp.checkpoint["id"] for tp in checkpointer.list(cfg, limit=2)]
        assert ids == ["cp3", "cp2"]

    def test_指定checkpoint_id可取历史版本(self, checkpointer):
        cfg = _config()
        checkpointer.put(cfg, _checkpoint("cp1", ["n1"]), {}, {})
        checkpointer.put(cfg, _checkpoint("cp2", ["n1", "n2"]), {}, {})
        tp = checkpointer.get_tuple(_config(checkpoint_id="cp1"))
        assert tp.checkpoint["id"] == "cp1"
        assert tp.checkpoint["channel_values"]["trace"] == ["n1"]

    def test_thread隔离_不同thread互不可见(self, checkpointer):
        checkpointer.put(_config("t1"), _checkpoint("cp1", ["a"]), {}, {})
        checkpointer.put(_config("t2"), _checkpoint("cp2", ["b"]), {}, {})
        assert checkpointer.get_tuple(_config("t1")).checkpoint["id"] == "cp1"
        assert checkpointer.get_tuple(_config("t3")) is None

    def test_put_writes后get_tuple携带pending_writes(self, checkpointer):
        cfg = _config()
        checkpointer.put(cfg, _checkpoint("cp1", []), {}, {})
        checkpointer.put_writes(
            cfg, [("human_decisions", {"1398597": "确认"})], "task1"
        )
        tp = checkpointer.get_tuple(cfg)
        writes = {(w[0], w[1]): w[2] for w in tp.pending_writes}
        assert writes[("task1", "human_decisions")] == {"1398597": "确认"}

    def test_serde非JSON类型往返不丢失(self, checkpointer):
        cfg = _config()
        checkpoint = _checkpoint("cp1", [])
        checkpoint["channel_values"] = {"obj": (1, {2, 3})}
        checkpointer.put(cfg, checkpoint, {}, {})
        tp = checkpointer.get_tuple(cfg)
        # JsonPlusSerializer 契约：set 保留，tuple 规范化为 list
        assert tp.checkpoint["channel_values"]["obj"] == [1, {2, 3}]

    def test_持久化到文件_新实例可读取(self, tmp_path):
        db_path = str(tmp_path / "ckpt.db")
        first = SQLiteCheckpointer(db_path)
        first.put(_config(), _checkpoint("cp1", ["n1"]), {}, {})
        reopened = SQLiteCheckpointer(db_path)
        tp = reopened.get_tuple(_config())
        assert tp.checkpoint["id"] == "cp1"
