# -*- coding: utf-8 -*-
"""pytest 路径引导：确保从 python/ai 目录执行 make unit 时可导入 fa_orchestrator 包。

pytest 从 python/ai/ 运行（见 python/Makefile unit target），
本文件把 fa_orchestrator/ 目录注入 sys.path，使 `import fa_orchestrator` 指向包目录。
"""
import sys
from pathlib import Path

_PKG_DIR = Path(__file__).resolve().parent
if str(_PKG_DIR) not in sys.path:
    sys.path.insert(0, str(_PKG_DIR))
