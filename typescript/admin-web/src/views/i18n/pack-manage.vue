<template>
  <div class="i18n-pack-container">
    <el-card class="box-card">
      <div slot="header" class="clearfix">
        <span>语言包管理</span>
      </div>

      <el-alert
        title="发布流程：选择语言包 → 选择目标环境 → 点击发布 → SDK 自动失效缓存并拉取新值"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom:16px"
      />

      <el-form ref="publishForm" :model="publishForm" :rules="publishRules" label-width="100px" size="medium">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="语言包 ID" prop="packID">
              <el-input-number v-model.number="publishForm.packID" :min="1" controls-position="right" style="width:100%" data-id="ia-input-pack-id" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="语言码" prop="langCode">
              <el-select v-model="publishForm.langCode" filterable allow-create placeholder="输入或选择" style="width:100%" data-id="ia-select-yuyanma-pack">
                <el-option label="简体中文" value="zh-CN" />
                <el-option label="繁體中文" value="zh-TW" />
                <el-option label="English" value="en-US" />
                <el-option label="日本語" value="ja-JP" />
                <el-option label="한국어" value="ko-KR" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="目标环境" prop="env">
              <el-select v-model="publishForm.env" style="width:100%" data-id="ia-select-huanjing-pack">
                <el-option label="开发 (dev)" value="dev" />
                <el-option label="测试 (test)" value="test" />
                <el-option label="预发 (staging)" value="staging" />
                <el-option label="生产 (prod)" value="prod" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item>
              <el-button
                v-permisaction="['i18n:pack:write']"
                type="primary"
                icon="el-icon-upload2"
                :loading="publishLoading"
                data-id="ia-btn-fabu-yueyanbao"
                @click="handlePublish"
              >发布语言包</el-button>
            </el-form-item>
          </el-col>
          <el-col :span="12" style="text-align:right">
            <el-form-item label="回滚到版本">
              <el-input-number
                v-model.number="rollbackVersion"
                :min="0"
                controls-position="right"
                size="small"
                style="width:120px;margin-right:8px"
                data-id="ia-input-huinban-banhao"
              />
              <el-button
                v-permisaction="['i18n:pack:write']"
                type="warning"
                size="small"
                icon="el-icon-refresh-left"
                :disabled="rollbackVersion <= 0"
                :loading="rollbackLoading"
                data-id="ia-btn-queren-huinban"
                @click="handleRollback"
              >回滚</el-button>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <!-- 发布结果 -->
      <el-card v-if="lastPublishResult" class="result-card" style="margin-top:16px" data-id="ia-card-fabu-jieguo-pack">
        <div slot="header"><span>最近发布结果</span></div>
        <el-descriptions :column="3" border size="medium">
          <el-descriptions-item label="Pack ID">{{ lastPublishResult.pack_id }}</el-descriptions-item>
          <el-descriptions-item label="语言码">{{ lastPublishResult.lang_code }}</el-descriptions-item>
          <el-descriptions-item label="版本号">
            <el-tag type="success" size="mini">v{{ lastPublishResult.version }}</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>
    </el-card>
  </div>
</template>

<script>
import { publishPack, rollbackPack } from '@/api/admin/i18n-pack'

export default {
  name: 'I18nPackManage',
  data() {
    return {
      publishLoading: false,
      rollbackLoading: false,
      lastPublishResult: null,
      rollbackVersion: 0,
      publishForm: {
        packID: undefined,
        langCode: '',
        env: 'dev'
      },
      publishRules: {
        packID: [{ required: true, message: '请输入语言包 ID', trigger: 'blur' }],
        langCode: [{ required: true, message: '请输入或选择语言码', trigger: 'change' }],
        env: [{ required: true, message: '请选择目标环境', trigger: 'change' }]
      }
    }
  },
  methods: {
    handlePublish() {
      this.$refs.publishForm.validate(valid => {
        if (!valid) return
        this.publishLoading = true
        publishPack({
          pack_id: this.publishForm.packID,
          lang_code: this.publishForm.langCode,
          env: this.publishForm.env
        }).then(res => {
          this.publishLoading = false
          if (res.code === 200) {
            this.$message.success('语言包发布成功')
            this.lastPublishResult = res.data || {}
          } else {
            this.$message.error(res.msg || '发布失败')
          }
        }).catch(() => {
          this.publishLoading = false
        })
      })
    },
    handleRollback() {
      if (this.rollbackVersion <= 0) {
        this.$message.warning('请输入有效的目标版本号')
        return
      }
      if (!this.publishForm.packID) {
        this.$message.warning('请先填写语言包 ID')
        return
      }
      this.$confirm(
        '确认将语言包回滚到版本 v' + this.rollbackVersion + '？此操作不可逆。',
        '回滚确认',
        { confirmButtonText: '确定回滚', cancelButtonText: '取消', type: 'warning' }
      ).then(() => {
        this.rollbackLoading = true
        return rollbackPack({
          pack_id: this.publishForm.packID,
          target_version: this.rollbackVersion
        })
      }).then(res => {
        this.rollbackLoading = false
        if (res.code === 200) {
          this.$message.success('回滚成功')
        } else {
          this.$message.error(res.msg || '回滚失败')
        }
      }).catch(() => {
        this.rollbackLoading = false
      })
    }
  }
}
</script>

<style scoped>
.i18n-pack-container { padding: 16px; }
.result-card { margin-top: 16px; }
</style>
