# Kubemanage 部署记录

## 当前状态

- 部署日期：2026-07-12
- Namespace：`kubemanage`
- 目标节点：`node234`（`10.90.1.234`）
- 校园网内部验收地址：<http://10.90.1.234:30240/>
- 校外浏览器入口：通过 SSH 本地端口转发后访问 <http://127.0.0.1:30240/>
- MySQL 数据卷：`nfs-csi`，10 Gi
- 当前版本：以 `kubectl get deployment -n kubemanage` 显示的镜像 tag 为准；MySQL `5.7`

所有正式 Pod 均固定调度到 `node234`。

## 浏览器访问

`10.90.1.234` 是校园网内部地址，校外浏览器不能直接访问。请在自己的电脑上通过当前跳板链路建立 SSH 隧道：

```bash
ssh -N -L 30240:10.90.1.234:30240 node191
```

上面的 `node191` 应是本机 SSH 配置中已经包含跳板机配置的别名。未配置别名时使用：

```bash
ssh -N -L 30240:10.90.1.234:30240 \
  -J <jump-user>@<jump-host> <node191-user>@10.90.1.191
```

保持 SSH 会话运行，然后浏览器访问：

```text
http://127.0.0.1:30240/
```

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

## 通用版本更新流程

以下流程适用于后端、前端同时发布的新版本。示例版本为 `v1.1.0`，以后必须使用未使用过的新 tag，不能覆盖旧 tag；清单使用 `IfNotPresent`，复用旧 tag 可能继续运行旧镜像。

### 1. 发布前检查和数据库备份

```bash
git pull origin master
git status --short
go test ./...                         # 在 kubemanage-another 目录执行
npm run build                         # 在 kubemanage-web 目录执行

kubectl config current-context
kubectl get nodes -o wide
kubectl get pods,pvc,svc -n kubemanage -o wide

BACKUP=kubemanage-$(date +%F-%H%M).sql
kubectl exec -n kubemanage kubemanage-mysql-0 -- \
  sh -c 'exec mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" \
  --single-transaction --routines --triggers kubemanage' \
  > "$BACKUP"
test -s "$BACKUP"
sha256sum "$BACKUP"
```

本项目后端使用 `Recreate` 策略，后端会有短暂中断。新版本首次启动还可能执行数据库字段迁移和 `rbac-catalog-v2` 权限目录迁移，备份不能省略。

### 2. 构建镜像

```bash
VERSION=v1.1.0

cd kubemanage-another
env GOCACHE=/tmp/kubemanage-go-cache CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags='-s -w' -o kubemanage-server ./cmd/main.go
docker build -f Dockerfile.runtime \
  -t docker.io/library/kubemanage-backend:$VERSION .
rm -f kubemanage-server

cd ../kubemanage-web
npm run build
docker build -t docker.io/library/kubemanage-web:$VERSION .
rm -rf dist
```

如果本地构建出的镜像显示为短名称，先补齐清单中的完整 tag：

```bash
docker tag kubemanage-backend:$VERSION docker.io/library/kubemanage-backend:$VERSION
docker tag kubemanage-web:$VERSION docker.io/library/kubemanage-web:$VERSION
```

### 3. 导入 node234 containerd

node234 当前不能稳定从公网 Registry 拉取镜像，必须先导入镜像：

```bash
docker save -o /tmp/kubemanage-$VERSION.tar \
  docker.io/library/kubemanage-backend:$VERSION \
  docker.io/library/kubemanage-web:$VERSION
```

若可以 SSH 到 node234，直接执行：

```bash
scp /tmp/kubemanage-$VERSION.tar node234:/tmp/
ssh node234 sudo ctr -n k8s.io images import /tmp/kubemanage-$VERSION.tar
```

若 SSH 不可用，可使用 node191 临时 HTTP 服务和固定到 node234 的一次性特权导入 Pod；导入完成后必须删除临时服务和 Pod，且不要把特权导入清单长期提交到仓库。

导入后确认两个完整镜像 tag 存在，再继续更新 Deployment。

### 4. 更新清单并先发布后端

将 `k8s/04-backend.yaml` 和 `k8s/05-web.yaml` 的镜像 tag 更新为 `$VERSION`。后端环境变量必须保留实际访问来源，例如：

```yaml
- name: KUBEMANAGE_ALLOWED_ORIGINS
  value: "http://10.90.1.234:30240,http://127.0.0.1:30240"
```

先更新后端并确认迁移完成：

```bash
kubectl apply -f k8s/04-backend.yaml
kubectl rollout status deployment/kubemanage-backend -n kubemanage --timeout=300s
kubectl logs -n kubemanage deployment/kubemanage-backend --tail=300

kubectl exec -n kubemanage kubemanage-mysql-0 -- \
  sh -c 'mysql -ukubemanage -p"$MYSQL_PASSWORD" kubemanage \
  -e "SELECT name, applied_at FROM sys_data_migrations \
  WHERE name='\''rbac-catalog-v2'\'';"'
```

首次 RBAC 迁移会重建内置角色 `111/222/2221` 的默认直授权限；自定义角色会保留，自定义菜单不会自动授予普通用户。

### 5. 发布前端并验收

```bash
kubectl apply -f k8s/05-web.yaml
kubectl rollout status deployment/kubemanage-web -n kubemanage --timeout=300s

kubectl get pods,svc,pvc -n kubemanage -o wide
kubectl get deployment -n kubemanage \
  -o custom-columns='NAME:.metadata.name,IMAGE:.spec.template.spec.containers[*].image,READY:.status.readyReplicas'
```

至少验证：页面 HTTP 200、默认账号登录 API、管理员菜单/API、普通用户和子角色权限、Kubernetes 节点列表、WebShell，以及 MySQL PVC 为 `Bound`。

### 6. 回滚

镜像兼容时可以回滚 Deployment：

```bash
kubectl rollout undo deployment/kubemanage-web -n kubemanage
kubectl rollout undo deployment/kubemanage-backend -n kubemanage
```

但 Deployment 回滚不会撤销 `AutoMigrate` 或 RBAC 数据迁移。若旧版本无法兼容新数据库，必须停止后端、使用升级前数据库备份恢复，再应用旧版本镜像和清单。Secret 不提交 Git，恢复数据库时不要覆盖现有 Secret。
