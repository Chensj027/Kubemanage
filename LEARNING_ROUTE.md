# Kubemanage 学习路线

目标：通过当前项目系统学习 Go Web 后端、Kubernetes client-go、Kubernetes 资源模型、RBAC 权限体系和前后端分离项目结构。

## 1. 先跑通项目，建立反馈闭环

先不要急着横向读完整个仓库，第一步是能把项目跑起来，并能观察一个请求从前端到后端再到 Kubernetes API 的完整路径。

重点确认：

- 后端如何启动 Gin 服务
- 配置文件如何加载
- MySQL 如何初始化
- Kubernetes client 如何初始化
- 前端如何通过 `/api` 调后端

建议先追踪一个具体动作：

```text
浏览器点击 Deployment 列表
-> 前端 API 函数发请求
-> 后端路由接收请求
-> controller 绑定和校验参数
-> service 调用 client-go
-> Kubernetes API 返回资源对象
-> 后端统一响应
-> 前端渲染表格
```

## 2. 吃透 Go 后端启动骨架

从后端入口开始读，不要直接跳进业务逻辑。

重点文件：

- `kubemanage-another/cmd/main.go`
- `kubemanage-another/cmd/app/server.go`
- `kubemanage-another/cmd/app/options/options.go`
- `kubemanage-another/cmd/app/config/config.go`
- `kubemanage-another/cmd/app/config/viper.go`
- `kubemanage-another/router/router.go`
- `kubemanage-another/middleware/middleware.go`

这一阶段重点掌握：

- Go package 组织方式
- Cobra 命令启动
- Gin engine 和 route group
- middleware 链式执行
- Viper 配置加载
- GORM 初始化
- 全局对象和依赖注册方式
- Go 中 interface 和 struct method 的用法

## 3. 以 Deployment 模块作为第一条主线

不要同时看 Pod、Service、Ingress、ConfigMap。先完整吃透 Deployment 模块，再类比其它资源。

前端重点文件：

- `kubemanage-web/src/views/deployment/Deployment.vue`
- `kubemanage-web/src/api/deployment.js`
- `kubemanage-web/src/utils/request.js`

后端重点文件：

- `kubemanage-another/controller/kubeController/kubeRouter.go`
- `kubemanage-another/controller/kubeController/deployment.go`
- `kubemanage-another/dto/kubeDto/deployment.go`
- `kubemanage-another/pkg/core/kubemanage/v1/kube/deployment.go`

建议按接口逐个追踪：

- Deployment 列表
- Deployment 详情
- 创建 Deployment
- 更新 Deployment
- 扩缩容 Deployment
- 重启 Deployment
- 删除 Deployment

每追完一个接口，都写出它的调用链：

```text
前端函数 -> URL -> 后端路由 -> controller 方法 -> DTO -> service 方法 -> client-go 方法
```

## 4. 学 Kubernetes client-go 和资源模型

Deployment 模块跑通后，重点转向 Kubernetes API 和 client-go。

重点文件：

- `kubemanage-another/pkg/core/kubemanage/v1/kube/init.go`
- `kubemanage-another/pkg/core/kubemanage/v1/kube/deployment.go`
- `kubemanage-another/pkg/core/kubemanage/v1/kube/pod.go`
- `kubemanage-another/pkg/core/kubemanage/v1/kube/service.go`
- `kubemanage-another/pkg/core/kubemanage/v1/kube/ingress.go`

重点理解这些 client-go 调用背后的资源含义：

```go
ClientSet.AppsV1().Deployments(namespace).List(...)
ClientSet.AppsV1().Deployments(namespace).Create(...)
ClientSet.CoreV1().Pods(namespace).Get(...)
ClientSet.CoreV1().Pods(namespace).GetLogs(...)
ClientSet.CoreV1().Services(namespace).List(...)
```

需要掌握的 Kubernetes 概念：

- Group / Version / Resource
- Namespace 级资源和集群级资源
- ObjectMeta / Spec / Status
- Deployment / ReplicaSet / Pod 的关系
- Service / Endpoint / Pod label selector 的关系
- InClusterConfig 和 kubeconfig 的区别
- ServiceAccount 和 Kubernetes RBAC

## 5. 再学习系统管理模块

Kubernetes 资源链路清楚后，再看用户、菜单、权限、审计这些系统模块。

重点文件和目录：

- `kubemanage-another/dao/factory.go`
- `kubemanage-another/dao/model/`
- `kubemanage-another/pkg/source/`
- `kubemanage-another/pkg/core/kubemanage/v1/system.go`
- `kubemanage-another/pkg/core/kubemanage/v1/sys/`
- `kubemanage-another/middleware/jwt.go`
- `kubemanage-another/middleware/casbin_rbac.go`
- `kubemanage-another/middleware/operation.go`

这一阶段重点掌握：

- GORM model 和 DAO 分层
- 数据库自动建表和初始化数据
- 用户登录与 JWT
- Casbin 权限校验
- 菜单权限和接口权限
- 操作审计如何记录

## 6. 建议练习任务

按小任务推进，比单纯阅读更容易吃透。

1. 给 Deployment 列表接口增加一个返回字段，例如 `availableReplicas`。
2. 给 Service 或 ConfigMap 列表增加一个筛选条件。
3. 新增一个只读接口：统计每个 Namespace 下 Pod 数量。
4. 给某个 controller 补充更严格的参数校验。
5. 给某个 Kubernetes service 方法拆出更容易测试的小函数。
6. 为某个接口补一个最小单元测试或集成测试。
7. 尝试把后端部署到 Kubernetes 集群内，使用 ServiceAccount 和 InClusterConfig。
8. 尝试给后端 Kubernetes 访问权限收敛 RBAC，而不是直接使用过大的 ClusterRole。

## 7. 吃透项目的判断标准

如果能做到下面几件事，说明这个项目已经基本吃透：

- 看到一个前端按钮，能追到后端具体的 client-go 调用。
- 想新增一种 Kubernetes 资源管理能力时，知道需要改哪些前端 API、路由、页面、后端 controller、DTO 和 service。
- 能解释哪些数据来自 Kubernetes API，哪些数据来自 MySQL。
- 能解释 JWT、Casbin、菜单权限和接口权限之间的关系。
- 能把后端部署进 Kubernetes，并说明为什么集群内应使用 ServiceAccount 和 RBAC。
- 能独立排查一个接口问题：前端请求、后端路由、参数绑定、权限中间件、Kubernetes API 错误、MySQL 错误。

## 8. 推荐阅读顺序总结

```text
README / DEPLOYMENT
-> 后端启动链路
-> 前端路由和 Axios
-> Deployment 完整请求链路
-> client-go 和 Kubernetes 资源模型
-> Pod / Service / Ingress / ConfigMap 等资源模块
-> 用户 / 权限 / 审计 / 菜单系统
-> 部署到 Kubernetes
-> 自己新增功能并重构小模块
```

