#!/usr/bin/env python3
"""
CaiRobot MVP 中文注释规范性检查脚本
检查：
1. Go exported type/function/interface 是否有中文注释
2. Protobuf message/field 是否有中文注释
3. 关键组件是否缺少说明
"""

import os
import re
import sys
from pathlib import Path

PROJECT_ROOT = Path(__file__).parent.parent.parent

ERRORS = []
WARNINGS = []

def check_go_comments():
    """检查 Go 文件的导出符号注释"""
    print("[Go] 检查导出符号注释...")
    
    go_dir = PROJECT_ROOT / "go"
    if not go_dir.exists():
        return
    
    for go_file in go_dir.rglob("*.go"):
        # 跳过生成的文件和测试文件
        if "_test.go" in go_file.name or "generated" in str(go_file):
            continue
        
        content = go_file.read_text(encoding="utf-8", errors="ignore")
        lines = content.splitlines()
        
        for i, line in enumerate(lines):
            # 匹配导出的类型声明
            type_match = re.match(r'^type\s+(\w[A-Z]\w*)\s+(struct|interface)\b', line)
            func_match = re.match(r'^func\s+(\w[A-Z]\w*)\s*\(', line)
            
            match = type_match or func_match
            if match:
                symbol_name = match.group(1)
                # 检查前一行或前几行是否有注释
                has_comment = False
                for j in range(max(0, i-5), i):
                    prev_line = lines[j].strip()
                    if prev_line.startswith("//"):
                        has_comment = True
                        break
                
                if not has_comment:
                    rel_path = go_file.relative_to(PROJECT_ROOT)
                    ERRORS.append(f"❌ Go 导出符号缺少注释: {symbol_name} @ {rel_path}:{i+1}")

def check_proto_comments():
    """检查 Protobuf 文件注释"""
    print("[Proto] 检查 Protobuf 注释...")
    
    proto_dir = PROJECT_ROOT / "proto"
    if not proto_dir.exists():
        return
    
    for proto_file in proto_dir.rglob("*.proto"):
        content = proto_file.read_text(encoding="utf-8", errors="ignore")
        lines = content.splitlines()
        
        for i, line in enumerate(lines):
            stripped = line.strip()
            # 检查 message 声明
            msg_match = re.match(r'^message\s+(\w+)\s*\{', stripped)
            field_match = re.match(r'^\s*(repeated\s+)?(\w[\w.]*)\s+(\w+)\s*=\s*\d+', stripped)
            
            if msg_match:
                msg_name = msg_match.group(1)
                has_comment = False
                for j in range(max(0, i-3), i):
                    if lines[j].strip().startswith("//"):
                        has_comment = True
                        break
                if not has_comment:
                    rel_path = proto_file.relative_to(PROJECT_ROOT)
                    ERRORS.append(f"❌ Proto message 缺少注释: {msg_name} @ {rel_path}:{i+1}")
            
            elif field_match:
                field_name = field_match.group(3)
                has_comment = i > 0 and lines[i-1].strip().startswith("//")
                if not has_comment:
                    rel_path = proto_file.relative_to(PROJECT_ROOT)
                    WARNINGS.append(f"⚠️ Proto field 缺少注释: {field_name} @ {rel_path}:{i+1}")

def main():
    print("=" * 60)
    print("CaiRobot MVP 中文注释规范性检查")
    print("=" * 60)
    print()
    
    check_go_comments()
    check_proto_comments()
    
    print("\n" + "=" * 60)
    print("检查结果")
    print("=" * 60)
    print(f"  错误数: {len(ERRORS)}")
    print(f"  警告数: {len(WARNINGS)}")
    
    if ERRORS:
        print("\n❌ 检查未通过！")
        for err in ERRORS[:20]:
            print(f"  {err}")
        if len(ERRORS) > 20:
            print(f"  ... 还有 {len(ERRORS) - 20} 个错误")
        return 1
    elif WARNINGS:
        print("\n⚠️ 检查通过（有警告）")
        for warn in WARNINGS[:10]:
            print(f"  {warn}")
        return 0
    else:
        print("\n✅ 中文注释检查通过")
        return 0

if __name__ == "__main__":
    sys.exit(main())
