# 用户、角色与权限

Kubemanage 使用 JWT + Casbin 实现身份认证和 RBAC，并将菜单权限与 API 权限分开保存。

## 请求链路

1. `POST /api/user/login` 校验 bcrypt 密码并签发 24 小时 JWT。
2. HTTP 请求使用 `Authorization: Bearer <token>`。
3. `JWTAuth` 校验签名、用户 UUID、冻结状态、当前角色和 `token_version`。
4. `CasbinHandler` 按 `角色 ID + Gin 路由模板 + HTTP 方法` 校验接口权限。
5. `GET /api/user/getinfo` 返回用户、有效菜单树和有效 API 权限，前端据此过滤路由、侧栏和操作按钮。

浏览器 WebShell 使用 WebSocket 子协议传递令牌：

```text
Sec-WebSocket-Protocol: kubemanage, <token>
```

WebShell 不再匿名开放，并且默认只接受同源 Origin。确需跨域时，通过环境变量显式配置：

```bash
KUBEMANAGE_ALLOWED_ORIGINS=https://console.example.com,https://ops.example.com
```

## 权限模型

- `sys_users.authority_id`：用户当前生效角色。
- `sys_authorities.parent_id`：父角色；子角色继承父角色的菜单和 API 权限。
- `sys_authority_menus`：角色直接拥有的菜单。
- `casbin_rule`：角色直接拥有的 API 策略及角色继承关系。
- `sys_apis`：权限配置页面使用的 API 清单。

### 内置角色定位

| 角色 ID | 名称 | 默认权限与用途 |
| --- | --- | --- |
| `111` | 管理员 | 平台超级管理员，菜单和 API 始终全部生效；权限页全选只读，不能被取消。内置管理员用户不能被冻结、删除或降权。 |
| `222` | 普通用户 | 业务使用基线，默认拥有除“系统管理”以外的菜单，以及安全的 Kubernetes 只读 API。WebShell 和扩缩容虽然使用 GET，但不会授予。 |
| `2221` | 普通用户子角色 | 示例扩展角色，默认不直授菜单或 API，通过父角色 `222` 继承业务基线；需要差异权限时只追加自身直授权限。 |

自定义角色也可设置父角色。权限页编辑的是当前角色的“直授”菜单和 API；父角色权限在运行时合并生效，但不会显示成子角色的直授勾选项，也不会在保存时复制到子角色。

角色菜单和 API 策略是两套独立配置：显示菜单不会自动授予接口权限，保存角色时应分别配置。

这里的 RBAC 控制平台功能权限。后端访问 Kubernetes 的资源范围仍由部署时的 ServiceAccount/ClusterRole 决定；如果需要“不同平台用户只能看到指定 Namespace”，应另建 Namespace 数据范围模型，不能复用角色继承关系冒充数据隔离。

## 用户管理 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/user/list` | 分页查询用户 |
| POST | `/api/user` | 创建用户 |
| PUT | `/api/user/:id` | 编辑用户资料和角色 |
| PUT | `/api/user/:id/enable` | 启用或冻结用户 |
| PUT | `/api/user/:id/set_auth` | 单独设置角色 |
| DELETE | `/api/user/:id/delete_user` | 删除用户 |
| POST | `/api/user/:id/change_pwd` | 用户修改自己的密码 |
| PUT | `/api/user/:id/reset_pwd` | 管理员设置新密码，body 为 `{"new_pwd":"..."}` |

角色、密码、启停状态发生变化，或者用户退出、被删除后，旧 JWT 会立即失效。

## 角色和授权 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/authority/getAuthorityList` | 角色列表 |
| POST | `/api/authority` | 创建角色 |
| PUT | `/api/authority/:authorityId` | 编辑角色和父角色 |
| DELETE | `/api/authority/:authorityId` | 删除未被使用的角色 |
| GET | `/api/menu/getBaseMenuTree` | 全量菜单树 |
| GET | `/api/menu/:authID/getMenuByAuthID` | 角色有效菜单树（包含继承） |
| GET | `/api/menu/:authID/getMenuByAuthID?scope=direct` | 角色直授菜单树，供角色权限页编辑 |
| POST | `/api/menu/add_menu_authority` | 覆盖保存角色菜单 |
| GET | `/api/sysApi/getAPiList` | API 清单 |
| GET | `/api/authority/getPolicyPathByAuthorityId` | 角色直接 API 策略 |
| POST | `/api/authority/updateCasbinByAuthority` | 覆盖保存角色直接 API 策略 |

## 升级说明

后端启动时会对现有数据库执行幂等 `AutoMigrate`，自动增加 `sys_users.token_version` 等新字段；系统 API 清单也会逐条补齐，不需要清空已有数据库。

首次升级到新版 RBAC 目录时，后端会执行一次 `rbac-catalog-v2` 数据迁移：

- 将旧的 `dashboard/cmdb/kubernetes/devops/setting` 菜单替换为当前前端实际使用的完整路由，例如 `/home`、`/workload/pod`、`/system/users`。
- 重建三个内置角色的默认菜单、API 和 `2221 -> 222` 继承关系。
- 保留无法识别的自定义菜单，但不会自动授予普通用户 `222`；迁移完成后写入 `sys_data_migrations` 标记，之后启动不会覆盖管理员的人工授权调整。
- 如果存量自定义菜单恰好占用了内置预留 ID `10001-10021`，迁移会报错并回滚，避免静默覆盖。

上线后，升级前签发的 Token 会因为令牌版本不匹配而失效，用户需要重新登录。

## 本地开发跨域

生产后端默认不再允许任意 Origin。浏览器直接跨域调用后端时，需要通过 `KUBEMANAGE_ALLOWED_ORIGINS` 显式配置可信来源。

本地开发使用 Vue dev server 的 `/api` 反向代理时不需要设置该变量：代理会移除浏览器的跨端口 Origin，再将请求转发到默认的 `http://127.0.0.1:6180/`。修改 `vue.config.js` 后必须重启前端开发服务；如后端不在默认地址，可设置 `KUBEMANAGE_DEV_BACKEND`。
