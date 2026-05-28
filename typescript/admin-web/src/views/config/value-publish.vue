<template>
  <div class="config-value-container">
    <el-card class="box-card">
      <div slot="header" class="clearfix">
        <span>配置值发布</span>
        <el-button
          style="float:right"
          type="primary"
          size="mini"
          icon="el-icon-refresh"
          @click="loadSchemaList"
        >刷新 Schema</el-button>
      </div>

      <el-alert
        title="操作流程：选择模块 → 填写字段值 → 点击发布 → 生效到对应环境"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 16px"
      />

      <el-form ref="publishForm" :model="form" :rules="formRules" label-width="100px" size="medium">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="模块 Key" prop="moduleKey">
              <el-select
                v-model="form.moduleKey"
                placeholder="请选择或输入模块 Key"
                filterable
                allow-create
                style="width:100%"
                @change="onModuleChange"
              >
                <el-option
                  v-for="m in moduleOptions"
                  :key="m"
                  :label="m"
                  :value="m"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="环境" prop="env">
              <el-select v-model="form.env" placeholder="选择环境" style="width:100%">
                <el-option label="开发 (dev)" value="dev" />
                <el-option label="测试 (test)" value="test" />
                <el-option label="预发 (staging)" value="staging" />
                <el-option label="生产 (prod)" value="prod" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 动态字段表单：按 schema 渲染 -->
        <div v-if="currentSchemas.length > 0" class="dynamic-fields">
          <el-divider content-position="left">配置字段（{{ currentSchemas.length }} 个）</el-divider>
          <el-row :gutter="20">
            <el-col v-for="field in currentSchemas" :key="field.id" :span="12">
              <el-form-item
                :label="field.fieldKey"
                :prop="'fields.' + field.fieldKey"
                :required="field.required"
              >
                <template slot="label">
                  {{ field.fieldKey }}
                  <el-tooltip v-if="field.description" :content="field.description" placement="top">
                    <i class="el-icon-info text-muted" style="cursor:pointer;margin-left:4px" />
                  </el-tooltip>
                  <el-tag size="mini" :type="fieldTypeTagType(field.fieldType)" style="margin-left:4px">
                    {{ field.fieldType }}
                  </el-tag>
                </template>

                <!-- string / json -->
                <el-input
                  v-if="field.fieldType === 'string' || field.fieldType === 'json'"
                  v-model="fieldValues[field.fieldKey]"
                  :placeholder="'默认: ' + (field.defaultValue || '-')"
                  @blur="validateField(field)"
                />

                <!-- int -->
                <el-input-number
                  v-else-if="field.fieldType === 'int'"
                  v-model.number="fieldValues[field.fieldKey]"
                  controls-position="right"
                  style="width:100%"
                  :placeholder="'默认: ' + (field.defaultValue || '0')"
                  @change="validateField(field)"
                />

                <!-- float -->
                <el-input-number
                  v-else-if="field.fieldType === 'float'"
                  v-model.number="fieldValues[field.fieldKey]"
                  :precision="2"
                  :step="0.1"
                  controls-position="right"
                  style="width:100%"
                  @change="validateField(field)"
                />

                <!-- bool -->
                <el-switch
                  v-else-if="field.fieldType === 'bool'"
                  v-model="fieldBoolValues[field.fieldKey]"
                  active-text="开"
                  inactive-text="关"
                />

                <!-- fallback -->
                <el-input
                  v-else
                  v-model="fieldValues[field.fieldKey]"
                  :placeholder="'默认: ' + (field.defaultValue || '-')"
                />
              </el-form-item>

              <!-- 字段级校验错误展示（10400） -->
              <div v-if="fieldErrors[field.fieldKey]" class="field-error-tip">
                <i class="el-icon-warning" /> {{ fieldErrors[field.fieldKey] }}
              </div>
            </el-col>
          </el-row>
        </div>

        <el-empty v-else-if="form.moduleKey && !schemaLoading" description="该模块暂无 Schema 定义" />

        <el-row :gutter="20" style="margin-top:24px">
          <el-col :span="24">
            <el-form-item>
              <el-button
                v-permisaction="['config:value:publish']"
                type="primary"
                icon="el-icon-upload2"
                :loading="publishLoading"
                :disabled="currentSchemas.length === 0"
                @click="handlePublish"
              >发布配置</el-button>
              <el-button icon="el-icon-refresh-left" @click="resetForm">重置</el-button>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <!-- 10400 校验错误弹窗 -->
      <el-dialog
        title="校验失败详情"
        :visible.sync="errorDialogVisible"
        width="560px"
      >
        <el-alert
          :title="errorMessage"
          type="error"
          :closable="false"
          show-icon
          style="margin-bottom:16px"
        />
        <el-table :data="errorList" border size="mini" max-height="300">
          <el-table-column label="字段" prop="field" width="150" />
          <el-table-column label="错误信息" prop="message" />
        </el-table>
        <div slot="footer">
          <el-button type="primary" @click="errorDialogVisible = false">知道了</el-button>
        </div>
      </el-dialog>

      <!-- 发布历史 -->
      <el-card v-if="showHistory" class="history-card" style="margin-top:16px">
        <div slot="header"><span>发布历史</span></div>
        <el-table :data="versionList" border size="small" max-height="250">
          <el-table-column label="版本号" prop="version" width="80" align="center" />
          <el-table-column label="环境" prop="env" width="90" align="center">
            <template slot-scope="scope">
              <el-tag size="mini" :type="envTagType(scope.row.env)">{{ scope.row.env }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="发布时间" prop="publishedAt" width="170" />
          <el-table-column label="操作人" prop="operator" width="100" />
        </el-table>
      </el-card>
    </el-card>
  </div>
</template>

<script>
import { listSchema } from '@/api/admin/config-schema'
import { publishValue, getValueVersions } from '@/api/admin/config-value'

export default {
  name: 'ConfigValuePublish',
  data() {
    return {
      schemaLoading: false,
      publishLoading: false,
      showHistory: false,
      allSchemas: [],
      currentSchemas: [],
      moduleOptions: [],
      fieldValues: {},
      fieldBoolValues: {},
      fieldErrors: {},
      versionList: [],
      form: {
        moduleKey: '',
        env: 'dev'
      },
      formRules: {
        moduleKey: [{ required: true, message: '请选择或输入模块 Key', trigger: 'change' }],
        env: [{ required: true, message: '请选择环境', trigger: 'change' }]
      },
      errorDialogVisible: false,
      errorMessage: '',
      errorList: []
    }
  },
  created() {
    this.loadSchemaList()
  },
  methods: {
    loadSchemaList() {
      this.schemaLoading = true
      listSchema({}).then(res => {
        this.allSchemas = res.data || []
        const modSet = new Set()
        this.allSchemas.forEach(s => { if (s.moduleKey) modSet.add(s.moduleKey) })
        this.moduleOptions = Array.from(modSet).sort()
        this.schemaLoading = false
        if (this.form.moduleKey) {
          this.filterSchemas(this.form.moduleKey)
        }
      }).catch(() => {
        this.schemaLoading = false
      })
    },
    onModuleChange(val) {
      this.filterSchemas(val)
      this.showHistory = false
    },
    filterSchemas(moduleKey) {
      this.currentSchemas = this.allSchemas.filter(s => s.moduleKey === moduleKey)
      this.fieldValues = {}
      this.fieldBoolValues = {}
      this.fieldErrors = {}
      this.currentSchemas.forEach(f => {
        if (f.fieldType === 'bool') {
          this.$set(this.fieldBoolValues, f.fieldKey, f.defaultValue === 'true')
        } else if (f.fieldType === 'int') {
          this.$set(this.fieldValues, f.fieldKey, parseInt(f.defaultValue) || 0)
        } else if (f.fieldType === 'float') {
          this.$set(this.fieldValues, f.fieldKey, parseFloat(f.defaultValue) || 0)
        } else {
          this.$set(this.fieldValues, f.fieldKey, f.defaultValue || '')
        }
      })
    },
    validateField(field) {
      const val = this.fieldValues[field.fieldKey]
      this.$delete(this.fieldErrors, field.fieldKey)
      if (field.required && (val === '' || val == null)) {
        this.$set(this.fieldErrors, field.fieldKey, '该字段为必填项')
        return
      }
      if (field.validator) {
        try {
          const rules = typeof field.validator === 'string' ? JSON.parse(field.validator) : field.validator
          if (rules.min != null && Number(val) < Number(rules.min)) {
            this.$set(this.fieldErrors, field.fieldKey, '不能小于 ' + rules.min)
            return
          }
          if (rules.max != null && Number(val) > Number(rules.max)) {
            this.$set(this.fieldErrors, field.fieldKey, '不能大于 ' + rules.max)
            return
          }
        } catch (e) {
          void 0
        }
      }
    },
    buildFieldsPayload() {
      const fields = []
      this.currentSchemas.forEach(f => {
        let value
        if (f.fieldType === 'bool') {
          value = this.fieldBoolValues[f.fieldKey]
        } else {
          value = this.fieldValues[f.fieldKey]
        }
        fields.push({ field_key: f.fieldKey, value })
      })
      return fields
    },
    handlePublish() {
      this.$refs.publishForm.validate(valid => {
        if (!valid) return

        const hasError = Object.keys(this.fieldErrors).length > 0
        if (hasError) {
          this.$message.warning('存在校验未通过的字段，请修正后重新提交')
          return
        }

        this.publishLoading = true
        publishValue({
          module_key: this.form.moduleKey,
          env: this.form.env,
          fields: this.buildFieldsPayload()
        }).then(res => {
          this.publishLoading = false
          if (res.code === 200) {
            this.$message.success('配置发布成功')
            this.loadVersionHistory()
          } else if (res.code === 10400) {
            this.handle10400Response(res)
          } else {
            this.$message.error(res.msg || '发布失败')
          }
        }).catch(() => {
          this.publishLoading = false
        })
      })
    },
    handle10400Response(res) {
      this.errorMessage = res.message || '参数校验失败'
      this.errorList = res.errors || []
      if (this.errorList.length > 0) {
        this.errorList.forEach(e => {
          if (e.field) {
            this.$set(this.fieldErrors, e.field, e.reason || e.message)
          }
        })
      }
      this.errorDialogVisible = true
    },
    loadVersionHistory() {
      getValueVersions({
        module_key: this.form.moduleKey,
        env: this.form.env
      }).then(res => {
        this.versionList = res.data && res.data.versions ? res.data.versions : []
        this.showHistory = true
      }).catch(() => { void 0 })
    },
    resetForm() {
      this.form.moduleKey = ''
      this.form.env = 'dev'
      this.currentSchemas = []
      this.fieldValues = {}
      this.fieldBoolValues = {}
      this.fieldErrors = {}
      this.showHistory = false
      this.versionList = []
      this.$refs.publishForm && this.$refs.publishForm.clearValidate()
    },
    fieldTypeTagType(type) {
      const map = { string: '', int: 'success', float: 'warning', bool: 'danger', json: 'info' }
      return map[type] || ''
    },
    envTagType(env) {
      const map = { dev: 'info', test: '', staging: 'warning', prod: 'danger' }
      return map[env] || ''
    }
  }
}
</script>

<style scoped>
.config-value-container { padding: 16px; }
.dynamic-fields { margin-top: 8px; }
.field-error-tip {
  color: #f56c6c;
  font-size: 12px;
  line-height: 1.4;
  padding: 4px 0;
}
.text-muted { color: #909399; }
.history-card { margin-top: 16px; }
</style>
