# Admin MVP E2E 截图/录屏测试方案（v2）

> **状态：待审核** | 提交人：Trae AI | 日期：2026-05-27
> **关联 PRD**：[PRD-10](../prd/PRD-10-Admin管理后台MVP.md) M6' 端到端验收 + M7' 文档

## 1. 背景

当前 `admin-mvp-test-report.md` 已记录 56 个 Go 后端测试用例全部 PASS 的文字证据，但缺少 **可视化测试痕迹**（截图、录屏）。按项目规则 [测试报告模板.md](./测试报告模板.md) §8~9 要求，每个用例应附带截图/视频证据文件路径。

## 2. 目标

为 M0'~M5' 交付的 admin 插件层（config_admin 12 用例 + i18n_admin 20 用例 = **32 个 HTTP handler 测试**）补充 E2E 截图/录屏证据，使测试报告达到「可追溯、可验证」标准。

**核心要求：**
1. admin-web **每个可交互 HTML 元素**（按钮、输入框、表格行、弹窗、表单项）必须绑定 `data-id` 属性
2. 测试用例步骤必须使用 **中文描述**
3. 截图/录屏文件名含 data-id 编号，与用例一一对应

## 3. 技术选型

| 维度 | 选型 | 理由 |
|------|------|------|
| 框架 | **Playwright (Node.js)** | 项目已有 Node.js 20 环境，Playwright 原生支持截图+录屏+video trace |
| 运行位置 | `tests/e2e/`（项目根目录） | 遵循 [testing.md](../../.trae/rules/testing.md) §3 E2E 测试目录规范 |
| 截图格式 | PNG（带 data-id 元数据水印） | 可读性强，支持嵌入 Markdown |
| 录屏格式 | WebM（关键用例） | 文件小，浏览器原生支持 |
| 证据存储 | `docs/reports/testing/evidence/` | 遵循 [docs.md](../../.trae/rules/docs.md) 报告目录规范 |

## 4. HTML 元素级 data-id 绑定规范（核心）

### 4.1 命名规则

```
data-id="{页面缩写}-{组件类型}-{操作或标识}"
```

**命名组成：**

| 组成部分 | 说明 | 可选值 |
|---------|------|--------|
| 页面缩写 | 所属功能模块 | `ca`=config-admin, `ia`=i18n-admin |
| 组件类型 | UI 元素类别 | `btn`=按钮, `input`=输入框, `select`=下拉, `table`=表格, `dialog`=弹窗, `form`=表单, `row`=行, `col`=列, `tab`=标签, `alert`=提示, `upload`=上传, `download`=下载 |
| 操作标识 | 具体操作含义 | 中文拼音或英文简写，见下表 |

### 4.2 各组件 data-id 清单

#### 4.2.1 config/schema-list.vue 元素绑定

| HTML 元素 | data-id 值 | 说明 |
|----------|------------|------|
| 搜索按钮 | `ca-btn-sousuo` | Schema 列表搜索 |
| 重置按钮 | `ca-btn-chongzhi` | 搜索条件重置 |
| 新增 Schema 按钮 | `ca-btn-xinzeng-schema` | 打开新增弹窗 |
| 删除按钮（工具栏） | `ca-btn-shanchu-toolbar` | 工具栏批量删除 |
| 表格行编辑按钮 | `ca-btn-bianji-row` | 行内编辑 |
| 表格行删除按钮 | `ca-btn-shanchu-row` | 行内删除 |
| 弹窗确定按钮 | `ca-btn-queding-dialog` | 新增/编辑弹窗提交 |
| 弹窗取消按钮 | `ca-btn-quxiao-dialog` | 关闭弹窗 |
| 模块 Key 输入框 | `ca-input-mokuai-key` | 搜索条件 |
| 弹窗模块 Key 输入框 | `ca-input-dialog-mokuai-key` | 新增/编辑表单 |
| 弹窗字段 Key 输入框 | `ca-input-dialog-ziduan-key` | 新增/编辑表单 |
| 弹窗字段类型下拉 | `ca-select-dialog-ziduan-leixing` | string/int/float/bool/json |
| 弹窗必填开关 | `ca-switch-dialog-bitian` | required 开关 |
| 弹窗默认值输入 | `ca-input-dialog-morenzhi` | 根据类型动态渲染 |
| 弹窗校验规则文本域 | `ca-textarea-dialog-jiaoyan-guize` | validator JSON |
| 弹窗描述文本域 | `ca-textarea-dialog-miaoshu` | description |
| Schema 列表表格 | `ca-table-schema-list` | 整个表格容器 |
| 分页组件 | `ca-pagination-schema` | 列表分页 |

