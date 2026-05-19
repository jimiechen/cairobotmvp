#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==> 生成 Protobuf Python 代码..."

mkdir -p proto/generated/python

PROTO_FILES=$(find proto -name "*.proto" -type f 2>/dev/null | sort)

if [ -z "$PROTO_FILES" ]; then
    echo "[py-proto] 警告：未找到 .proto 文件"
    exit 0
fi

if command -v grpc_python_plugin >/dev/null 2>&1 || python3 -c "import grpc_tools.protoc" 2>/dev/null; then
    python3 -m grpc_tools.protoc \
      -Iproto \
      --python_out=proto/generated/python \
      --grpc_python_out=proto/generated/python \
      $PROTO_FILES
    touch proto/generated/python/__init__.py
    for dir in $(find proto/generated/python -type d); do
        touch "$dir/__init__.py" 2>/dev/null || true
    done
    echo "==> Python Protobuf 代码生成完成，输出到 proto/generated/python/"
else
    echo "[py-proto] 跳过：grpcio-tools 未安装"
    echo "安装方式：pip install grpcio-tools"
fi
