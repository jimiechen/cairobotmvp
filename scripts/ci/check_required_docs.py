#!/usr/bin/env python3
"""
检查关键文档是否存在。

用于 GitHub Actions CI 的 docs-check job。
"""

import os
import sys

REQUIRED_DOCS = [
    "AGENTS.md",
    ".trae/rules/tdd.md",
    ".trae/rules/testing.md",
    ".trae/rules/review.md",
    ".trae/rules/docs.md",
    ".trae/rules/reporting.md",
    "docs/prd/README.md",
    "docs/api/协议编号注册表.md",
    "docs/api/protobuf规范.md",
    "docs/api/openapi-protobuf映射规范.md",
    "docs/wiki/LLM-WIKI.md",
    ".github/workflows/ci.yml",
    ".github/pull_request_template.md",
]

def main():
    missing = []
    
    for doc in REQUIRED_DOCS:
        if not os.path.exists(doc):
            missing.append(doc)
            print(f"[缺失] {doc}")
        else:
            print(f"[存在] {doc}")
    
    if missing:
        print(f"\n错误：缺少 {len(missing)} 个关键文档")
        for doc in missing:
            print(f"  - {doc}")
        sys.exit(1)
    
    print(f"\n成功：所有 {len(REQUIRED_DOCS)} 个关键文档都存在")
    sys.exit(0)

if __name__ == "__main__":
    main()