#### 4.2.2 config/value-publish.vue 元素绑定

| HTML 元素 | data-id 值 | 说明 |
|----------|------------|------|
| 模块 Key 下拉选择 | `ca-select-value-mokuai` | 选择配置模块 |
| 环境选择下拉 | `ca-select-value-huanjing` | dev/test/staging/prod |
| 发布配置按钮 | `ca-btn-fabu-peizhi` | 提交发布 |
| 重置按钮 | `ca-btn-chongzhi-form` | 重置发布表单 |
| 动态字段输入框（通用） | `ca-input-dongtai-{fieldKey}` | 按 schema 渲染的每个字段 |
| 字段错误提示 | `ca-error-{fieldKey}` | 10400 校验失败红色提示 |
| 10400 错误弹窗 | `ca-dialog-10400-cuowu` | 校验失败详情弹窗 |
| 错误弹窗关闭按钮 | `ca-btn-guanbi-10400` | 关闭 10400 弹窗 |
| 版本历史表格 | `ca-table-version-history` | 发布历史列表 |
| 刷新 Schema 按钮 | `ca-btn-shuaxin-schema` | 重新加载 schema |

#### 4.2.3 i18n/string-list.vue 元素绑定

| HTML 元素 | data-id 值 | 说明 |
|----------|------------|------|
| 语言包 ID 输入 | `ia-input-yuyanbao-id` | pack ID 数字输入 |
| 查询按钮 | `ia-btn-chaxun-string` | 字符串查询 |
| 重置按钮 | `ia-btn-chongzhi-string` | 条件重置 |
| 新增字符串按钮 | `ia-btn-xinzeng-string` | 打开新增弹窗 |
| 删除按钮（工具栏） | `ia-btn-shanchu-string-toolbar` | 批量删除 |
| 行内编辑按钮 | `ia-btn-bianji-string` | 编辑字符串 |
| 行内删除按钮 | `ia-btn-shanchu-string` | 删除字符串 |
| 弹窗 String Key 输入 | `ia-input-dialog-string-key` | 字符串唯一标识 |
| 弹窗字符串值文本域 | `ia-textarea-dialog-string-value` | 翻译内容 |
| 弹窗模板类型下拉 | `ia-select-dialog-moban-leixing` | plain/named/icu |
| 弹窗分组输入 | `ia-input-dialog-fenzu` | groupName |
| 弹窗参数 Schema 文本域 | `a-textarea-dialog-canshu-schema` | named 类型参数定义 |
| 弹窗预览示例文本域 | `ia-textarea-dialog-yulan-shili` | 模板预览 |
| 弹窗确定按钮 | `ia-btn-queding-string-dialog` | 提交字符串 |
| 弹窗取消按钮 | `ia-btn-quxiao-string-dialog` | 关闭弹窗 |
| 字符串列表表格 | `ia-table-string-list` | 整个表格 |
| String Key popover | `ia-popover-string-preview` | 长文本预览 |

#### 4.2.4 i18n/pack-manage.vue 元素绑定

| HTML 元素 | data-id 值 | 说明 |
|----------|------------|------|
| 语言包 ID 输入 | `ia-input-pack-id` | pack ID |
| 语言码下拉 | `ia-select-pack-yuyanma` | zh-CN/en-US 等 |
| 目标环境下拉 | `ia-select-pack-huanjing` | dev/test/staging/prod |
| 发布语言包按钮 | `ia-btn-fabu-yuyanbao` | 提交发布 |
| 回滚版本号输入 | `ia-input-huinban-banhao` | 目标版本号 |
| 回滚确认按钮 | `ia-btn-queren-huinban` | 执行回滚 |
| 发布结果卡片 | `ia-card-fabu-jieguo` | PackID/LangCode/Version 展示 |

