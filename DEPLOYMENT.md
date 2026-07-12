# Kubemanage 部署记录

## 当前状态

- 部署日期：2026-07-12
- Namespace：`kubemanage`
- 目标节点：`node234`（`10.90.1.234`）
- 访问地址：<http://10.90.1.234:30240/>
- MySQL 数据卷：`nfs-csi`，10 Gi
- 当前版本：后端/前端 `v1.0.0`，MySQL `5.7`

所有正式 Pod 均固定调度到 `node234`。

## 部署文件

```text
kubemanage-web/Dockerfile
kubemanage-web/nginx.conf
kubemanage-another/Dockerfile.runtime
k8s/00-namespace.yaml
k8s/01-backend-rbac.yaml
k8s/02-backend-config.yaml
k8s/03-mysql.yaml
k8s/04-backend.yaml
k8s/05-web.yaml
```

Secret 不保存在 Git 中：

- `kubemanage-secrets`：MySQL root 密码、应用密码和 JWT 密钥

## 首次部署

### 1. 构建镜像

后端：

```bash
cd kubemanage-another
go test ./...
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags='-s -w' -o kubemanage-server ./cmd/main.go
docker build -f Dockerfile.runtime -t docker.io/library/kubemanage-backend:v1.0.0 .
rm -f kubemanage-server
```

前端：

```bash
cd kubemanage-web
npm run build
docker build -t docker.io/library/kubemanage-web:v1.0.0 .
```

MySQL 使用：

```bash
docker tag mysql:5.7 docker.io/library/kubemanage-mysql:5.7
```

### 2. 将镜像导入 node234

当前 node234 从公网 Registry 拉取镜像层会返回 405，因此正式 Pod 使用 `imagePullPolicy: IfNotPresent`，镜像需预先导入 node234 的 containerd：

```bash
docker save -o /tmp/kubemanage-images.tar \
  docker.io/library/kubemanage-backend:v1.0.0 \
  docker.io/library/kubemanage-web:v1.0.0 \
  docker.io/library/kubemanage-mysql:5.7

# 将 tar 传到 node234 后执行
sudo ctr -n k8s.io images import /tmp/kubemanage-images.tar
```

清单中的镜像名必须与导入后的镜像 tag 完全一致。本次实际部署通过 node191 临时 HTTP 服务和一次性特权 Pod 完成导入，临时资源已删除。

### 3. 创建 Secret

```bash
kubectl apply -f k8s/00-namespace.yaml

kubectl -n kubemanage create secret generic kubemanage-secrets \
  --from-literal=mysql-root-password="$(openssl rand -hex 24)" \
  --from-literal=mysql-password="$(openssl rand -hex 24)" \
  --from-literal=jwt-secret="$(openssl rand -hex 32)"
```

### 4. 应用资源

```bash
kubectl apply -f k8s/
```

## 验收

```bash
kubectl get pods -n kubemanage -o wide
kubectl get pvc,svc -n kubemanage
kubectl logs -n kubemanage deploy/kubemanage-backend
```

验收结果：

- 前端页面 HTTP 200
- 默认账号登录 API 返回 200
- Kubernetes 节点列表 API 返回 200，并读取到 6 个节点
- WebShell WebSocket 返回 `101 Switching Protocols`
- MySQL PVC 为 `Bound`
- MySQL、后端、前端均为 `Running/Ready`

初始登录账号来自项目内置数据：

```text
admin / kubemanage
```

首次登录后必须立即修改默认密码。

## 发布新版本

1. 使用新 tag 构建镜像，例如 `v1.0.1`。
2. 将新镜像导入 node234 containerd。
3. 更新 `k8s/04-backend.yaml`、`k8s/05-web.yaml` 中的 tag。
4. 执行：

```bash
kubectl apply -f k8s/04-backend.yaml
kubectl apply -f k8s/05-web.yaml
kubectl rollout status deployment/kubemanage-backend -n kubemanage
kubectl rollout status deployment/kubemanage-web -n kubemanage
```

回滚：

```bash
kubectl rollout undo deployment/kubemanage-backend -n kubemanage
kubectl rollout undo deployment/kubemanage-web -n kubemanage
```

数据库变更不能依赖 Deployment 回滚，升级前应单独备份 PVC 数据。
