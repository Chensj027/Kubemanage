import { createRouter, createWebHashHistory } from 'vue-router'
import HomeView from "@/views/home/HomeView";
import NProgress from 'nprogress';
import 'nprogress/nprogress.css';
import Layout from '../layout/Layout'
import { getUserInfo } from '@/api/system'
import { authState, clearAuth, hasMenuPath, setAuthInfo } from '@/utils/auth'

export const routes = [
    {
        path: '/system',
        name: '系统管理',
        component: Layout,
        icon: 'Setting',
        meta: {title: '系统管理', requireAuth: true},
        children: [
            { path: '/system/users', name: '用户管理', icon: 'User', meta: {title: '用户管理', requireAuth: true, menuAliases: ['/user']}, component: () => import('@/views/system/Users.vue') },
            { path: '/system/roles', name: '角色权限', icon: 'Lock', meta: {title: '角色权限', requireAuth: true, menuAliases: ['/authority']}, component: () => import('@/views/system/Roles.vue') }
        ]
    },
    {
        path: '/404',
        component: () => import('@/views/common/404.vue'),
        meta: {title:"404" ,requireAuth:false}
    },
    {
        path: '/:pathMatch(.*)',
        redirect: '404'
    },
    {
        path: '/',
        redirect: '/home' //重定向
    },
    {
        path: '/login',  //url路径
        component: () => import('@/views/login/Login.vue'),  //视图组件
        icon: "odometer",  //图标
        meta: {title: "登录", requireAuth: false},  //meta元信息
    },
    {
        path: '/home',
        component: Layout,
        meta: {title:"概要" ,requireAuth:true},
        children: [
            {
                path: '/home',
                name: '概要',
                icon: 'Help',
                meta: {title: '概要', requireAuth: true},
                component: HomeView,
            }
        ]
    },
    {
        path: '/workflow',
        component: Layout,
        icon: "VideoPlay",
        children: [
            {
                path: "/workflow",
                name: "工作流",
                icon: "VideoPlay",
                meta: {title: "工作流", requireAuth: true},
                component: () => import('@/views/workflow/Workflow.vue')
            }
        ]
    },
    {
        path: '/workload',
        name: '工作负载',
        component:Layout,
        icon: 'menu',
        meta: {title:"工作负载" ,requireAuth:true},
        children: [
            {
                path: '/workload/deployment',
                name: 'Deployment',
                icon: 'el-icon-s-data',
                meta: {title:"Deployment" ,requireAuth:true},
                component: () => import('@/views/deployment/Deployment'),
            },
            {
                path: '/workload/pod',
                name: 'Pod',
                icon: 'el-icon-document-add',
                meta: {title:"Deployment" ,requireAuth:true},
                component: () => import('@/views/pod/Pod'),
            },
            {
                path: "/workload/daemonset",
                name: "DaemonSet",
                icon: "el-icon-document-add",
                meta: {title: "DaemonSet", requireAuth: true},
                component: () => import("@/views/daemonset/DaemonSet.vue")
            },
            {
                path: "/workload/deamonset",
                redirect: "/workload/daemonset",
                meta: {requireAuth: true, hidden: true, permissionPath: '/workload/daemonset'}
            },
            {
                path: "/workload/statefulset",
                name: "StatefulSet",
                icon: "el-icon-document-add",
                meta: {title: "StatefulSets", requireAuth: true},
                component: () => import("@/views/statefulset/StatefulSet.vue")
            }
        ]
    },
    {
        path: "/loadbalance",
        name: "负载均衡",
        component: Layout,
        icon: "files",
        meta: {title: "负载均衡", requireAuth: true},
        children: [
            {
                path: "/loadbalance/service",
                name: "Service",
                icon: "el-icon-s-data",
                meta: {title: "Service", requireAuth: true},
                component: () => import("@/views/service/Service.vue")
            },
            {
                path: "/loadbalance/ingress",
                name: "Ingress",
                icon: "el-icon-document-add",
                meta: {title: "Ingress", requireAuth: true},
                component: () => import("@/views/ingress/Ingress.vue")
            }
        ]
    },
    {
        path: "/storage",
        name: "存储与配置",
        component: Layout,
        icon: "tickets",
        meta: {title: "存储与配置", requireAuth: true},
        children: [
            {
                path: "/storage/configmap",
                name: "Configmap",
                icon: "el-icon-document-add",
                meta: {title: "Configmap", requireAuth: true},
                component: () => import("@/views/configmap/ConfigMap.vue")
            },
            {
                path: "/storage/secret",
                name: "Secret",
                icon: "el-icon-document-add",
                meta: {title: "Secret", requireAuth: true},
                component: () => import("@/views/secret/Secret.vue")
            },
            {
                path: "/storage/persistentvolumeclaim",
                name: "PVC",
                icon: "el-icon-s-data",
                meta: {title: "PersistentVolumeClaim", requireAuth: true},
                component: () => import("@/views/persistentvolumeclaim/PersistentVolumeClaim.vue")
            }
        ]
    },
    {
        path: "/cluster",
        name: "集群",
        component: Layout,
        icon: "Cpu",
        meta: {title: "集群", requireAuth: true},
        children: [
            {
                path: "/cluster/node",
                name: "Node",
                icon: "el-icon-ship",
                meta: {title: "Node", requireAuth: true},
                component: () => import("@/views/node/Node.vue")
            },
            {
                path: "/cluster/namespace",
                name: "Namespace",
                icon: "el-icon-ship",
                meta: {title: "Namespace", requireAuth: true},
                component: () => import("@/views/namespace/Namespace.vue")
            },
            {
                path: "/cluster/persistentvolume",
                name: "PersistentVolume",
                icon: "el-icon-ship",
                meta: {title: "PersistemtVolume", requireAuth: true},
                component: () => import("@/views/persistentvolume/PersistentVolume.vue")
            }
        ]
    },
    {
        path: "/monitor",
        name: "监控",
        component: Layout,
        icon: "Odometer",
        meta: {title: "监控", requireAuth: true},
        children: [
            {
                path: "/monitor/grafana",
                name: "Grafana",
                icon: "Odometer",
                meta: {title: "Grafana", requireAuth: true, external: true},
                component: () => import("@/views/monitor/GrafanaEntry.vue")
            }
        ]
    }
]


