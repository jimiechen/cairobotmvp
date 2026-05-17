#!/usr/bin/env python3
"""
检查测试报告、日报、蒸馏和 LLM Wiki 是否存在。

用于 GitHub Actions CI 的 report-check job。
"""

import os
import sys

REQUIRED_DIRS = [
    "docs/reports/testing",
    "docs/reports/daily",
    "docs/reports/distilled",
]

REQUIRED_FILES = [
    "docs/wiki/LLM-WIKI.md",
]

def main():
    print("=== 报告检查 ===\n")
    
    warnings = []
    errors = []
    
    # 检查目录
    for dir_path in REQUIRED_DIRS:
        if not os.path.exists(dir_path):
            print(f"[缺失] 目录：{dir_path}")
            warnings.append(f"目录不存在：{dir_path}")
            # 创建目录
            os.makedirs(dir_path, exist_ok=True)
            print(f"[创建] 目录：{dir_path}")
        else:
            print(f"[存在] 目录：{dir_path}")
    
    # 检查文件
    for file_path in REQUIRED_FILES:
        if not os.path.exists(file_path):
            print(f"[缺失] 文件：{file_path}")
            errors.append(f"文件不存在：{file_path}")
        else:
            print(f"[存在] 文件：{file_path}")
    
    # 检查是否有实际报告
    testing_reports = []
    if os.path.exists("docs/reports/testing"):
        for root, dirs, files in os.walk("docs/reports/testing"):
            for f in files:
                if f.endswith('.md') or f.endswith('.html'):
                    testing_reports.append(os.path.join(root, f))
    
    daily_reports = []
    if os.path.exists("docs/reports/daily"):
        for root, dirs, files in os.walk("docs/reports/daily"):
            for f in files:
                if f.endswith('.md'):
                    daily_reports.append(os.path.join(root, f))
    
    distilled_reports = []
    if os.path.exists("docs/reports/distilled"):
        for root, dirs, files in os.walk("docs/reports/distilled"):
            for f in files:
                if f.endswith('.md'):
                    distilled_reports.append(os.path.join(root, f))
    
    print(f"\n测试报告数量：{len(testing_reports)}")
    print(f"日报数量：{len(daily_reports)}")
    print(f"蒸馏报告数量：{len(distilled_reports)}")
    
    if testing_reports:
        print("\n测试报告：")
        for r in testing_reports:
            print(f"  - {r}")
    
    if daily_reports:
        print("\n日报：")
        for r in daily_reports:
            print(f"  - {r}")
    
    if distilled_reports:
        print("\n蒸馏报告：")
        for r in distilled_reports:
            print(f"  - {r}")
    
    # 输出结果
    if errors:
        print(f"\n错误：{len(errors)} 个文件缺失")
        for e in errors:
            print(f"  - {e}")
        sys.exit(1)
    
    if warnings:
        print(f"\n警告：{len(warnings)} 个目录缺失（已自动创建）")
        for w in warnings:
            print(f"  - {w}")
    
    if not testing_reports:
        print("\n警告：当前没有测试报告")
    
    if not daily_reports:
        print("警告：当前没有日报")
    
    if not distilled_reports:
        print("警告：当前没有蒸馏报告")
    
    print("\n成功：报告检查通过")
    sys.exit(0)

if __name__ == "__main__":
    main()
