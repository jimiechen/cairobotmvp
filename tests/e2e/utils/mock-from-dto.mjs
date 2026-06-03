import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const GO_DTO_FILES = [
  '../../../go/admin/app/admin/config_admin/models/dto.go',
  '../../../go/admin/app/admin/i18n_admin/models/dto.go',
]

function extractGoStructs(content) {
  const structs = []
  const structRegex = /type\s+(\w+)\s+struct\s*\{([\s\S]*?)^\n/mg
  let match

  while ((match = structRegex.exec(content)) !== null) {
    const structName = match[1]
    const body = match[2]
    const fields = []
    const fieldRegex = /^\s+(\w+)\s+([\[\]\w.*]+)\s+`json:"([^"]*)"(?:\s+binding:"([^"]*)")?/gm
    let fieldMatch

    while ((fieldMatch = fieldRegex.exec(body)) !== null) {
      fields.push({
        name: fieldMatch[1],
        goType: fieldMatch[2],
        jsonTag: fieldMatch[3],
        bindingTag: fieldMatch[4] || '',
      })
    }

    if (fields.length > 0) {
      structs.push({ name: structName, fields })
    }
  }

  return structs
}

function generateMockValue(goType, fieldName) {
  if (goType === 'string') {
    const valueMap = {
      module_key: 'app.server.timeout',
      field_key: 'port',
      field_type: 'int',
      default_value: '8080',
      validator: '{"min":0,"max":65535}',
      description: '服务端口号',
      client_scope: '*',
      string_key: 'greeting.hello',
      string_value: '你好，世界！',
      group_name: 'common',
      template_type: 'plain',
      params_schema: '{"name":"string"}',
      preview_sample: 'Hello {name}',
      message: '操作成功',
      reason: '字段值超出范围',
      lang_code: 'zh-CN',
      env: 'dev',
      data: 'string_key,string_value,group_name,template_type\n',
    }
    return valueMap[fieldName] || `mock-${fieldName}`
  }

  if (goType === 'int' || goType === 'int64') {
    const intMap = {
      id: 1, ID: 1, PackID: 1, pack_id: 1,
      sort_order: 0, Version: 1, version: 1,
      TotalRows: 10, total_rows: 10,
      SuccessCount: 8, success_count: 8,
      FailCount: 2, fail_count: 2,
      RowNum: 3, row_num: 3,
      FieldCount: 5, field_count: 5,
      TargetVersion: 2, target_version: 2,
    }
    return intMap[fieldName] !== undefined ? intMap[fieldName] : 42
  }

  if (goType === 'bool') {
    const boolMap = {
      IsRequired: true, is_required: true,
      IsSecret: false, is_secret: false,
      IsEnabled: true, is_enabled: true,
    }
    return boolMap[fieldName] !== undefined ? boolMap[fieldName] : true
  }

  if (goType.startsWith('[]')) {
    const innerType = goType.replace('[]', '')
    if (innerType === 'PublishFieldItem') {
      return [
        { field_key: 'port', value: 8080 },
        { field_key: 'timeout', value: 30 },
      ]
    }
    if (innerType === 'ValidationErrorItem') {
      return [{ field: 'port', reason: '不能大于 65535' }]
    }
    if (innerType === 'ImportErrorItem') {
      return [{ row_num: 3, reason: 'string_key 不能为空' }]
    }
    return [generateMockValue(innerType, fieldName + '_0')]
  }

  if (goType === 'interface{}') {
    return 'mock-value'
  }

  return null
}

function structToMock(struct) {
  const mock = {}
  for (const field of struct.fields) {
    if (field.jsonTag === '-' || field.jsonTag === '') continue
    mock[field.jsonTag] = generateMockValue(field.goType, field.name)
  }
  return mock
}

const output = {
  config_admin: {},
  i18n_admin: {},
}

for (let i = 0; i < GO_DTO_FILES.length; i++) {
  const goFile = GO_DTO_FILES[i]
  const filePath = path.resolve(__dirname, goFile)
  if (!fs.existsSync(filePath)) {
    console.warn(`⚠️  文件不存在: ${filePath}`)
    continue
  }

  const content = fs.readFileSync(filePath, 'utf-8')
  const structs = extractGoStructs(content)
  const moduleKey = i === 0 ? 'config_admin' : 'i18n_admin'

  for (const s of structs) {
    output[moduleKey][s.name] = structToMock(s)
  }

  console.log(`✅ 解析 ${goFile}: ${structs.length} 个结构体`)
}

const fixturesDir = path.resolve(__dirname, '../fixtures')
fs.mkdirSync(fixturesDir, { recursive: true })

fs.writeFileSync(
  path.join(fixturesDir, 'config-dto-mock.json'),
  JSON.stringify(output.config_admin, null, 2),
  'utf-8'
)
fs.writeFileSync(
  path.join(fixturesDir, 'i18n-dto-mock.json'),
  JSON.stringify(output.i18n_admin, null, 2),
  'utf-8'
)

const testDataSet = {
  config_schemas: [
    {
      id: 1, module_key: 'app.server', field_key: 'port',
      field_type: 'int', default_value: '8080',
      validator: '{"min":0,"max":65535}', is_required: true,
      is_secret: false, description: '服务监听端口',
      client_scope: '*', sort_order: 1, is_enabled: true,
    },
    {
      id: 2, module_key: 'app.server', field_key: 'timeout',
      field_type: 'int', default_value: '30',
      validator: '{"min":1,"max":300}', is_required: true,
      is_secret: false, description: '请求超时时间(秒)',
      client_scope: '*', sort_order: 2, is_enabled: true,
    },
    {
      id: 3, module_key: 'app.server', field_key: 'debug_mode',
      field_type: 'bool', default_value: 'false',
      validator: '', is_required: false,
      is_secret: false, description: '调试模式开关',
      client_scope: '*', sort_order: 3, is_enabled: true,
    },
    {
      id: 4, module_key: 'cache.redis', field_key: 'host',
      field_type: 'string', default_value: '127.0.0.1',
      validator: '', is_required: true,
      is_secret: false, description: 'Redis 主机地址',
      client_scope: '*', sort_order: 1, is_enabled: true,
    },
    {
      id: 5, module_key: 'cache.redis', field_key: 'db_index',
      field_type: 'int', default_value: '0',
      validator: '{"min":0,"max":15}', is_required: true,
      is_secret: false, description: 'Redis 数据库索引',
      client_scope: '*', sort_order: 2, is_enabled: true,
    },
  ],
  config_versions: [
    { version: 1, module_key: 'app.server', env: 'dev', field_count: 3, published_at: '2026-05-27 10:00:00', operator: 'admin' },
    { version: 2, module_key: 'app.server', env: 'dev', field_count: 3, published_at: '2026-05-27 11:00:00', operator: 'admin' },
  ],
  i18n_strings: [
    {
      id: 1, pack_id: 1, string_key: 'greeting.hello',
      string_value: '你好，世界！', group_name: 'common',
      template_type: 'plain', operation_type: 'create', version: 1,
    },
    {
      id: 2, pack_id: 1, string_key: 'error.notFound',
      string_value: '页面未找到', group_name: 'error',
      template_type: 'plain', operation_type: 'create', version: 1,
    },
    {
      id: 3, pack_id: 1, string_key: 'greeting.welcome',
      string_value: '欢迎回来，{name}！您有{count}条新消息。',
      group_name: 'common', template_type: 'named',
      params_schema: '{"name":"string","count":"number"}',
      preview_sample: '欢迎回来，张三！您有5条新消息。',
      operation_type: 'create', version: 1,
    },
  ],
  i18n_pack_result: { pack_id: 1, lang_code: 'zh-CN', version: 3 },
  import_result_success: {
    code: 200, message: '导入完成',
    total_rows: 5, success_count: 5, fail_count: 0, errors: [],
  },
  import_result_partial: {
    code: 10400, message: '存在部分行校验未通过',
    total_rows: 5, success_count: 3, fail_count: 2,
    errors: [
      { row_num: 3, reason: 'string_key 不能为空' },
      { row_num: 5, reason: 'template_type 值非法: xxx' },
    ],
  },
  validation_error_10400: {
    code: 10400, message: '参数校验失败',
    errors: [{ field: 'port', reason: '不能大于 65535' }],
  },
  template_error_10400: {
    code: 10400, message: '模板格式校验失败',
    errors: [{ reason: 'named 模板类型必须提供 params_schema' }],
  },
}

fs.writeFileSync(
  path.join(fixturesDir, 'test-data.json'),
  JSON.stringify(testDataSet, null, 2),
  'utf-8'
)

console.log(`\n✅ Mock 数据生成完成:`)
console.log(`   ${path.join(fixturesDir, 'config-dto-mock.json')}`)
console.log(`   ${path.join(fixturesDir, 'i18n-dto-mock.json')}`)
console.log(`   ${path.join(fixturesDir, 'test-data.json')}\n`)
