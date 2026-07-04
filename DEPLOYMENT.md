# Kubemanage 正式部署文档

本文档用于将当前项目正式部署到 Kubernetes 集群中的 `node234` 节点。

当前项目结构：

- `kubemanage-another/`: 后端，Go + Gin，默认监听 `:6180`
- `kubemanage-web/`: 前端，Vue CLI，正式部署时构建为静态文件

目标节点：

- 节点名：`node234`
- 节点 IP：`10.90.1.234`
- 调度标签：`kubernetes.io/hostname: node234`

## 部署目标

正式部署后，用户访问前端入口，由前端 Nginx 代理 `/api` 到后端服务：

```text
Browser -> kubemanage-web -> /api -> kubemanage-backend -> Kubernetes API / MySQL
```

推荐先用 NodePort 暴露：

```text
http://10.90.1.234:30240/
```

如果配置校园网 DNS 和 Ingress，可以使用域名：

```text
http://kubemanage.xxx.edu.cn/
```

## 不建议使用开发端口

当前开发模式使用：

```text
http://10.90.1.191:5240/
```

`5240` 是 Vue 开发服务器端口，只适合开发调试。正式部署不应使用 `npm run serve`，而应使用前端静态构建产物和 Nginx 容器。

## 部署前提

集群中已确认：

- `node234` 状态为 `Ready`
- 集群存在默认 StorageClass：`nfs-csi`
- 集群存在 IngressClass：`nginx`
- 后端代码已支持 `rest.InClusterConfig()`，部署进集群后可使用 ServiceAccount 访问 Kubernetes API

本地或 CI 环境需要：

- 能构建后端镜像
- 能构建前端镜像
- 能将镜像推送到集群节点可拉取的镜像仓库，或手动导入到 `node234` 的 containerd
- 有 `kubectl` 集群管理权限

## 推荐命名

Namespace：

```text
kubemanage
```

镜像示例：

```text
<registry>/kubemanage-backend:v1.0.0
<registry>/kubemanage-web:v1.0.0
```

如暂无镜像仓库，也可以先在 `node234` 本机导入镜像，并在 Deployment 中设置：

```yaml
imagePullPolicy: IfNotPresent
```

长期维护建议使用镜像仓库。

## 后端配置

后端正式部署时不要使用默认 MySQL `127.0.0.1`，应通过 ConfigMap 或 Secret 指向集群内 MySQL Service。

示例 `config.yaml`：

```yaml
default:
  listenAddr: ":6180"
  webSocketListenAddr: ""
  JWTSecret: "change-me"
  expireTime: 10

mysql:
  host: "kubemanage-mysql"
  port: "3306"
  user: "kubemanage"
  password: "" # 生产环境建议由 Secret 注入，不要明文写入仓库
  name: "kubemanage"
  maxOpenConns: 100
  maxLifetime: 20
  maxIdleConns: 10

log:
  level: "info"
  filename: "/tmp/kubemanage.log"
  max_size: 200
  max_age: 30
  max_backups: 7
```

更好的做法是让后端支持从环境变量读取数据库密码；如果暂时不改代码，可以将完整配置作为 Kubernetes Secret 挂载到 Pod 中。

## Kubernetes 权限

后端需要访问 Kubernetes API。正式部署时使用 ServiceAccount，不挂载 master 节点的 kubeconfig。

最小权限应按实际功能收敛。当前平台涉及 Namespace、Node、Deployment、Pod、Service、Ingress、ConfigMap、Secret、PV、PVC 等管理操作，初期可以使用较宽的 ClusterRole，稳定后再缩小权限。

