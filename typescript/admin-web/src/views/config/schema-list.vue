<template>
  <div class="config-schema-container">
    <el-card class="box-card">
      <div slot="header" class="clearfix">
        <span>配置 Schema 管理</span>
      </div>

      <el-form ref="queryForm" :model="queryParams" :inline="true" size="small">
        <el-form-item label="模块Key" prop="moduleKey">
          <el-input
            v-model="queryParams.moduleKey"
            placeholder="请输入模块标识"
            clearable
            style="width: 200px"
            @keyup.enter.native="handleQuery"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="el-icon-search" @click="handleQuery">搜索</el-button>
          <el-button icon="el-icon-refresh" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <el-row :gutter="10" class="mb8">
        <el-col :span="1.5">
          <el-button
            v-permisaction="['config:schema:add']"
            type="primary"
            icon="el-icon-plus"
            size="mini"
            @click="handleAdd"
          >新增 Schema</el-button>
        </el-col>
        <el-col :span="1.5">
          <el-button
            v-permisaction="['config:schema:delete']"
            type="danger"
            icon="el-icon-delete"
            size="mini"
            :disabled="!currentRow"
            @click="handleDelete"
          >删除</el-button>
        </el-col>
      </el-row>

      <el-table
        v-loading="loading"
        :data="schemaList"
        border
        stripe
        highlight-current-row
        @current-change="handleCurrentChange"
      >
        <el-table-column label="ID" prop="id" width="70" align="center" />
        <el-table-column label="模块 Key" prop="moduleKey" min-width="140" show-overflow-tooltip />
        <el-table-column label="字段 Key" prop="fieldKey" min-width="140" show-overflow-tooltip />
        <el-table-column label="字段类型" prop="fieldType" width="100" align="center">
          <template slot-scope="scope">
            <el-tag :type="fieldTypeTagType(scope.row.fieldType)" size="mini">
              {{ scope.row.fieldType }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="必填" prop="required" width="70" align="center">
          <template slot-scope="scope">
            <i :class="scope.row.required ? 'el-icon-success text-success' : 'el-icon-remove text-muted'" />
          </template>
        </el-table-column>
        <el-table-column label="默认值" prop="defaultValue" min-width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="120" align="center" fixed="right">
          <template slot-scope="scope">
            <el-button
              v-permisaction="['config:schema:edit']"
              type="text"
              size="mini"
              icon="el-icon-edit"
              @click="handleEdit(scope.row)"
            >编辑</el-button>
            <el-button type="text" size="mini" icon="el-icon-delete" class="text-danger" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <pagination
        v-show="total > 0"
        :total="total"
        :page.sync="queryParams.pageNum"
        :limit.sync="queryParams.pageSize"
        @pagination="getList"
      />

      <!-- 新增/编辑对话框 -->
      <el-dialog :title="dialogTitle" :visible.sync="dialogVisible" width="560px" :close-on-click-modal="false">
        <el-form ref="schemaForm" :model="form" :rules="formRules" label-width="100px">
          <el-form-item label="模块 Key" prop="moduleKey">
            <el-input v-model="form.moduleKey" placeholder="如：app.server" :disabled="isEdit" />
          </el-form-item>
          <el-form-item label="字段 Key" prop="fieldKey">
            <el-input v-model="form.fieldKey" placeholder="如：timeout、port" :disabled="isEdit" />
          </el-form-item>
          <el-form-item label="字段类型" prop="fieldType">
            <el-select v-model="form.fieldType" placeholder="选择字段类型" @change="onFieldTypeChange">
              <el-option label="字符串 (string)" value="string" />
              <el-option label="整数 (int)" value="int" />
              <el-option label="浮点数 (float)" value="float" />
              <el-option label="布尔值 (bool)" value="bool" />
              <el-option label="JSON (json)" value="json" />
            </el-select>
          </el-form-item>
          <el-form-item label="是否必填" prop="required">
            <el-switch v-model="form.required" active-text="是" inactive-text="否" />
          </el-form-item>
          <el-form-item label="默认值" prop="defaultValue">
            <el-input
              v-if="form.fieldType === 'string' || form.fieldType === 'json'"
              v-model="form.defaultValue"
              :placeholder="form.fieldType === 'json' ? 'JSON 格式' : '请输入默认值'"
            />
            <el-input-number
              v-else-if="form.fieldType === 'int'"
              v-model.number="form.defaultValueInt"
              controls-position="right"
            />
            <el-input-number
              v-else-if="form.fieldType === 'float'"
              v-model.number="form.defaultValueFloat"
              :precision="2"
              :step="0.1"
              controls-position="right"
            />
            <el-switch
              v-else-if="form.fieldType === 'bool'"
              v-model="form.defaultValueBool"
              active-text="true"
              inactive-text="false"
            />
            <el-input v-else v-model="form.defaultValue" placeholder="请输入默认值" />
          </el-form-item>
          <el-form-item v-if="form.fieldType === 'int' || form.fieldType === 'float'" label="取值范围" prop="validator">
            <el-row :gutter="10">
              <el-col :span="11">
                <el-input v-model="form.validatorMin" placeholder="最小值" />
              </el-col>
              <el-col :span="2" style="text-align:center;line-height:32px">~</el-col>
              <el-col :span="11">
                <el-input v-model="form.validatorMax" placeholder="最大值" />
              </el-col>
            </el-row>
          </el-form-item>
          <el-form-item label="校验规则 JSON" prop="validator">
            <el-input
              v-model="form.validator"
              type="textarea"
              :rows="3"
              placeholder="如：{&quot;min&quot;:0,&quot;max&quot;:100} 或 {&quot;pattern&quot;:&quot;^https://&quot;}"
            />
          </el-form-item>
          <el-form-item label="描述" prop="description">
            <el-input v-model="form.description" type="textarea" :rows="2" placeholder="字段用途说明" />
          </el-form-item>
        </el-form>
        <div slot="footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
import { listSchema, createSchema, updateSchema, deleteSchema } from '@/api/admin/config-schema'

const defaultForm = () => ({
  id: undefined,
  moduleKey: '',
  fieldKey: '',
  fieldType: 'string',
  required: false,
  defaultValue: '',
  defaultValueInt: 0,
  defaultValueFloat: 0.0,
  defaultValueBool: false,
  validator: '',
  validatorMin: '',
  validatorMax: '',
  description: ''
})

export default {
  name: 'ConfigSchemaList',
  data() {
    return {
      loading: false,
      schemaList: [],
      total: 0,
      currentRow: null,
      queryParams: {
        moduleKey: '',
        pageNum: 1,
        pageSize: 10
      },
      dialogVisible: false,
      dialogTitle: '新增 Schema',
      isEdit: false,
      submitLoading: false,
      form: defaultForm(),
      formRules: {
        moduleKey: [{ required: true, message: '模块 Key 不能为空', trigger: 'blur' }],
        fieldKey: [{ required: true, message: '字段 Key 不能为空', trigger: 'blur' }],
        fieldType: [{ required: true, message: '请选择字段类型', trigger: 'change' }]
      }
    }
  },
  created() {
    this.getList()
  },
  methods: {
    getList() {
      this.loading = true
      listSchema(this.queryParams).then(res => {
        this.schemaList = res.data || []
        this.total = Array.isArray(res.data) ? res.data.length : 0
        this.loading = false
      }).catch(() => {
        this.loading = false
      })
    },
    handleQuery() {
      this.queryParams.pageNum = 1
      this.getList()
    },
    resetQuery() {
      this.queryParams.moduleKey = ''
      this.handleQuery()
    },
    handleCurrentChange(row) {
      this.currentRow = row
    },
    handleAdd() {
      this.form = defaultForm()
      this.isEdit = false
      this.dialogTitle = '新增 Schema'
      this.dialogVisible = true
      this.$nextTick(() => {
        this.$refs.schemaForm && this.$refs.schemaForm.clearValidate()
      })
    },
    handleEdit(row) {
      this.form = { ...defaultForm(), ...row }
      if (row.fieldType === 'int') {
        this.form.defaultValueInt = parseInt(row.defaultValue) || 0
      } else if (row.fieldType === 'float') {
        this.form.defaultValueFloat = parseFloat(row.defaultValue) || 0
      } else if (row.fieldType === 'bool') {
        this.form.defaultValueBool = row.defaultValue === 'true' || row.defaultValue === true
      }
      this.isEdit = true
      this.dialogTitle = '编辑 Schema'
      this.dialogVisible = true
    },
    handleDelete(row) {
      const targetId = row ? row.id : (this.currentRow ? this.currentRow.id : null)
      if (!targetId) {
        this.$message.warning('请先选择要删除的记录')
        return
      }
      this.$confirm('确认删除该 Schema？删除后不可恢复', '警告', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        return deleteSchema(targetId)
      }).then(res => {
        if (res.code === 200) {
          this.$message.success('删除成功')
          this.getList()
        } else {
          this.$message.error(res.msg || '删除失败')
        }
      }).catch(() => {})
    },
    onFieldTypeChange(val) {
      this.form.defaultValue = ''
    },
    buildSubmitData() {
      const data = {
        moduleKey: this.form.moduleKey,
        fieldKey: this.form.fieldKey,
        fieldType: this.form.fieldType,
        required: this.form.required,
        validator: this.form.validator,
        description: this.form.description
      }
      switch (this.form.fieldType) {
        case 'int':
          data.defaultValue = String(this.form.defaultValueInt || 0)
          break
        case 'float':
          data.defaultValue = String(this.form.defaultValueFloat || 0)
          break
        case 'bool':
          data.defaultValue = String(this.form.defaultValueBool)
          break
        default:
          data.defaultValue = this.form.defaultValue
      }
      if (!this.isEdit && data.validator === '' && (this.form.fieldType === 'int' || this.form.fieldType === 'float')) {
        const parts = []
        if (this.form.validatorMin !== '') parts.push('"min":' + this.form.validatorMin)
        if (this.form.validatorMax !== '') parts.push('"max":' + this.form.validatorMax)
        if (parts.length > 0) data.validator = '{' + parts.join(',') + '}'
      }
      if (this.isEdit) {
        data.id = this.form.id
      }
      return data
    },
    handleSubmit() {
      this.$refs.schemaForm.validate(valid => {
        if (!valid) return
        this.submitLoading = true
        const data = this.buildSubmitData()
        const action = this.isEdit ? updateSchema(data) : createSchema(data)
        action.then(res => {
          this.submitLoading = false
          if (res.code === 200) {
            this.$message.success(res.msg || (this.isEdit ? '更新成功' : '创建成功'))
            this.dialogVisible = false
            this.getList()
          } else if (res.code === 10400) {
            this.showValidationError(res)
          } else {
            this.$message.error(res.msg || '操作失败')
          }
        }).catch(() => {
          this.submitLoading = false
        })
      })
    },
    showValidationError(res) {
      const errors = res.errors || []
      if (errors.length > 0) {
        const msgs = errors.map(e => e.field + ': ' + e.message).join('\n')
        this.$message({ message: msgs, type: 'error', duration: 5000 })
      } else {
        this.$message.error(res.message || '参数校验失败')
      }
    },
    fieldTypeTagType(type) {
      const map = { string: '', int: 'success', float: 'warning', bool: 'danger', json: 'info' }
      return map[type] || ''
    }
  }
}
</script>

<style scoped>
.config-schema-container { padding: 16px; }
.mb8 { margin-bottom: 8px; }
.text-success { color: #67c23a; }
.text-danger { color: #f56c6c; }
.text-muted { color: #909399; }
</style>
