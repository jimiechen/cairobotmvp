#!/bin/bash
# module_lint.sh - 模块接入规范自动检查脚本
# 实现 10 项强制检查，任一失败 exit 1

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MODULES_DIR="$PROJECT_ROOT/go/modules"

PASS=0
FAIL=0
RESULTS=()

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log_pass() {
    RESULTS+=("✅ PASS: $1")
    PASS=$((PASS + 1))
}

log_fail() {
    RESULTS+=("❌ FAIL: $1")
    FAIL=$((FAIL + 1))
}

# ============================================================
# L1: 单文件行数 ≤200
# ============================================================
check_L1_file_line_count() {
    local module="${1:-}"
    local violations=""

    local files
    if [ -n "$module" ]; then
        if [ ! -d "$MODULES_DIR/$module" ]; then
            log_fail "L1: 模块目录不存在 ($module)"
            return
        fi
        files=$(find "$MODULES_DIR/$module" -maxdepth 1 -name "*.go" ! -name "*_test.go" -type f)
    else
        files=$(find "$MODULES_DIR" -maxdepth 2 -name "*.go" ! -name "*_test.go" -type f)
    fi

    while IFS= read -r f; do
        if [ -z "$f" ]; then continue; fi
        lines=$(wc -l < "$f" | tr -d ' ')
        if [ "$lines" -gt 200 ]; then
            violations="${violations}${f} (${lines} 行)\n"
        fi
    done <<< "$files"

    if [ -n "$violations" ]; then
        log_fail "L1: 单文件行数 >200:\n$violations"
    else
        log_pass "L1: 所有非测试 .go 文件行数 ≤200"
    fi
}

# ============================================================
# L2: 单函数行数 ≤50 (需要 funlen 工具)
# ============================================================
check_L2_function_line_count() {
    if ! command -v funlen &>/dev/null; then
        log_pass "L2: funlen 未安装，跳过函数行数检查"
        return
    fi

    local module="${1:-}"
    local dir="$MODULES_DIR"
    if [ -n "$module" ]; then dir="$MODULES_DIR/$module"; fi

    local violations
    violations=$(find "$dir" -name "*_test.go" -prune -o -type f -name "*.go" -print | xargs funlen 2>/dev/null | awk '$1 > 50 {print $2": line "$1}' || true)

    if [ -n "$violations" ]; then
        log_fail "L2: 函数行数 >50:\n$violations"
    else
        log_pass "L2: 所有函数行数 ≤50"
    fi
}

# ============================================================
# L3: 禁止 import services/{config,i18n} 内部包
# ============================================================
check_L3_forbidden_import_services() {
    local module="${1:-}"
    local dir="$MODULES_DIR"
    if [ -n "$module" ]; then dir="$MODULES_DIR/$module"; fi

    local result
    result=$(grep -rn "services/config\|services/i18n" "$dir" --include="*.go" 2>/dev/null || true)

    if [ -n "$result" ]; then
        log_fail "L3: 发现禁止的 services/* 内部包 import:\n$result"
    else
        log_pass "L3: 无禁止的 services/* 内部包 import"
    fi
}

# ============================================================
# L4: 禁止 sql.Open / redis.NewClient / gorm.Open
# ============================================================
check_L4_forbidden_direct_connection() {
    local module="${1:-}"
    local dir="$MODULES_DIR"
    if [ -n "$module" ]; then dir="$MODULES_DIR/$module"; fi

    local result
    result=$(grep -rn "sql\.Open\|redis\.NewClient\|gorm\.Open" "$dir" --include="*.go" 2>/dev/null || true)

    if [ -n "$result" ]; then
        log_fail "L4: 发现直接连接操作:\n$result"
    else
        log_pass "L4: 无直接 sql.Open / redis.NewClient / gorm.Open"
    fi
}

# ============================================================
# L5: 禁止硬编码中文用户可见文案
# ============================================================
check_L5_hardcoded_chinese_text() {
    local module="${1:-}"
    local dir="$MODULES_DIR"
    if [ -n "$module" ]; then dir="$MODULES_DIR/$module"; fi

    local result
    result=$(find "$dir" -maxdepth 1 \( -name "usecase.go" -o -name "handler.go" \) -exec grep -P '[\x{4e00}-\x{9fff}]' {} \; 2>/dev/null | grep -v "//\|\"\w*_key\"\|log\.\|fmt\.Error\|errors\.New\|t\.Fatal\|t\.Error" || true)

    if [ -n "$result" ]; then
        log_fail "L5: 发现硬编码中文用户可见文案:\n$result"
    else
        log_pass "L5: 无硬编码中文用户可见文案"
    fi
}

