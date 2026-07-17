package sys

import (
	"context"
	"sync"
	"testing"

	"github.com/glebarez/sqlite"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"

	"github.com/noovertime7/kubemanage/dao"
	"github.com/noovertime7/kubemanage/dao/model"
	"github.com/noovertime7/kubemanage/dto"
	"github.com/noovertime7/kubemanage/pkg"
)

func newRBACServiceTestDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.SysAuthority{}, &model.SysUser{}, &model.SysBaseMenu{}, &model.SysAuthorityMenu{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	return db
}

func resetCasbinForTest(t *testing.T) {
	t.Helper()
	once = sync.Once{}
	cachedEnforcer = nil
	t.Cleanup(func() {
		once = sync.Once{}
		cachedEnforcer = nil
	})
}

func TestRoleInheritanceAppliesAPIAndMenus(t *testing.T) {
	resetCasbinForTest(t)
	db := newRBACServiceTestDB(t, "role_inheritance")
	if err := db.Create([]model.SysAuthority{
		{AuthorityId: 10, AuthorityName: "parent"},
		{AuthorityId: 20, AuthorityName: "child", ParentId: 10},
	}).Error; err != nil {
		t.Fatalf("create roles: %v", err)
	}
	if err := db.Create([]model.SysBaseMenu{
		{ID: 1, ParentId: "0", Path: "parent-menu", Name: "parent-menu"},
		{ID: 2, ParentId: "0", Path: "child-menu", Name: "child-menu"},
	}).Error; err != nil {
		t.Fatalf("create menus: %v", err)
	}
	if err := db.Create([]model.SysAuthorityMenu{
		{MenuId: "1", AuthorityId: "10"},
		{MenuId: "2", AuthorityId: "20"},
	}).Error; err != nil {
		t.Fatalf("create role menus: %v", err)
	}

	factory := dao.NewShareDaoFactory(db)
	casbinService := NewCasbinService(factory)
	if err := casbinService.UpdateCasbin(10, []dto.CasbinInfo{{Path: "/api/parent", Method: "GET"}}); err != nil {
		t.Fatalf("set parent policy: %v", err)
	}
	allowed, err := casbinService.Casbin().Enforce("20", "/api/parent", "GET")
	if err != nil {
		t.Fatalf("enforce inherited policy: %v", err)
	}
	if !allowed {
		t.Fatal("child role did not inherit parent API permission")
	}

	permissions := casbinService.GetImplicitPolicyPathByAuthorityId(20)
	if len(permissions) != 1 || permissions[0].Path != "/api/parent" {
		t.Fatalf("implicit permissions = %#v", permissions)
	}
	menus, err := NewMenuService(factory).GetMenuByAuthorityID(context.Background(), 20)
	if err != nil {
		t.Fatalf("get inherited menus: %v", err)
	}
	if len(menus) != 2 {
		t.Fatalf("child menus = %#v, want parent and child menus", menus)
	}
	directMenus, err := NewMenuService(factory).GetDirectMenuByAuthorityID(context.Background(), 20)
	if err != nil {
		t.Fatalf("get direct menus: %v", err)
	}
	if len(directMenus) != 1 || directMenus[0].Path != "child-menu" {
		t.Fatalf("child direct menus = %#v, want only child-menu", directMenus)
	}
}

func TestAdminMenuPermissionsCannotBeReduced(t *testing.T) {
	db := newRBACServiceTestDB(t, "admin_menu_permissions")
	if err := db.Create(&model.SysAuthority{AuthorityId: pkg.AdminDefaultAuth, AuthorityName: "admin"}).Error; err != nil {
		t.Fatalf("create admin role: %v", err)
	}
	menus := []model.SysBaseMenu{
		{ID: 1, ParentId: "0", Path: "/home", Name: "home"},
		{ID: 2, ParentId: "0", Path: "/system", Name: "system"},
	}
	if err := db.Create(&menus).Error; err != nil {
		t.Fatalf("create menus: %v", err)
	}
	factory := dao.NewShareDaoFactory(db)
	service := NewMenuService(factory)
	if err := service.AddMenuAuthority(context.Background(), nil, pkg.AdminDefaultAuth); err != nil {
		t.Fatalf("save admin menus: %v", err)
	}
	directMenus, err := service.GetDirectMenuByAuthorityID(context.Background(), pkg.AdminDefaultAuth)
	if err != nil {
		t.Fatalf("get admin menus: %v", err)
	}
	if len(directMenus) != len(menus) {
		t.Fatalf("admin menus=%d want=%d", len(directMenus), len(menus))
	}
}

func TestValidateClaimsRejectsChangedUserState(t *testing.T) {
	db := newRBACServiceTestDB(t, "claims_validation")
	userUUID := uuid.NewV4()
	user := model.SysUser{
		UUID:         userUUID,
		UserName:     "tester",
		AuthorityId:  10,
		Enable:       1,
		TokenVersion: 3,
	}
	if err := db.Create(&model.SysAuthority{AuthorityId: 10, AuthorityName: "tester"}).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := NewUserService(dao.NewShareDaoFactory(db))
	claims := &pkg.CustomClaims{BaseClaims: pkg.BaseClaims{
		UUID:         userUUID,
		ID:           user.ID,
		AuthorityId:  10,
		TokenVersion: 3,
	}}
	if err := service.ValidateClaims(context.Background(), claims); err != nil {
		t.Fatalf("valid claims rejected: %v", err)
	}

	stale := *claims
	stale.TokenVersion = 2
	if err := service.ValidateClaims(context.Background(), &stale); err == nil {
		t.Fatal("stale token version was accepted")
	}
	if err := db.Model(&user).Updates(map[string]interface{}{"enable": 2, "token_version": 4}).Error; err != nil {
		t.Fatalf("freeze user: %v", err)
	}
	if err := service.ValidateClaims(context.Background(), claims); err == nil {
		t.Fatal("frozen user token was accepted")
	}
}
