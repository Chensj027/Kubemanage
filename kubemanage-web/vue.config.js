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
      }
    }
  },
  transpileDependencies: true,
  // 关闭语法检测
  lintOnSave: false
})
