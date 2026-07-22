package model

import (
	"context"
	"strconv"
	"time"

	"github.com/noovertime7/kubemanage/pkg"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const grafanaMenuV1Migration = "grafana-menu-v1"

// grafanaMenuIDs 是本次迁移涉及的「监控」目录与 Grafana 入口菜单 ID，
// 其字段定义以 init.go 的 SysBaseMenuEntities（ID 10022/10023）为准。
var grafanaMenuIDs = []int{10022, 10023}

// grafanaMenuMigration 为「存量库」增量补齐 监控/Grafana 菜单。
//
// 背景：SysBaseMenu.InitData（空表才插）与 rbacCatalogMigration（标记 rbac-catalog-v2，
// 一次性）都是一次性播种器，对已初始化的库不会再插入新菜单。新装环境由这两者直接覆盖
// （replaceRoleMenus 会把新菜单发给内置角色），因此本迁移在新装环境是幂等空操作；
// 只有存量库会真正执行插入，并仅为管理员追加菜单关联，其它角色由管理员在
// 「系统管理/角色权限」页面自助勾选。
type grafanaMenuMigration struct{}

func init() {
	RegisterInitializer(GrafanaMenuMigrationOrder, &grafanaMenuMigration{})
}

func (*grafanaMenuMigration) TableName() string { return (&dataMigration{}).TableName() }

func (*grafanaMenuMigration) MigrateTable(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(&dataMigration{})
}

func (*grafanaMenuMigration) IsInitData(ctx context.Context, db *gorm.DB) (bool, error) {
	var count int64
	err := db.WithContext(ctx).Model(&dataMigration{}).Where("name = ?", grafanaMenuV1Migration).Count(&count).Error
	return count > 0, err
}

func (*grafanaMenuMigration) TableCreated(ctx context.Context, db *gorm.DB) bool {
	return db.WithContext(ctx).Migrator().HasTable(&dataMigration{})
}

func (m *grafanaMenuMigration) InitData(ctx context.Context, db *gorm.DB) error {
	done, err := m.IsInitData(ctx, db)
	if err != nil || done {
		return err
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 事务内抢占迁移标记，保证并发启动只执行一次；失败回滚可安全重试。
		marker := dataMigration{Name: grafanaMenuV1Migration, AppliedAt: time.Now()}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&marker)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}

		// 1. upsert 监控目录与 Grafana 入口菜单（字段以 SysBaseMenuEntities 为准）。
		for _, menu := range grafanaMenuEntities() {
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

		// 2. 仅为管理员追加菜单关联（不删除既有关联、不影响其它角色）。
		adminID := pkg.AdminDefaultAuthStr
		for _, id := range grafanaMenuIDs {
			link := SysAuthorityMenu{MenuId: strconv.Itoa(id), AuthorityId: adminID}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&link).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// grafanaMenuEntities 从共享目录 SysBaseMenuEntities 中挑出本次迁移涉及的菜单，
// 避免在此重复声明菜单字段。
func grafanaMenuEntities() []SysBaseMenu {
	want := make(map[int]struct{}, len(grafanaMenuIDs))
	for _, id := range grafanaMenuIDs {
		want[id] = struct{}{}
	}
	out := make([]SysBaseMenu, 0, len(grafanaMenuIDs))
	for _, menu := range SysBaseMenuEntities {
		if _, ok := want[menu.ID]; ok {
			out = append(out, menu)
		}
	}
	return out
}
