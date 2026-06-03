# Admin MVP E2E 截图/录屏测试方案（v3 — 主控审核对齐版）

> **状态：待审核** | 提交人：Trae AI | 日期：2026-05-27
> **关联 PRD**：[PRD-10](../prd/PRD-10-Admin管理后台MVP.md) M6' 端到端验收 + M7' 文档
> **主控指导意见来源**：[TabAI会话_1779890854089.md](../../tabbit/inbox/2026/05/TabAI会话_1779890854089.md)

## 1. 背景

当前 `admin-mvp-test-report.md` 已记录 56 个 Go 后端测试用例全部 PASS 的文字证据，但缺少 **可视化测试痕迹**（截图、录屏）。按项目规则 [测试报告模板.md](./测试报告模板.md) §8~9 要求，每个用例应附带截图/视频证据文件路径。

## 2. 目标

为 M0'~M5' 交付的 admin 插件层（config_admin 12 用例 + i18n_admin 20 用例 = **32 个 HTTP handler 测试**）补充 E2E 截图/录屏证据，使测试报告达到「可追溯、可验证」标准。

**定位：内部工具的轻量自检，非生产级追溯系统。**

**核心要求：**
1. admin-web **每个可交互 HTML 元素**必须绑定 `data-id` 属性（四段式命名）
2. 测试用例步骤使用 **中文描述**
3. 截图/录屏 **一律存本地**（`tests/e2e/evidence/`），不上传 OSS
4. **无水印、无 index.json、无外部存储依赖**
5. 录屏数量 **≥ 8 个**
6. Mock 数据 **必须从 Go DTO 自动派生**，禁止手写

## 3. 技术选型

| 维度 | 选型 | 理由 |
|------|------|------|
| 框架 | **Playwright (Node.js)** | 项目已有 Node.js 20 环境，原生支持截图+录屏 |
| 运行位置 | `tests/e2e/`（项目根目录） | 遵循 [testing.md](../../.trae/rules/testing.md) §3 E2E 规范 |
| 截图格式 | PNG（无水印） | 轻量本地存储 |
| 录屏格式 | WebM | 关键用例操作回放 |
| 证据存储 | `tests/e2e/evidence/`（本地 + .gitignore） | 内部工具定位，CI artifact 可选兜底 |

## 4. data-id 绑定规范（四段式）

### 4.1 命名规则

```
data-id="{页面缩写}-{组件类型}-{操作或标识}"
```

**四段式组成：**

| 段位 | 说明 | 取值范围 |
|------|------|---------|
| 页面缩写 | 所属功能模块 | `ca`=config-admin, `ia`=i18n-admin |
| 组件类型 | UI 元素类别 | `btn`=按钮, `input`=输入框, `select`=下拉, `table`=表格, `dialog`=弹窗, `switch`=开关, `textarea`=文本域, `upload`=上传区, `card`=卡片, `result`=结果展示, `pagination`=分页 |
| 操作标识 | 具体操作含义 | 中文拼音（如 sousuo/xinzeng/bianji/shanchu/fabu/daoru/daochu）或英文简写（如 mokuai/ziduan/yuyanma/huanjing） |
| 后缀（可选） | 同类型元素区分 | `-row`（行内）, `-toolbar`（工具栏）, `-dialog`（弹窗内）, `-dongtai-{fieldKey}`（动态字段）|

**段位约束：**
- 段位数 ≥ 3（除全局唯一豁免外，如 `ca-table-schema-list` 为 3 段可接受）
- 全部小写，中划线分隔
- 禁止使用三段以下命名

### 4.2 各组件 data-id 清单（四段式）

#### 4.2.1 config/schema-list.vue 元素绑定

| HTML 元素 | data-id 值（四段式） | 说明 |
|----------|---------------------|------|
| 搜索按钮 | `ca-btn-sousuo` | Schema 列表搜索 |
| 重置按钮 | `ca-btn-chongzhi` | 搜索条件重置 |
| 新增 Schema 按钮 | `ca-btn-xinzeng-schema` | 打开新增弹窗 |
| 工具栏删除按钮 | `ca-btn-shanchu-schema-toolbar` | 批量删除 Schema |
| 行内编辑按钮 | `ca-btn-bianji-schema-row` | 行内编辑 Schema |
| 行内删除按钮 | `ca-btn-shanchu-schema-row` | 行内删除 Schema |
| 弹窗确定按钮 | `ca-btn-queding-schema-dialog` | 新增/编辑弹窗提交 |
| 弹窗取消按钮 | `ca-btn-quxiao-schema-dialog` | 关闭弹窗 |
| 搜索模块 Key 输入框 | `ca-input-mokuai-key` | 搜索条件 |
| 弹窗模块 Key 输入框 | `ca-input-mokuai-key-dialog` | 新增/编辑表单 |
| 弹窗字段 Key 输入框 | `ca-input-ziduan-key-dialog` | 新增/编辑表单 |
| 弹窗字段类型下拉 | `ca-select-ziduan-leixing-dialog` | string/int/float/bool/json |
| 弹窗必填开关 | `ca-switch-bitian-dialog` | required 开关 |
| 弹窗默认值输入 | `ca-input-morenzhi-dialog` | 根据字段类型渲染 |
| 弹窗校验规则文本域 | `ca-textarea-jiaoyan-guize-dialog` | validator JSON |
| 弹窗描述文本域 | `ca-textarea-miaoshu-dialog` | description |
| Schema 列表表格 | `ca-table-schema-list` | 整个表格容器 |
| 分页组件 | `ca-pagination-schema-list` | 列表分页 |

