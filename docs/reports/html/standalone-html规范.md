# Standalone HTML 规范

## 1. 目的

Standalone HTML 用于提供可离线查看的完整测试报告，便于归档和分享。

## 2. 文件要求

### 2.1 单文件 HTML

- 必须是单个 HTML 文件
- CSS 必须内嵌在 &lt;style&gt; 标签中
- JavaScript 必须内嵌在 &lt;script&gt; 标签中，如无必要可不使用
- 图片优先使用相对路径，也可使用 base64

### 2.2 兼容性

- 支持现代浏览器
- 支持离线打开
- 不依赖外部 CDN 资源

## 3. 报告结构

Standalone HTML 必须包含以下部分：

| 章节 | 说明 |
|---|---|
| 1. 基本信息 | 测试报告 ID、日期、相关 PRD/Issue |
| 2. 测试环境 | OS、依赖、设备等 |
| 3. 测试对象 | 说明本次测试验证什么 |
| 4. 测试用例执行结果 | 用例列表及结果 |
| 5. 详细测试步骤 | 步骤、输入、期望、实际、证据 |
| 6. 截图证据 | 截图展示 |
| 7. 视频证据 | 视频链接或说明 |
| 8. Bug 与风险 | Bug 列表和风险说明 |
| 9. 结论 | 总体结论 |

## 4. 存放位置

Standalone HTML 存放于：

```text
docs/reports/html/
```

文件命名建议：

```text
REPORT-XXX_YYYY-MM-DD.html
```

示例：

```text
REPORT-001_2026-05-17.html
```

## 5. 生成时机

以下情况必须生成 Standalone HTML：

- 验收测试
- 重大功能测试
- Bug 验证
- 报告需要分享给外部人员

## 6. 索引更新

新增 Standalone HTML 后，请同步更新：

- [docs/wiki/测试索引.md](../../wiki/测试索引.md)
