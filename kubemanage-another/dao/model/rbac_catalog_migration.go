package model

import (
	"context"
	"fmt"
	"strconv"
	"time"

	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/noovertime7/kubemanage/pkg"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const rbacCatalogV2Migration = "rbac-catalog-v2"

var legacyMenuPaths = []string{
	"dashboard", "cmdb", "kubernetes", "devops", "setting", "host", "secret",
	"cluster", "deployment", "service", "node", "config", "events", "authority",
	"user", "operation",
}

type dataMigration struct {
	Name      string    `gorm:"column:name;primaryKey;size:128"`
	AppliedAt time.Time `gorm:"column:applied_at;not null"`
}

func (*dataMigration) TableName() string { return "sys_data_migrations" }

type rbacCatalogMigration struct{}

func init() {
	RegisterInitializer(RBACCatalogMigrationOrder, &rbacCatalogMigration{})
}

func (*rbacCatalogMigration) TableName() string { return (&dataMigration{}).TableName() }

func (*rbacCatalogMigration) MigrateTable(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(&dataMigration{})
}

func (*rbacCatalogMigration) IsInitData(ctx context.Context, db *gorm.DB) (bool, error) {
	var count int64
	err := db.WithContext(ctx).Model(&dataMigration{}).Where("name = ?", rbacCatalogV2Migration).Count(&count).Error
	return count > 0, err
}

func (m *rbacCatalogMigration) InitData(ctx context.Context, db *gorm.DB) error {
	done, err := m.IsInitData(ctx, db)
	if err != nil || done {
		return err
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 在事务开始时抢占唯一迁移标记，并发启动时只有一个实例能够执行迁移。
		// 迁移失败会回滚事务和标记，后续启动仍可安全重试。
		marker := dataMigration{Name: rbacCatalogV2Migration, AppliedAt: time.Now()}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&marker)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}

		if err := migrateBuiltinRoles(tx); err != nil {
			return err
		}
		if err := migrateMenuCatalog(tx); err != nil {
			return err
		}
		if err := migrateBuiltinPolicies(tx); err != nil {
			return err
		}
		return nil
	})
}

func (*rbacCatalogMigration) TableCreated(ctx context.Context, db *gorm.DB) bool {
	return db.WithContext(ctx).Migrator().HasTable(&dataMigration{})
}

func migrateBuiltinRoles(tx *gorm.DB) error {
	for _, role := range SysAuthorityEntities {
		item := role
		if err := tx.Where("authority_id = ?", role.AuthorityId).
			Assign(map[string]interface{}{
				"authority_name": role.AuthorityName,
				"parent_id":      role.ParentId,
				"default_router": role.DefaultRouter,
			}).FirstOrCreate(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateMenuCatalog(tx *gorm.DB) error {
	canonicalPaths := make([]string, 0, len(SysBaseMenuEntities))
	canonicalIDs := make(map[int]struct{}, len(SysBaseMenuEntities))
	for _, menu := range SysBaseMenuEntities {
		canonicalPaths = append(canonicalPaths, menu.Path)
		canonicalIDs[menu.ID] = struct{}{}
	}
	var reserved []SysBaseMenu
	if err := tx.Unscoped().Where("id IN ?", mapKeys(canonicalIDs)).Find(&reserved).Error; err != nil {
		return err
	}
	canonicalPathByID := make(map[int]string, len(SysBaseMenuEntities))
	for _, menu := range SysBaseMenuEntities {
		canonicalPathByID[menu.ID] = menu.Path
	}
	for _, menu := range reserved {
		if expected := canonicalPathByID[menu.ID]; menu.Path != expected {
			return fmt.Errorf("内置菜单ID %d已被路由 %s 占用", menu.ID, menu.Path)
		}
	}

	var obsolete []SysBaseMenu
	if err := tx.Unscoped().Where("path IN ?", legacyMenuPaths).Or("path IN ?", canonicalPaths).Find(&obsolete).Error; err != nil {
		return err
	}
	obsoleteIDs := make([]int, 0, len(obsolete))
	for _, menu := range obsolete {
		if _, canonical := canonicalIDs[menu.ID]; !canonical || menu.DeletedAt.Valid {
			obsoleteIDs = append(obsoleteIDs, menu.ID)
		}
	}
	if len(obsoleteIDs) > 0 {
		if err := tx.Where("sys_base_menu_id IN ?", obsoleteIDs).Delete(&SysAuthorityMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("id IN ?", obsoleteIDs).Delete(&SysBaseMenu{}).Error; err != nil {
			return err
		}
	}

	for _, menu := range SysBaseMenuEntities {
		item := menu
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"menu_level", "parent_id", "path", "name", "hidden", "disabled", "sort",
				"active_name", "keep_alive", "title", "icon", "close_tab", "deleted_at", "updated_at",
			}),
		}).Create(&item).Error; err != nil {
			return err
		}
	}

	var allMenus []SysBaseMenu
	if err := tx.Order("id").Find(&allMenus).Error; err != nil {
		return err
	}
	userMenuIDs := make([]int, 0, len(SysBaseMenuEntities))
	for _, menu := range SysBaseMenuEntities {
		if menu.Path != "/system" && menu.ParentId != "10019" {
			userMenuIDs = append(userMenuIDs, menu.ID)
		}
	}
	var userMenus []SysBaseMenu
	if err := tx.Where("id IN ?", userMenuIDs).Order("id").Find(&userMenus).Error; err != nil {
		return err
	}
	roles := map[uint][]SysBaseMenu{
		pkg.AdminDefaultAuth:   allMenus,
		pkg.UserDefaultAuth:    userMenus,
		pkg.UserSubDefaultAuth: nil,
	}
	for authorityID, menus := range roles {
		if err := replaceRoleMenus(tx, authorityID, menus); err != nil {
			return err
		}
	}
	return nil
}

func replaceRoleMenus(tx *gorm.DB, authorityID uint, menus []SysBaseMenu) error {
	roleID := strconv.Itoa(int(authorityID))
	if err := tx.Where("sys_authority_authority_id = ?", roleID).Delete(&SysAuthorityMenu{}).Error; err != nil {
		return err
	}
	links := make([]SysAuthorityMenu, 0, len(menus))
	for _, menu := range menus {
		links = append(links, SysAuthorityMenu{MenuId: strconv.Itoa(menu.ID), AuthorityId: roleID})
	}
	if len(links) == 0 {
		return nil
	}
	return tx.Create(&links).Error
}

func mapKeys(values map[int]struct{}) []int {
	keys := make([]int, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func migrateBuiltinPolicies(tx *gorm.DB) error {
	builtinIDs := []string{pkg.AdminDefaultAuthStr, pkg.UserDefaultAuthStr, pkg.UserSubDefaultAuthStr}
	if err := tx.Where("ptype = ? AND v0 IN ?", "p", builtinIDs).Delete(&adapter.CasbinRule{}).Error; err != nil {
		return err
	}
	if err := tx.Where("ptype = ? AND v0 IN ?", "g", builtinIDs).Delete(&adapter.CasbinRule{}).Error; err != nil {
		return err
	}
	rules := buildCasbinRule(SysApis)
	rules = append(rules, adapter.CasbinRule{
		Ptype: "g",
		V0:    strconv.Itoa(int(pkg.UserSubDefaultAuth)),
		V1:    strconv.Itoa(int(pkg.UserDefaultAuth)),
	})
	return tx.Create(&rules).Error
}
