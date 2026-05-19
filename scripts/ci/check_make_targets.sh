#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==========================================="
echo "CaiRobot MVP Make Target 完整性检查"
echo "==========================================="
echo ""

ERRORS=0

ROOT_TARGETS="help bootstrap proto lint test unit integration coverage build package docs rules testcase-check comment-check ci clean"
GO_TARGETS="proto tars unit integration coverage lint build package clean help"
TS_TARGETS="install proto lint unit coverage build package clean help"
PY_TARGETS="install proto lint unit coverage build package clean help"

check_targets() {
    local dir=$1
    local expected=$2
    local label=$3
    
    if [ ! -f "$dir/Makefile" ]; then
        echo "[$label] ⚠️ Makefile 不存在，跳过检查"
        return
    fi
    
    echo "[$label] 检查 Make targets..."
    
    for target in $expected; do
        if grep -q "^${target}:" "$dir/Makefile"; then
            echo "  ✅ $target"
        else
            echo "  ❌ $target (缺失)"
            ERRORS=$((ERRORS + 1))
        fi
    done
}

check_targets "." "$ROOT_TARGETS" "Root"
check_targets "go" "$GO_TARGETS" "Go"
check_targets "typescript" "$TS_TARGETS" "TypeScript"
check_targets "python" "$PY_TARGETS" "Python"

echo ""
if [ $ERRORS -gt 0 ]; then
    echo "❌ 发现 $ERRORS 个缺失 target"
    exit 1
else
    echo "✅ 所有必需 target 已就绪"
    exit 0
fi
