package model

import (
	"context"
	"github.com/noovertime7/kubemanage/pkg"
	"gorm.io/gorm"
)

func init() {
	RegisterInitializer(MenuAuthorityOrder, &MenuAuthority{})
}

type MenuAuthority struct{}

func (i MenuAuthority) TableName() string {
	return "sys_menu_authorities"
}

func (i *MenuAuthority) MigrateTable(ctx context.Context, db *gorm.DB) error {
	//数据由gorm填充不需要手动迁移
	return nil
}

func (i *MenuAuthority) InitData(ctx context.Context, db *gorm.DB) error {
	var (
		adminRole   = SysAuthorityEntities[0]
		userRole    = SysAuthorityEntities[1]
		userSubRole = SysAuthorityEntities[2]
		err         error
		ok          bool
	)
	ok, err = i.IsInitData(ctx, db)
	if err != nil || ok {
		return err
	}
	// 管理员拥有全部菜单；普通用户拥有非系统菜单；
	// 普通用户子角色不直接授予菜单，而是继承角色 222。
	if err = db.WithContext(ctx).Model(&adminRole).Association("SysBaseMenus").Replace(SysBaseMenuEntities); err != nil {
		return err
	}
	userMenus := make([]SysBaseMenu, 0, len(SysBaseMenuEntities))
	for _, menu := range SysBaseMenuEntities {
		if menu.Path != "/system" && menu.ParentId != "10019" {
			userMenus = append(userMenus, menu)
		}
	}
	if err = db.WithContext(ctx).Model(&userRole).Association("SysBaseMenus").Replace(userMenus); err != nil {
		return err
	}
	if err = db.WithContext(ctx).Model(&userSubRole).Association("SysBaseMenus").Clear(); err != nil {
		return err
	}
	return nil
}

func (i *MenuAuthority) IsInitData(ctx context.Context, db *gorm.DB) (bool, error) {
	auth := &SysAuthority{}
	if err := db.WithContext(ctx).Model(auth).Where("authority_id = ?", pkg.AdminDefaultAuth).Preload("SysBaseMenus").Find(auth).Error; err != nil {
		return false, err
	}
	return len(auth.SysBaseMenus) > 0, nil
}

func (i *MenuAuthority) TableCreated(ctx context.Context, db *gorm.DB) bool {
	return false // always replace
}