#### 4.2.2 config/value-publish.vue 元素绑定

| HTML 元素 | data-id 值（四段式） | 说明 |
|----------|---------------------|------|
| 模块 Key 下拉选择 | `ca-select-mokuai-value` | 选择配置模块 |
| 环境选择下拉 | `ca-select-huanjing-value` | dev/test/staging/prod |
| 发布配置按钮 | `ca-btn-fabu-peizhi` | 提交发布 |
| 重置按钮 | `ca-btn-chongzhi-form-value` | 重置发布表单 |
| 动态字段输入框 | `ca-input-dongtai-{fieldKey}` | 按 schema 渲染的每个字段 |
| 字段错误提示 | `ca-error-{fieldKey}` | 10400 校验失败红色提示 |
| 10400 错误弹窗 | `ca-dialog-10400-cuowu` | 校验失败详情弹窗 |
| 错误弹窗关闭按钮 | `ca-btn-guanbi-10400-dialog` | 关闭 10400 弹窗 |
| 版本历史表格 | `ca-table-version-history` | 发布历史列表 |
| 刷新 Schema 按钮 | `ca-btn-shuaxin-schema` | 重新加载 schema |

#### 4.2.3 i18n/string-list.vue 元素绑定

| HTML 元素 | data-id 值（四段式） | 说明 |
|----------|---------------------|------|
| 语言包 ID 输入 | `ia-input-yuyanbao-id` | pack ID 数字输入 |
| 查询按钮 | `ia-btn-chaxun-string` | 字符串查询 |
| 重置按钮 | `ia-btn-chongzhi-string` | 条件重置 |
| 新增字符串按钮 | `ia-btn-xinzeng-string` | 打开新增弹窗 |
| 工具栏删除按钮 | `ia-btn-shanchu-string-toolbar` | 批量删除字符串 |
| 行内编辑按钮 | `ia-btn-bianji-string-row` | 编辑字符串 |
| 行内删除按钮 | `ia-btn-shanchu-string-row` | 删除字符串 |
| 弹窗 String Key 输入 | `ia-input-string-key-dialog` | 字符串唯一标识 |
| 弹窗字符串值文本域 | `ia-textarea-string-value-dialog` | 翻译内容 |
| 弹窗模板类型下拉 | `ia-select-moban-leixing-dialog` | plain/named/icu |
| 弹窗分组输入 | `ia-input-fenzu-dialog` | groupName |
| 弹窗参数 Schema 文本域 | `ia-textarea-canshu-schema-dialog` | named 类型参数定义 |
| 弹窗预览示例文本域 | `ia-textarea-yulan-shili-dialog` | 模板预览 |
| 弹窗确定按钮 | `ia-btn-queding-string-dialog` | 提交字符串 |
| 弹窗取消按钮 | `ia-btn-quxiao-string-dialog` | 关闭弹窗 |
| 字符串列表表格 | `ia-table-string-list` | 整个表格 |
| String Key popover | `ia-popover-string-preview` | 长文本预览 |

#### 4.2.4 i18n/pack-manage.vue 元素绑定

| HTML 元素 | data-id 值（四段式） | 说明 |
|----------|---------------------|------|
| 语言包 ID 输入 | `ia-input-pack-id` | pack ID |
| 语言码下拉 | `ia-select-yuyanma-pack` | zh-CN/en-US/ja-JP/ko-KR |
| 目标环境下拉 | `ia-select-huanjing-pack` | dev/test/staging/prod |
| 发布语言包按钮 | `ia-btn-fabu-yueyanbao` | 提交发布 |
| 回滚版本号输入 | `ia-input-huinban-banhao` | 目标版本号 |
| 回滚确认按钮 | `ia-btn-queren-huinban` | 执行回滚 |
| 发布结果卡片 | `ia-card-fabu-jieguo-pack` | PackID/LangCode/Version 展示 |

#### 4.2.5 i18n/import-export.vue 元素绑定

