const { defineConfig } = require('@vue/cli-service')
const backendTarget = process.env.KUBEMANAGE_DEV_BACKEND || 'http://127.0.0.1:6180/'

module.exports = defineConfig({
  devServer: {
    host: '127.0.0.1',
    port: 5240,
    open: false,
    proxy: {
      '/api': {
        target: backendTarget,
        ws: true,//代理websocked
        changeOrigin: true,//虚拟的站点需要更管origin
        // 浏览器只与 Vue 开发服务器进行同源通信；转发到后端后属于服务端代理，
        // 不应继续携带来自 5240 端口的浏览器来源请求头，否则会被生产环境的跨域校验拒绝。
        onProxyReq(proxyReq) {
          proxyReq.removeHeader('origin')
        },
        onProxyReqWs(proxyReq) {
          proxyReq.removeHeader('origin')
        },
        // pathRewrite: {
        //   '^/api': ''//重写路径
        // }
      },
      // Grafana 监控：与生产一致，转发到后端；由后端做 SSO 会话校验 + 反代 Grafana。
      // 前提：后端 grafana.upstream 指向对 dev 主机可达的 Grafana，例如
      //   KUBEMANAGE_GRAFANA_UPSTREAM=http://10.90.1.234:30300
      '/grafana': {
        target: backendTarget,
        ws: true,
        changeOrigin: true,
      }
    }
  },
  transpileDependencies: true,
  // 关闭语法检测
  lintOnSave: false
})
