package model

import (
	"database/sql"

	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/noovertime7/kubemanage/pkg"
	uuid "github.com/satori/go.uuid"
)

// 初始化顺序
const (
	SysAuthorityOrder = iota
	SysUserOrder
	SysBaseMenuOrder
	MenuAuthorityOrder
	SysApisInitOrder
	CasbinInitOrder
	RBACCatalogMigrationOrder
	OperatorationOrder
	WorkFlowOrder
	GrafanaMenuMigrationOrder
)

// SysUserEntities 用户初始化数据
var (
	SysUserEntities = []*SysUser{
		{
			UUID:         uuid.NewV4(),
			UserName:     "admin",
			Password:     "$2a$14$Zfb6w0UDBFMN0.nJeVXCUO3zH/iWKGtbBYyIzDDRnC..EgTS0Et0S",
			NickName:     "admin",
			SideMode:     "dark",
			Avatar:       "https://qmplusimg.henrongyi.top/gva_header.jpg",
			BaseColor:    "#fff",
			ActiveColor:  "#1890ff",
			AuthorityId:  pkg.AdminDefaultAuth,
			Phone:        "12345678901",
			Email:        "test@qq.com",
			Enable:       1,
			TokenVersion: 1,
			Status:       sql.NullInt64{Int64: 0, Valid: true},
		},
		{
			UUID:         uuid.NewV4(),
			UserName:     "chenteng",
			Password:     "$2a$14$yLCxKYP46M2NRnXujYe3mOfNe00GtBtjpaLM2eIzYCzYKQXqzsuka",
			NickName:     "chenteng",
			SideMode:     "dark",
			Avatar:       "https://qmplusimg.henrongyi.top/gva_header.jpg",
			BaseColor:    "#fff",
			ActiveColor:  "#1890ff",
			AuthorityId:  pkg.UserDefaultAuth,
			Phone:        "12345678901",
			Email:        "test@qq.com",
			Enable:       1,
			TokenVersion: 1,
			Status:       sql.NullInt64{Int64: 0, Valid: true},
		},
		{
			UUID:         uuid.NewV4(),
			UserName:     "chentengsub",
			Password:     "$2a$14$MPINiht5QO2wlR3DynizXOtuqcNAOrNZdrSUKXrbjqcKbK.jcfyAW",
			NickName:     "chentengsub",
			SideMode:     "dark",
			Avatar:       "https://qmplusimg.henrongyi.top/gva_header.jpg",
			BaseColor:    "#fff",
			ActiveColor:  "#1890ff",
			AuthorityId:  pkg.UserSubDefaultAuth,
			Phone:        "12345678901",
			Email:        "test@qq.com",
			Enable:       1,
			TokenVersion: 1,
			Status:       sql.NullInt64{Int64: 0, Valid: true},
		},
	}
)

