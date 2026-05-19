#!/usr/bin/env python3
"""
CaiRobot MVP 目录布局一致性检查脚本
验证项目目录是否符合 ADR-0012 多语言 Monorepo 布局规范
"""

import os
import sys
from pathlib import Path

PROJECT_ROOT = Path(__file__).parent.parent.parent

REQUIRED_DIRS = [
    "docs/prd",
    "docs/adr", 
    "docs/api",
    "docs/testing",
    "docs/reports",
    "docs/wiki",
    "proto",
    "scripts",
]

OPTIONAL_DIRS = [
    "go",
    "python",
    "typescript",
    "services",
    "web",
    "tars",
    "firmware",
    "app",
    "hardware",
]

FORBIDDEN_IN_ROOT = [
    "*.md",
    "*.tmp",
    "*draft*",
    "*new*",
    "*final*",
]

ERRORS = []
WARNINGS = []

def main():
    print("=" * 60)
    print("CaiRobot MVP 目录布局检查")
    print("=" * 60)
    print()
    
    print("[1] 检查必备目录...")
    for dir_name in REQUIRED_DIRS:
        dir_path = PROJECT_ROOT / dir_name
        if dir_path.exists() and dir_path.is_dir():
            print(f"  ✅ {dir_name}/")
        else:
            ERRORS.append(f"❌ 缺少必备目录: {dir_name}/")
            print(f"  ❌ {dir_name}/ (缺失)")
    
    print("\n[2] 检查可选目录状态...")
    for dir_name in OPTIONAL_DIRS:
        dir_path = PROJECT_ROOT / dir_name
        if dir_path.exists():
            print(f"  ✅ {dir_name}/ (存在)")
        else:
            print(f"  ⏭️  {dir_name}/ (暂未创建)")
    
    print("\n[3] 检查根目录禁止文件...")
    import glob
    for pattern in FORBIDDEN_IN_ROOT:
        matches = glob.glob(str(PROJECT_ROOT / pattern))
        for m in matches:
            file_name = Path(m).name
            if file_name not in ("README.md", "AGENTS.md", ".gitignore", "LICENSE", "Makefile"):
                ERRORS.append(f"❌ 根目录不应有文件: {file_name}")
                print(f"  ❌ {file_name}")
    
    print("\n[4] 检查 scripts/ 子目录...")
    expected_scripts_subdirs = ["ci", "proto", "coverage", "dev"]
    for sub in expected_scripts_subdirs:
        if (PROJECT_ROOT / "scripts" / sub).exists():
            print(f"  ✅ scripts/{sub}/")
        else:
            WARNINGS.append(f"⚠️ scripts/{sub}/ 不存在")
            print(f"  ⚭️ scripts/{sub}/ (缺失)")
    
    print("\n" + "=" * 60)
    print(f"结果: {len(ERRORS)} 错误, {len(WARNINGS)} 警告")
    
    if ERRORS:
        print("\n❌ 目录布局检查未通过！")
        for err in ERRORS:
            print(f"  {err}")
        return 1
    else:
        print("\n✅ 目录布局检查通过")
        return 0

if __name__ == "__main__":
    sys.exit(main())