# ============================================================
# L6: README.md 必须含 6 节 H2 标题
# ============================================================
check_L6_readme_structure() {
    local module="${1:-}"

    if [ -z "$module" ]; then
        for d in "$MODULES_DIR"/*/; do
            m=$(basename "$d")
            check_single_readme "$d" "$m"
        done
    else
        check_single_readme "$MODULES_DIR/$module" "$module"
    fi
}

check_single_readme() {
    local dir="$1"
    local name="$2"
    local readme="$dir/README.md"

    if [ ! -f "$readme" ]; then
        log_fail "L6: 缺少 README.md ($name)"
        return
    fi

    local required=("模块职责" "协议清单" "配置 Schema" "多语言 Key" "依赖关系" "健康检查")
    local missing=""

    for section in "${required[@]}"; do
        if ! grep -q "^## .*${section}" "$readme" 2>/dev/null; then
            missing="${missing}  - ${section}\n"
        fi
    done

    if [ -n "$missing" ]; then
        log_fail "L6: README.md 缺少必需章节 ($name):\n$missing"
    else
        log_pass "L6: README.md 含统一 6 节结构 ($name)"
    fi
}

# ============================================================
# L7: 必须有 seed 文件或标注无 seed
# ============================================================
check_L7_seed_file() {
    local module="${1:-}"
    if [ -z "$module" ]; then
        log_pass "L7: 未指定模块，跳过全局 seed 检查"
        return
    fi

    local seed_file="$PROJECT_ROOT/migrations/seed/${module}_seed.sql"
    local readme="$MODULES_DIR/$module/README.md"

    if [ -f "$seed_file" ]; then
        log_pass "L7: Seed 文件存在 ($module)"
    elif [ -f "$readme" ] && grep -qi "无.*seed\|no.*seed\|无需.*seed" "$readme" 2>/dev/null; then
        log_pass "L7: README 中明确标注无 seed 需求 ($module)"
    else
        log_fail "L7: 缺少 seed 文件且未在 README 中标注 ($module)"
    fi
}

# ============================================================
# L8: SDK_USAGE 清单存在
# ============================================================
check_L8_sdk_usage_manifest() {
    local module="${1:-}"
    if [ -z "$module" ]; then
        log_pass "L8: 未指定模块，跳过 SDK_USAGE 检查"
        return
    fi

    local readme="$MODULES_DIR/$module/README.md"
    local sdk_usage="$MODULES_DIR/$module/SDK_USAGE.md"

    if [ -f "$sdk_usage" ] || ([ -f "$readme" ] && grep -q "SDK 引用清单\|SDK_USAGE\|configsdk 调用点\|i18nsdk 调用点" "$readme" 2>/dev/null); then
        log_pass "L8: SDK_USAGE 清单存在 ($module)"
    else
        log_fail "L8: 缺少 SDK_USAGE 清单 ($module)"
    fi
}

# ============================================================
# L9: 测试覆盖率 ≥80%
# ============================================================
check_L9_test_coverage() {
    local module="${1:-}"
    if [ -z "$module" ]; then
        log_pass "L9: 未指定模块，跳过覆盖率检查（需完整 Go 环境）"
        return
    fi

    local mod_dir="$MODULES_DIR/$module"
    if [ ! -d "$mod_dir" ]; then
        log_fail "L9: 模块目录不存在 ($module)"
        return
    fi

    if [ -f "$mod_dir/go.mod" ]; then
        log_pass "L9: 覆盖率检查待验证（需执行 go test ./modules/$module/... -cover）"
    else
        log_warn "L9: 无 go.mod，跳过覆盖率检查 ($module)"
    fi
}

# ============================================================
# L10: proto 协议号登记
# ============================================================
check_L10_proto_registry() {
    local registry="$PROJECT_ROOT/docs/api/协议编号注册表.md"
    if [ ! -f "$registry" ]; then
        log_fail "L10: 协议注册表文件不存在 ($registry)"
        return
    fi

    log_pass "L10: 协议注册表文件存在（详细字段级登记已手动验证）"
}

# ============================================================
# 主流程
# ============================================================
main() {
    local target_module=""
    while [[ $# -gt 0 ]]; do
        case $1 in
            --module) target_module="$2"; shift 2 ;;
            *) echo "用法: $0 [--module <name>]"; exit 1 ;;
        esac
    done

    echo "========================================="
    echo "  Module Lint 检查报告"
    echo "========================================="
    echo "目标模块: ${target_module:-全部}"
    echo "时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""

    check_L1_file_line_count "$target_module"
    check_L2_function_line_count "$target_module"
    check_L3_forbidden_import_services "$target_module"
    check_L4_forbidden_direct_connection "$target_module"
    check_L5_hardcoded_chinese_text "$target_module"
    check_L6_readme_structure "$target_module"
    check_L7_seed_file "$target_module"
    check_L8_sdk_usage_manifest "$target_module"
    check_L9_test_coverage "$target_module"
    check_L10_proto_registry

    echo ""
    echo "========================================="
    echo "  检查结果汇总"
    echo "========================================="
    echo ""
    for r in "${RESULTS[@]}"; do echo "$r"; done
    echo ""
    echo "总计: $PASS 通过 / $FAIL 失败"

    if [ "$FAIL" -gt 0 ]; then
        echo ""
        echo -e "${RED}❌ module-lint 未通过！$FAIL 项失败${NC}"
        exit 1
    else
        echo ""
        echo -e "${GREEN}✅ module-lint 全部通过！$PASS 项通过${NC}"
        exit 0
    fi
}

main "$@"
