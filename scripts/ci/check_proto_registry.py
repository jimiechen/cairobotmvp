#!/usr/bin/env python3
"""
检查协议编号唯一性和注册表同步。

用于 GitHub Actions CI 的 proto-check job。

功能：
1. 扫描 proto/ 下所有 .proto 文件
2. 提取每个 message 内 enum Type 的 max 和 min
3. 检查 max + min 是否重复
4. 检查 max + min 是否登记到 docs/api/协议编号注册表.md
"""

import os
import re
import sys

PROTO_DIR = "proto"
REGISTRY_FILE = "docs/api/协议编号注册表.md"

def extract_max_min_from_proto(proto_file):
    """从 proto 文件提取所有 message 的 max 和 min。"""
    results = []
    
    with open(proto_file, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # 匹配 message 块
    message_pattern = re.compile(
        r'message\s+(\w+)\s*\{[^}]*enum\s+Type\s*\{([^}]*)\}',
        re.DOTALL
    )
    
    for match in message_pattern.finditer(content):
        message_name = match.group(1)
        enum_content = match.group(2)
        
        # 提取 max
        max_match = re.search(r'max\s*=\s*(\d+)', enum_content)
        # 提取 min
        min_match = re.search(r'min\s*=\s*(\d+)', enum_content)
        
        if max_match and min_match:
            max_val = int(max_match.group(1))
            min_val = int(min_match.group(1))
            results.append({
                'message': message_name,
                'max': max_val,
                'min': min_val,
                'file': proto_file
            })
    
    return results

def parse_registry():
    """解析协议编号注册表。"""
    registered = []
    
    if not os.path.exists(REGISTRY_FILE):
        return registered
    
    with open(REGISTRY_FILE, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # 匹配表格中的编号行
    # 格式：| max | min | ... |
    pattern = re.compile(r'\|\s*(\d+)\s*\|\s*(\d+)\s*\|')
    
    for match in pattern.finditer(content):
        max_val = int(match.group(1))
        min_val = int(match.group(2))
        registered.append((max_val, min_val))
    
    return registered

def main():
    print("=== 协议编号检查 ===\n")
    
    # 1. 扫描 proto 文件
    all_protos = []
    for root, dirs, files in os.walk(PROTO_DIR):
        for file in files:
            if file.endswith('.proto'):
                all_protos.append(os.path.join(root, file))
    
    print(f"扫描到 {len(all_protos)} 个 proto 文件\n")
    
    # 2. 提取所有 max + min
    all_entries = []
    for proto_file in all_protos:
        entries = extract_max_min_from_proto(proto_file)
        all_entries.extend(entries)
        for entry in entries:
            print(f"  发现：{entry['message']} (max={entry['max']}, min={entry['min']}) in {entry['file']}")
    
    print(f"\n共发现 {len(all_entries)} 个协议编号\n")
    
    # 3. 检查重复
    seen = {}
    duplicates = []
    for entry in all_entries:
        key = (entry['max'], entry['min'])
        if key in seen:
            duplicates.append({
                'key': key,
                'first': seen[key],
                'second': entry
            })
        else:
            seen[key] = entry
    
    if duplicates:
        print("错误：发现重复的协议编号！")
        for dup in duplicates:
            print(f"  max={dup['key'][0]}, min={dup['key'][1]}:")
            print(f"    - {dup['first']['message']} in {dup['first']['file']}")
            print(f"    - {dup['second']['message']} in {dup['second']['file']}")
        sys.exit(1)
    
    print("成功：没有重复的协议编号\n")
    
    # 4. 检查注册表
    registered = parse_registry()
    print(f"注册表中有 {len(registered)} 个编号\n")
    
    unregistered = []
    for entry in all_entries:
        key = (entry['max'], entry['min'])
        if key not in registered:
            unregistered.append(entry)
    
    if unregistered:
        print("警告：以下协议编号未登记到注册表：")
        for entry in unregistered:
            print(f"  - {entry['message']} (max={entry['max']}, min={entry['min']}) in {entry['file']}")
        print(f"\n请更新 {REGISTRY_FILE}")
        sys.exit(1)
    
    print("成功：所有协议编号都已登记\n")
    
    # 5. 输出已检查的编号列表
    print("=== 已检查的协议编号 ===")
    for entry in sorted(all_entries, key=lambda x: (x['max'], x['min'])):
        print(f"  {entry['max']}:{entry['min']} -> {entry['message']}")
    
    print(f"\n成功：协议编号检查通过，共 {len(all_entries)} 个编号")
    sys.exit(0)

if __name__ == "__main__":
    main()