#### 4.2.5 i18n/import-export.vue 元素绑定

| HTML 元素 | data-id 值 | 说明 |
|----------|------------|------|
| 导入目标 Pack ID 输入 | `ia-import-input-pack-id` | CSV 导入目标 |
| 文件上传区域 | `ia-upload-csv-wenjian` | el-upload drag 区域 |
| 开始导入按钮 | `ia-btn-kaishi-daoru` | 触发导入 |
| 导入重置按钮 | `ia-btn-chongzhi-daoru` | 重置导入区 |
| 导入结果展示 | `ia-result-daoru-jieguo` | el-result 成功/失败 |
| 导入错误详情表格 | `ia-table-daoru-cuowu` | 失败行明细 |
| 导出源 Pack ID 输入 | `ia-export-input-pack-id` | CSV 导出源 |
| 导出 CSV 按钮 | `ia-btn-daochu-csv` | 触发导出下载 |
| 导出说明列表 | `ul-daochu-shuoming` | 导出格式说明 |

### 4.3 Vue 模板绑定示例

```vue
<!-- ✅ 正确：每个交互元素都有 data-id -->
<el-button
  v-permisaction="['config:schema:add']"
  type="primary"
  icon="el-icon-plus"
  size="mini"
  data-id="ca-btn-xinzeng-schema"
  @click="handleAdd"
>新增 Schema</el-button>

<el-input
  v-model="form.moduleKey"
  placeholder="请输入模块标识"
  data-id="ca-input-dialog-mokuai-key"
/>

<!-- ❌ 错误：缺少 data-id -->
<el-button type="primary" @click="handleAdd">新增</el-button>
```

### 4.4 grep 自检命令

```bash
# 检查所有 Vue 文件的 data-id 覆盖率
# 应覆盖所有 el-button / el-input / el-select / el-upload 等交互元素
grep -rn 'data-id=' typescript/admin-web/src/views/config/*.vue \
            typescript/admin-web/src/views/i18n/*.vue \
         | wc -l
# 期望：≥ 60 个 data-id 绑定点
```

## 5. 截图/录屏文件命名规范

### 5.1 文件名格式

```
{模块缩写}-{用例序号}-{中文场景描述}-{时间戳}.{png|webm}
```

**示例：**

| 文件名 | 对应用例 | 场景说明 |
|--------|---------|---------|
| `ca-01-Schema列表正常渲染-20260527.png` | ca-01 | Schema 列表页加载后正常显示数据 |
| `ca-09-10400校验错误字段展示-20260527.png` | ca-09 | 配置值发布校验失败时 10400 弹窗截图 |
| `ia-14-CSV导入成功结果展示-20260527.png` | ia-14 | CSV 上传导入后的成功结果页 |
| `ia-14-CSV导入成功操作录屏-20260527.webm` | ia-14 | CSV 导入完整操作流程录屏 |

### 5.2 data-id 水印格式

每张截图右下角叠加半透明水印：

```
[data-id: ca-01] [2026-05-27 14:30:22] [env: test]
```

## 6. 目录结构

```
tests/e2e/
├── playwright.config.ts          # Playwright 配置
├── package.json                  # 依赖隔离（不污染 admin-web）
├── utils/
│   ├── evidence.ts               # 截图/录屏工具函数（带 data-id 水印）
│   ├── test-helpers.ts           # HTTP 请求 mock / API 启动辅助
│   └── data-id-check.ts          # data-id 覆盖率自检工具
├── fixtures/
│   └── test-data.json            # 固定测试数据（Schema/Value/String/Pack）
└── specs/
    ├── config-admin.spec.ts      # config_admin 12 个用例（中文步骤）
    └── i18n-admin.spec.ts        # i18n_admin 20 个用例（中文步骤）

docs/reports/testing/
├── evidence/                     # 截图/录屏输出目录（gitignore）
│   ├── ca-01-Schema列表正常渲染-*.png
│   ├── ca-02-空ModuleKey查询结果-*.png
│   ├── ...
│   └── ia-19-无效PackID导出报错-*.png
└── admin-mvp-test-report.md      # 更新后含图片引用 + 中文步骤
```

## 7. 核心工具函数设计

