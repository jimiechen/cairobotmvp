#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==> 检查生成代码是否已提交到 Git..."

HAS_CHANGES=false

for dir in proto/generated/go proto/generated/ts proto/generated/python proto/generated/tarsgo; do
    if [ -d "$dir" ]; then
        if ! git diff --quiet -- "$dir" 2>/dev/null; then
            echo "❌ 错误：$dir 有未提交的生成代码变更"
            HAS_CHANGES=true
        fi
        if ! git ls-files --error-unmatch -- "$dir" >/dev/null 2>&1; then
            if [ -n "$(ls -A "$dir" 2>/dev/null)" ]; then
                echo "❌ 错误：$dir 包含文件但未被 Git 追踪"
                HAS_CHANGES=true
            fi
        fi
    fi
done

if [ "$HAS_CHANGES" = true ]; then
    echo ""
    echo "请执行以下命令更新生成代码："
    echo "  make proto"
    echo "  git add proto/generated/"
    echo "  git commit -m 'chore(proto): 更新生成代码'"
    exit 1
else
    echo "✅ 所有生成代码已正确提交"
    exit 0
fi
