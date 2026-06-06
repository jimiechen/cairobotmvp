# E2E Artifact 产物规范

> **生成时间**：2026-06-05
> **阶段**：Phase 1 — 定义产物结构
> **目标**：定义从 proto-tester CLI 输出到验收要求的 artifact 转换规范

---

## 一、每 Case 必备产物清单

| # | 文件名 | 格式 | 来源 | 必要性 | 当前可用性 |
|---|--------|------|------|--------|-----------|
| 1 | request.raw.json | JSON | proto-tester 生成 / wrapper 脚本 | **必须** | 部分可用（需 wrapper 补全） |
| 2 | response.raw.json | JSON | proto-tester 原生输出 | **必须** | 可用 |
| 3 | response.raw.bin | Protobuf binary | proto-tester 原生输出 | **必须** | 可用 |
| 4 | trace.log | 文本 | trace CLI 或 grep | **必须** | DEGRADED（grep 替代） |
| 5 | gateway.log | 文本 | Gateway stdout/stderr | **必须** | 需启动时重定向 |
| 6 | service.log | 文本 | 各 service stdout | 推荐 | 取决于日志配置 |
| 7 | assertion.json | JSON | 断言引擎生成 | **必须** | runner 部分生成 |
| 8 | screenshot.png | 图像 | capture 命令 | UI case 必须 | 需 vite dev server |

---

## 二、proto-tester 原始输出 vs 验收需求 差距分析

### 2.1 send 命令原始输出

