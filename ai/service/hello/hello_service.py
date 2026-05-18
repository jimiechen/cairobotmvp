"""Hello Service 主模块"""

from datetime import datetime, timezone
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()


class Result(BaseModel):
    """通用返回结果"""
    code: int = 10200
    message: str = "success"


class HelloResponse(BaseModel):
    """Hello 响应"""
    result: Result
    message: str
    timestamp: str


@app.get("/hello")
def hello_endpoint() -> HelloResponse:
    """Hello 端点"""
    return HelloResponse(
        result=Result(code=10200, message="success"),
        message="Hello, World!",
        timestamp=datetime.now(timezone.utc).isoformat()
    )


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8081)
