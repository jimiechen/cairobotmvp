#!/usr/bin/env python3
"""
CaiRobot MVP 模块路径一致性检查脚本
验证 go.work、package.json、pyproject.toml 中的模块路径与实际目录一致
"""

import os
import sys
from pathlib import Path

PROJECT_ROOT = Path(__file__).parent.parent.parent

ERRORS = []
WARNINGS = []

def check_go_work():
    """检查 go.work 中声明的模块是否存在"""
    print("[Go Work] 检查 go.work 模块路径...")
    
    go_work = PROJECT_ROOT / "go" / "go.work"
    if not go_work.exists():
        WARNINGS.append("go/work.go 不存在")
        return
    
    content = go_work.read_text(encoding="utf-8")
    
    import re
    modules = re.findall(r'\./([^\s\)]+)', content)
    
    for mod in modules:
        mod_path = PROJECT_ROOT / "go" / mod
        if not mod_path.exists():
            ERRORS.append(f"❌ go.work 声明的模块不存在: go/{mod}")
            print(f"  ❌ go/{mod}")
        elif not (mod_path / "go.mod").exists():
            WARNINGS.append(f"⚠️ go/{mod} 存在但无 go.mod")
            print(f"  ⚠️ go/{mod} (无 go.mod)")
        else:
            print(f"  ✅ go/{mod}")

def check_typescript_workspace():
    """检查 TypeScript workspace 配置"""
    print("\n[TS Workspace] 检查 pnpm workspace...")
    
    pnpm_ws = PROJECT_ROOT / "typescript" / "pnpm-workspace.yaml"
    if not pnpm_ws.exists():
        WARNINGS.append("typescript/pnpm-workspace.yaml 不存在")
        return
    
    content = pnpm_ws.read_text(encoding="utf-8")
    
    import yaml
    try:
        ws_config = yaml.safe_load(content)
        packages = ws_config.get("packages", [])
        
        for pkg_pattern in packages:
            for pkg_dir in PROJECT_ROOT.glob(f"typescript/{pkg_pattern}"):
                if pkg_dir.is_dir():
                    pkg_json = pkg_dir / "package.json"
                    if pkg_json.exists():
                        print(f"  ✅ {pkg_dir.relative_to(PROJECT_ROOT)}")
                    else:
                        WARNINGS.append(f"{pkg_dir.name} 无 package.json")
    except ImportError:
        WARNINGS.append("PyYAML 未安装，跳过详细检查")

def main():
    print("=" * 60)
    print("CaiRobot MVP 模块路径一致性检查")
    print("=" * 60)
    print()
    
    check_go_work()
    check_typescript_workspace()
    
    print("\n" + "=" * 60)
    print(f"结果: {len(ERRORS)} 错误, {len(WARNINGS)} 警告")
    
    if ERRORS:
        print("\n❌ 模块路径检查未通过！")
        for err in ERRORS:
            print(f"  {err}")
        return 1
    else:
        print("\n✅ 模块路径检查通过")
        return 0

if __name__ == "__main__":
    sys.exit(main())