### 7.1 evidence.ts — 基于 data-id 的截图封装

```typescript
import { Page } from '@playwright/test'
import path from 'path'

export interface EvidenceOptions {
  /** 对应测试报告中的 data-id 编号，如 "ca-01" */
  dataId: string
  /** 中文场景描述 */
  description: string
  /** 是否全页截图 */
  fullPage?: boolean
  /** 是否录制视频 */
  recordVideo?: boolean
}

const EVIDENCE_DIR = '../../docs/reports/testing/evidence'

/**
 * 带data-id水印的截图
 * 文件名格式：{dataId}-{description}-{timestamp}.png
 */
export async function takeEvidence(
  page: Page,
  options: EvidenceOptions
): Promise<string> {
  const { dataId, description, fullPage = true } = options
  const ts = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19)
  const safeDesc = description.replace(/[/\\?%*:|"<>]/g, '-')
  const filename = `${dataId}-${safeDesc}-${ts}.png`
  const filepath = path.join(__dirname, EVIDENCE_DIR, filename)

  await page.screenshot({
    path: filepath,
    fullPage,
  })

  return filepath
}

/**
 * 通过 data-id 定位元素并点击，同时截图记录操作前状态
 */
export async function clickAndScreenshot(
  page: Page,
  dataId: string,
  options: EvidenceOptions
): Promise<void> {
  await takeEvidence(page, { ...options, description: options.description + '-操作前' })
  await page.locator(`[data-id="${dataId}"]`).click()
  await page.waitForTimeout(500) // 等待动画完成
  await takeEvidence(page, { ...options, description: options.description + '-操作后' })
}
```

### 7.2 data-id-check.ts — 覆盖率自检

```typescript
/**
 * 检查指定 Vue 文件中是否所有交互元素都绑定了 data-id
 * 未绑定的元素输出 WARNING
 */
export function checkDataIdCoverage(vueContent: string, filePath: string): void {
  const buttons = (vueContent.match(/<el-button/g) || []).length
  const inputs = (vueContent.match(/<el-input/g) || []).length
  const selects = (vueContent.match(/<el-select/g) || []).length
  const dataIds = (vueContent.match(/data-id=/g) || []).length
  const expected = buttons + inputs + selects
  const coverage = expected > 0 ? (dataIds / expected * 100).toFixed(1) : '100.0'

  console.log(`[${filePath}] 交互元素=${expected}, data-id=${dataIds}, 覆盖率=${coverage}%`)
  if (Number(coverage) < 90) {
    console.warn(`⚠️ ${filePath} data-id 覆盖率低于 90%，请补充绑定`)
  }
}
```

## 8. 测试策略

### 8.1 Mock Server 模式

由于 Go 后端和 Vue 前端是独立进程，E2E 测试采用 **Mock Server 模式**：

```
Playwright (Chromium)
  ├─ 访问 localhost:8080（前端 dev server）
  ├─ 通过 page.route() 拦截 XHR → 返回预设 Mock JSON
  ├─ 通过 page.locator('[data-id="xxx"]') 定位元素
  ├─ 模拟点击/输入/上传操作
  ├─ 验证 DOM 渲染正确性（断言文本/可见性）
  └─ 截图/录屏存入 evidence/
```

**原因：**
- Go 后端需要数据库 + Redis 才能运行完整链路，E2E 环境搭建成本高
- Handler 层逻辑已在单元测试中覆盖（32/32 PASS）
- E2E 重点验证 **前端渲染 + 10400 错误展示 + 用户交互流程** 的视觉正确性

### 8.2 用例映射表（32 个 → 中文步骤 + data-id）

#### config_admin（12 个）

