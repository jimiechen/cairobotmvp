from fastapi import FastAPI
from datetime import datetime

app = FastAPI()


@app.get("/hello")
async def hello(name: str = "World"):
    message = f"Hello, {name}!"
    return {
        "result": {
            "code": 0,
            "msg": "success"
        },
        "message": message,
        "timestamp": int(datetime.now().timestamp())
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8081)