| HTML 元素 | data-id 值（四段式） | 说明 |
|----------|---------------------|------|
| 导入目标 Pack ID 输入 | `ia-import-input-pack-id` | CSV 导入目标 |
| 文件上传区域 | `ia-upload-csv-wenjian` | el-upload drag 区域 |
| 开始导入按钮 | `ia-btn-kaishi-daoru` | 触发导入 |
| 导入重置按钮 | `ia-btn-chongzhi-daoru` | 重置导入区 |
| 导入结果展示 | `ia-result-daoru-jieguo` | el-result 成功/失败 |
| 导入错误详情表格 | `ia-table-daoru-cuowu` | 失败行明细 |
| 导出源 Pack ID 输入 | `ia-export-input-pack-id` | CSV 导出源 |
| 导出 CSV 按钮 | `ia-btn-daochu-csv` | 触发导出下载 |

### 4.3 Vue 模板绑定示例

```vue
<!-- ✅ 正确：每个交互元素都有四段式 data-id -->
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
  data-id="ca-input-mokuai-key-dialog"
/>

<el-upload
  drag
  action=""
  data-id="ia-upload-csv-wenjian"
>
  <i class="el-icon-upload" />
</el-upload>

<!-- ❌ 错误 1：缺少 data-id -->
<el-button type="primary" @click="handleAdd">新增</el-button>

<!-- ❌ 错误 2：data-id 仅两段 -->
<el-button data-id="btn-add" @click="handleAdd">新增</el-button>
```

### 4.4 自检命令

```bash
# 1. data-id 覆盖率检查（期望 ≥60 个绑定点）
grep -rc 'data-id=' typescript/admin-web/src/views/config/*.vue \
              typescript/admin-web/src/views/i18n/*.vue \
          | awk -F: '{sum+=$2} END {print "data-id 总数:", sum}'

# 2. 段位合规性检查（统计 ≤2 段的违规项）
grep -roh 'data-id="[^"]*"' typescript/admin-web/src/views/config/*.vue \
              typescript/admin-web/src/views/i18n/*.vue \
      | awk -F'" '{n=split($2,a,"-"); if(n<3) print "⚠️ 段位不足:", $2}' \
      | wc -l
# 期望输出：0

# 3. 双向校验：Vue 中存在的 data-id 必须在 spec 中被引用
# （见 T-5 步骤详细说明）
```

## 5. 截图/录屏文件规范

### 5.1 文件名格式

```
{dataId编号}-{中文场景描述}-{时间戳}.{png|webm}
```

**示例：**

| 文件名 | 对应用例 | 场景说明 |
|--------|---------|---------|
| `ca-01-Schema列表正常渲染-20260527.png` | ca-01 | Schema 列表页加载后正常显示数据 |
| `ca-09-10400校验错误字段展示-20260527.png` | ca-09 | 配置值发布校验失败时 10400 弹窗截图 |
| `ia-14-CSV导入成功结果展示-20260527.png` | ia-14 | CSV 上传导入后的成功结果页 |
| `ia-14-CSV导入成功操作录屏-20260527.webm` | ia-14 | CSV 导入完整操作流程录屏 |

### 5.2 存储约定（无 OSS）

```text
存储路径：tests/e2e/evidence/
.gitignore：tests/e2e/evidence/*.png + tests/e2e/evidence/*.webm
占位文件：tests/e2e/evidence/.gitkeep（空目录进 git）
清理策略：本地不主动清理，开发者按需 rm
CI artifact：可选，仅 CI 失败或 main 分支合并时自动归档 7 天
测试报告引用：相对路径 ![desc](../../../tests/e2e/evidence/{filename})
```

**无水印、无 OSS、无 index.json。**

## 6. 目录结构（v3 简化版）

```text
tests/e2e/
├── playwright.config.ts                  # 全局配置（video=on-first-retry）
├── package.json                          # 依赖隔离（仅 @playwright/test + @types/node）
├── evidence/                             # 本地证据目录（.gitignore）
│   └── .gitkeep                           # 空目录占位
├── utils/
│   ├── evidence.ts                       # 截图工具（35 行，零水印，零 DOM 注入）
│   ├── data-id-coverage-check.ts         # Vue 端覆盖率自检
│   ├── data-id-bidirectional-check.mjs   # 双向校验 hard gate
│   ├── mock-from-dto.mjs                 # 从 Go DTO 自动派生 Mock 数据
│   └── test-helpers.ts                   # API mock / 启动辅助
├── fixtures/
│   ├── config-dto-mock.json              # ← mock-from-dto.mjs 自动生成
│   ├── i18n-dto-mock.json                # ← mock-from-dto.mjs 自动生成
│   └── test-data.json                    # 兼容性入口
└── specs/
    ├── config-admin.spec.ts               # 12 个中文用例
    └── i18n-admin.spec.ts                 # 20 个中文用例

docs/reports/testing/
└── admin-mvp-test-report.md              # 更新后含相对路径图片引用
```

