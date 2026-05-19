#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==========================================="
echo "CaiRobot MVP 开发工具链检查"
echo "==========================================="
echo ""

MISSING_TOOLS=0
WARNINGS=0

check_tool() {
    local name=$1
    local cmd=$2
    local install_cmd=$3
    local required=$4
    
    if command -v "$cmd" >/dev/null 2>&1; then
        VERSION=$($cmd --version 2>/dev/null || $cmd version 2>/dev/null || echo "unknown")
        echo "  ✅ $name: $(echo "$VERSION" | head -1)"
    else
        if [ "$required" = "true" ]; then
            echo "  ❌ $name: 未安装 ($cmd)"
            echo "     安装方式: $install_cmd"
            MISSING_TOOLS=$((MISSING_TOOLS + 1))
        else
            echo "  ⚠️  $name: 未安装 ($cmd) — 可选"
            echo "     安装方式: $install_cmd"
            WARNINGS=$((WARNINGS + 1))
        fi
    fi
}

echo "[必需工具]"
check_tool "Go" "go" "https://go.dev/dl/" "true"
check_tool "protoc" "protoc" "https://github.com/protocolbuffers/protobuf/releases" "true"
check_tool "Node.js" "node" "https://nodejs.org/" "true"
check_tool "pnpm" "pnpm" "npm install -g pnpm" "true"
check_tool "Python3" "python3" "https://www.python.org/downloads/" "true"
check_tool "pip" "pip3" "python3 -m ensurepip --upgrade" "true"

echo ""
echo "[Protobuf 插件]"
check_tool "protoc-gen-go" "protoc-gen-go" "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" "true"
check_tool "protoc-gen-go-grpc" "protoc-gen-go-grpc" "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" "true"
check_tool "protoc-gen-ts" "protoc-gen-ts" "npm install -g protoc-gen-ts @bufbuild/protobuf" "false"
check_tool "tars2go" "tars2go" "go install github.com/TarsCloud/TarsGo/tars/tools/tars2go@latest" "false"

echo ""
echo "[Lint / 测试工具]"
check_tool "golangci-lint" "golangci-lint" "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" "false"
check_tool "ruff" "ruff" "pip install ruff" "false"
check_tool "jest" "jest" "npm install -g jest" "false"

echo ""
echo "[Go 代理设置]"
if [ -n "${GOPROXY:-}" ]; then
    echo "  ✅ GOPROXY=$GOPROXY"
else
    echo "  ⚠️  GOPROXY 未设置，建议 export GOPROXY=https://goproxy.cn,direct"
fi

echo ""
if [ $MISSING_TOOLS -gt 0 ]; then
    echo "❌ 缺少 $MISSING_TOOLS 个必需工具，请安装后重试"
    echo ""
    echo "快速安装命令汇总："
    echo "  Go:         https://go.dev/dl/"
    echo "  protoc:     brew install protobuf  (macOS)"
    echo "  Node.js:    brew install node       (macOS)"
    echo "  pnpm:       npm install -g pnpm"
    echo "  Python3:    brew install python@3.11 (macOS)"
    exit 1
else
    echo "✅ 所有必要工具已就绪"
    if [ $WARNINGS -gt 0 ]; then
        echo "  （有 $WARNINGS 个可选工具未安装）"
    fi
    exit 0
fi
