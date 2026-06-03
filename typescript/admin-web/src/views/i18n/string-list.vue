<template>
  <div class="i18n-string-container">
    <el-card class="box-card">
      <div slot="header" class="clearfix">
        <span>语言字符串管理</span>
      </div>

      <el-form ref="queryForm" :model="queryParams" :inline="true" size="small">
        <el-form-item label="语言包 ID" prop="packID">
          <el-input-number
            v-model="queryParams.packID"
            :min="1"
            controls-position="right"
            placeholder="请输入 Pack ID"
            style="width:180px"
            data-id="ia-input-yuyanbao-id"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="el-icon-search" data-id="ia-btn-chaxun-string" @click="handleQuery">查询</el-button>
          <el-button icon="el-icon-refresh" data-id="ia-btn-chongzhi-string" @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <el-row :gutter="10" class="mb8">
        <el-col :span="1.5">
          <el-button
            v-permisaction="['i18n:string:add']"
            type="primary"
            icon="el-icon-plus"
            size="mini"
            :disabled="!queryParams.packID"
            data-id="ia-btn-xinzeng-string"
            @click="handleAdd"
          >新增字符串</el-button>
        </el-col>
        <el-col :span="1.5">
          <el-button
            v-permisaction="['i18n:string:delete']"
            type="danger"
            icon="el-icon-delete"
            size="mini"
            :disabled="!currentRow"
            data-id="ia-btn-shanchu-string-toolbar"
            @click="handleDelete"
          >删除</el-button>
        </el-col>
      </el-row>

      <el-table
        v-loading="loading"
        :data="stringList"
        border
        stripe
        highlight-current-row
        data-id="ia-table-string-list"
        @current-change="handleCurrentChange"
      >
        <el-table-column label="ID" prop="id" width="70" align="center" />
        <el-table-column label="String Key" prop="stringKey" min-width="160" show-overflow-tooltip />
        <el-table-column label="字符串值" prop="stringValue" min-width="200" show-overflow-tooltip>
          <template slot-scope="scope">
            <el-popover trigger="hover" placement="top" width="300" data-id="ia-popover-string-preview">
              <p style="word-break:break-all">{{ scope.row.stringValue }}</p>
              <div slot="reference" class="name-wrapper">
                {{ scope.row.stringValue | truncate(40) }}
              </div>
            </el-popover>
          </template>
        </el-table-column>
        <el-table-column label="分组" prop="groupName" width="100" align="center" />
        <el-table-column label="模板类型" prop="templateType" width="90" align="center">
          <template slot-scope="scope">
            <el-tag :type="templateTagType(scope.row.templateType)" size="mini">
              {{ scope.row.templateType || 'plain' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="版本" prop="version" width="60" align="center" />
        <el-table-column label="操作" width="120" align="center" fixed="right">
          <template slot-scope="scope">
            <el-button
              v-permisaction="['i18n:string:edit']"
              type="text"
              size="mini"
              icon="el-icon-edit"
              data-id="ia-btn-bianji-string-row"
              @click="handleEdit(scope.row)"
            >编辑</el-button>
            <el-button
              v-permisaction="['i18n:string:delete']"
              type="text"
              size="mini"
              icon="el-icon-delete"
              class="text-danger"
              data-id="ia-btn-shanchu-string-row"
              @click="handleDelete(scope.row)"
            >删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 新增/编辑对话框 -->
      <el-dialog :title="dialogTitle" :visible.sync="dialogVisible" width="600px" :close-on-click-modal="false">
        <el-form ref="stringForm" :model="form" :rules="formRules" label-width="100px">
          <el-form-item label="语言包 ID" prop="packID">
            <el-input-number v-model.number="form.packID" :min="1" disabled controls-position="right" style="width:100%" />
          </el-form-item>
          <el-form-item label="String Key" prop="stringKey">
            <el-input v-model="form.stringKey" placeholder="如：greeting.hello、error.notFound" :disabled="isEdit" data-id="ia-input-string-key-dialog" />
          </el-form-item>
          <el-form-item label="字符串值" prop="stringValue">
            <el-input
              v-model="form.stringValue"
              type="textarea"
              :rows="3"
              placeholder="请输入翻译文本"
              data-id="ia-textarea-string-value-dialog"
            />
          </el-form-item>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="模板类型" prop="templateType">
                <el-select v-model="form.templateType" placeholder="选择类型" style="width:100%" data-id="ia-select-moban-leixing-dialog">
                  <el-option label="纯文本 (plain)" value="plain" />
                  <el-option label="命名参数 (named)" value="named" />
                  <el-option label="ICU 格式 (icu)" value="icu" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="分组" prop="groupName">
                <el-input v-model="form.groupName" placeholder="如：common、error" data-id="ia-input-fenzu-dialog" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item v-if="form.templateType === 'named'" label="参数 Schema">
            <el-input
              v-model="form.paramsSchema"
              type="textarea"
              :rows="2"
              placeholder="如：{&quot;name&quot;:&quot;string&quot;,&quot;count&quot;:&quot;number&quot;}"
              data-id="ia-textarea-canshu-schema-dialog"
            />
          </el-form-item>
          <el-form-item v-if="form.templateType !== 'plain'" label="预览示例">
            <el-input
              v-model="form.previewSample"
              type="textarea"
              :rows="2"
              placeholder="如：Hello {name}, you have {count} messages"
              data-id="ia-textarea-yulan-shili-dialog"
            />
          </el-form-item>
        </el-form>
        <div slot="footer">
          <el-button data-id="ia-btn-quxiao-string-dialog" @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitLoading" data-id="ia-btn-queding-string-dialog" @click="handleSubmit">确定</el-button>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
import { listStrings, createString, updateString, deleteString } from '@/api/admin/i18n-string'

const defaultForm = () => ({
  id: undefined,
  packID: undefined,
  stringKey: '',
  stringValue: '',
  groupName: '',
  templateType: 'plain',
  paramsSchema: '',
  previewSample: ''
})

export default {
  name: 'I18nStringList',
  filters: {
    truncate(len) {
      return function(val) {
        if (!val) return ''
        return val.length > len ? val.substring(0, len) + '...' : val
      }
    }
  },
  data() {
    return {
      loading: false,
      stringList: [],
      currentRow: null,
      queryParams: {
        packID: undefined
      },
      dialogVisible: false,
      dialogTitle: '新增字符串',
      isEdit: false,
      submitLoading: false,
      form: defaultForm(),
      formRules: {
        packID: [{ required: true, message: '请先选择语言包', trigger: 'blur' }],
        stringKey: [{ required: true, message: 'String Key 不能为空', trigger: 'blur' }],
        stringValue: [{ required: true, message: '字符串值不能为空', trigger: 'blur' }]
      }
    }
  },
  methods: {
    getList() {
      if (!this.queryParams.packID || this.queryParams.packID <= 0) {
        this.stringList = []
        return
      }
      this.loading = true
      listStrings(this.queryParams.packID).then(res => {
        this.stringList = res.data || []
        this.loading = false
      }).catch(() => {
        this.loading = false
      })
    },
    handleQuery() {
      this.getList()
    },
    resetQuery() {
      this.queryParams.packID = undefined
      this.stringList = []
    },
    handleCurrentChange(row) {
      this.currentRow = row
    },
    handleAdd() {
      this.form = defaultForm()
      this.form.packID = this.queryParams.packID
      this.isEdit = false
      this.dialogTitle = '新增字符串'
      this.dialogVisible = true
      this.$nextTick(() => {
        this.$refs.stringForm && this.$refs.stringForm.clearValidate()
      })
    },
    handleEdit(row) {
      this.form = { ...defaultForm(), ...row }
      this.isEdit = true
      this.dialogTitle = '编辑字符串'
      this.dialogVisible = true
    },
    handleDelete(row) {
      const targetId = row ? row.id : (this.currentRow ? this.currentRow.id : null)
      if (!targetId) {
        this.$message.warning('请先选择要删除的记录')
        return
      }
      this.$confirm('确认删除该字符串？', '警告', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        return deleteString(targetId)
      }).then(res => {
        if (res.code === 200) {
          this.$message.success('删除成功')
          this.getList()
        } else {
          this.$message.error(res.msg || '删除失败')
        }
      }).catch(() => {})
    },
    handleSubmit() {
      this.$refs.stringForm.validate(valid => {
        if (!valid) return
        this.submitLoading = true
        const data = {
          pack_id: this.form.packID,
          string_key: this.form.stringKey,
          string_value: this.form.stringValue,
          group_name: this.form.groupName,
          template_type: this.form.templateType,
          params_schema: this.form.paramsSchema,
          preview_sample: this.form.previewSample
        }
        if (this.isEdit) data.id = this.form.id

        const action = this.isEdit ? updateString(data) : createString(data)
        action.then(res => {
          this.submitLoading = false
          if (res.code === 200) {
            this.$message.success(res.msg || (this.isEdit ? '更新成功' : '创建成功'))
            this.dialogVisible = false
            this.getList()
          } else if (res.code === 10400) {
            this.showTemplateError(res)
          } else {
            this.$message.error(res.msg || '操作失败')
          }
        }).catch(() => {
          this.submitLoading = false
        })
      })
    },
    showTemplateError(res) {
      const errors = res.errors || []
      const msgs = errors.length > 0
        ? errors.map(e => e.reason).join('\n')
        : (res.message || '模板校验失败')
      this.$message({ message: msgs, type: 'error', duration: 5000 })
    },
    templateTagType(type) {
      const map = { plain: '', named: 'success', icu: 'warning' }
      return map[type] || ''
    }
  }
}
</script>

<style scoped>
.i18n-string-container { padding: 16px; }
.mb8 { margin-bottom: 8px; }
.text-danger { color: #f56c6c; }
.name-wrapper { cursor: pointer; }
</style>
