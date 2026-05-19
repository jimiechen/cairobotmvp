#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==> 生成 Python 覆盖率报告..."

OUTPUT_DIR="docs/reports/coverage/python"
mkdir -p "$OUTPUT_DIR"

AI_DIR="python/ai"

if [ -d "$AI_DIR" ]; then
    echo "[py-coverage] 处理 AI 模块..."
    cd "$AI_DIR"
    
    if python3 -m pytest --version >/dev/null 2>&1; then
        python3 -m pytest \
            --cov=. \
            --cov-report=html:"../../$OUTPUT_DIR" \
            --cov-report=term-missing \
            --cov-report=xml:"../../$OUTPUT_DIR/coverage.xml" \
            -v 2>/dev/null || echo "  警告：pytest 执行失败"
    else
        echo "  跳过：pytest 未安装"
    fi
    
    cd ../..
else
    echo "[py-coverage] 跳过：python/ai 目录不存在"
fi

echo "==> Python 覆盖率报告已生成到 $OUTPUT_DIR/"