**当前输出**（[send.ts](../../../typescript/proto-tester/src/cli/commands/send.ts#L101-L117)）：

| 文件 | 路径模板 | 内容 |
|------|----------|------|
| Markdown 报告 | `{outputDir}/send-{max}-{min}-{timestamp}.md` | 协议名、payload JSON、HTTP Status、业务码、耗时 |
| 终端摘要 | stdout | 8 行文本摘要 |

**缺失项**：

| 缺失内容 | 说明 | 影响 |
|----------|------|------|
| request raw body（base64） | CLI 发送后不保存请求体原文 | 无法回放请求 |
| request headers 完整列表 | 不记录完整 header set | 无法确认 Content-Type / Authorization |
| traceId 值 | 终端显示但报告中可能不含 | trace 关联困难 |
| response bin 文件 | 仅在 memory 中 decode，不落盘 | 无法离线分析响应二进制 |
| token 脱敏记录 | 不记录 token 是否传入及脱敏状态 | 安全审计缺口 |

### 2.2 run 命令原始输出

**当前输出**（[runner.ts](../../../typescript/proto-tester/src/cli/runner.ts#L177-L199)）：

| 文件 | 路径模板 | 内容 |
|------|----------|------|
| Markdown 摘要 | `{outputDir}/summary-{timestamp}.md` | 通过/失败统计 |
| JUnit XML | `{outputDir}/junit-{timestamp}.xml` | CI 集成格式 |
| traces-index.json | `{outputDir}/traces-index.json` | caseId → traceId 映射 |

**缺失项**：同 send 命令，且缺少每 case 的独立 request/response 详细报告。

---

## 三、request.raw.json 规范定义

### 3.1 目标结构

```json
{
  "caseId": "E2E-MVP-002",
  "timestamp": "2026-06-05T10:30:00.000Z",
  "request": {
    "method": "POST",
    "url": "http://localhost:8080/api/hello",
    "headers": {
      "Content-Type": "application/octet-stream",
      "Authorization": "Bearer ***REDACTED***",
      "Accept": "*/*"
    },
    "messagePacket": {
      "maxType": 2100,
      "minType": 2097,
      "platform": 1,
      "extend": {
        "method": "com.mineplanet.pojo.health.ServiceHealthCheckRequest",
        "traceId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "requestId": "f7e6d5c4-b3a2-1098-fedc-ba0987654321"
      },
      "dataEncoding": "json-string-utf8-bytes",
      "dataLength": 2,
      "dataBase64": "e30="           // {} 的 UTF-8 bytes base64
    },
    "rawBodyBase64": "<protobuf binary base64>"  // 整个 MessagePacket 序列化后的 base64
  },
  "meta": {
    "cliCommand": "send --max 2100 --min 2097",
    "cliVersion": "2.0.0",
    "env": "dev",
    "tokenProvided": false,
    "tokenRedacted": true
  }
}
```

### 3.2 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| caseId | string | 是 | 用例 ID |
| timestamp | string (ISO 8601) | 是 | 请求发送时间 |
| request.method | string | 是 | HTTP 方法（固定 POST） |
| request.url | string | 是 | 完整请求 URL |
| request.headers | object | 是 | 请求头字典 |
| request.headers.Authorization | string | 条件 | 有 token 时为 `***REDACTED***`；无 token 时省略此字段 |
| request.messagePacket | object | 是 | 解码后的 MessagePacket 结构 |
| request.messagePacket.dataEncoding | string | 是 | data 字段的编码方式标识 |
| request.messagePacket.dataBase64 | string (base64) | 是 | data 字段 base64（便于离线分析） |
| request.rawBodyBase64 | string (base64) | 推荐 | 整个 HTTP body 的 base64 |
| meta.cliCommand | string | 是 | 实际执行的 CLI 命令 |
| meta.tokenProvided | boolean | 是 | 是否提供了 --token |
| meta.tokenRedacted | boolean | 是 | token 是否已脱敏 |

---

## 四、response.raw.json 规范定义

### 4.1 目标结构

```json
{
  "caseId": "E2E-MVP-002",
  "timestamp": "2026-06-05T10:30:00.100Z",
  "response": {
    "httpStatus": 200,
    "headers": {
      "Content-Type": "application/octet-stream"
    },
    "messagePacket": {
      "maxType": 2100,
      "minType": 2098,
      "platform": 0,
      "extend": {
        "code": "10200",
        "traceId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "requestId": "f7e6d5c4-b3a2-1098-fedc-ba0987654321"
      },
      "dataLength": 45,
      "dataBase64": "eyJoZWFsdGgiOnRydWUsInNlcnZlciI6Im9rIn0="
    },
    "rawBodyBase64": "<response protobuf binary base64>"
  },
  "performance": {
    "durationMs": 100,
    "dataSizeBytes": 128
  },
  "businessCode": 10200,
  "businessCodeLabel": "SUCCESS"
}
```

---

## 五、assertion.json 规范定义

### 5.1 目标结构

```json
{
  "caseId": "E2E-MVP-002",
  "assertions": [
    {
      "field": "httpStatus",
      "matcher": "exact",
      "expected": 200,
      "actual": 200,
      "passed": true
    },
    {
      "field": "businessCode",
      "matcher": "exact",
      "expected": 10200,
      "actual": 10200,
      "passed": true
    },
    {
      "field": "response.maxType",
      "matcher": "exact",
      "expected": 2100,
      "actual": 2100,
      "passed": true
    },
    {
      "field": "response.minType",
      "matcher": "exact",
      "expected": 2098,
      "actual": 2098,
      "passed": true
    },
    {
      "field": "extend.traceId",
      "matcher": "exists",
      "expected": null,
      "actual": "a1b2c3d4-...",
      "passed": true
    }
  ],
  "summary": {
    "total": 5,
    "passed": 5,
    "failed": 0,
    "result": "PASS"
  },
  "warnings": [],
  "notes": []
}
```

---

## 六、Wrapper 脚本方案

由于 proto-tester CLI 当前无法直接输出完整的 `request.raw.json`（缺少 raw body base64 和 headers 记录），需要 wrapper 脚本补全。

### 6.1 方案：e2e-wrapper.sh

```bash
#!/bin/bash
# e2e-wrapper.sh — 包裹 proto-tester send，补充 artifact 输出
#
# 用法：
#   ./e2e-wrapper.sh --case-id E2E-MVP-002 --max 2100 --min 2097 [--token "xxx"]
#
# 输出：
#   artifacts/{caseId}/request.raw.json
#   artifacts/{caseId}/response.raw.json
#   artifacts/{caseId}/assertion.json

set -euo pipefail

CASE_ID=""
MAX=""
MIN=""
TOKEN=""
OUTPUT_DIR="./e2e-artifacts"
GATEWAY_URL="http://localhost:8080"

# 解析参数
while [[ $# -gt 0 ]]; do
  case $1 in
    --case-id) CASE_ID="$2"; shift 2 ;;
    --max) MAX="$2"; shift 2 ;;
    --min) MIN="$2"; shift 2 ;;
    --token) TOKEN="$2"; shift 2 ;;
    --gateway) GATEWAY_URL="$2"; shift 2 ;;
    --output-dir) OUTPUT_DIR="$2"; shift 2 ;;
    *) echo "未知参数: $1"; exit 1 ;;
  esac
done

CASE_DIR="${OUTPUT_DIR}/${CASE_ID}"
mkdir -p "${CASE_DIR}"

TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%S.000Z")

# ===== 1. 构造 request.raw.json =====
TOKEN_LINE=""
if [ -n "${TOKEN}" ]; then
  TOKEN_LINE='"Authorization": "***REDACTED***",'
fi

cat > "${CASE_DIR}/request.raw.json" << REQEOF
{
  "caseId": "${CASE_ID}",
  "timestamp": "${TIMESTAMP}",
  "request": {
    "method": "POST",
    "url": "${GATEWAY_URL}/api/hello",
    "headers": {
      "Content-Type": "application/octet-stream",
      ${TOKEN_LINE:-}
      "Accept": "*/*"
    },
    "messagePacket": {
      "maxType": ${MAX},
      "minType": ${MIN},
      "platform": 1,
      "dataEncoding": "json-string-utf8-bytes"
    },
    "meta": {
      "cliCommand": "send --max ${MAX} --min ${MIN}${TOKEN:+ --token <redacted>}",
      "cliVersion": "2.0.0",
      "env": "dev",
      "tokenProvided": $([ -n "${TOKEN}" ] && echo "true" || echo "false"),
      "tokenRedacted": true
    }
  }
}
REQEOF

# ===== 2. 执行 proto-tester send 并捕获输出 =====
SEND_ARGS="--max ${MAX} --min ${MIN} --outputDir ${CASE_DIR}"
if [ -n "${TOKEN}" ]; then
  SEND_ARGS="${SEND_ARGS} --token ${TOKEN}"
fi

EXIT_CODE=0
STDOUT_FILE=$(mktemp)
STDERR_FILE=$(mktemp)

npx proto-tester send ${SEND_ARGS} > "${STDOUT_FILE}" 2>${STDERR_FILE} || EXIT_CODE=$?

# ===== 3. 从 proto-tester 报告提取信息构造 response.raw.json =====
# proto-tester 输出 .md 和 .json 响应文件，解析最新的一个
RESPONSE_JSON=$(ls -t "${CASE_DIR}"/response-*.json 2>/dev/null | head -1 || echo "")

if [ -n "${RESPONSE_JSON}" ] && [ -f "${RESPONSE_JSON}" ]; then
  cp "${RESPONSE_JSON}" "${CASE_DIR}/response.raw.json"
else
  # 回退：从 stdout 提取
  cat > "${CASE_DIR}/response.raw.json" << RESPEOF
{
  "caseId": "${CASE_ID}",
  "timestamp": "${TIMESTAMP}",
  "exitCode": ${EXIT_CODE},
  "stdout": "$(cat "${STDOUT_FILE}")",
  "stderr": "$(cat "${STDERR_FILE}")"
}
RESPEOF
fi

# ===== 4. 构造 assertion.json =====
# 从 response 中提取 businessCode 进行断言
BUSINESS_CODE=$(grep -o '"businessCode": [0-9]*' "${CASE_DIR}/response.raw.json" 2>/dev/null | grep -o '[0-9]*' || echo "unknown")
HTTP_STATUS=$(grep -o '"status": [0-9]*' "${CASE_DIR}/response.raw.json" 2>/dev/null | grep -o '[0-9]*' || echo "unknown")

cat > "${CASE_DIR}/assertion.json" << ASSERTEOF
{
  "caseId": "${CASE_ID}",
  "assertions": [
    {
      "field": "exitCode",
      "matcher": "exact",
      "expected": 0,
      "actual": ${EXIT_CODE},
      "passed": $([ ${EXIT_CODE} -eq 0 ] && echo "true" || echo "false")
    },
    {
      "field": "httpStatus",
      "matcher": "exact",
      "expected": 200,
      "actual": ${HTTP_STATUS},
      "passed": $([ "${HTTP_STATUS}" = "200" ] && echo "true" || echo "false")
    },
    {
      "field": "businessCode",
      "matcher": "exact",
      "expected": 10200,
      "actual": ${BUSINESS_CODE},
      "passed": $([ "${BUSINESS_CODE}" = "10200" ] && echo "true" || echo "false")
    }
  ],
  "cliExitCode": ${EXIT_CODE},
  "cliStdout": $(cat "${STDOUT_FILE}" | jq -Rs .),
  "cliStderr": $(cat "${STDERR_FILE}" | jq -Rs .)
}
ASSERTEOF

# 清理临时文件
rm -f "${STDOUT_FILE}" "${STDERR_FILE}"

echo "Artifact output: ${CASE_DIR}/"
ls -la "${CASE_DIR}/"
```

### 6.2 工具缺口清单

| 缺口 | 影响 | 修复建议 | 优先级 |
|------|------|----------|--------|
| send 命令不输出 request raw body base64 | 无法离线回放请求 | encodePacket 后在发送前 dump base64 | P1 |
| send 命令不输出 response .bin 文件 | 无法离线分析响应二进制 | writeFileSync response-*.bin 在 decode 前 | P1 |
| send 命令不记录完整 headers | 无法审计 Authorization 等 | 在 sendRequest 入口 log config.headers | P2 |
| run 命令不输出 per-case detail report | 批量执行缺乏单 case 详情 | 每个 executeSingleCase 结果写独立 JSON | P1 |
| token 脱敏不统一 | 安全风险 | 统一在 apiClient 层脱敏后再记录 | P2 |

---

## 七、目录结构规范

```
e2e-artifacts/
├── {RUN_ID}/                        # 每次执行的唯一 ID（时间戳或 git sha）
│   ├── preflight/
│   │   └── E2E-PREFLIGHT.md
│   ├── mvp/
│   │   ├── E2E-MVP-001/
│   │   │   ├── request.raw.json     # Wrapper 生成或 CLI 增强
│   │   │   ├── response.raw.json    # proto-tester 原生
│   │   │   ├── response.raw.bin     # 待 CLI 增强
│   │   │   ├── assertion.json       # 断言结果
│   │   │   └── trace.log            # grep 或 trace CLI 输出
│   │   ├── E2E-MVP-002/
│   │   │   └── ...
│   │   ├── E2E-MVP-004/
│   │   │   └── ...
│   │   ├── summary-run.json         # run 命令汇总
│   │   ├── junit-run.xml            # JUnit 格式
│   │   └── traces-index.json        # caseId → traceId 映射
│   ├── regression/
│   │   └── ...（同 mvp 结构）
│   ├── capture/
│   │   └── E2E-MVP-013/
│   │       ├── screenshot.png
│   │       ├── video.mp4
│   │       └── capture-report.json
│   ├── logs/
│   │   ├── gateway.log              # Gateway 进程 stdout/stderr 重定向
│   │   └── service.log              # Service 层日志（如有）
│   └── RUN-REPORT.md                # 最终验收报告
```

---

## 八、验收 Checklist

对每个 case，验收时必须确认以下 artifact 存在且合法：

- [ ] `request.raw.json` 存在且包含 method/url/headers/messagePacket/meta
- [ ] `response.raw.json` 存在且包含 httpStatus/businessCode/messagePacket
- [ ] `assertion.json` 存在且 summary.result = PASS 或有明确 FAIL 理由
- [ ] traceId 在 request 和 response 中一致（MVP-010 专用）
- [ ] token 已脱敏（auth_required=true 的 case）
- [ ] gateway.log 中可检索到对应 traceId（DEGRADED 时用 grep）

---

*规范完毕。建议 Phase 2 执行前先实现 wrapper 脚本 baseline。*
