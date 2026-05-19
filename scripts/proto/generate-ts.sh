#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==> 生成 Protobuf TypeScript 代码..."

mkdir -p proto/generated/ts

PROTO_FILES=$(find proto -name "*.proto" -type f 2>/dev/null | sort)

if [ -z "$PROTO_FILES" ]; then
    echo "[ts-proto] 警告：未找到 .proto 文件"
    exit 0
fi

if command -v protoc-gen-ts >/dev/null 2>&1 || command -v protoc-gen-es >/dev/null 2>&1; then
    protoc \
      --ts_out=proto/generated/ts \
      --ts_opt=eslint_disable=false \
      -Iproto \
      $PROTO_FILES
    echo "==> TypeScript Protobuf 代码生成完成，输出到 proto/generated/ts/"
else
    echo "[ts-proto] 跳过：protoc-gen-ts / protoc-gen-es 未安装"
    echo "安装方式：npm install -g protoc-gen-ts @bufbuild/protobuf"
fi
