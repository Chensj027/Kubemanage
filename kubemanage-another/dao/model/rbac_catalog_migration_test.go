package model

import (
	"context"
	"strconv"
	"testing"

	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/glebarez/sqlite"
	"github.com/noovertime7/kubemanage/pkg"
	"gorm.io/gorm"
)

func TestRBACCatalogMigrationIsOneTimeAndPreservesCustomMenus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:rbac_catalog_migration?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&SysAuthority{}, &SysBaseMenu{}, &SysAuthorityMenu{}, &SysApi{}, &adapter.CasbinRule{}, &dataMigration{}); err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&SysAuthorityEntities).Error; err != nil {
		t.Fatal(err)
	}
	legacy := []SysBaseMenu{
		{ID: 1, ParentId: "0", Path: "dashboard", Name: "旧仪表盘"},
		{ID: 2, ParentId: "0", Path: "cmdb", Name: "旧资产中心"},
		{ID: 77, ParentId: "0", Path: "/custom", Name: "自定义"},
	}
	if err = db.Create(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Create([]SysAuthorityMenu{
		{MenuId: "1", AuthorityId: pkg.AdminDefaultAuthStr},
		{MenuId: "2", AuthorityId: pkg.UserDefaultAuthStr},
		{MenuId: "77", AuthorityId: pkg.AdminDefaultAuthStr},
		{MenuId: "1", AuthorityId: pkg.UserSubDefaultAuthStr},
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&SysApis).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Create([]adapter.CasbinRule{
		{Ptype: "p", V0: pkg.AdminDefaultAuthStr, V1: "/old-admin", V2: "GET"},
		{Ptype: "p", V0: pkg.UserDefaultAuthStr, V1: "/old-user", V2: "POST"},
		{Ptype: "p", V0: pkg.UserSubDefaultAuthStr, V1: "/old-child", V2: "GET"},
	}).Error; err != nil {
		t.Fatal(err)
	}

	migration := &rbacCatalogMigration{}
	if err = migration.InitData(context.Background(), db); err != nil {
		t.Fatal(err)
	}

	var legacyCount int64
	if err = db.Model(&SysBaseMenu{}).Where("path IN ?", legacyMenuPaths).Count(&legacyCount).Error; err != nil || legacyCount != 0 {
		t.Fatalf("legacy menus remain: count=%d err=%v", legacyCount, err)
	}
	var customCount int64
	if err = db.Model(&SysBaseMenu{}).Where("path = ?", "/custom").Count(&customCount).Error; err != nil || customCount != 1 {
		t.Fatalf("custom menu was not preserved: count=%d err=%v", customCount, err)
	}
	assertDirectMenuCount(t, db, pkg.AdminDefaultAuth, len(SysBaseMenuEntities)+1)
	assertDirectMenuCount(t, db, pkg.UserDefaultAuth, len(SysBaseMenuEntities)-3)
	assertDirectMenuCount(t, db, pkg.UserSubDefaultAuth, 0)

	var adminPolicies int64
	if err = db.Model(&adapter.CasbinRule{}).Where("ptype = ? AND v0 = ?", "p", pkg.AdminDefaultAuthStr).Count(&adminPolicies).Error; err != nil || adminPolicies != int64(len(SysApis)) {
		t.Fatalf("admin policies=%d want=%d err=%v", adminPolicies, len(SysApis), err)
	}
	var childPolicies int64
	if err = db.Model(&adapter.CasbinRule{}).Where("ptype = ? AND v0 = ?", "p", pkg.UserSubDefaultAuthStr).Count(&childPolicies).Error; err != nil || childPolicies != 0 {
		t.Fatalf("child direct policies=%d err=%v", childPolicies, err)
	}
	var unsafeUserPolicies int64
	if err = db.Model(&adapter.CasbinRule{}).Where("ptype = ? AND v0 = ? AND v1 IN ?", "p", pkg.UserDefaultAuthStr, []string{"/api/k8s/pod/webshell", "/api/k8s/deployment/scale"}).Count(&unsafeUserPolicies).Error; err != nil || unsafeUserPolicies != 0 {
		t.Fatalf("unsafe read policies=%d err=%v", unsafeUserPolicies, err)
	}
	var grouping int64
	if err = db.Model(&adapter.CasbinRule{}).Where("ptype = ? AND v0 = ? AND v1 = ?", "g", pkg.UserSubDefaultAuthStr, pkg.UserDefaultAuthStr).Count(&grouping).Error; err != nil || grouping != 1 {
		t.Fatalf("child grouping=%d err=%v", grouping, err)
	}

	// 第二次启动不能恢复管理员已经调整过的权限。
	if err = db.Where("sys_authority_authority_id = ? AND sys_base_menu_id = ?", pkg.UserDefaultAuth, 10001).Delete(&SysAuthorityMenu{}).Error; err != nil {
		t.Fatal(err)
	}
	customPolicy := adapter.CasbinRule{Ptype: "p", V0: pkg.UserDefaultAuthStr, V1: "/custom/read", V2: "GET"}
	if err = db.Create(&customPolicy).Error; err != nil {
		t.Fatal(err)
	}
	if err = migration.InitData(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	assertDirectMenuCount(t, db, pkg.UserDefaultAuth, len(SysBaseMenuEntities)-4)
	var customPolicyCount int64
	if err = db.Model(&adapter.CasbinRule{}).Where("ptype = ? AND v0 = ? AND v1 = ?", "p", pkg.UserDefaultAuthStr, "/custom/read").Count(&customPolicyCount).Error; err != nil || customPolicyCount != 1 {
		t.Fatalf("second migration overwrote custom policy: count=%d err=%v", customPolicyCount, err)
	}
}

func assertDirectMenuCount(t *testing.T, db *gorm.DB, authorityID uint, want int) {
	t.Helper()
	var count int64
	err := db.Model(&SysAuthorityMenu{}).Where("sys_authority_authority_id = ?", strconv.Itoa(int(authorityID))).Count(&count).Error
	if err != nil || count != int64(want) {
		t.Fatalf("role %d direct menu count=%d want=%d err=%v", authorityID, count, want, err)
	}
}
