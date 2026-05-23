#!/usr/bin/env python3
"""
Protobuf 生成代码校验（CI 用）。

功能：
1. 校验 proto/ 下 .proto 文件的协议编号唯一性和注册表一致性
2. 校验各语言生成代码文件是否存在（不需要 protoc）

CI 调用方式：make proto-check 或 python3 scripts/ci/check_proto_registry.py

设计原则：
- 本地开发：make proto → 生成代码 → git commit 提交生成代码
- CI 检查：make proto-check → 只校验生成文件是否存在和注册表一致
- CI 环境不安装 protoc，只做校验
"""

import os
import re
import sys

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
PROTO_DIR = os.path.join(PROJECT_ROOT, "proto")
REGISTRY_FILE = os.path.join(PROJECT_ROOT, "docs", "api", "协议编号注册表.md")

# 各语言生成代码的预期路径（相对于 PROJECT_ROOT）
EXPECTED_GENERATED = {
    "go": [
        "proto/generated/go/go.mod",
        "proto/generated/go/go.sum",
        "proto/generated/go/base/health.pb.go",
        "proto/generated/go/base/hello.pb.go",
        "proto/generated/go/base/message.pb.go",
        "proto/generated/go/base/result.pb.go",
        "proto/generated/go/base/app_config.pb.go",
        "proto/generated/go/base/i18n.pb.go",
    ],
    "typescript": [
        "proto/generated/ts/base/health.ts",
        "proto/generated/ts/base/hello.ts",
        "proto/generated/ts/base/message.ts",
        "proto/generated/ts/base/result.ts",
    ],
    "python": [
        "proto/generated/python/base/health_pb2.py",
        "proto/generated/python/base/hello_pb2.py",
        "proto/generated/python/base/message_pb2.py",
        "proto/generated/python/base/result_pb2.py",
        "proto/generated/python/base/health_pb2_grpc.py",
        "proto/generated/python/base/hello_pb2_grpc.py",
        "proto/generated/python/base/message_pb2_grpc.py",
        "proto/generated/python/base/result_pb2_grpc.py",
    ],
}

# 禁止存在的路径（历史遗留问题，防止重复生成）
FORBIDDEN_PATHS = [
    "proto/generated/go/proto/base",
]


def extract_max_min_from_proto(proto_file):
    """从 proto 文件提取所有 message 的 max 和 min。"""
    results = []
    
    with open(proto_file, 'r', encoding='utf-8') as f:
        content = f.read()
    
    message_pattern = re.compile(
        r'message\s+(\w+)\s*\{[^}]*enum\s+Type\s*\{([^}]*)\}',
        re.DOTALL
    )
    
    for match in message_pattern.finditer(content):
        message_name = match.group(1)
        enum_content = match.group(2)
        
        max_match = re.search(r'max\s*=\s*(\d+)', enum_content)
        min_match = re.search(r'min\s*=\s*(\d+)', enum_content)
        
        if max_match and min_match:
            results.append({
                'message': message_name,
                'max': int(max_match.group(1)),
                'min': int(min_match.group(1)),
                'file': proto_file,
            })
    
    return results


def parse_registry():
    """解析协议编号注册表。"""
    registered = []
    
    if not os.path.exists(REGISTRY_FILE):
        return registered
    
    with open(REGISTRY_FILE, 'r', encoding='utf-8') as f:
        content = f.read()
    
    pattern = re.compile(r'\|\s*(\d+)\s*\|\s*(\d+)\s*\|')
    
    for match in pattern.finditer(content):
        registered.append((int(match.group(1)), int(match.group(2))))
    
    return registered


def check_go_package_consistency():
    """检查 .proto 文件中的 go_package 是否与目录结构一致。"""
    print(" [1] 检查 go_package 一致性...\n")
    
    errors = []
    
    for root, dirs, files in os.walk(PROTO_DIR):
        for file in files:
            if not file.endswith('.proto'):
                continue
            
            proto_file = os.path.join(root, file)
            rel_path = os.path.relpath(proto_file, PROJECT_ROOT)
            
            with open(proto_file, 'r', encoding='utf-8') as f:
                content = f.read()
            
            go_package_match = re.search(
                r'option\s+go_package\s*=\s*"([^"]+)"',
                content
            )
            
            if not go_package_match:
                continue
            
            go_package = go_package_match.group(1)
            
            # 期望的 go_package: github.com/jimiechen/mineplanet/protocols/generated/go/<proto_dir>
            # 同一目录下的所有 proto 文件共享同一个 go_package
            proto_rel = os.path.relpath(proto_file, PROTO_DIR)
            proto_dir = os.path.dirname(proto_rel)
            expected_suffix = proto_dir.replace(os.sep, '/')
            expected = f"github.com/jimiechen/mineplanet/protocols/generated/go/{expected_suffix}"
            
            if go_package != expected:
                errors.append({
                    'file': rel_path,
                    'actual': go_package,
                    'expected': expected,
                })
    
    if errors:
        print("❌ go_package 配置错误：")
        for err in errors:
            print(f"  文件: {err['file']}")
            print(f"    当前: {err['actual']}")
            print(f"    期望: {err['expected']}")
        print("")
        return False
    
    print("✅ 所有 .proto 文件的 go_package 配置正确\n")
    return True


