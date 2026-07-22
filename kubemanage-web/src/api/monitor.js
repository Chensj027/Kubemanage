import request from '@/utils/request'

// 换取一次性票据（后端据 Bearer JWT 校验并按角色映射）
export const grafanaTicket = () => request({ url: '/api/monitor/grafana/ticket', method: 'post' })

// enterGrafana 完成 SSO 交接并在新标签打开 Grafana。
// 先同步开空白标签（避免异步后 window.open 被弹窗拦截），再取票据设置其地址。
export async function enterGrafana() {
  const tab = window.open('', '_blank')
  try {
    const res = await grafanaTicket()
    const ticket = res && res.data && res.data.ticket
    if (!ticket) throw new Error('未获取到票据')
    const url = `/grafana/sso?ticket=${encodeURIComponent(ticket)}`
    if (tab) tab.location.href = url
    else window.location.href = url
  } catch (e) {
    if (tab) tab.close()
    // 具体错误（如“权限不足”）已由 request 响应拦截器弹出
  }
}
