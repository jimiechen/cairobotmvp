#!/usr/bin/env bash
set -euo pipefail

# Go 模块 CI 检查脚本
# 从项目根目录运行

cd "$(dirname "$0")/../.."

echo "==> 检查 Go 模块..."
cd go

modules=(
  "gateway/proto-gateway"
  "tars/system"
)

failed=0

for module in "${modules[@]}"; do
  echo "==> go test ./$module/..."
  if ! (cd "$module" && go test ./...); then
    echo "FAILED: $module"
    failed=1
  fi
done

if [ $failed -eq 1 ]; then
  echo ""
  echo "部分模块测试失败"
  exit 1
fi

echo ""
echo "所有 Go 模块测试通过"