**撤销项（v2 已有，v3 删除）：**
- ~~utils/oss-uploader.ts~~
- ~~docs/reports/testing/evidence/~~（整目录迁移到 tests/e2e/evidence/）
- ~~evidence-index.json~~
- ~~injectWatermark / removeWatermark~~
- ~~@cairobot/oss-client 依赖~~

## 7. 核心工具函数设计（v3 简化版）

### 7.1 evidence.ts — 本地截图（35 行，无水印）

```typescript
import { Page } from '@playwright/test'
import path from 'path'
import fs from 'fs/promises'

export interface EvidenceOptions {
  /** 对应测试报告中的 data-id 编号，如 "ca-01" */
  dataId: string
  /** 中文场景描述 */
  description: string
  /** 是否全页截图，默认 true */
  fullPage?: boolean
}

const EVIDENCE_DIR = path.resolve(__dirname, '../evidence')

/**
 * 本地截图保存（无水印、无 DOM 注入、无外部依赖）
 * 文件名：{dataId}-{description}-{timestamp}.png
 */
export async function takeEvidence(
  page: Page,
  options: EvidenceOptions
): Promise<string> {
  const { dataId, description, fullPage = true } = options
  const ts = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19)
  const safeDesc = description.replace(/[/\\?%*:|"<>]/g, '-')
  const filename = `${dataId}-${safeDesc}-${ts}.png`
  const filepath = path.join(EVIDENCE_DIR, filename)

  await fs.mkdir(EVIDENCE_DIR, { recursive: true })
  await page.screenshot({ path: filepath, fullPage })
  return filepath
}

/**
 * 通过 data-id 定位元素，操作前后各截一张
 */
export async function clickAndScreenshot(
  page: Page,
  dataId: string,
  options: EvidenceOptions
): Promise<void> {
  await takeEvidence(page, {
    ...options,
    description: options.description + '-操作前',
  })
  await page.locator(`[data-id="${dataId}"]`).click()
  await page.waitForTimeout(500)
  await takeEvidence(page, {
    ...options,
    description: options.description + '-操作后',
  })
}
```

**与 v2 差异：删除 injectWatermark / removeWatermark / nowIso8601CST / page.evaluate 全部逻辑，纯 Playwright 原生 screenshot。**

### 7.2 mock-from-dto.mjs — 从 Go DTO 自动派生 Mock

```javascript
/**
 * 从 Go DTO 定义自动生成 Mock JSON
 * 禁止手写 Mock 数据，确保 Mock 与 DTO 结构一致
 *
 * 输入：Go 结构体定义（从 .go 文件提取）
 * 输出：fixtures/{config,i18n}-dto-mock.json
 */

import fs from 'fs'
import path from 'path'

const GO_DTO_FILES = [
  '../go/admin/app/admin/config_admin/models/dto.go',
  '../go/admin/app/admin/i18n_admin/models/dto.go',
]

function extractGoStructs(content) {
  // 解析 Go type Xxx struct { ... } 定义
  // 返回字段名+类型的映射
}

function generateMockJSON(structs) {
  // 根据 Go 类型映射生成合理 Mock 值：
  // string → "test-value", int → 42, bool → true
  // []T → [mock T], map[string]T → {"key": mock T}
}

// 主函数：读取 Go DTO 文件 → 提取结构体 → 生成 Mock JSON
for (const goFile of GO_DTO_FILES) {
  const content = fs.readFileSync(path.resolve(__dirname, goFile), 'utf-8')
  const structs = extractGoStructs(content)
  const mockData = generateMockJSON(structs)
  const outPath = path.resolve(__dirname, '../fixtures/', path.basename(goFile, '.go') + '-mock.json')
  fs.writeFileSync(outPath, JSON.stringify(mockData, null, 2))
  console.log(`✅ 生成 ${outPath}`)
}
```

### 7.3 data-id-bidirectional-check.mjs — 双向校验 hard gate

