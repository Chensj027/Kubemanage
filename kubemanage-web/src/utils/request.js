import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from "@/router";
import { clearAuth } from '@/utils/auth'


// 新建axios对象
const  httpClient = axios.create({
    VUE_APP_BASE_API: '/api',
    timeout: 10000,
    validateStatus(status) {
        return status >= 200 && status < 504
    }
})

httpClient.defaults.retry = 1
httpClient.defaults.retryDelay = 1000
httpClient.defaults.shoudRetry = true // 是否重试

// 请求拦截器
httpClient.interceptors.request.use(
    config => {
        // 添加header
        config.headers["Content-Type"] = 'application/json;charset=UTF-8'
        config.headers["Accept-Language"] = 'zh-CN'
        const token = localStorage.getItem('token')
        if (token) config.headers.Authorization = `Bearer ${token}`
        // 处理post请求
        if (config.method === 'POST') {
            if (!config.data) {
                config.data = []
            }
        }
        return config
    },
    err => {
        return Promise.reject(err)
    }
)


// 响应拦截器
httpClient.interceptors.response.use(
    response => {
        const renewedToken = response.headers && response.headers['new-token']
        if (renewedToken) localStorage.setItem('token', renewedToken)
        const code = response.data && response.data.code
        if (response.status !== 200 || code !== 200 ) {
            if ([11002, 10103].includes(code) || response.status === 401) {
                ElMessage({
                    message: '登录已过期，请重新登陆',
                    type: 'warning',
                })
                clearAuth()
                router.push('/login')
                return Promise.reject(response.data)
            }
            if (code === 405 || response.status === 403) {
                ElMessage.warning((response.data && response.data.msg) || '权限不足，请联系管理员')
                return Promise.reject(response.data)
            }
            ElMessage.error((response.data && response.data.msg) || `请求失败 (${response.status})`)
            return Promise.reject(response.data)
        }else {
            return response.data
        }
    },
    err => {
        ElMessage.error(err.message || '网络异常')
        return Promise.reject(err)
    }
)


export default httpClient
