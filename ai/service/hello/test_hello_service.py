"""Hello Service 测试模块"""

import pytest
from fastapi.testclient import TestClient
from hello_service import app

client = TestClient(app)


def test_hello_endpoint_returns_200():
    """测试 /hello 接口返回 200"""
    response = client.get("/hello")
    assert response.status_code == 200


def test_hello_endpoint_returns_json():
    """测试 /hello 接口返回 JSON"""
    response = client.get("/hello")
    assert response.headers["content-type"] == "application/json"


def test_hello_endpoint_contains_message():
    """测试 /hello 接口响应包含 message 字段"""
    response = client.get("/hello")
    data = response.json()
    assert "message" in data
    assert isinstance(data["message"], str)
    assert len(data["message"]) > 0


def test_hello_endpoint_has_timestamp():
    """测试 /hello 接口响应包含 timestamp 字段"""
    response = client.get("/hello")
    data = response.json()
    assert "timestamp" in data