```javascript
/**
 * 双向校验：
 * 1. Vue 中声明的 data-id 必须在 spec 中被引用（Vue→spec 方向）
 * 2. spec 中引用的 data-id 必须在 Vue 中声明（spec→Vue 方向）
 *
 * 任一方向未命中 → CI 失败（hard gate）
 */
import fs from 'fs'
import path from 'path'

const VUE_DIRS = [
  '../../../typescript/admin-web/src/views/config/',
  '../../../typescript/admin-web/src/views/i18n/',
]
const SPEC_DIR = './specs/'

function extractVueDataIds() {
  const ids = new Set()
  for (const dir of VUE_DIRS) {
    const files = fs.readdirSync(path.resolve(__dirname, dir))
    for (const f of files.filter(x => x.endsWith('.vue'))) {
      const content = fs.readFileSync(path.resolve(__dirname, dir, f), 'utf-8')
      for (const m of content.matchAll(/data-id="([^"]+)"/g)) {
        ids.add(m[1])
      }
    }
  }
  return ids
}

function extractSpecDataIds() {
  const ids = new Set()
  const files = fs.readdirSync(path.resolve(__dirname, SPEC_DIR)).filter(x => x.endsWith('.spec.ts'))
  for (const f of files) {
    const content = fs.readFileSync(path.resolve(__dirname, SPEC_DIR, f), 'utf-8')
    for (const m of content.matchAll(/data-id="([^"]+)"/g)) {
      ids.add(m[1])
    }
  }
  return ids
}

const vueIds = extractVueDataIds()
const specIds = extractSpecDataIds()

// Vue→spec 方向：Vue 有但 spec 没引用
const vueOnly = [...vueIds].filter(id => !specIds.has(id))
if (vueOnly.length > 0) {
  console.error(`❌ Vue 中存在 ${vueOnly.length} 个 data-id 未被 spec 引用：`)
  vueOnly.forEach(id => console.error(`   未引用: ${id}`))
  process.exit(1)
}

// spec→Vue 方向：spec 引用了但 Vue 不存在
const specOnly = [...specIds].filter(id => !vueIds.has(id))
if (specOnly.length > 0) {
  console.error(`❌ Spec 中引用了 ${specOnly.length} 个 data-id 在 Vue 中不存在：`)
  specOnly.forEach(id => console.error(`   缺失: ${id}`))
  process.exit(1)
}

console.log(`✅ 双向校验通过：Vue=${vueIds.size}, Spec=${specIds.size}, 100% 命中`)
```

## 8. 测试策略

### 8.1 Mock Server 模式

```
Playwright (Chromium)
  ├─ 访问 localhost:8080（前端 dev server）
  ├─ page.route() 拦截 XHR → 返回 mock-from-dto.mjs 生成的 JSON
  ├─ page.locator('[data-id="xxx"]') 定位元素（四段式）
  ├─ 模拟点击/输入/上传操作
  ├─ 断言 DOM 渲染正确性
  └─ takeEvidence() 截图存入 tests/e2e/evidence/
```

**Mock 数据来源：必须从 Go DTO 文件（dto.go）自动派生，禁止手写 JSON。**

### 8.2 用例映射表（32 个 → 中文步骤 + data-id + 录屏标记）

#### config_admin（12 个）

| data-id | 用例名称 | 中文测试步骤 | 涉及 data-id 元素 | 录屏 |
|---------|---------|-------------|------------------|:----:|
| ca-01 | Schema 列表正常渲染 | 1.打开 Schema 列表页<br>2.确认表格正常显示数据<br>3.截图记录初始状态 | `ca-table-schema-list`, `ca-pagination-schema-list` | 否 |
| ca-02 | 空 ModuleKey 查询 | 1.清空模块 Key 输入框<br>2.点击搜索按钮<br>3.确认表格为空或显示无数据 | `ca-input-mokuai-key`, `ca-btn-sousuo`, `ca-table-schema-list` | 否 |
| ca-03 | 正常创建 Schema | 1.点击新增 Schema 按钮<br>2.在弹窗填写 moduleKey=app.server, fieldKey=port<br>3.选择字段类型=int<br>4.点击确定按钮<br>5.确认成功提示并截图 | `ca-btn-xinzeng-schema`, `ca-input-mokuai-key-dialog`, `ca-input-ziduan-key-dialog`, `ca-select-ziduan-leixing-dialog`, `ca-btn-queding-schema-dialog` | **是** |
| ca-04 | 创建时 Body 为空 | 1.点击新增 Schema 按钮<br>2.不填写任何内容直接点确定<br>3.确认必填项校验红色提示出现 | `ca-btn-xinzeng-schema`, `ca-btn-queding-schema-dialog` | 否 |
| ca-05 | 正常更新 Schema | 1.点击某行的编辑按钮<br>2.确认弹窗回显当前数据<br>3.修改字段值<br>4.点击确定<br>5.确认更新成功 | `ca-btn-bianji-schema-row`, `ca-input-ziduan-key-dialog`, `ca-btn-queding-schema-dialog` | 否 |
| ca-06 | 无效 ID 删除 | 1.传入 id=abc 触发删除请求<br>2.确认返回参数错误提示 | `ca-btn-shanchu-schema-row` | 否 |
| ca-07 | 正常删除 Schema | 1.选中一行数据<br>2.点击删除按钮<br>3.在确认对话框中点击确定<br>4.确认列表已刷新 | `ca-btn-shanchu-schema-row`, `ca-btn-shanchu-schema-toolbar` | **是** |
| ca-08 | 正常发布配置值 | 1.打开配置值发布页<br>2.选择模块 Key 和环境<br>3.确认动态字段按 Schema 渲染<br>4.填写各字段值<br>5.点击发布按钮<br>6.确认发布成功并显示版本号 | `ca-select-mokuai-value`, `ca-select-huanjing-value`, `ca-input-dongtai-*`, `ca-btn-fabu-peizhi` | **是** |
| ca-09 | 10400 校验错误展示 | 1.打开配置值发布页<br>2.填写超出范围的数值（如 port=99999）<br>3.点击发布<br>4.确认弹出 10400 错误详情弹窗<br>5.确认字段级错误映射到对应输入框下方 | `ca-input-dongtai-port`, `ca-btn-fabu-peizhi`, `ca-dialog-10400-cuowu`, `ca-error-port` | **是（重点）** |
| ca-10 | 空 Fields 发布 | 1.不选择任何模块直接点发布<br>2.确认返回空 fields 错误 | `ca-btn-fabu-peizhi` | 否 |
| ca-11 | 缺 ModuleKey 查版本 | 1.不选模块直接查版本历史<br>2.确认返回缺少参数错误 | （版本历史区域） | 否 |
| ca-12 | 正常查询版本历史 | 1.选择有效模块和环境<br>2.确认版本历史表格正常渲染 | `ca-select-mokuai-value`, `ca-table-version-history` | 否 |

