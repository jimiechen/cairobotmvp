# -*- coding: utf-8 -*-
"""LLM 统一网关契约测试（PRD-12 F-03）。

用本地 Mock OpenAI Server（标准库 http.server）验证契约：
- POST {base_url}/chat/completions，Bearer 认证
- 请求体含 model/temperature/messages
- 响应解析为 langchain AIMessage
- 统一超时/重试；密钥不落日志（安全红线）
禁止真实大模型调用。
"""
import json
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

import pytest
from langchain_core.messages import AIMessage

from fa_orchestrator.llm_gateway import GatewayError, LlmGateway


class _MockOpenAIServer:
    """OpenAI 兼容 Mock Server：可编程返回序列，记录收到的请求。"""

    def __init__(self, responses=None):
        self.requests = []
        self.responses = list(responses or [200, {"choices": [{"message": {"role": "assistant", "content": "ok"}}]}])
        self._server = ThreadingHTTPServer(("127.0.0.1", 0), self._make_handler())
        self._thread = threading.Thread(target=self._server.serve_forever, daemon=True)

    @property
    def port(self):
        return self._server.server_address[1]

    def _make_handler(self):
        mock = self

        class _Handler(BaseHTTPRequestHandler):
            def do_POST(self):  # noqa: N802
                length = int(self.headers.get("Content-Length", 0))
                body = json.loads(self.rfile.read(length) or b"{}")
                mock.requests.append({"path": self.path, "headers": dict(self.headers), "body": body})
                status, payload = mock.responses.pop(0) if isinstance(mock.responses[0], tuple) else mock.responses[0]
                data = json.dumps(payload).encode("utf-8")
                self.send_response(status)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(data)))
                self.end_headers()
                self.wfile.write(data)

            def log_message(self, *args):  # 静默访问日志
                pass

        return _Handler

    def __enter__(self):
        self._thread.start()
        return self

    def __exit__(self, *exc):
        self._server.shutdown()
        self._server.server_close()


def _ok_response(content="mock 回复"):
    return (200, {"choices": [{"message": {"role": "assistant", "content": content}}]})


class TestContract:
    def test_发送OpenAI兼容请求(self):
        with _MockOpenAIServer(responses=[_ok_response()]) as server:
            gateway = LlmGateway(
                base_url=f"http://127.0.0.1:{server.port}",
                api_key="test-key-123",
                model="test-model",
                temperature=0.1,
            )
            gateway.complete([{"role": "user", "content": "你好"}])

        sent = server.requests[0]
        assert sent["path"] == "/chat/completions"
        assert sent["headers"].get("Authorization") == "Bearer test-key-123"
        assert sent["body"]["model"] == "test-model"
        assert sent["body"]["temperature"] == 0.1
        assert sent["body"]["messages"] == [{"role": "user", "content": "你好"}]

    def test_响应解析为AIMessage(self):
        with _MockOpenAIServer(responses=[_ok_response("计划已生成")]) as server:
            gateway = LlmGateway(
                base_url=f"http://127.0.0.1:{server.port}",
                api_key="k", model="m",
            )
            reply = gateway.complete([{"role": "user", "content": "生成计划"}])
        assert isinstance(reply, AIMessage)
        assert reply.content == "计划已生成"


class TestRetryAndError:
    def test_服务端错误后重试成功(self):
        server_error = (500, {"error": "boom"})
        with _MockOpenAIServer(responses=[server_error, _ok_response("重试成功")]) as server:
            gateway = LlmGateway(
                base_url=f"http://127.0.0.1:{server.port}",
                api_key="k", model="m", max_retries=2,
            )
            reply = gateway.complete([{"role": "user", "content": "hi"}])
        assert reply.content == "重试成功"
        assert len(server.requests) == 2

    def test_重试耗尽抛GatewayError(self):
        server_error = (500, {"error": "boom"})
        with _MockOpenAIServer(responses=[server_error, server_error, server_error]) as server:
            gateway = LlmGateway(
                base_url=f"http://127.0.0.1:{server.port}",
                api_key="k", model="m", max_retries=2,
            )
            with pytest.raises(GatewayError):
                gateway.complete([{"role": "user", "content": "hi"}])
        assert len(server.requests) == 3  # 首次 + 2 次重试


class TestConfigFromEnv:
    def test_环境变量注入配置(self, monkeypatch):
        monkeypatch.setenv("FA_LLM_BASE_URL", "http://llm.local/v1")
        monkeypatch.setenv("FA_LLM_API_KEY", "env-key")
        monkeypatch.setenv("FA_LLM_MODEL", "env-model")
        gateway = LlmGateway.from_env()
        assert gateway.base_url == "http://llm.local/v1"
        assert gateway.model == "env-model"

    def test_缺环境变量报错(self, monkeypatch):
        monkeypatch.delenv("FA_LLM_BASE_URL", raising=False)
        with pytest.raises(GatewayError):
            LlmGateway.from_env()


class TestSecretSafety:
    def test_密钥不落日志(self, capsys):
        with _MockOpenAIServer(responses=[_ok_response()]) as server:
            gateway = LlmGateway(
                base_url=f"http://127.0.0.1:{server.port}",
                api_key="SUPER-SECRET-KEY", model="m",
            )
            gateway.complete([{"role": "user", "content": "hi"}])
        captured = capsys.readouterr()
        assert "SUPER-SECRET-KEY" not in captured.out + captured.err
