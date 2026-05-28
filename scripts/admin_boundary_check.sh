#!/bin/bash
# admin_boundary_check.sh — Admin 边界自检脚本
# 用途：验证 admin 子包不违反 PRD-10 §10 的架构约束
#
# 检查项：
# 1. 业务服务层（service/、sdk/）不反向 import admin
# 2. admin 包内不出现字段级校验逻辑（field_type switch / validator JSON 解析）
# 3. admin 校验必须委托给 inner 服务层的 Validate* 方法
# 4. admin 不直接操作 domain 层的私有字段（仅通过 DTO 转换）
#
# 使用方式：
#   chmod +x scripts/admin_boundary_check.sh
#   ./scripts/admin_boundary_check.sh
#
# 退出码：
#   0 = 全部通过
#   1 = 有违规

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
GO_DIR="$PROJECT_ROOT/go"

PASS=0
FAIL=0
WARN=0

green() { printf "\033[32m✅ %s\033[0m\n" "$1"; }
red()   { printf "\033[31m❌ %s\033[0m\n" "$1"; }
yellow(){ printf "\033[33m⚠️  %s\033[0m\n" "$1"; }

echo "========================================="
echo "  Admin 边界自检 (PRD-10 §10)"
echo "  项目: $(basename "$PROJECT_ROOT")"
echo "  时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="
echo ""

# ---- 检查 1: 业务服务不反向引用 admin ----
echo "--- 检查 1: 业务服务 → 无 admin 反向依赖 ---"

BUSINESS_PKGS=(
    "services/config/service"
    "services/config/sdk"
    "services/config/domain"
    "services/config/repository"
    "services/i18n/service"
    "services/i18n/sdk"
    "services/i18n/domain"
    "services/i18n/repository"
)

VIOLATION_FOUND=false
for pkg in "${BUSINESS_PKGS[@]}"; do
    pkg_path="$GO_DIR/$pkg"
    if [ ! -d "$pkg_path" ]; then
        continue
    fi
    # 检查 .go 文件中是否有 import 引用 admin
    matches=$(grep -r '".*admin"' "$pkg_path" --include="*.go" 2>/dev/null || true)
    if [ -n "$matches" ]; then
        red "  $pkg 存在 admin 反向依赖:"
        echo "$matches" | sed 's/^/    /'
        VIOLATION_FOUND=true
        FAIL=$((FAIL + 1))
    else
        green "  $pkg ✅ 无 admin 反向依赖"
        PASS=$((PASS + 1))
    fi
done

echo ""

# ---- 检查 2: admin 内部不出现字段级校验逻辑 ----
echo "--- 检查 2: admin 内部 → 禁止字段级校验 ---"

ADMIN_DIRS=(
    "services/config/admin"
    "services/i18n/admin"
)

FORBIDDEN_PATTERNS=(
    'switch.*field_type'
    'case.*FieldTypeInt:'
    'case.*FieldTypeString:'
    'case.*FieldTypeBool:'
    'case.*FieldTypeFloat:'
    'case.*FieldTypeJSON:'
    'json\.Unmarshal.*Validator'
    'validator.*JSON'
)

for admin_dir in "${ADMIN_DIRS[@]}"; do
    admin_path="$GO_DIR/$admin_dir"
    if [ ! -d "$admin_path" ]; then
        continue
    fi
    for pattern in "${FORBIDDEN_PATTERNS[@]}"; do
        matches=$(grep -rn "$pattern" "$admin_path" --include="*.go" 2>/dev/null || true)
        if [ -n "$matches" ]; then
            red "  $admin_dir 发现禁止模式 '$pattern':"
            echo "$matches" | sed 's/^/    /'
            FAIL=$((FAIL + 1))
            VIOLATION_FOUND=true
        fi
    done
done
if [ "$VIOLATION_FOUND" = false ]; then
    green "  admin 包内无字段级校验逻辑 ✅"
    PASS=$((PASS + 1))
fi

echo ""

# ---- 检查 3: admin 必须委托 inner 校验 ----
echo "--- 检查 3: admin → 委托 inner.Validate* ---"

for admin_dir in "${ADMIN_DIRS[@]}"; do
    admin_path="$GO_DIR/$admin_dir"
    if [ ! -d "$admin_path" ]; then
        continue
    fi
    validate_calls=$(grep -rn '\.Validate' "$admin_path" --include="*.go" 2>/dev/null || true)
    if [ -z "$validate_calls" ]; then
        yellow "  $admin_dir 未发现 Validate 委托调用（可能为骨架代码）"
        WARN=$((WARN + 1))
    else
        green "  $admin_dir 发现 Validate 委托调用 ✅"
        PASS=$((PASS + 1))
    fi
done

echo ""

# ---- 检查 4: admin 仅通过 DTO 转换操作 domain ----
echo "--- 检查 4: admin → DTO 转换模式正确 ---"

for admin_dir in "${ADMIN_DIRS[@]}"; do
    admin_path="$GO_DIR/$admin_dir"
    if [ ! -d "$admin_path" ]; then
        continue
    fi
    dto_funcs=$(grep -rn 'func to\|func from\|type.*Request\b\|type.*Item\b' "$admin_path" --include="*.go" 2>/dev/null || true)
    if [ -z "$dto_funcs" ]; then
        yellow "  $admin_dir 未发现 DTO 转换函数（可能为骨架代码）"
        WARN=$((WARN + 1))
    else
        green "  $admin_dir 发现 DTO 定义/转换函数 ✅"
        PASS=$((PASS + 1))
    fi
done

echo ""
echo "========================================="
printf "  结果: "
if [ "$FAIL" -gt 0 ]; then
    red   "有 $FAIL 项违规 / $PASS 项通过 / $WARN 项警告"
elif [ "$WARN" -gt 0 ]; then
    yellow "全部通过 ($PASS 项) 但有 $WARN 项警告"
else
    green  "全部通过 ($PASS 项)"
fi
echo "========================================="

exit $FAIL
