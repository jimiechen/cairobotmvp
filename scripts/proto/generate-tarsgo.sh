#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==> 生成 TarsGo/Tars2Go 代码..."

TARS_PROTO_DIR="tars/protocol"

mkdir -p proto/generated/tarsgo

TARS_FILES=$(find tars/protocol/tars -name "*.tars" -type f 2>/dev/null | sort)

if [ -z "$TARS_FILES" ]; then
    echo "[tarsgo] 警告：未找到 .tars 文件"
    exit 0
fi

if command -v tars2go >/dev/null 2>&1; then
    for tars_file in $TARS_FILES; do
        echo "[tarsgo] 处理: $tars_file"
        tars2go --outdir=proto/generated/tarsgo "$tars_file" || true
    done
    echo "==> TarsGo 代码生成完成，输出到 proto/generated/tarsgo/"
elif command -v protoc >/dev/null 2>&1 && [ -d "$TARS_PROTO_DIR/proto-adapter" ]; then
    echo "[tarsgo] 使用 protoc + 自定义插件..."
    protoc \
      --tars_out=proto/generated/tarsgo \
      -I"$TARS_PROTO_DIR" \
      -Iproto \
      $TARS_FILES
    echo "==> TarsGo 代码生成完成（protoc 模式）"
else
    echo "[tarsgo] 跳过：tars2go 工具未安装"
    echo "安装方式：go install github.com/TarsCloud/TarsGo/tars/tools/tars2go@latest"
fi