def check_forbidden_paths():
    """检查是否存在禁止的生成代码路径（历史遗留问题）。"""
    print(" [2] 检查禁止路径...\n")
    
    found = []
    for path in FORBIDDEN_PATHS:
        full_path = os.path.join(PROJECT_ROOT, path)
        if os.path.exists(full_path):
            found.append(path)
    
    if found:
        print("❌ 发现禁止的生成代码路径（历史遗留）：")
        for path in found:
            print(f"   - {path}")
        print("\n  请删除上述目录，统一使用 proto/generated/go/base/ 路径\n")
        return False
    
    print("✅ 未发现禁止路径\n")
    return True


def check_generated_files():
    """检查各语言生成代码文件是否存在。"""
    print(" [3] 检查生成代码文件存在性...\n")
    
    all_exist = True
    stats = {"total": 0, "found": 0, "missing": 0}
    
    for lang, files in EXPECTED_GENERATED.items():
        print(f"  {lang} 生成代码：")
        for rel_path in files:
            full_path = os.path.join(PROJECT_ROOT, rel_path)
            stats["total"] += 1
            if os.path.exists(full_path):
                stats["found"] += 1
                print(f"    ✅ {rel_path}")
            else:
                stats["missing"] += 1
                all_exist = False
                print(f"    ❌ {rel_path} 不存在")
        print("")
    
    if not all_exist:
        print(f"❌ 缺少 {stats['missing']} 个生成代码文件")
        print("   请在本地执行 make proto 生成代码后提交到 Git\n")
        return False
    
    print(f"✅ 所有 {stats['total']} 个生成代码文件均存在\n")
    return True


def check_protocol_numbers():
    """检查协议编号唯一性和注册表一致性。"""
    print(" [3] 检查协议编号唯一性与注册表...\n")
    
    # 扫描 proto 文件
    all_protos = []
    for root, dirs, files in os.walk(PROTO_DIR):
        for file in files:
            if file.endswith('.proto'):
                all_protos.append(os.path.join(root, file))
    
    if not all_protos:
        print("⚠️  未找到 .proto 文件，跳过协议编号检查\n")
        return True
    
    print(f"  扫描到 {len(all_protos)} 个 proto 文件")
    
    # 提取所有 max + min
    all_entries = []
    for proto_file in all_protos:
        entries = extract_max_min_from_proto(proto_file)
        all_entries.extend(entries)
    
    if not all_entries:
        print("  ⚠️  未发现协议编号定义，跳过检查\n")
        return True
    
    print(f"  发现 {len(all_entries)} 个协议编号\n")
    
    # 检查重复
    seen = {}
    duplicates = []
    for entry in all_entries:
        key = (entry['max'], entry['min'])
        if key in seen:
            duplicates.append({'key': key, 'first': seen[key], 'second': entry})
        else:
            seen[key] = entry
    
    if duplicates:
        print("❌ 错误：发现重复的协议编号！")
        for dup in duplicates:
            print(f"  max={dup['key'][0]}, min={dup['key'][1]}:")
            print(f"    - {dup['first']['message']} in {dup['first']['file']}")
            print(f"    - {dup['second']['message']} in {dup['second']['file']}")
        return False
    
    print("✅ 协议编号无重复")
    
    # 检查注册表
    registered = parse_registry()
    unregistered = [e for e in all_entries if (e['max'], e['min']) not in registered]
    
    if unregistered:
        print("\n❌ 以下协议编号未登记到注册表：")
        for entry in unregistered:
            print(f"  - {entry['message']} (max={entry['max']}, min={entry['min']})")
        print(f"\n  请更新 {REGISTRY_FILE}\n")
        return False
    
    print("✅ 所有协议编号已登记\n")
    return True


def main():
    print("=" * 60)
    print(" Protobuf 生成代码校验（CI 用，无需 protoc）")
    print("=" * 60)
    print("")
    
    errors = []
    
    # 1. 检查 go_package 一致性
    if not check_go_package_consistency():
        errors.append("go_package 配置错误")
    
    # 2. 检查禁止路径
    if not check_forbidden_paths():
        errors.append("存在禁止的生成代码路径")
    
    # 3. 检查生成文件存在性
    if not check_generated_files():
        errors.append("生成代码文件缺失")
    
    # 4. 检查协议编号
    if not check_protocol_numbers():
        errors.append("协议编号检查失败")
    
    # 输出汇总
    print("=" * 60)
    print(" 校验结果汇总")
    print("=" * 60)
    
    if errors:
        print(f"\n❌ 校验未通过（{len(errors)} 项错误）：")
        for err in errors:
            print(f"   - {err}")
        print("\n  请在本地执行以下命令修复：")
        print("    1. make proto          # 生成 Protobuf 代码")
        print("    2. git add proto/generated/")
        print("    3. git commit -m 'chore(proto): 更新生成代码'")
        print("    4. git push\n")
        sys.exit(1)
    else:
        print("\n✅ Proto 校验全部通过")
        print("  - go_package 配置：正确")
        print("  - 生成代码路径：无历史遗留问题")
        print("  - 生成代码文件：存在且完整")
        print("  - 协议编号：无重复、已登记\n")
        sys.exit(0)


if __name__ == "__main__":
    main()
