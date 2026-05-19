#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==> 生成 Go 覆盖率报告..."

OUTPUT_DIR="docs/reports/coverage/go"
mkdir -p "$OUTPUT_DIR"

GO_WORK="go/go.work"

if [ ! -f "$GO_WORK" ]; then
    echo "[go-coverage] 跳过：go.work 不存在"
    exit 0
fi

TOTAL_COVERAGE=0
MODULE_COUNT=0

for mod in gateway/proto-gateway tars/system services; do
    if [ -f "$mod/go.mod" ]; then
        echo "[go-coverage] 处理模块: $mod"
        cd "$mod"
        
        if go test ./... -coverprofile=cover.out -covermode=count >/dev/null 2>&1; then
            if [ -s cover.out ]; then
                go tool cover -html=cover.out -o "../../$OUTPUT_DIR/${mod//\//-}_coverage.html" 2>/dev/null || true
                
                PCT=$(go tool cover -func=cover.out 2>/dev/null | grep total | awk '{print $$3}' | tr -d '%')
                echo "  覆盖率: ${PCT:-N/A}%"
                
                if [ -n "$PCT" ] && [ "$PCT" != "N/A" ]; then
                    TOTAL_COVERAGE=$(echo "$TOTAL_COVERAGE + $PCT" | bc 2>/dev/null || echo "$TOTAL_COVERAGE")
                    MODULE_COUNT=$((MODULE_COUNT + 1))
                fi
            fi
        else
            echo "  警告：测试未通过，跳过覆盖率收集"
        fi
        
        rm -f cover.out
        cd ..
    fi
done

if [ $MODULE_COUNT -gt 0 ] && [ -n "$TOTAL_COVERAGE" ]; then
    AVG=$(echo "scale=1; $TOTAL_COVERAGE / $MODULE_COUNT" | bc 2>/dev/null || echo "N/A")
    echo ""
    echo "==> Go 平均覆盖率: ${AVG}%"
else
    echo "==> Go 覆盖率报告已生成到 $OUTPUT_DIR/"
fi