#### i18n_admin（20 个）

| data-id | 用例名称 | 中文测试步骤 | 涉及 data-id 元素 | 录屏 |
|---------|---------|-------------|------------------|:----:|
| ia-01 | 正常创建字符串 | 1.打开字符串管理页<br>2.填入 pack_id=1<br>3.点击新增字符串按钮<br>4.填写 string_key=greeting.hello, string_value=你好<br>5.选择模板类型=plain<br>6.点击确定<br>7.确认创建成功 | `ia-input-yuyanbao-id`, `ia-btn-xinzeng-string`, `ia-input-string-key-dialog`, `ia-textarea-string-value-dialog`, `ia-select-moban-leixing-dialog`, `ia-btn-queding-string-dialog` | **是** |
| ia-02 | 缺少必填字段创建 | 1.点击新增字符串按钮<br>2.不填任何内容直接点确定<br>3.确认必填项校验红色提示 | `ia-btn-xinzeng-string`, `ia-btn-queding-string-dialog` | 否 |
| ia-03 | 10400 模板校验错误 | 1.点击新增字符串按钮<br>2.填写 string_value=Hello {name}<br>3.选择模板类型=named（未提供 params_schema）<br>4.点击确定<br>5.确认返回 10400 模板错误提示 | `ia-textarea-string-value-dialog`, `ia-select-moban-leixing-dialog`, `ia-btn-queding-string-dialog` | **是（重点）** |
| ia-04 | 正常更新字符串 | 1.点击某行的编辑按钮<br>2.确认弹窗回显当前数据<br>3.修改字符串值<br>4.点击确定<br>5.确认更新成功 | `ia-btn-bianji-string-row`, `ia-textarea-string-value-dialog`, `ia-btn-queding-string-dialog` | 否 |
| ia-05 | 空 Body 更新 | 1.模拟发送空的 PUT 请求体<br>2.确认返回 400 错误 | — | 否 |
| ia-06 | 无效 ID 删除字符串 | 1.传入 id=abc 触发删除<br>2.确认返回参数错误提示 | `ia-btn-shanchu-string-row` | 否 |
| ia-07 | 正常删除字符串 | 1.选中一行字符串<br>2.点击删除按钮<br>3.确认对话框点确定<br>4.确认列表刷新 | `ia-btn-shanchu-string-row`, `ia-btn-shanchu-string-toolbar` | 否 |
| ia-08 | 正常查询字符串列表 | 1.填入 pack_id=1<br>2.点击查询<br>3.确认表格渲染字符串数据 | `ia-input-yuyanbao-id`, `ia-btn-chaxun-string`, `ia-table-string-list` | 否 |
| ia-09 | 缺 PackId 查询 | 1.不填 pack_id 直接查<br>2.确认返回缺少参数错误 | `ia-btn-chaxun-string` | 否 |
| ia-10 | 正常发布语言包 | 1.打开语言包管理页<br>2.填入 pack_id=1, 选择 lang_code=zh-CN, env=dev<br>3.点击发布按钮<br>4.确认返回版本号信息 | `ia-input-pack-id`, `ia-select-yuyanma-pack`, `ia-select-huanjing-pack`, `ia-btn-fabu-yueyanbao`, `ia-card-fabu-jieguo-pack` | **是** |
| ia-11 | 空 Body 发布语言包 | 1.不填任何内容点发布<br>2.确认返回 400 错误 | `ia-btn-fabu-yueyanbao` | 否 |
| ia-12 | 正常回滚语言包 | 1.填入 pack_id=1<br>2.输入回滚版本号=3<br>3.点击回滚确认<br>4.确认回滚成功 | `ia-input-pack-id`, `ia-input-huinban-banhao`, `ia-btn-queren-huinban` | 否 |
| ia-13 | 空 Body 回滚 | 1.不填版本号点回滚<br>2.确认返回 400 错误 | `ia-btn-queren-huinban` | 否 |
| ia-14 | CSV 正常导入 | 1.填入目标 pack_id=1<br>2.将测试 CSV 文件拖入上传区域<br>3.点击开始导入<br>4.确认显示导入成功结果（total/success/fail） | `ia-import-input-pack-id`, `ia-upload-csv-wenjian`, `ia-btn-kaishi-daoru`, `ia-result-daoru-jieguo` | **是（重点）** |
| ia-15 | CSV 部分失败返回 10400 | 1.上传包含错误行的 CSV<br>2.点击开始导入<br>3.确认显示 code=10400 部分失败结果<br>4.确认错误明细表格展示行号和原因 | `ia-upload-csv-wenjian`, `ia-btn-kaishi-daoru`, `ia-result-daoru-jieguo`, `ia-table-daoru-cuowu` | **是（重点）** |
| ia-16 | 缺文件导入 | 1.不选择文件直接点开始导入<br>2.确认返回「请先选择文件」提示 | `ia-btn-kaishi-daoru` | 否 |
| ia-17 | 无效 PackId 导入 | 1.填入 pack_id=abc<br>2.点开始导入<br>3.确认返回参数错误提示 | `ia-import-input-pack-id`, `ia-btn-kaishi-daoru` | 否 |
| ia-18 | CSV 正常导出 | 1.填入有效的 pack_id=1<br>2.点击导出 CSV 按钮<br>3.确认浏览器触发下载（文件名 strings_1.csv） | `ia-export-input-pack-id`, `ia-btn-daochu-csv` | 否 |
| ia-19 | 无效 PackId 导出 | 1.填入 pack_id=abc<br>2.点导出<br>3.确认返回参数错误 | `ia-export-input-pack-id`, `ia-btn-daochu-csv` | 否 |