// 进度条配置
NProgress.inc(0.2)
//全局进度条的配置
NProgress.configure({
    easing: 'ease', // 动画方式
    speed: 1000, // 递增进度条的速度
    showSpinner: false, // 是否显示加载ico
    trickleSpeed: 200, // 自动递增间隔
    minimum: 0.3, // 更改启动时使用的最小百分比
    parent: 'body', //指定进度条的父容器
})
const router = createRouter({
    history: createWebHashHistory(),
    routes
})

router.beforeEach(async (to,from,next) => {
    // 启动进度条
    NProgress.start()
    // 设置头部
    if (to.meta.title) {
        document.title = to.meta.title
    }else {
        document.title = "Kubernetes"
    }
    const token = getToken()
    if (!token) return to.path === '/login' ? next() : next('/login')
    if (to.path === '/login') return next('/home')
    if (!authState.loaded) {
        try {
            const res = await getUserInfo()
            setAuthInfo(res.data || {})
        } catch (_) {
            clearAuth()
            return next('/login')
        }
    }
    if (!isRouteAllowed(to)) return next('/404')
    next()
})


const getToken = () => {
    return localStorage.getItem('token')
}

export function isRouteAllowed(route) {
    const meta = route.meta || {}
    if (!meta.requireAuth) return true
    return hasMenuPath(meta.permissionPath || route.path, meta.menuAliases || [])
}

export function visibleRoutes() {
    return routes.filter(route => route.children).map(route => ({
        ...route,
        children: (route.children || []).filter(child => !child.meta?.hidden && isRouteAllowed({ ...child, meta: child.meta || {} }))
    })).filter(route => route.children.length)
}


router.afterEach(() => {
    NProgress.done()
})


export default router