示例：

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubemanage-backend
  namespace: kubemanage
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubemanage-backend
rules:
  - apiGroups: ["", "apps", "batch", "networking.k8s.io"]
    resources: ["*"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kubemanage-backend
subjects:
  - kind: ServiceAccount
    name: kubemanage-backend
    namespace: kubemanage
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubemanage-backend
```

## 固定部署到 node234

所有 Kubemanage 相关工作负载都固定调度到 `node234`：

```yaml
nodeSelector:
  kubernetes.io/hostname: node234
```

如果后续不希望 MySQL 也固定在 `node234`，可以只给前端和后端加这个 `nodeSelector`，MySQL 交给集群调度。

## MySQL

正式部署建议在集群中使用 StatefulSet + PVC。

如果已有稳定 MySQL，也可以直接使用外部 MySQL，但需要保证：

- 后端 Pod 能访问 MySQL 地址
- 数据库 `kubemanage` 已创建
- 账号权限足够
- 凭据通过 Secret 管理

示例 Service 名：

```text
kubemanage-mysql
```

后端配置中 `mysql.host` 填：

```text
kubemanage-mysql
```

## 前端 Nginx 代理

前端正式镜像应使用 Nginx 托管 `dist/`。Nginx 需要代理 `/api` 到后端，并支持 WebSocket upgrade。

示例 `nginx.conf`：

```nginx
server {
  listen 80;
  server_name _;

  root /usr/share/nginx/html;
  index index.html;

  location / {
    try_files $uri $uri/ /index.html;
  }

  location /api/ {
    proxy_pass http://kubemanage-backend:6180/api/;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
  }
}
```

前端 WebSocket 地址已经改为基于当前访问域名动态生成：

```js
const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
const k8sTerminalWs = wsProtocol + "//" + window.location.host + "/api/k8s/pod/webshell";
```

因此正式部署时只要 `/api` 代理正确，WebSocket 不需要单独写死 IP。

## 镜像构建

后端：

```bash
docker build -t <registry>/kubemanage-backend:v1.0.0 ./kubemanage-another
docker push <registry>/kubemanage-backend:v1.0.0
```

前端需要先准备生产 `Dockerfile` 和 `nginx.conf`。建议 Dockerfile：

```dockerfile
FROM node:20-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
```

构建：

```bash
docker build -t <registry>/kubemanage-web:v1.0.0 ./kubemanage-web
docker push <registry>/kubemanage-web:v1.0.0
```

## 后端 Deployment

示例：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubemanage-backend
  namespace: kubemanage
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubemanage-backend
  template:
    metadata:
      labels:
        app: kubemanage-backend
    spec:
      serviceAccountName: kubemanage-backend
      nodeSelector:
        kubernetes.io/hostname: node234
      containers:
        - name: backend
          image: <registry>/kubemanage-backend:v1.0.0
          imagePullPolicy: IfNotPresent
          args:
            - --configFile=/etc/kubemanage/config.yaml
          ports:
            - containerPort: 6180
          volumeMounts:
            - name: config
              mountPath: /etc/kubemanage
              readOnly: true
      volumes:
        - name: config
          secret:
            secretName: kubemanage-backend-config
```

后端 Service：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: kubemanage-backend
  namespace: kubemanage
spec:
  type: ClusterIP
  selector:
    app: kubemanage-backend
  ports:
    - name: http
      port: 6180
      targetPort: 6180
```

## 前端 Deployment 和 NodePort

前端 Deployment：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubemanage-web
  namespace: kubemanage
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubemanage-web
  template:
    metadata:
      labels:
        app: kubemanage-web
    spec:
      nodeSelector:
        kubernetes.io/hostname: node234
      containers:
        - name: web
          image: <registry>/kubemanage-web:v1.0.0
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 80
```

前端 NodePort Service：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: kubemanage-web
  namespace: kubemanage
spec:
  type: NodePort
  selector:
    app: kubemanage-web
  ports:
    - name: http
      port: 80
      targetPort: 80
      nodePort: 30240
```

访问：

```text
http://10.90.1.234:30240/
```

## 域名访问

如果希望不输入 IP，需要校园网 DNS 或本机 hosts 将域名解析到 `10.90.1.234`。

临时本机 hosts：

```text
10.90.1.234 kubemanage.local
```

访问：

```text
http://kubemanage.local:30240/
```

正式校园网域名：

```text
kubemanage.xxx.edu.cn -> 10.90.1.234
```

如果希望不带端口，使用 Ingress：

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kubemanage-web
  namespace: kubemanage
spec:
  ingressClassName: nginx
  rules:
    - host: kubemanage.xxx.edu.cn
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: kubemanage-web
                port:
                  number: 80
```

访问：

```text
http://kubemanage.xxx.edu.cn/
```

前提是 ingress-nginx 的入口已经暴露到校园网可访问的 `80/443`。如果没有，需要先调整 ingress-nginx 的 Service 或使用 hostNetwork/NodePort。

## 部署命令

建议将清单放到：

```text
k8s/
```

部署：

```bash
kubectl create namespace kubemanage
kubectl apply -f k8s/
```

查看状态：

```bash
kubectl -n kubemanage get pod -o wide
kubectl -n kubemanage get svc
kubectl -n kubemanage logs deploy/kubemanage-backend
kubectl -n kubemanage logs deploy/kubemanage-web
```

确认 Pod 是否在 `node234`：

```bash
kubectl -n kubemanage get pod -o wide
```

## 发布新版本

开发仍在本地或开发节点进行。正式环境只运行稳定镜像。

发布流程：

```bash
docker build -t <registry>/kubemanage-backend:v1.0.1 ./kubemanage-another
docker build -t <registry>/kubemanage-web:v1.0.1 ./kubemanage-web
docker push <registry>/kubemanage-backend:v1.0.1
docker push <registry>/kubemanage-web:v1.0.1

kubectl -n kubemanage set image deployment/kubemanage-backend backend=<registry>/kubemanage-backend:v1.0.1
kubectl -n kubemanage set image deployment/kubemanage-web web=<registry>/kubemanage-web:v1.0.1
```

查看滚动更新：

```bash
kubectl -n kubemanage rollout status deployment/kubemanage-backend
kubectl -n kubemanage rollout status deployment/kubemanage-web
```

## 回滚

查看发布历史：

```bash
kubectl -n kubemanage rollout history deployment/kubemanage-backend
kubectl -n kubemanage rollout history deployment/kubemanage-web
```

回滚：

```bash
kubectl -n kubemanage rollout undo deployment/kubemanage-backend
kubectl -n kubemanage rollout undo deployment/kubemanage-web
```

## 开发与生产分离

正式部署后仍可以继续开发：

- 生产环境：`node234` 上运行固定版本镜像
- 开发环境：当前目录继续改代码、运行 `npm run serve` 和 `go run cmd/main.go`
- 发布方式：新镜像 tag + `kubectl set image`

不要在生产 Pod 内直接改代码。

不要让生产 Pod 挂载当前源码目录。

生产数据库和开发数据库应尽量分开；如果必须共用，任何表结构变更前必须先备份。

## 需要补充的文件

正式落地前建议新增：

```text
kubemanage-web/Dockerfile
kubemanage-web/nginx.conf
k8s/namespace.yaml
k8s/mysql.yaml
k8s/backend-rbac.yaml
k8s/backend-config-secret.yaml
k8s/backend.yaml
k8s/web.yaml
k8s/ingress.yaml
```

其中 `backend-config-secret.yaml` 不应提交真实密码到公开仓库。
