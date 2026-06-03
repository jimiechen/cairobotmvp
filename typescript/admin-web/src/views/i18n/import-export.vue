<template>
  <div class="i18n-import-export-container">
    <el-card class="box-card">
      <div slot="header" class="clearfix">
        <span>CSV 导入导出</span>
      </div>

      <el-alert
        title="CSV 格式要求：首行为表头（string_key, string_value, group_name, template_type），后续为数据行。单次导入上限 5MB。"
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom:16px"
      />

      <el-row :gutter="24">
        <!-- 导入区域 -->
        <el-col :span="12">
          <el-card shadow="never" class="section-card">
            <div slot="header">
              <i class="el-icon-upload2" /> CSV 导入
            </div>
            <el-form ref="importForm" :model="importForm" :rules="importRules" label-width="90px" size="medium">
              <el-form-item label="目标 Pack ID" prop="packID">
                <el-input-number v-model.number="importForm.packID" :min="1" controls-position="right" style="width:100%" data-id="ia-import-input-pack-id" />
              </el-form-item>
              <el-form-item label="选择文件">
                <el-upload
                  ref="uploadRef"
                  :auto-upload="false"
                  :limit="1"
                  :on-change="onFileChange"
                  :on-remove="onFileRemove"
                  accept=".csv"
                  drag
                  action=""
                  data-id="ia-upload-csv-wenjian"
                >
                  <i class="el-icon-upload" />
                  <div class="el-upload__text">将 CSV 文件拖到此处，或<em>点击上传</em></div>
                  <div slot="tip" class="el-upload__tip">只能上传 .csv 文件，且不超过 5MB</div>
                </el-upload>
              </el-form-item>
              <el-form-item>
                <el-button
                  v-permisaction="['i18n:string:write']"
                  type="primary"
                  icon="el-icon-upload2"
                  :loading="importLoading"
                  :disabled="!uploadFile"
                  data-id="ia-btn-kaishi-daoru"
                  @click="handleImport"
                >开始导入</el-button>
                <el-button icon="el-icon-refresh-left" data-id="ia-btn-chongzhi-daoru" @click="resetImport">重置</el-button>
              </el-form-item>
            </el-form>

            <!-- 导入结果 -->
            <el-divider v-if="importResult" content-position="left">导入结果</el-divider>
            <div v-if="importResult" class="import-result">
              <el-result
                :icon="importResult.fail_count > 0 ? 'warning' : 'success'"
                :title="importResult.fail_count > 0 ? '部分成功' : '全部成功'"
                data-id="ia-result-daoru-jieguo"
              >
                <template slot="subTitle">
                  <p>总行数：<strong>{{ importResult.total_rows }}</strong></p>
                  <p>成功：<strong style="color:#67c23a">{{ importResult.success_count }}</strong></p>
                  <p>失败：<strong :style="importResult.fail_count > 0 ? 'color:#f56c6c' : ''">{{ importResult.fail_count }}</strong></p>
                </template>
                <template slot="extra">
                  <el-table
                    v-if="importResult.errors && importResult.errors.length > 0"
                    :data="importResult.errors"
                    border
                    size="mini"
                    max-height="200"
                    data-id="ia-table-daoru-cuowu"
                  >
                    <el-table-column label="行号" prop="row_num" width="70" align="center" />
                    <el-table-column label="原因" prop="reason" />
                  </el-table>
                </template>
              </el-result>
            </div>
          </el-card>
        </el-col>

        <!-- 导出区域 -->
        <el-col :span="12">
          <el-card shadow="never" class="section-card">
            <div slot="header">
              <i class="el-icon-download" /> CSV 导出
            </div>
            <el-form ref="exportForm" :model="exportForm" :rules="exportRules" label-width="90px" size="medium">
              <el-form-item label="源 Pack ID" prop="packID">
                <el-input-number v-model.number="exportForm.packID" :min="1" controls-position="right" style="width:100%" data-id="ia-export-input-pack-id" />
              </el-form-item>
              <el-form-item>
                <el-button
                  v-permisaction="['i18n:string:read']"
                  type="success"
                  icon="el-icon-download"
                  :loading="exportLoading"
                  data-id="ia-btn-daochu-csv"
                  @click="handleExport"
                >导出 CSV</el-button>
              </el-form-item>
            </el-form>

            <el-divider content-position="left">导出说明</el-divider>
            <ul class="export-tips">
              <li>导出文件名为 strings_{packID}.csv</li>
              <li>包含表头：string_key, string_value, group_name, template_type</li>
              <li>编码格式：UTF-8 with BOM（Excel 兼容）</li>
              <li>可直接修改后重新导入</li>
            </ul>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script>