### 8.3 录屏清单（8 个，必做）

| 序号 | data-id | 用例名称 | 录屏理由 |
|-----|---------|---------|---------|
| 1 | ca-03 | 正常创建 Schema | 完整 CRUD 弹窗交互流程 |
| 2 | ca-07 | 正常删除 Schema | 确认对话框 + 列表刷新联动 |
| 3 | ca-08 | 正常发布配置值 | 动态表单渲染 + 版本号展示 |
| 4 | ca-09 | 10400 校验错误展示 | **核心业务价值：字段级错误映射** |
| 5 | ia-01 | 正常创建字符串 | 字符串 CRUD 弹窗交互 |
| 6 | ia-03 | 10400 模板校验错误 | **核心业务价值：模板错误展示** |
| 7 | ia-10 | 正常发布语言包 | 结果卡片展示 |
| 8 | ia-14 | CSV 正常导入 | **拖拽上传 + el-result 展示** |
| 9 | ia-15 | CSV 部分失败 10400 | **部分失败 warning + 错误明细表** |

> 注：录屏数 ≥ 8，上表列出 9 个候选，实施时可按实际情况调整。

## 9. 实施步骤

| 步骤 | 内容 | 产出 | 工作量 |
|------|------|------|-------|
| T-1 | 为 5 个 Vue 组件补充四段式 data-id（~60 个绑定点） | 5 个 .vue 文件修改 | 1 天 |
| T-2 | 安装 Playwright + 初始化配置 + 工具函数（evidence.ts/mock-from-dto/双向校验） | `playwright.config.ts`, `utils/*.{ts,mjs}`, `package.json` | 0.5 天 |
| T-3 | 编写 config-admin.spec.ts（12 个中文用例 + 3 录屏） | 12 张截图 + 3 个录屏 | 1.5 天 |
| T-4 | 编写 i18n-admin.spec.ts（20 个中文用例 + 5 录屏） | 19 张截图 + 5 个录屏 | 1.5 天 |
| T-5 | 运行 data-id 双向校验 + 覆盖率自检 + 全量 E2E 测试 | 自检报告 + `tests/e2e/evidence/*.{png,webm}` | 0.5 天 |
| T-6 | 更新测试报告（嵌入相对路径截图引用 + 中文步骤） | `admin-mvp-test-report.md` v2 | 0.5 天 |
| **合计** | | **31 张截图 + ≥8 个录屏 + 1 份更新后的测试报告 + 5 个 Vue 组件增强** | **5.5 天** |

## 10. 依赖清单

```json
{
  "devDependencies": {
    "@playwright/test": "^1.48.0",
    "@types/node": "^20.0.0"
  }
}
```

