# -*- coding: utf-8 -*-
"""LLM 统一网关（PRD-12 F-03）：OpenAI 兼容 API 外接大模型。

职责边界：
- 负责：chat/completions 调用、Bearer 认证、统一超时/重试、响应转 AIMessage。
- 不负责：提示词构建（planner_agent.py）、业务逻辑；不内置本地模型。

安全约束（ADR-017 §9）：
- base_url/api_key/model 全部配置化（环境变量注入），密钥不进代码、不进日志。
- 单元测试一律 Mock，禁止真实大模型调用。
"""
import json
import logging
import time
import urllib.error
import urllib.request

from langchain_core.messages import AIMessage

# 环境变量名（密钥与接入点全配置化，禁止硬编码）
ENV_BASE_URL = "FA_LLM_BASE_URL"
ENV_API_KEY = "FA_LLM_API_KEY"
ENV_MODEL = "FA_LLM_MODEL"

# 重试间隔基数（秒）：统一超时/重试策略的一部分
RETRY_BACKOFF_SECONDS = 0.5

logger = logging.getLogger(__name__)


class GatewayError(Exception):
    """网关配置缺失或调用失败（重试耗尽）时抛出。"""


class LlmGateway:
    """OpenAI 兼容 LLM 网关。

    状态约束：线程安全（无共享可变状态）；api_key 仅用于请求头，禁止打印。
    """

    def __init__(self, base_url, api_key, model, temperature=0.2,
                 timeout_seconds=30.0, max_retries=2):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.model = model
        self.temperature = temperature
        self.timeout_seconds = timeout_seconds
        self.max_retries = max_retries

    @classmethod
    def from_env(cls) -> "LlmGateway":
        """从环境变量构建网关（FA_LLM_BASE_URL/FA_LLM_API_KEY/FA_LLM_MODEL）。

        失败行为：任一必填环境变量缺失抛 GatewayError。
        """
        import os

        base_url = os.environ.get(ENV_BASE_URL)
        api_key = os.environ.get(ENV_API_KEY)
        model = os.environ.get(ENV_MODEL)
        missing = [
            name for name, value in (
                (ENV_BASE_URL, base_url), (ENV_API_KEY, api_key), (ENV_MODEL, model),
            ) if not value
        ]
        if missing:
            raise GatewayError(f"缺少 LLM 网关环境变量: {', '.join(missing)}")
        return cls(base_url=base_url, api_key=api_key, model=model)

    def complete(self, messages: list) -> AIMessage:
        """调用 chat/completions 并解析为 AIMessage。

        输入：OpenAI 格式消息列表（[{"role": ..., "content": ...}]）。
        输出：AIMessage（content 为回复文本）。
        失败行为：5xx 按统一重试策略重试，耗尽抛 GatewayError；密钥不落日志。
        """
        url = f"{self.base_url}/chat/completions"
        payload = {
            "model": self.model,
            "temperature": self.temperature,
            "messages": messages,
        }
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.api_key}",
        }
        body = json.dumps(payload, ensure_ascii=False).encode("utf-8")

        attempts = self.max_retries + 1
        for attempt in range(1, attempts + 1):
            try:
                return self._request_once(url, headers, body)
            except urllib.error.HTTPError as error:
                # 仅对 5xx 服务端错误重试，4xx 属请求契约错误直接失败
                if error.code < 500 or attempt == attempts:
                    raise GatewayError(f"LLM 网关调用失败: HTTP {error.code}") from error
                logger.warning("LLM 网关 HTTP %s，第 %s 次重试", error.code, attempt)
            time.sleep(RETRY_BACKOFF_SECONDS * attempt)
        raise GatewayError("LLM 网关调用失败: 重试耗尽")

    def _request_once(self, url, headers, body) -> AIMessage:
        """单次请求（不重试），记录日志时脱敏（不含 api_key）。"""
        logger.info(
            "LLM 网关请求 base_url=%s model=%s body_bytes=%s",
            self.base_url, self.model, len(body),
        )
        request = urllib.request.Request(url, data=body, headers=headers, method="POST")
        with urllib.request.urlopen(request, timeout=self.timeout_seconds) as response:
            data = json.loads(response.read().decode("utf-8"))
        content = data["choices"][0]["message"]["content"]
        return AIMessage(content=content)
