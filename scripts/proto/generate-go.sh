#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==> 生成 Protobuf Go 代码..."

mkdir -p proto/generated/go

PROTO_FILES=$(find proto -name "*.proto" -type f 2>/dev/null | sort)

if [ -z "$PROTO_FILES" ]; then
    echo "[go-proto] 警告：未找到 .proto 文件"
    exit 0
fi

# module 路径前缀：与 proto/generated/go/go.mod 中的 module 声明一致
# 新增子包时只需在 proto/ 下新增 .proto 文件，无需修改此变量
GO_MODULE_PREFIX="github.com/jimiechen/mineplanet/protocols/generated/go"

protoc \
  --go_out=proto/generated/go \
  --go_opt=paths=source_relative \
  --go-grpc_out=proto/generated/go \
  --go-grpc_opt=paths=source_relative \
  $(for f in $PROTO_FILES; do echo "--go_opt=M${f#proto/}=${GO_MODULE_PREFIX}/${f%.*}" | sed 's|/|.|g'; done) \
  -Iproto \
  $PROTO_FILES

echo "==> Go Protobuf 代码生成完成，输出到 proto/generated/go/"