| data-id | 用例名称 | 中文测试步骤 | 涉及 data-id 元素 | 是否录屏 |
|---------|---------|-------------|------------------|---------|
| ca-01 | Schema 列表正常渲染 | 1.打开 Schema 列表页<br>2.确认表格正常显示数据<br>3.截图记录初始状态 | `ca-table-schema-list`, `ca-pagination-schema` | 否 |
| ca-02 | 空 ModuleKey 查询 | 1.清空模块 Key 输入框<br>2.点击搜索按钮<br>3.确认表格为空或显示无数据 | `ca-input-mokuai-key`, `ca-btn-sousuo`, `ca-table-schema-list` | 否 |
| ca-03 | 正常创建 Schema | 1.点击新增 Schema 按钮<br>2.在弹窗填写 moduleKey=app.server, fieldKey=port<br>3.选择字段类型=int<br>4.点击确定按钮<br>5.确认成功提示并截图 | `ca-btn-xinzeng-schema`, `ca-input-dialog-mokuai-key`, `ca-input-dialog-ziduan-key`, `ca-select-dialog-ziduan-leixing`, `ca-btn-queding-dialog` | 是 |
| ca-04 | 创建时 Body 为空 | 1.点击新增 Schema 按钮<br>2.不填写任何内容直接点确定<br>3.确认必填项校验红色提示出现 | `ca-btn-xinzeng-schema`, `ca-btn-queding-dialog` | 否 |
| ca-05 | 正常更新 Schema | 1.点击某行的编辑按钮<br>2.确认弹窗回显当前数据<br>3.修改字段值<br>4.点击确定<br>5.确认更新成功 | `ca-btn-bianji-row`, `ca-input-dialog-ziduan-key`, `ca-btn-queding-dialog` | 否 |
| ca-06 | 无效 ID 删除 | 1.在 URL 参数中传入 id=abc<br>2.触发删除请求<br>3.确认返回参数错误提示 | `ca-btn-shanchu-row` | 否 |
| ca-07 | 正常删除 Schema | 1.选中一行数据<br>2.点击删除按钮<br>3.在确认对话框中点击确定<br>4.确认列表已刷新 | `ca-btn-shanchu-row`, `ca-btn-shanchu-toolbar` | 是 |
| ca-08 | 正常发布配置值 | 1.打开配置值发布页<br>2.选择模块 Key 和环境<br>3.确认动态字段按 Schema 渲染<br>4.填写各字段值<br>5.点击发布按钮<br>6.确认发布成功并显示版本号 | `ca-select-value-mokuai`, `ca-select-value-huanjing`, `ca-input-dongtai-*`, `ca-btn-fabu-peizhi` | 是 |
| ca-09 | 10400 校验错误展示 | 1.打开配置值发布页<br>2.填写超出范围的数值（如 port=99999）<br>3.点击发布<br>4.确认弹出 10400 错误详情弹窗<br>5.确认字段级错误映射到对应输入框下方 | `ca-input-dongtai-port`, `ca-btn-fabu-peizhi`, `ca-dialog-10400-cuowu`, `ca-error-port` | **是（重点）** |
| ca-10 | 空 Fields 发布 | 1.不选择任何模块直接点发布<br>2.确认返回空 fields 错误 | `ca-btn-fabu-peizhi` | 否 |
| ca-11 | 缺 ModuleKey 查版本 | 1.不选模块直接查版本历史<br>2.确认返回缺少参数错误 | （版本历史区域） | 否 |
| ca-12 | 正常查询版本历史 | 1.选择有效模块和环境<br>2.确认版本历史表格正常渲染 | `ca-select-value-mokuai`, `ca-table-version-history` | 否 |

#### i18n_admin（20 个）