// SysBaseMenuEntities 是与 kubemanage-web 共用的路由目录。
// 内置菜单使用预留 ID 区间，升级替换旧菜单目录时，避免与历史数据中的小号自增 ID 冲突。
var (
	SysBaseMenuEntities = []SysBaseMenu{
		{ID: 10001, MenuLevel: 0, ParentId: "0", Path: "/home", Name: "概要", Sort: 1, Meta: Meta{Title: "概要", Icon: "Help"}},
		{ID: 10002, MenuLevel: 0, ParentId: "0", Path: "/workflow", Name: "工作流", Sort: 2, Meta: Meta{Title: "工作流", Icon: "VideoPlay"}},

		{ID: 10003, MenuLevel: 0, ParentId: "0", Path: "/workload", Name: "工作负载", Sort: 3, Meta: Meta{Title: "工作负载", Icon: "Menu"}},
		{ID: 10004, MenuLevel: 1, ParentId: "10003", Path: "/workload/deployment", Name: "Deployment", Sort: 1, Meta: Meta{Title: "Deployment"}},
		{ID: 10005, MenuLevel: 1, ParentId: "10003", Path: "/workload/pod", Name: "Pod", Sort: 2, Meta: Meta{Title: "Pod"}},
		{ID: 10006, MenuLevel: 1, ParentId: "10003", Path: "/workload/daemonset", Name: "DaemonSet", Sort: 3, Meta: Meta{Title: "DaemonSet"}},
		{ID: 10007, MenuLevel: 1, ParentId: "10003", Path: "/workload/statefulset", Name: "StatefulSet", Sort: 4, Meta: Meta{Title: "StatefulSet"}},

		{ID: 10008, MenuLevel: 0, ParentId: "0", Path: "/loadbalance", Name: "负载均衡", Sort: 4, Meta: Meta{Title: "负载均衡", Icon: "Files"}},
		{ID: 10009, MenuLevel: 1, ParentId: "10008", Path: "/loadbalance/service", Name: "Service", Sort: 1, Meta: Meta{Title: "Service"}},
		{ID: 10010, MenuLevel: 1, ParentId: "10008", Path: "/loadbalance/ingress", Name: "Ingress", Sort: 2, Meta: Meta{Title: "Ingress"}},

		{ID: 10011, MenuLevel: 0, ParentId: "0", Path: "/storage", Name: "存储与配置", Sort: 5, Meta: Meta{Title: "存储与配置", Icon: "Tickets"}},
		{ID: 10012, MenuLevel: 1, ParentId: "10011", Path: "/storage/configmap", Name: "Configmap", Sort: 1, Meta: Meta{Title: "Configmap"}},
		{ID: 10013, MenuLevel: 1, ParentId: "10011", Path: "/storage/secret", Name: "Secret", Sort: 2, Meta: Meta{Title: "Secret"}},
		{ID: 10014, MenuLevel: 1, ParentId: "10011", Path: "/storage/persistentvolumeclaim", Name: "PVC", Sort: 3, Meta: Meta{Title: "PersistentVolumeClaim"}},

		{ID: 10015, MenuLevel: 0, ParentId: "0", Path: "/cluster", Name: "集群", Sort: 6, Meta: Meta{Title: "集群", Icon: "Cpu"}},
		{ID: 10016, MenuLevel: 1, ParentId: "10015", Path: "/cluster/node", Name: "Node", Sort: 1, Meta: Meta{Title: "Node"}},
		{ID: 10017, MenuLevel: 1, ParentId: "10015", Path: "/cluster/namespace", Name: "Namespace", Sort: 2, Meta: Meta{Title: "Namespace"}},
		{ID: 10018, MenuLevel: 1, ParentId: "10015", Path: "/cluster/persistentvolume", Name: "PersistentVolume", Sort: 3, Meta: Meta{Title: "PersistentVolume"}},

		{ID: 10019, MenuLevel: 0, ParentId: "0", Path: "/system", Name: "系统管理", Sort: 7, Meta: Meta{Title: "系统管理", Icon: "Setting"}},
		{ID: 10020, MenuLevel: 1, ParentId: "10019", Path: "/system/users", Name: "用户管理", Sort: 1, Meta: Meta{Title: "用户管理"}},
		{ID: 10021, MenuLevel: 1, ParentId: "10019", Path: "/system/roles", Name: "角色权限", Sort: 2, Meta: Meta{Title: "角色权限"}},

		{ID: 10022, MenuLevel: 0, ParentId: "0", Path: "/monitor", Name: "监控", Sort: 8, Meta: Meta{Title: "监控", Icon: "Odometer"}},
		{ID: 10023, MenuLevel: 1, ParentId: "10022", Path: "/monitor/grafana", Name: "Grafana", Sort: 1, Meta: Meta{Title: "Grafana"}},
	}
)

// SysAuthorityEntities 角色初始化数据
var (
	SysAuthorityEntities = []SysAuthority{
		{
			AuthorityId:   pkg.AdminDefaultAuth,
			AuthorityName: "管理员",
			DefaultRouter: "/home",
			ParentId:      0,
		},
		{
			AuthorityId:   pkg.UserDefaultAuth,
			AuthorityName: "普通用户",
			DefaultRouter: "/home",
			ParentId:      0,
		},
		{
			AuthorityId:   pkg.UserSubDefaultAuth,
			AuthorityName: "普通用户子角色",
			DefaultRouter: "/home",
			ParentId:      222,
		},
	}
)

var CasbinApi = buildCasbinRule(SysApis)

// buildCasbinRule 构建角色casbin api
func buildCasbinRule(apis []SysApi) []adapter.CasbinRule {
	out := make([]adapter.CasbinRule, 0, len(apis)*2)
	for _, api := range apis {
		out = append(out, adapter.CasbinRule{
			Ptype: "p",
			V0:    pkg.AdminDefaultAuthStr,
			V1:    api.Path,
			V2:    api.Method,
		})
		if isDefaultUserAPI(api) {
			out = append(out, adapter.CasbinRule{
				Ptype: "p",
				V0:    pkg.UserDefaultAuthStr,
				V1:    api.Path,
				V2:    api.Method,
			})
		}
	}
	// 子角色不直接授予策略，其生效权限全部来自 222 -> 2221 的角色继承关系。
	return out
}

