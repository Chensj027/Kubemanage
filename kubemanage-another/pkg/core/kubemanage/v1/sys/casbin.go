package sys

import (
	"strconv"
	"sync"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/pkg/errors"

	"github.com/noovertime7/kubemanage/dao"
	daomodel "github.com/noovertime7/kubemanage/dao/model"
	"github.com/noovertime7/kubemanage/dto"
	"github.com/noovertime7/kubemanage/pkg"
)

type CasbinServiceGetter interface {
	CasbinService() CasbinService
}

type CasbinService interface {
	UpdateCasbin(AuthorityID uint, casbinInfos []dto.CasbinInfo) error
	UpdateCasbinApi(oldPath string, newPath string, oldMethod string, newMethod string) error
	GetPolicyPathByAuthorityId(AuthorityID uint) (pathMaps []dto.CasbinInfo)
	GetImplicitPolicyPathByAuthorityId(AuthorityID uint) (pathMaps []dto.CasbinInfo)
	SyncRoleParent(AuthorityID, ParentID uint) error
	RemoveRole(AuthorityID uint) error
	Casbin() *casbin.CachedEnforcer
}

type casbinService struct {
	factory dao.ShareDaoFactory
}

func NewCasbinService(factory dao.ShareDaoFactory) CasbinService {
	return &casbinService{factory: factory}
}

func (c *casbinService) UpdateCasbin(AuthorityID uint, casbinInfos []dto.CasbinInfo) error {
	if AuthorityID == pkg.AdminDefaultAuth {
		var err error
		casbinInfos, err = c.allAPIPolicies()
		if err != nil {
			return err
		}
	}
	authorityId := strconv.Itoa(int(AuthorityID))
	e := c.Casbin()
	if e == nil {
		return errors.New("Casbin初始化失败")
	}
	if _, err := e.RemoveFilteredPolicy(0, authorityId); err != nil {
		return err
	}
	var rules [][]string
	for _, v := range casbinInfos {
		rules = append(rules, []string{authorityId, v.Path, v.Method})
	}
	if len(rules) > 0 {
		success, err := e.AddPolicies(rules)
		if err != nil {
			return err
		}
		if !success {
			return errors.New("存在相同api,添加失败,请联系管理员")
		}
	}
	err := e.InvalidateCache()
	if err != nil {
		return err
	}
	return nil
}

func (c *casbinService) UpdateCasbinApi(oldPath string, newPath string, oldMethod string, newMethod string) error {
	err := c.factory.GetDB().Model(&gormadapter.CasbinRule{}).Where("v1 = ? AND v2 = ?", oldPath, oldMethod).Updates(map[string]interface{}{
		"v1": newPath,
		"v2": newMethod,
	}).Error
	e := c.Casbin()
	err = e.InvalidateCache()
	if err != nil {
		return err
	}
	return err
}

func (c *casbinService) GetPolicyPathByAuthorityId(AuthorityID uint) (pathMaps []dto.CasbinInfo) {
	// 角色 111 设计为不可降权，匹配器始终允许其访问，
	// 因此这里返回完整 API 目录，避免展示旧版本遗留的不完整策略列表。
	if AuthorityID == pkg.AdminDefaultAuth {
		pathMaps, _ = c.allAPIPolicies()
		return pathMaps
	}
	e := c.Casbin()
	authorityId := strconv.Itoa(int(AuthorityID))
	list := e.GetFilteredPolicy(0, authorityId)
	for _, v := range list {
		pathMaps = append(pathMaps, dto.CasbinInfo{
			Path:   v[1],
			Method: v[2],
		})
	}
	return pathMaps
}

func (c *casbinService) allAPIPolicies() ([]dto.CasbinInfo, error) {
	var apis []daomodel.SysApi
	if err := c.factory.GetDB().Order("api_group, path, method").Find(&apis).Error; err != nil {
		return nil, err
	}
	policies := make([]dto.CasbinInfo, 0, len(apis))
	for _, api := range apis {
		policies = append(policies, dto.CasbinInfo{Path: api.Path, Method: api.Method})
	}
	return policies, nil
}

