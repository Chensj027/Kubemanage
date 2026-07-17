# Kubemanage

Kubemanage 是一个 Kubernetes 管理平台完整项目仓库，包含前端和后端两部分：

- `kubemanage-web/`: Vue 3 前端
- `kubemanage-another/`: Go + Gin 后端

本仓库已整理为单一 monorepo，前端和后端不再作为独立 Git 子仓库维护。

## 功能概览

- Kubernetes 资源管理
- Namespace、Node、Deployment、Pod、Service、Ingress 等资源查看与操作
- Pod 日志查看
- Pod WebShell
- RBAC 权限管理
- 用户、菜单、API、操作审计等管理功能

## 技术栈

前端：

- Vue 3
- Vue Router
- Element Plus
- Axios
- Xterm.js
- ECharts

后端：

- Go
- Gin
- GORM
- MySQL
- client-go
- Casbin

## 项目结构

```text
.
├── DEPLOYMENT.md          # 正式部署文档
├── README.md              # 项目总说明
├── kubemanage-another/    # 后端项目
└── kubemanage-web/        # 前端项目
```

## 本地开发

### 后端

后端默认读取 `kubemanage-another/config.yaml`。该文件包含本地环境配置，已被根目录 `.gitignore` 忽略，不会提交到仓库。

首次使用可以从示例配置复制：

```bash
cp kubemanage-another/config.example.yaml kubemanage-another/config.yaml
```

然后按本地 MySQL 和 Kubernetes 环境修改：

```bash
cd kubemanage-another
go run cmd/main.go
```

默认后端地址：

```text
http://127.0.0.1:6180/
```

### 前端

```bash
cd kubemanage-web
npm install
npm run serve -- --host 0.0.0.0 --port 5240
```

默认前端开发地址：

```text
http://127.0.0.1:5240/
```

如果在局域网访问，请使用当前机器的局域网 IP。

## 生产构建

前端：

```bash
cd kubemanage-web
npm run build
```

后端：

```bash
cd kubemanage-another
go build -o app ./cmd/main.go
```

## 正式部署

正式部署建议使用 Kubernetes Deployment + Service + PVC，并将前端构建为 Nginx 静态服务。

当前规划部署节点：

```text
node234 / 10.90.1.234
```

详细部署流程见：

[DEPLOYMENT.md](./DEPLOYMENT.md)

用户、角色、菜单和 API 权限的实现与接口说明见：

[AUTHORIZATION.md](./AUTHORIZATION.md)

## 配置与安全

- 不要提交真实数据库密码、token、kubeconfig 或其它密钥。
- `kubemanage-another/config.yaml` 是本地配置文件，默认被忽略。
- 仓库中提供 `kubemanage-another/config.example.yaml` 作为模板。
- 正式部署时建议使用 Kubernetes Secret 管理后端配置。
- 后端部署到集群内时优先使用 ServiceAccount 和 RBAC，不建议挂载 master 节点 kubeconfig。

## 默认账号

根据原项目说明，初始化后的默认账号为：

```text
admin / kubemanage
```

实际账号以当前数据库初始化结果为准。

## 说明

本仓库整合自前端与后端项目，后续开发、提交和发布都以当前根目录 Git 仓库为准。