func isDefaultUserAPI(api SysApi) bool {
	if api.ApiGroup == "Kubernetes" && api.Method == "GET" {
		return api.Path != "/api/k8s/pod/webshell" && api.Path != "/api/k8s/deployment/scale"
	}
	if api.ApiGroup != "用户" {
		return false
	}
	switch api.Path {
	case "/api/user/login", "/api/user/loginout", "/api/user/getinfo", "/api/user/:id/change_pwd":
		return true
	default:
		return false
	}
}

var SysApis = []SysApi{
	// api接口
	{Path: "/api/sysApi/getAPiList", Description: "获取系统API列表", ApiGroup: "系统", Method: "GET"},

	// 用户相关接口
	{Path: "/api/user/login", Description: "用户登录", ApiGroup: "用户", Method: "POST"},
	{Path: "/api/user/loginout", Description: "用户退出", ApiGroup: "用户", Method: "GET"},
	{Path: "/api/user/getinfo", Description: "获取用户信息", ApiGroup: "用户", Method: "GET"},
	{Path: "/api/user/:id/set_auth", Description: "设置用户权限", ApiGroup: "用户", Method: "PUT"},
	{Path: "/api/user/:id/delete_user", Description: "删除用户", ApiGroup: "用户", Method: "DELETE"},
	{Path: "/api/user/:id/change_pwd", Description: "修改密码", ApiGroup: "用户", Method: "POST"},
	{Path: "/api/user/:id/reset_pwd", Description: "重置密码", ApiGroup: "用户", Method: "PUT"},
	{Path: "/api/user/list", Description: "用户列表", ApiGroup: "用户", Method: "GET"},
	{Path: "/api/user", Description: "创建用户", ApiGroup: "用户", Method: "POST"},
	{Path: "/api/user/:id", Description: "编辑用户", ApiGroup: "用户", Method: "PUT"},
	{Path: "/api/user/:id/enable", Description: "启停用户", ApiGroup: "用户", Method: "PUT"},
	// 操作审计接口
	{Path: "/api/operation/get_operations", Description: "查询操作记录列表", ApiGroup: "操作审计", Method: "GET"},
	{Path: "/api/operation/:id/delete_operation", Description: "删除单条记录", ApiGroup: "操作审计", Method: "DELETE"},
	{Path: "/api/operation/delete_operations", Description: "批量删除记录", ApiGroup: "操作审计", Method: "POST"},
	// Other
	{Path: "/api/swagger/*any", Description: "swagger文档", ApiGroup: "Other", Method: "GET"},
	// 菜单接口
	{Path: "/api/menu/:authID/getMenuByAuthID", Description: "根据角色获取菜单", ApiGroup: "菜单", Method: "GET"},
	{Path: "/api/menu/getBaseMenuTree", Description: "获取菜单总树", ApiGroup: "菜单", Method: "GET"},
	{Path: "/api/menu/add_base_menu", Description: "添加菜单", ApiGroup: "菜单", Method: "POST"},
	{Path: "/api/menu/add_menu_authority", Description: "添加角色", ApiGroup: "菜单", Method: "POST"},
	// 权限RBAC接口
	{Path: "/api/authority/getPolicyPathByAuthorityId", Description: "获取角色api权限", ApiGroup: "权限", Method: "GET"},
	{Path: "/api/authority/updateCasbinByAuthority", Description: "更改角色api权限", ApiGroup: "用户", Method: "POST"},
	{Path: "/api/authority/getAuthorityList", Description: "获取角色列表", ApiGroup: "权限", Method: "GET"},
	{Path: "/api/authority", Description: "创建角色", ApiGroup: "权限", Method: "POST"},
	{Path: "/api/authority/:authorityId", Description: "编辑角色", ApiGroup: "权限", Method: "PUT"},
	{Path: "/api/authority/:authorityId", Description: "删除角色", ApiGroup: "权限", Method: "DELETE"},
	// K8S相关接口
	{Path: "/api/k8s/deployment/create", Description: "创建deployment", ApiGroup: "Kubernetes", Method: "POST"},
	{Path: "/api/k8s/deployment/del", Description: "删除deployment", ApiGroup: "Kubernetes", Method: "DELETE"},
	{Path: "/api/k8s/deployment/update", Description: "更新deployment", ApiGroup: "Kubernetes", Method: "PUT"},
	{Path: "/api/k8s/deployment/list", Description: "查询deployment列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/deployment/detail", Description: "查询deployment详情", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/deployment/restart", Description: "重启deployment", ApiGroup: "Kubernetes", Method: "PUT"},
	{Path: "/api/k8s/deployment/scale", Description: "deployment扩缩容", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/deployment/numnp", Description: "查询deployment数量信息", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/pod/list", Description: "查询pod列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/pod/detail", Description: "查询pod详情", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/pod/del", Description: "删除pod", ApiGroup: "Kubernetes", Method: "DELETE"},
	{Path: "/api/k8s/pod/update", Description: "更新pod", ApiGroup: "Kubernetes", Method: "PUT"},
	{Path: "/api/k8s/pod/container", Description: "获取Pod内容器名", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/pod/log", Description: "获取容器日志", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/pod/numnp", Description: "查询pod数量信息", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/pod/webshell", Description: "web终端", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/daemonset/del", Description: "删除daemonset", ApiGroup: "Kubernetes", Method: "DELETE"},
	{Path: "/api/k8s/daemonset/update", Description: "更新daemonset", ApiGroup: "Kubernetes", Method: "PUT"},
	{Path: "/api/k8s/daemonset/list", Description: "查询daemonset列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/daemonset/detail", Description: "查询daemonset详情", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/statefulset/del", Description: "删除statefulset", ApiGroup: "Kubernetes", Method: "DELETE"},
	{Path: "/api/k8s/statefulset/update", Description: "更新statefulset", ApiGroup: "Kubernetes", Method: "PUT"},
	{Path: "/api/k8s/statefulset/list", Description: "查询statefulset列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/statefulset/detail", Description: "查询statefulset详情", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/node/list", Description: "查询node列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/node/detail", Description: "查询node详情", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/namespace/create", Description: "创建namespace", ApiGroup: "Kubernetes", Method: "PUT"},
	{Path: "/api/k8s/namespace/del", Description: "删除namespace", ApiGroup: "Kubernetes", Method: "DELETE"},
	{Path: "/api/k8s/namespace/list", Description: "查询namespace列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/namespace/detail", Description: "查询namespace详情", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/persistentvolume/del", Description: "删除persistentvolume", ApiGroup: "Kubernetes", Method: "DELETE"},
	{Path: "/api/k8s/persistentvolume/list", Description: "查询persistentvolume列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/persistentvolume/detail", Description: "查询persistentvolume详情", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/service/create", Description: "创建service", ApiGroup: "Kubernetes", Method: "POST"},
	{Path: "/api/k8s/service/del", Description: "删除service", ApiGroup: "Kubernetes", Method: "DELETE"},
	{Path: "/api/k8s/service/update", Description: "更新service", ApiGroup: "Kubernetes", Method: "PUT"},
	{Path: "/api/k8s/service/list", Description: "查询service列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/service/detail", Description: "查询service详情", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/service/numnp", Description: "查询service数量信息", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/ingress/create", Description: "创建ingress", ApiGroup: "Kubernetes", Method: "PUT"},
	{Path: "/api/k8s/ingress/del", Description: "删除ingress", ApiGroup: "Kubernetes", Method: "DELETE"},
	{Path: "/api/k8s/ingress/update", Description: "更新ingress", ApiGroup: "Kubernetes", Method: "PUT"},
	{Path: "/api/k8s/ingress/list", Description: "查询ingress列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/ingress/detail", Description: "查询ingress详情", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/ingress/numnp", Description: "查询ingress数量信息", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/configmap/del", Description: "删除configmap", ApiGroup: "Kubernetes", Method: "DELETE"},
	{Path: "/api/k8s/configmap/update", Description: "更新configmap", ApiGroup: "Kubernetes", Method: "PUT"},
	{Path: "/api/k8s/configmap/list", Description: "查询configmap列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/configmap/detail", Description: "查询configmap详情", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/persistentvolumeclaim/del", Description: "删除persistentvolumeclaim", ApiGroup: "Kubernetes", Method: "DELETE"},
	{Path: "/api/k8s/persistentvolumeclaim/update", Description: "更新persistentvolumeclaim", ApiGroup: "Kubernetes", Method: "PUT"},
	{Path: "/api/k8s/persistentvolumeclaim/list", Description: "查询persistentvolumeclaim列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/persistentvolumeclaim/detail", Description: "查询persistentvolumeclaim详情", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/secret/del", Description: "删除secret", ApiGroup: "Kubernetes", Method: "DELETE"},
	{Path: "/api/k8s/secret/update", Description: "更新secret", ApiGroup: "Kubernetes", Method: "PUT"},
	{Path: "/api/k8s/secret/list", Description: "查询secret列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/secret/detail", Description: "查询secret详情", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/workflow/create", Description: "创建workflow", ApiGroup: "Kubernetes", Method: "POST"},
	{Path: "/api/k8s/workflow/del", Description: "删除workflow", ApiGroup: "Kubernetes", Method: "DELETE"},
	{Path: "/api/k8s/workflow/list", Description: "查询workflow列表", ApiGroup: "Kubernetes", Method: "GET"},
	{Path: "/api/k8s/workflow/id", Description: "查看workflow", ApiGroup: "Kubernetes", Method: "GET"},
}