**禁止引入：**
- ❌ OSS SDK（如 @cairobot/oss-client、ali-oss、aws-sdk）
- ❌ 水印库（如 canvas、sharp、jimp）
- ❌ index.json 生成器

## 11. 与现有测试的关系

| 现有测试 | 职责 | 不变 |
|---------|------|------|
| Go 单元测试（56 个） | 验证 handler 逻辑正确性 | 保持不变 |
| pnpm build + eslint | 验证前端编译正确性 | 保持不变 |
| grep 自检 | 验证边界铁律 | 保持不变 |
| **新增：Playwright E2E** | **验证 UI 渲染 + 用户交互 + 10400 错误展示 + data-id 可追溯性** | **新增** |

三者互补，不重复。

## 12. 风险评估

| 风险 ID | 等级 | 影响 | 应对 |
|---------|------|------|------|
| R-DATAID-DRIFT | **P1** | data-id 与 Vue 代码漂移导致定位失败 | 双向校验 CI hard gate（data-id-bidirectional-check.mjs），100% 命中才允许 merge |
| R-MOCK-DRIFT | P2 | 手写 Mock 与 Go DTO 结构不一致 | **禁止手写**，强制 mock-from-dto.mjs 从 dto.go 自动派生 |
| R-LOCAL-DISK | P3 | 本地证据丢失（误删/换机） | CI artifact 兜底（可选，main 分支合并触发，保留 7 天）；开发本地按需重跑 |
| R-REVIEW-FRICTION | P3 | PR 评审者看不到截图 | 评审者本地 git pull + 跑一次 E2E 即可；若团队反馈强烈不便 → S2 再考虑接 OSS（**非本批次范围**） |
| 前端 dev server 未启动 | P2 | E2E 无法访问页面 | 测试脚本自动检测端口，未启动则 SKIP 并 WARN |

**撤销风险（v2 已有，v3 删除）：**
- ~~R-OSS-DOWN~~（已无 OSS 依赖）

## 13. 验收标准（12 条）

### 必须通过（9 条）

- [ ] **A1**: 5 个 Vue 组件的所有交互元素（button/input/select/upload/switch/textarea）均已绑定 **四段式 data-id**
- [ ] **A2**: data-id 覆盖率自检 **≥90%**（交互元素总数 vs data-id 绑定数）
- [ ] **A3**: data-id **双向校验 100% 命中**（Vue 声明 ↔ spec 引用，hard gate）
- [ ] **A4**: **31 张截图**全部生成在 `tests/e2e/evidence/`，文件名含 data-id + 中文场景描述
- [ ] **A5**: **≥8 个录屏**（WebM 格式）生成
- [ ] **A6**: 所有测试步骤使用 **中文描述**
- [ ] **A7**: 测试报告 `admin-mvp-test-report.md` 用 **相对路径**引用截图（如 `../../../tests/e2e/evidence/ca-01-xxx.png`）
- [ ] **A8**: `npx playwright test` 全量通过（exit code = 0）
- [ ] **A9**: 不影响现有 56 个 Go 测试和 11 个前端文件的 build/lint 结果

### 新增通过（3 条）

- [ ] **B1**: `tests/e2e/evidence/*.png` 与 `*.webm` 已被 **.gitignore 排除**（不会误提交到仓库）
- [ ] **B2**: Mock 数据由 **mock-from-dto.mjs** 从 Go DTO 自动派生，**禁止手写 JSON**
- [ ] **B3**: evidence.ts **无水印逻辑**（无 injectWatermark/removeWatermark/page.evaluate/DOM 注入）

### 明确撤销（2 条）

- [ ] **C1**: ~~每张截图含 data-id 半透明水印~~ → **撤回，无水印**
- [ ] **C2**: ~~evidence-index.json 含全部截图索引~~ → **撤回，无 index.json**

## 14. 全程禁止事项

| # | 禁止事项 | 原因 |
|---|---------|------|
| 1 | 禁止照抄 v2 方案的 evidence.ts（含 4 处 bug：水印逻辑、filename 拼接错误等） | 使用 v3 主控简化版（35 行） |
| 2 | 禁止手写 Mock JSON | 必须用 mock-from-dto.mjs 从 Go DTO 派生 |
| 3 | 禁止 spec.ts 引用的 data-id 在 Vue 中不存在 | 双向校验 hard gate 会拦截 |
| 4 | 禁止截图直接进 git 仓库 | .gitignore 强制排除 *.png + *.webm |
| 5 | 禁止录屏数量少于 8 个 | 验收标准 A5 硬性要求 |
| 6 | 禁止 data-id 三段以下命名 | 验收标准 A1 四段式要求 |
| 7 | 禁止引入 OSS SDK / 水印库 / index.json 生成器 | 简化要求，内部工具定位 |
| 8 | 禁止 `docs/reports/testing/evidence/` 目录残留 | v3 已迁移到 `tests/e2e/evidence/` |
