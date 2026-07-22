<template>
  <div class="grafana-entry">
    <el-card class="entry-card" shadow="hover">
      <div class="entry-icon">
        <el-icon :size="48"><Odometer /></el-icon>
      </div>
      <h2 class="entry-title">监控面板 · Grafana</h2>
      <p class="entry-desc">正在为你进入监控面板…</p>
      <el-button type="primary" size="large" @click="go">进入 Grafana 监控</el-button>
      <p class="entry-hint">若浏览器拦截了新标签页，请点击上方按钮。</p>
    </el-card>
  </div>
</template>

<script>
// 该页面仅作为深链兜底（直接访问 /#/monitor/grafana 时）；
// 正常入口在侧边菜单点击时已直接走 SSO 交接，不经过本页。
import { enterGrafana } from '@/api/monitor'

export default {
  name: 'GrafanaEntry',
  methods: {
    go() {
      enterGrafana()
    }
  },
  mounted() {
    enterGrafana()
  }
}
</script>

<style scoped>
.grafana-entry {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 8vh;
}
.entry-card {
  max-width: 480px;
  text-align: center;
  padding: 16px 8px;
}
.entry-icon {
  color: #20a0ff;
  margin-bottom: 8px;
}
.entry-title {
  margin: 8px 0 12px;
  font-size: 20px;
  color: #303133;
}
.entry-desc {
  color: #606266;
  font-size: 14px;
  line-height: 1.7;
  margin-bottom: 20px;
}
.entry-hint {
  margin-top: 16px;
  font-size: 12px;
  color: #909399;
}
</style>
