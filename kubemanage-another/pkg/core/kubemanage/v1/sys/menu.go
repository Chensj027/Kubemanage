package sys

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/noovertime7/kubemanage/dao"
	"github.com/noovertime7/kubemanage/dao/model"
	"github.com/noovertime7/kubemanage/dto"
	"github.com/noovertime7/kubemanage/pkg"
)

// MenuGetter MenuService对象获取器
type MenuGetter interface {
	// Menu 获取一个新的MenuService对象
	Menu() MenuService
}

// MenuService Menu菜单相关操作的Service方法
type MenuService interface {
	GetMenuByAuthorityID(ctx context.Context, authorityId uint) ([]model.SysMenu, error)
	GetDirectMenuByAuthorityID(ctx context.Context, authorityId uint) ([]model.SysMenu, error)
	GetBassMenu(ctx context.Context) ([]model.SysBaseMenu, error)
	AddBaseMenu(ctx context.Context, in *dto.AddSysMenusInput) error
	AddMenuAuthority(ctx context.Context, menus []model.SysBaseMenu, authorityId uint) error
}

// menuService MenuService接口的实现类
type menuService struct {
	factory dao.ShareDaoFactory
}

func NewMenuService(factory dao.ShareDaoFactory) *menuService {
	return &menuService{factory: factory}
}

// GetBassMenu 获取全量的菜单
func (m *menuService) GetBassMenu(ctx context.Context) ([]model.SysBaseMenu, error) {
	treeMap, err := m.getBaseMenuTreeMap(ctx)
	if err != nil {
		return nil, err
	}
	menus := treeMap["0"]
	for i := 0; i < len(menus); i++ {
		if err := m.getBaseChildrenList(&menus[i], treeMap); err != nil {
			return nil, err
		}
	}
	return menus, nil
}

func (m *menuService) GetMenuByAuthorityID(ctx context.Context, authorityId uint) ([]model.SysMenu, error) {
	return m.getMenusByAuthorityID(ctx, authorityId, true)
}

// GetDirectMenuByAuthorityID 仅返回直接授予指定角色的菜单。
// 角色管理页面使用此视图，避免保存直授权限时把继承菜单复制到子角色。
func (m *menuService) GetDirectMenuByAuthorityID(ctx context.Context, authorityId uint) ([]model.SysMenu, error) {
	return m.getMenusByAuthorityID(ctx, authorityId, false)
}

func (m *menuService) getMenusByAuthorityID(ctx context.Context, authorityId uint, inherited bool) ([]model.SysMenu, error) {
	menuTree, err := m.getMenuTree(ctx, authorityId, inherited)
	if err != nil {
		return nil, err
	}
	//parent_id = 0 ,代表所有跟路由
	menus := menuTree["0"]
	for i := 0; i < len(menus); i++ {
		err = m.getChildrenList(&menus[i], menuTree)
	}
	return menus, nil
}

// AddBaseMenu 添加基础路由
func (m *menuService) AddBaseMenu(ctx context.Context, in *dto.AddSysMenusInput) error {
	menuInfo := &model.SysBaseMenu{
		ParentId: in.ParentId,
		Name:     in.Name,
		Path:     in.Path,
		Hidden:   in.Hidden,
		Sort:     in.Sort,
		Meta:     in.Meta,
	}
	menu, err := m.factory.BaseMenu().Find(ctx, menuInfo)
	if !errors.Is(err, gorm.ErrRecordNotFound) && menu.ID != 0 {
		return errors.New("存在重复名称菜单，请修改菜单名称")
	}
	return m.factory.BaseMenu().Save(ctx, menuInfo)
}

// AddMenuAuthority 为角色增加menu树
func (m *menuService) AddMenuAuthority(ctx context.Context, menus []model.SysBaseMenu, authorityId uint) error {
	if _, err := m.factory.Authority().Find(ctx, &model.SysAuthority{AuthorityId: authorityId}); err != nil {
		return errors.New("角色不存在")
	}
	// 管理员菜单与 API 一样不可降权。即使绕过前端直接调用接口，
	// 服务端也会将其恢复为完整菜单目录。
	if authorityId == pkg.AdminDefaultAuth {
		allMenus, err := m.factory.BaseMenu().FindList(ctx, nil)
		if err != nil {
			return err
		}
		menus = allMenus
	}
	auth := &model.SysAuthority{AuthorityId: authorityId, SysBaseMenus: menus}
	return m.factory.Authority().SetMenuAuthority(ctx, auth)
}

