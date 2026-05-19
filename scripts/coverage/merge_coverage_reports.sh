#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==> 合并覆盖率报告..."

SUMMARY_FILE="docs/reports/coverage/coverage-summary.md"
mkdir -p docs/reports/coverage

cat > "$SUMMARY_FILE" << 'EOF'
# CaiRobot MVP 测试覆盖率汇总报告

> 自动生成于 $(date '+%Y-%m-%d %H:%M:%S')

## 覆盖率概览

| 语言 | 模块 | 状态 | 报告位置 |
|---|---|---|---|
| Go | go/gateway/proto-gateway | 待检测 | coverage-go/gateway-proto-gateway_coverage.html |
| Go | go/tars/system | 待检测 | coverage-go/tars-system_coverage.html |
| Go | go/services/hello-service | 待检测 | coverage-go/services_coverage.html |
| TypeScript | web | 待检测 | typescript/web/ |
| TypeScript | admin-web | 待检测 | typescript/admin-web/ |
| Python | ai/hello | 待检测 | python/index.html |

## 覆盖率阈值要求

| 模块 | 最低阈值 | 当前值 | 状态 |
|---|---|---|---|
| Gateway 核心模块 | 60% | - | ⏳ 待检测 |
| Tars 核心模块 | 60% | - | ⏳ 待检测 |
| AI 服务 | 50% | - | ⏳ 待检测 |
| Web 前端 | 40% | - | ⏳ 待检测 |

## 说明

当前阶段覆盖率先要求**生成报告**，不强制高阈值。
Gateway/Tars 核心模块建议最低 **60%**。

---

*由 `bash scripts/coverage/merge_coverage_reports.sh` 自动生成*
EOF

echo "==> 覆盖率汇总报告已生成: $SUMMARY_FILE"
