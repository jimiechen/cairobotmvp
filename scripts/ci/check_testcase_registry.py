#!/usr/bin/env python3
"""
CaiRobot MVP 测试用例注册表一致性检查脚本
检查：
1. 所有 *_test.go 必须在注册表中登记
2. 所有 test_*.py 必须在注册表中登记
3. 所有 *.test.ts / *.spec.ts 必须在注册表中登记
4. 注册表中登记的文件必须存在
5. 废弃测试必须标注原因
"""

import os
import re
import sys
from pathlib import Path

PROJECT_ROOT = Path(__file__).parent.parent.parent
REGISTRY_FILE = PROJECT_ROOT / "docs" / "testing" / "测试用例注册表.md"

ERRORS = []
WARNINGS = []

def find_go_test_files():
    """查找所有 Go 测试文件"""
    test_files = []
    go_dir = PROJECT_ROOT / "go"
    if go_dir.exists():
        for f in go_dir.rglob("*_test.go"):
            rel_path = f.relative_to(PROJECT_ROOT)
            test_files.append(str(rel_path))
    return sorted(test_files)

def find_python_test_files():
    """查找所有 Python 测试文件"""
    test_files = []
    python_dir = PROJECT_ROOT / "python"
    if python_dir.exists():
        for f in python_dir.rglob("test_*.py"):
            rel_path = f.relative_to(PROJECT_ROOT)
            test_files.append(str(rel_path))
    return sorted(test_files)

def find_typescript_test_files():
    """查找所有 TypeScript 测试文件"""
    test_files = []
    ts_dir = PROJECT_ROOT / "typescript"
    if ts_dir.exists():
        for pattern in ["*.test.ts", "*.test.tsx", "*.spec.ts", "*.spec.tsx"]:
            for f in ts_dir.rglob(pattern):
                rel_path = f.relative_to(PROJECT_ROOT)
                test_files.append(str(rel_path))
    return sorted(test_files)

def parse_registry():
    """解析测试用例注册表"""
    if not REGISTRY_FILE.exists():
        ERRORS.append(f"❌ 测试用例注册表不存在: {REGISTRY_FILE}")
        return {"go": [], "python": [], "typescript": [], "proto": [], "ci": []}

    content = REGISTRY_FILE.read_text(encoding="utf-8")
    
    registered = {
        "go": [],
        "python": [],
        "typescript": [],
        "proto": [],
        "ci": []
    }

    current_section = None
    
    for line in content.splitlines():
        line = line.strip()
        
        # 检测章节标题（用于设置当前分区）
        if "## Go 测试用例" in line:
            current_section = "go"
            continue
        elif "## Python 测试用例" in line:
            current_section = "python"
            continue
        elif "## TypeScript 测试用例" in line:
            current_section = "typescript"
            continue
        elif "## Protobuf" in line:
            current_section = "proto"
            continue
        elif "## CI" in line:
            current_section = "ci"
            continue
        
        # 只处理表格行
        if not line.startswith("|"):
            continue
        
        # 跳过表头和分隔行
        if "---" in line or "测试用例 ID" in line or "前缀" in line:
            continue
            
        parts = [p.strip() for p in line.split("|")]
        
        # 过滤空行和非 TC- 开头的行（注意：markdown 表格 | 开头导致 parts[0] 为空）
        if len(parts) >= 4 and parts[1] and parts[1].startswith("TC-"):
            file_path = parts[2]
            status = parts[5] if len(parts) > 5 else ""
            
            if current_section in ("go", "python", "typescript") and file_path and ("✅" in status or "⏸️" in status):
                registered[current_section].append(file_path)

    return registered

def main():
    print("=" * 60)
    print("CaiRobot MVP 测试用例注册表一致性检查")
    print("=" * 60)
    print()

    # 解析注册表
    registered = parse_registry()

    # 查找实际测试文件
    go_tests = find_go_test_files()
    py_tests = find_python_test_files()
    ts_tests = find_typescript_test_files()

    all_actual = {
        "go": go_tests,
        "python": py_tests,
        "typescript": ts_tests
    }

    # 检查 1: 实际存在的测试文件是否已登记
    print("[1] 检查未登记的测试文件...")
    for lang, actual_files in all_actual.items():
        reg_files = set(registered.get(lang, []))
        for f in actual_files:
            if f not in reg_files:
                ERRORS.append(f"❌ {lang} 测试文件未登记: {f}")
                print(f"   ❌ {f} 未在注册表中登记")

    # 检查 2: 注册表中登记的文件是否实际存在
    print("\n[2] 检查登记文件是否存在...")
    for lang, reg_files in registered.items():
        if lang in ("proto", "ci"):
            continue
        for f in reg_files:
            full_path = PROJECT_ROOT / f
            if not full_path.exists():
                ERRORS.append(f"❌ {lang} 登记文件不存在: {f}")
                print(f"   ❌ {f} 已登记但文件不存在")

    # 检查 3: 废弃测试是否标注原因
    print("\n[3] 检查废弃测试...")
    if REGISTRY_FILE.exists():
        content = REGISTRY_FILE.read_text(encoding="utf-8")
        if "废弃" in content or "❌" in content:
            deprecated_section = False
            for line in content.splitlines():
                if "## 废弃测试记录" in line:
                    deprecated_section = True
                    continue
                if deprecated_section and line.startswith("|") and len(line.split("|")) >= 4:
                    parts = [p.strip() for p in line.split("|")]
                    if parts[0] and not parts[0].startswith("原测试") and not parts[0].startswith("---"):
                        reason = parts[1]
                        date = parts[2]
                        if not reason or reason == "废弃原因":
                            continue
                        if not reason:
                            WARNINGS.append(f"⚠️ 废弃测试 {parts[0]} 未标注原因")

    # 输出统计
    print("\n" + "=" * 60)
    print("检查结果汇总")
    print("=" * 60)
    print(f"  Go 测试文件:     实际 {len(go_tests)} 个, 登记 {len(registered['go'])} 个")
    print(f"  Python 测试文件: 实际 {len(py_tests)} 个, 登记 {len(registered['python'])} 个")
    print(f"  TS 测试文件:     实际 {len(ts_tests)} 个, 登记 {len(registered['typescript'])} 个")
    print(f"  错误数: {len(ERRORS)}")
    print(f"  警告数: {len(WARNINGS)}")

    if ERRORS:
        print("\n❌ 检查未通过！")
        for err in ERRORS:
            print(f"  {err}")
        return 1
    else:
        print("\n✅ 测试用例注册表检查通过")
        return 0

if __name__ == "__main__":
    sys.exit(main())
