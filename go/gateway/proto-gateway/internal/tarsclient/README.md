# TarsClient

## 职责

TarsClient 统一调用 Tars bytes 接口。

## 方法签名

```tars
int Method(vector<byte> request, map<string,string> extend, out vector<byte> response);
```

## 约束

- 不允许业务代码绕过 Router 硬编码 Tars 目标
- 不允许外部调用方传入 Tars app/server/servant/method
- Tars timeout、server not found、invoke failed 必须转换为项目统一错误码
- traceId/requestId 必须透传

## 错误映射

| Tars 异常          | 项目统一错误码 |
| ---------------- | ------- |
| timeout          | 10504   |
| server not found | 10404   |
| invoke failed    | 10500   |

## 相关文档

- [docs/api/tars规范.md](../../../docs/api/tars规范.md)