| data-id | 用例名称 | 中文测试步骤 | 涉及 data-id 元素 | 是否录屏 |
|---------|---------|-------------|------------------|---------|
| ia-01 | 正常创建字符串 | 1.打开字符串管理页<br>2.填入 pack_id=1<br>3.点击新增字符串按钮<br>4.填写 string_key=greeting.hello, string_value=你好<br>5.选择模板类型=plain<br>6.点击确定<br>7.确认创建成功 | `ia-input-yuyanbao-id`, `ia-btn-xinzeng-string`, `ia-input-dialog-string-key`, `ia-textarea-dialog-string-value`, `ia-select-dialog-moban-leixing`, `ia-btn-queding-string-dialog` | 是 |
| ia-02 | 缺少必填字段创建 | 1.点击新增字符串按钮<br>2.不填任何内容直接点确定<br>3.确认必填项校验红色提示 | `ia-btn-xinzeng-string`, `ia-btn-queding-string-dialog` | 否 |
| ia-03 | 10400 模板校验错误 | 1.点击新增字符串按钮<br>2.填写 string_value=Hello {name}<br>3.选择模板类型=named（未提供 params_schema）<br>4.点击确定<br>5.确认返回 10400 模板错误提示 | `ia-textarea-dialog-string-value`, `ia-select-dialog-moban-leixing`, `ia-btn-queding-string-dialog` | **是（重点）** |
| ia-04 | 正常更新字符串 | 1.点击某行的编辑按钮<br>2.确认弹窗回显当前数据<br>3.修改字符串值<br>4.点击确定<br>5.确认更新成功 | `ia-btn-bianji-string`, `ia-textarea-dialog-string-value`, `ia-btn-queding-string-dialog` | 否 |
| ia-05 | 空 Body 更新 | 1.模拟发送空的 PUT 请求体<br>2.确认返回 400 错误 | — | 否 |
| ia-06 | 无效 ID 删除字符串 | 1.传入 id=abc 触发删除<br>2.确认返回参数错误提示 | `ia-btn-shanchu-string` | 否 |
| ia-07 | 正常删除字符串 | 1.选中一行字符串<br>2.点击删除按钮<br>3.确认对话框点确定<br>4.确认列表刷新 | `ia-btn-shanchu-string`, `ia-btn-shanchu-string-toolbar` | 否 |
| ia-08 | 正常查询字符串列表 | 1.填入 pack_id=1<br>2.点击查询<br>3.确认表格渲染字符串数据 | `ia-input-yuyanbao-id`, `ia-btn-chaxun-string`, `ia-table-string-list` | 否 |
| ia-09 | 缺 PackId 查询 | 1.不填 pack_id 直接查<br>2.确认返回缺少参数错误 | `ia-btn-chaxun-string` | 否 |
| ia-10 | 正常发布语言包 | 1.打开语言包管理页<br>2.填入 pack_id=1, 选择 lang_code=zh-CN, env=dev<br>3.点击发布按钮<br>4.确认返回版本号信息 | `ia-input-pack-id`, `ia-select-pack-yuyanma`, `ia-select-pack-huanjing`, `ia-btn-fabu-yuyanbao`, `ia-card-fabu-jieguo` | 是 |
| ia-11 | 空 Body 发布语言包 | 1.不填任何内容点发布<br>2.确认返回 400 错误 | `ia-btn-fabu-yuyanbao` | 否 |
| ia-12 | 正常回滚语言包 | 1.填入 pack_id=1<br>2.输入回滚版本号=3<br>3.点击回滚确认<br>4.确认回滚成功 | `ia-input-pack-id`, `ia-input-huinban-banhao`, `ia-btn-queren-huinban` | 否 |
| ia-13 | 空 Body 回滚 | 1.不填版本号点回滚<br>2.确认返回 400 错误 | `ia-btn-queren-huinban` | 否 |
| ia-14 | CSV 正常导入 | 1.填入目标 pack_id=1<br>2.将测试 CSV 文件拖入上传区域<br>3.点击开始导入<br>4.确认显示导入成功结果（total/success/fail） | `ia-import-input-pack-id`, `ia-upload-csv-wenjian`, `ia-btn-kaishi-daoru`, `ia-result-daoru-jieguo` | **是（重点）** |
| ia-15 | CSV 部分失败返回 10400 | 1.上传包含错误行的 CSV<br>2.点击开始导入<br>3.确认显示 code=10400 部分失败结果<br>4.确认错误明细表格展示行号和原因 | `ia-upload-csv-wenjian`, `ia-btn-kaishi-daoru`, `ia-result-daoru-jieguo`, `ia-table-daoru-cuowu` | **是（重点）** |
| ia-16 | 缺文件导入 | 1.不选择文件直接点开始导入<br>2.确认返回「请先选择文件」提示 | `ia-btn-kaishi-daoru` | 否 |
| ia-17 | 无效 PackId 导入 | 1.填入 pack_id=abc<br>2.点开始导入<br>3.确认返回参数错误提示 | `ia-import-input-pack-id`, `ia-btn-kaishi-daoru` | 否 |
| ia-18 | CSV 正常导出 | 1.填入有效的 pack_id=1<br>2.点击导出 CSV 按钮<br>3.确认浏览器触发下载（文件名 strings_1.csv） | `ia-export-input-pack-id`, `ia-btn-daochu-csv` | 是 |
| ia-19 | 无效 PackId 导出 | 1.填入 pack_id=abc<br>2.点导出<br>3.确认返回参数错误 | `ia-export-input-pack-id`, `ia-btn-daochu-csv` | 否 |

