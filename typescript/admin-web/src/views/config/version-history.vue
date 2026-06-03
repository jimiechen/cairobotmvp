<template>
  <div class="app-container">
    <el-card>
      <div slot="header" class="clearfix">
        <span>版本历史</span>
      </div>
      <el-alert title="配置版本发布记录将在此展示" type="info" :closable="false" show-icon />
      <el-table :data="[]" border style="margin-top: 16px">
        <el-table-column prop="version" label="版本号" width="100" />
        <el-table-column prop="moduleKey" label="模块" width="150" />
        <el-table-column prop="env" label="环境" width="80" />
        <el-table-column prop="fieldCount" label="字段数" width="80" />
        <el-table-column prop="operator" label="操作人" width="120" />
        <el-table-column prop="createdAt" label="发布时间" />
      </el-table>
    </el-card>
  </div>
</template>

<script>
export default {
  name: 'VersionHistory',
  data() {
    return {}
  }
}
</script>