func (m *menuService) getMenuTree(ctx context.Context, authorityId uint, inherited bool) (map[string][]model.SysMenu, error) {
	var allMenus []model.SysMenu
	treeMap := make(map[string][]model.SysMenu)
	// 111 是不可降权的超级管理员。读取时也直接使用完整菜单目录，
	// 避免数据库被手工改动后管理员菜单与其 API 硬放行语义不一致。
	if authorityId == pkg.AdminDefaultAuth {
		baseMenus, err := m.factory.BaseMenu().FindList(ctx, nil)
		if err != nil {
			return nil, err
		}
		for i := range baseMenus {
			allMenus = append(allMenus, model.SysMenu{
				SysBaseMenu: baseMenus[i],
				AuthorityId: authorityId,
				MenuId:      strconv.Itoa(baseMenus[i].ID),
			})
		}
		for _, menu := range allMenus {
			treeMap[menu.ParentId] = append(treeMap[menu.ParentId], menu)
		}
		return treeMap, nil
	}
	authorityIDs := []uint{authorityId}
	if inherited {
		var err error
		authorityIDs, err = m.getAuthorityChain(ctx, authorityId)
		if err != nil {
			return nil, err
		}
	}
	menuIDSet := make(map[string]struct{})
	for _, id := range authorityIDs {
		authorityMenus, err := m.factory.AuthorityMenu().FindList(ctx, &model.SysAuthorityMenu{AuthorityId: strconv.Itoa(int(id))})
		if err != nil {
			return nil, err
		}
		for i := range authorityMenus {
			menuIDSet[authorityMenus[i].MenuId] = struct{}{}
		}
	}
	menuIDs := make([]string, 0, len(menuIDSet))
	for id := range menuIDSet {
		menuIDs = append(menuIDs, id)
	}
	baseMenus, err := m.factory.BaseMenu().FindIn(ctx, menuIDs)
	if err != nil {
		return nil, err
	}
	for i := range baseMenus {
		allMenus = append(allMenus, model.SysMenu{
			SysBaseMenu: *baseMenus[i],
			AuthorityId: authorityId,
			MenuId:      strconv.Itoa(baseMenus[i].ID),
		})
	}
	for _, v := range allMenus {
		treeMap[v.ParentId] = append(treeMap[v.ParentId], v)
	}
	return treeMap, nil
}

func (m *menuService) getAuthorityChain(ctx context.Context, authorityID uint) ([]uint, error) {
	chain := make([]uint, 0, 4)
	visited := make(map[uint]struct{})
	current := authorityID
	for current != 0 {
		if _, exists := visited[current]; exists {
			return nil, errors.New("角色继承关系存在循环")
		}
		visited[current] = struct{}{}
		chain = append(chain, current)
		var role model.SysAuthority
		if err := m.factory.GetDB().WithContext(ctx).Select("authority_id", "parent_id").Where("authority_id = ?", current).First(&role).Error; err != nil {
			return nil, errors.Wrap(err, "角色不存在")
		}
		current = role.ParentId
	}
	return chain, nil
}

func (m *menuService) getChildrenList(menu *model.SysMenu, treeMap map[string][]model.SysMenu) error {
	// treeMap中包含所有路由
	menu.Children = treeMap[menu.MenuId]
	for i := 0; i < len(menu.Children); i++ {
		if err := m.getChildrenList(&menu.Children[i], treeMap); err != nil {
			return err
		}
	}
	return nil
}

func (m *menuService) getBaseChildrenList(menu *model.SysBaseMenu, treeMap map[string][]model.SysBaseMenu) (err error) {
	menu.Children = treeMap[strconv.Itoa(menu.ID)]
	for i := 0; i < len(menu.Children); i++ {
		err = m.getBaseChildrenList(&menu.Children[i], treeMap)
	}
	return err
}

func (m *menuService) getBaseMenuTreeMap(ctx context.Context) (treeMap map[string][]model.SysBaseMenu, err error) {
	var menuDB *model.SysBaseMenu
	treeMap = make(map[string][]model.SysBaseMenu)
	allMenus, err := m.factory.BaseMenu().FindList(ctx, menuDB)
	if err != nil {
		return nil, err
	}
	for _, v := range allMenus {
		treeMap[v.ParentId] = append(treeMap[v.ParentId], v)
	}
	return treeMap, err
}
