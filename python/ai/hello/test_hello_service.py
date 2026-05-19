import pytest
from fastapi.testclient import TestClient
from hello_service import app

client = TestClient(app)


def test_hello_endpoint():
    response = client.get("/hello?name=CaiRobot")
    assert response.status_code == 200
