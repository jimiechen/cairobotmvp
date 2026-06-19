#!/bin/bash
# check_routes_sync.sh — CI 校验 routes.yaml 两副本是否一致
# 用途：CI Pipeline 中检查根目录 configs/gateway/routes.yaml 与
#       运行时目录 go/gateway/proto-gateway/configs/gateway/routes.yaml 的 SHA256 是否一致
# 退出码：0=一致, 1=不一致或文件缺失
#
# 相关文档：
# - ADR: routes.yaml 三阶段治理方案（短期 Makefile sync / 中期 CI hash / 长期单一权威源）
# - Phase 1 MVP-P0 Integration Plan Task 6

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

SRC="${PROJECT_ROOT}/configs/gateway/routes.yaml"
DST="${PROJECT_ROOT}/go/gateway/proto-gateway/configs/gateway/routes.yaml"

echo "==> 校验 routes.yaml 同步状态 ..."
echo "    源(编辑): $SRC"
echo "    目标(运行时): $DST"

# 检查源文件存在
if [ ! -f "$SRC" ]; then
    echo "❌ 源文件不存在: $SRC"
    echo "   请确认 configs/gateway/routes.yaml 已提交到仓库"
    exit 1
fi

# 检查目标文件存在
if [ ! -f "$DST" ]; then
    echo "❌ 目标文件不存在: $DST"
    echo "   请执行 make routes-sync 同步"
    exit 1
fi

# SHA256 比较
SRC_HASH=$(shasum -a 256 "$SRC" | awk '{print $1}')
DST_HASH=$(shasum -a 256 "$DST" | awk '{print $1}')

if [ "$SRC_HASH" = "$DST_HASH" ]; then
    echo "✅ routes.yaml 一致 (SHA256=${SRC_HASH:0:16}...)"
    exit 0
else
    echo "❌ routes.yaml 不一致！"
    echo "   源文件 SHA256: ${SRC_HASH}"
    echo "   目标文件 SHA256: ${DST_HASH}"
    echo ""
    echo "   修复方法：make routes-sync"
    echo "   或手动：cp $SRC $DST"
    exit 1
fi
