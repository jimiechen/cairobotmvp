# AI

CaiRobot MVP AI 相关代码。

## 目录结构

```
ai/
├── README.md
└── service/          # AI 服务
    ├── README.md
    ├── app/         # 应用代码
    │   ├── api/     # API 层
    │   ├── core/    # 核心逻辑
    │   ├── prompts/ # 提示词模板
    │   ├── safety/  # 安全策略
    │   ├── inference/ # 推理封装
    │   └── schemas/ # 数据模型
    ├── tests/       # 测试
    │   ├── unit/    # 单元测试
    │   ├── safety/  # 安全测试
    │   └── contract/ # 契约测试
    └── fixtures/    # 测试数据
```

## 相关文档

- [PRD-04-AI服务系统.md](../docs/prd/PRD-04-AI服务系统.md)
- [ADR-0004-AI服务使用Python.md](../docs/adr/ADR-0004-AI服务使用Python.md)
- [proto/ai_service/](../proto/ai_service/)
