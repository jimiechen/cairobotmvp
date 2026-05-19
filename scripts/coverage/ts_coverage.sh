#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==> 生成 TypeScript 覆盖率报告..."

OUTPUT_DIR="docs/reports/coverage/typescript"
mkdir -p "$OUTPUT_DIR"

for pkg in web admin-web; do
    PKG_DIR="typescript/$pkg"
    
    if [ -f "$PKG_DIR/package.json" ]; then
        echo "[ts-coverage] 处理包: $pkg"
        cd "$PKG_DIR"
        
        if command -v npx >/dev/null 2>&1; then
            npx jest --coverage --coverageDirectory="../../$OUTPUT_DIR/$pkg" \
                --passWithNoTests --coverageReporters=html --coverageReporters=text \
                --coverageReporters=json-summary 2>/dev/null || echo "  跳过：Jest 未配置"
        else
            echo "  跳过：npx 不可用"
        fi
        
        cd ../..
    fi
done

echo "==> TypeScript 覆盖率报告已生成到 $OUTPUT_DIR/"