import { importCSV, exportCSV } from '@/api/admin/i18n-import-export'

export default {
  name: 'I18nImportExport',
  data() {
    return {
      uploadFile: null,
      importLoading: false,
      exportLoading: false,
      importResult: null,
      importForm: {
        packID: undefined
      },
      importRules: {
        packID: [{ required: true, message: '请输入目标 Pack ID', trigger: 'blur' }]
      },
      exportForm: {
        packID: undefined
      },
      exportRules: {
        packID: [{ required: true, message: '请输入源 Pack ID', trigger: 'blur' }]
      }
    }
  },
  methods: {
    onFileChange(file) {
      const isCSV = file.raw.type === 'text/csv' ||
        file.raw.name.endsWith('.csv') ||
        file.raw.type === 'application/vnd.ms-excel'
      const isLt5M = file.size / 1024 / 1024 < 5
      if (!isCSV) {
        this.$message.error('只能上传 CSV 格式文件')
        this.uploadFile = null
        return
      }
      if (!isLt5M) {
        this.$message.error('文件大小不能超过 5MB')
        this.uploadFile = null
        return
      }
      this.uploadFile = file.raw
    },
    onFileRemove() {
      this.uploadFile = null
    },
    handleImport() {
      this.$refs.importForm.validate(valid => {
        if (!valid) return
        if (!this.uploadFile) {
          this.$message.warning('请先选择 CSV 文件')
          return
        }
        this.importLoading = true
        this.importResult = null
        importCSV(this.uploadFile, this.importForm.packID).then(res => {
          this.importLoading = false
          if (res.code === 200 || res.code === 10400) {
            this.importResult = {
              total_rows: res.total_rows || 0,
              success_count: res.success_count || 0,
              fail_count: res.fail_count || 0,
              errors: res.errors || []
            }
            if (res.code === 10400) {
              this.$message.warning('存在部分行校验未通过，请查看详情')
            } else {
              this.$message.success('导入完成')
            }
          } else {
            this.$message.error(res.msg || '导入失败')
          }
        }).catch(() => {
          this.importLoading = false
        })
      })
    },
    handleExport() {
      this.$refs.exportForm.validate(valid => {
        if (!valid) return
        this.exportLoading = true
        exportCSV(this.exportForm.packID).then(res => {
          this.exportLoading = false
          const blob = new Blob([res], { type: 'text/csv;charset=utf-8;' })
          const link = document.createElement('a')
          link.href = URL.createObjectURL(blob)
          link.download = 'strings_' + this.exportForm.packID + '.csv'
          document.body.appendChild(link)
          link.click()
          document.body.removeChild(link)
          URL.revokeObjectURL(link.href)
          this.$message.success('导出成功')
        }).catch(() => {
          this.exportLoading = false
          this.$message.error('导出失败')
        })
      })
    },
    resetImport() {
      this.importForm.packID = undefined
      this.uploadFile = null
      this.importResult = null
      if (this.$refs.uploadRef) {
        this.$refs.uploadRef.clearFiles()
      }
      this.$refs.importForm && this.$refs.importForm.clearValidate()
    }
  }
}
</script>

<style scoped>
.i18n-import-export-container { padding: 16px; }
.section-card { margin-bottom: 0; }
.import-result { margin-top: 12px; }
.export-tips {
  margin: 0;
  padding-left: 20px;
  color: #909399;
  font-size: 13px;
  line-height: 2;
}
</style>