**必须录屏的重点用例（5 个）：**
- `ca-09`：10400 配置值校验错误展示流程
- `ia-03`：10400 模板校验错误展示流程
- `ia-14`：CSV 拖拽上传导入完整交互
- `ia-15`：CSV 部分失败 10400 结果展示
- `ia-08`：正常发布语言包（含结果卡片展示）

## 9. 实施步骤

| 步骤 | 内容 | 产出 | 预估工作量 |
|------|------|------|-----------|
| T-1 | 为 5 个 Vue 组件补充 data-id 属性（~60 个绑定点） | 5 个 .vue 文件修改 | 中 |
| T-2 | 安装 Playwright + 初始化配置 + 工具函数 | `playwright.config.ts`, `utils/*.ts`, `package.json` | 小 |
| T-3 | 编写 config-admin.spec.ts（12 个中文用例） | 12 张截图 + 3 个录屏 | 中 |
| T-4 | 编写 i18n-admin.spec.ts（20 个中文用例） | 19 张截图 + 2 个录屏 | 中 |
| T-5 | 运行 data-id 覆盖率自检 + 全量 E2E 测试 | 自检报告 + evidence 目录 | 小 |
| T-6 | 更新测试报告（嵌入截图引用 + 中文步骤） | `admin-mvp-test-report.md` v2 | 小 |

**合计产出：31 张截图 + 5 个录屏 + 1 份更新后的测试报告 + 5 个 Vue 组件 data-id 增强**

## 10. 依赖清单

```json
{
  "devDependencies": {
    "@playwright/test": "^1.48.0",
    "@types/node": "^20.0.0"
  }
}
```

不需要安装浏览器（Playwright 首次运行自动下载 Chromium）。

## 11. 与现有测试的关系

| 现有测试 | 职责 | 不变 |
|---------|------|------|
| Go 单元测试（56 个） | 验证 handler 逻辑正确性 | 保持不变 |
| pnpm build + eslint | 验证前端编译正确性 | 保持不变 |
| grep 自检 | 验证边界铁律 | 保持不变 |
| **新增：Playwright E2E** | **验证 UI 渲染 + 用户交互 + 10400 错误展示 + data-id 可追溯性** | **新增** |

三者互补，不重复。

## 12. 风险评估

| 风险 | 等级 | 影响 | 应对 |
|------|------|------|------|
| 前端 dev server 需要启动 | P2 | 需要 `pnpm run dev` 在后台运行 | 测试脚本自动检测端口，未启动则跳过并 WARN |
| Mock 数据与真实 API 结构不一致 | P2 | 截图可能与真实环境有差异 | Mock 数据从 DTO 定义同步生成 |
| Chromium 下载慢 | P3 | 首次安装耗时 | 使用系统已装 Chrome 作为 fallback |
| 截图文件体积（31+ 张 PNG） | P3 | Git 仓库膨胀 | `docs/reports/testing/evidence/` 加入 `.gitignore`，仅保留缩略图索引 |
| data-id 遗漏导致定位失败 | P2 | Playwright 无法找到元素 | 实施 T-5 data-id 覆盖率自检作为 gate |

## 13. 验收标准

- [ ] 5 个 Vue 组件的所有交互元素（button/input/select/upload 等）均已绑定 data-id
- [ ] data-id 覆盖率自检 ≥ 90%
- [ ] 31 张截图全部生成，文件名含合法 data-id + 中文场景描述
- [ ] 5 个关键用例录屏生成（WebM 格式）
- [ ] 每张截图含 data-id 半透明水印
- [ ] 所有测试步骤使用中文描述
- [ ] 测试报告 `admin-mvp-test-report.md` 更新，每个用例行内嵌截图引用 + 中文步骤
- [ ] `npx playwright test` 全量通过（exit code 0）
- [ ] 不影响现有 56 个 Go 测试和 11 个前端文件的 build/lint 结果
