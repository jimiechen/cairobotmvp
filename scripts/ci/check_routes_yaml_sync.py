#!/usr/bin/env python3
"""
check_routes_yaml_sync.py — 校验 Gateway routes.yaml 双副本一致性

用途：
  防止 configs/gateway/routes.yaml（权威版本）与
  go/gateway/proto-gateway/configs/gateway/routes.yaml（开发副本）
  出现漂移，避免再次触发 BUG-E2E-001 类路由缺失问题。

调用方式：
  make rules          # 通过 make rules 编排调用
  python3 scripts/ci/check_routes_yaml_sync.py  # 直接执行

退出码：
  0   两文件一致（或开发副本为符号链接指向权威版本）
  1   两文件不一致，阻断构建
  2   权威版本不存在

关联文档：
  - BUG-E2E-003 (E2E-ISSUES.md)
  - ADR-0012 (polyglot monorepo directory layout)
"""

import os
import sys
import difflib
import pathlib

# 项目根目录（以本脚本位置为基准推算）
SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
PROJECT_ROOT = SCRIPT_DIR.parent.parent

# 两个副本路径（相对于项目根）
CANONICAL_REL = "configs/gateway/routes.yaml"
DEV_COPY_REL = "go/gateway/proto-gateway/configs/gateway/routes.yaml"


def main():
    canonical = PROJECT_ROOT / CANONICAL_REL
    dev_copy = PROJECT_ROOT / DEV_COPY_REL

    # 检查权威版本是否存在
    if not canonical.exists():
        print(f"ERROR: 权威版本不存在: {canonical}")
        print("       请确认 configs/gateway/routes.yaml 已提交到仓库")
        return 2

    # 开发副本不存在 → 提示但不算失败（可能是首次 checkout）
    if not dev_copy.exists():
        print(f"WARN: 开发副本不存在: {dev_copy}")
        print("      建议从权威版本复制或创建符号链接:")
        print(f"      cp {canonical} {dev_copy}")
        print(f"      ln -sf ../../../{CANONICAL_REL} {dev_copy}")
        # 不阻断：首次构建可能尚未同步
        return 0

    # 开发副本是符号链接 → 检查链接目标
    if dev_copy.is_symlink():
        target = os.readlink(dev_copy)
        # 解析相对路径
        resolved = (dev_copy.parent / target).resolve()
        canonical_resolved = canonical.resolve()
        if resolved == canonical_resolved:
            print(f"OK: {DEV_COPY_REL} -> 符号链接 -> {target} (指向权威版本)")
            return 0
        else:
            print(f"ERROR: {DEV_COPY_REL} 是符号链接但指向错误目标")
            print(f"       链接目标: {resolved}")
            print(f"       期望指向: {canonical_resolved}")
            return 1

    # 常规文件对比
    try:
        canonical_text = canonical.read_text(encoding="utf-8")
        dev_copy_text = dev_copy.read_text(encoding="utf-8")
    except Exception as e:
        print(f"ERROR: 读取文件失败: {e}")
        return 1

    if canonical_text == dev_copy_text:
        line_count = canonical_text.count("\n") + 1
        route_count = canonical_text.lower().count("request_max:")
        print(f"OK: routes.yaml 双副本一致 ({line_count} 行, {route_count} 条路由)")
        return 0

    # 不一致 → 输出 diff
    print(f"FAIL: routes.yaml 双副本不一致!")
    print(f"\n  权威版本: {canonical} ({canonical.stat().st_size} bytes)")
    print(f"  开发副本: {dev_copy} ({dev_copy.stat().st_size} bytes)")
    print("\n  差异:")

    canonical_lines = canonical_text.splitlines(keepends=True)
    dev_lines = dev_copy_text.splitlines(keepends=True)

    diff = list(difflib.unified_diff(
        canonical_lines,
        dev_lines,
        fromfile=f"权威 ({CANONICAL_REL})",
        tofile=f"开发副本 ({DEV_COPY_REL})",
        lineterm="",
    ))

    for line in diff:
        # 限制输出行数，避免刷屏
        print(f"  {line}")

    print("\n  修复命令:")
    print(f"    cp {canonical} {dev_copy}")
    print(f"    # 或使用符号链接（推荐）:")
    print(f"    rm {dev_copy} && ln -sf ../../../{CANONICAL_REL} {dev_copy}")

    return 1


if __name__ == "__main__":
    sys.exit(main())