func (c *casbinService) GetImplicitPolicyPathByAuthorityId(AuthorityID uint) (pathMaps []dto.CasbinInfo) {
	e := c.Casbin()
	if e == nil {
		return nil
	}
	list, err := e.GetImplicitPermissionsForUser(strconv.Itoa(int(AuthorityID)))
	if err != nil {
		return nil
	}
	seen := make(map[string]struct{}, len(list))
	for _, rule := range list {
		if len(rule) < 3 {
			continue
		}
		key := rule[1] + "\x00" + rule[2]
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		pathMaps = append(pathMaps, dto.CasbinInfo{Path: rule[1], Method: rule[2]})
	}
	return pathMaps
}

func (c *casbinService) SyncRoleParent(AuthorityID, ParentID uint) error {
	e := c.Casbin()
	if e == nil {
		return errors.New("Casbin初始化失败")
	}
	roleID := strconv.Itoa(int(AuthorityID))
	if _, err := e.RemoveFilteredGroupingPolicy(0, roleID); err != nil {
		return err
	}
	if ParentID != 0 {
		if _, err := e.AddGroupingPolicy(roleID, strconv.Itoa(int(ParentID))); err != nil {
			return err
		}
	}
	return e.InvalidateCache()
}

func (c *casbinService) RemoveRole(AuthorityID uint) error {
	e := c.Casbin()
	if e == nil {
		return errors.New("Casbin初始化失败")
	}
	roleID := strconv.Itoa(int(AuthorityID))
	if _, err := e.RemoveFilteredPolicy(0, roleID); err != nil {
		return err
	}
	if _, err := e.RemoveFilteredGroupingPolicy(0, roleID); err != nil {
		return err
	}
	if _, err := e.RemoveFilteredGroupingPolicy(1, roleID); err != nil {
		return err
	}
	return e.InvalidateCache()
}

func (c *casbinService) ClearCasbin(v int, p ...string) bool {
	e := c.Casbin()
	success, _ := e.RemoveFilteredPolicy(v, p...)
	return success
}

var (
	cachedEnforcer *casbin.CachedEnforcer
	once           sync.Once
)

func (c *casbinService) Casbin() *casbin.CachedEnforcer {
	once.Do(func() {
		a, err := gormadapter.NewAdapterByDB(c.factory.GetDB())
		if err != nil {
			return
		}
		text := `
		[request_definition]
		r = sub, obj, act
		
		[policy_definition]
		p = sub, obj, act
		
		[role_definition]
		g = _, _
		
		[policy_effect]
		e = some(where (p.eft == allow))
		
		[matchers]
			m = (g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)) || r.sub == "111"
			`
		m, err := casbinmodel.NewModelFromString(text)
		if err != nil {
			return
		}
		cachedEnforcer, err = casbin.NewCachedEnforcer(m, a)
		if err != nil {
			cachedEnforcer = nil
			return
		}
		cachedEnforcer.SetExpireTime(60 * 60)
		if err := cachedEnforcer.LoadPolicy(); err != nil {
			cachedEnforcer = nil
			return
		}

		// ParentId 是角色继承关系的唯一数据源。
		// 启动时重建 Casbin 角色分组策略，使存量数据库和修改后的角色始终与 sys_authorities 一致。
		for _, grouping := range cachedEnforcer.GetGroupingPolicy() {
			_, _ = cachedEnforcer.RemoveGroupingPolicy(grouping)
		}
		var roles []daomodel.SysAuthority
		if err := c.factory.GetDB().Select("authority_id", "parent_id").Find(&roles).Error; err == nil {
			var groupings [][]string
			for _, role := range roles {
				if role.ParentId != 0 && role.ParentId != role.AuthorityId {
					groupings = append(groupings, []string{strconv.Itoa(int(role.AuthorityId)), strconv.Itoa(int(role.ParentId))})
				}
			}
			if len(groupings) > 0 {
				_, _ = cachedEnforcer.AddGroupingPolicies(groupings)
			}
		}
	})
	return cachedEnforcer
}
