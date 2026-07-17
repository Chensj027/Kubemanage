package authority

import (
	"context"
	"fmt"
	"github.com/noovertime7/kubemanage/dao/model"
	"github.com/noovertime7/kubemanage/dto"
	"gorm.io/gorm"
)

type Authority interface {
	Find(ctx context.Context, authInfo *model.SysAuthority) (*model.SysAuthority, error)
	FindList(ctx context.Context, authInfo *model.SysAuthority) ([]*model.SysAuthority, error)
	Save(ctx context.Context, authInfo *model.SysAuthority) error
	Updates(ctx context.Context, authInfo *model.SysAuthority) error
	Delete(ctx context.Context, authorityId uint) error

	SetMenuAuthority(ctx context.Context, authInfo *model.SysAuthority) error
	PageList(ctx context.Context, params dto.PageInfo) ([]model.SysAuthority, int64, error)
}

var _ Authority = &authority{}

type authority struct {
	db *gorm.DB
}

func NewAuthority(db *gorm.DB) *authority {
	return &authority{db: db}
}

func (a *authority) Find(ctx context.Context, authInfo *model.SysAuthority) (*model.SysAuthority, error) {
	out := &model.SysAuthority{}
	return out, a.db.WithContext(ctx).Where(authInfo).First(out).Error
}

func (a *authority) FindList(ctx context.Context, authInfo *model.SysAuthority) ([]*model.SysAuthority, error) {
	var out []*model.SysAuthority
	return out, a.db.WithContext(ctx).Where(&authInfo).Find(&out).Error
}

func (a *authority) PageList(ctx context.Context, params dto.PageInfo) ([]model.SysAuthority, int64, error) {
	var total int64 = 0
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 200 {
		params.PageSize = 20
	}
	limit := params.PageSize
	offset := params.PageSize * (params.Page - 1)
	query := a.db.WithContext(ctx)
	var list []model.SysAuthority
	// 如果有条件搜索 下方会自动创建搜索语句
	if params.Keyword != "" {
		query = query.Where("authority_name = ?", params.Keyword)
	}
	if err := query.Find(&list).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("authority_id desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (a *authority) Save(ctx context.Context, authInfo *model.SysAuthority) error {
	return a.db.WithContext(ctx).Create(authInfo).Error
}

func (a *authority) Updates(ctx context.Context, authInfo *model.SysAuthority) error {
	return a.db.WithContext(ctx).Model(&model.SysAuthority{}).Where("authority_id = ?", authInfo.AuthorityId).
		Select("authority_name", "parent_id", "default_router").Updates(authInfo).Error
}
func (a *authority) Delete(ctx context.Context, authorityId uint) error {
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var n int64
		if e := tx.Model(&model.SysUser{}).Where("authority_id = ?", authorityId).Count(&n).Error; e != nil {
			return e
		}
		if n > 0 {
			return fmt.Errorf("角色仍被 %d 个用户使用", n)
		}
		if e := tx.Model(&model.SysAuthority{}).Where("parent_id = ?", authorityId).Count(&n).Error; e != nil {
			return e
		}
		if n > 0 {
			return fmt.Errorf("角色仍被 %d 个子角色继承", n)
		}
		if e := tx.Where("sys_authority_authority_id = ?", authorityId).Delete(&model.SysAuthorityMenu{}).Error; e != nil {
			return e
		}
		return tx.Where("authority_id = ?", authorityId).Delete(&model.SysAuthority{}).Error
	})
}

// SetMenuAuthority 菜单与角色绑定
func (a *authority) SetMenuAuthority(ctx context.Context, authInfo *model.SysAuthority) error {
	var s model.SysAuthority
	a.db.WithContext(ctx).Preload("SysBaseMenus").First(&s, "authority_id = ?", authInfo.AuthorityId)
	return a.db.WithContext(ctx).Model(&s).Association("SysBaseMenus").Replace(&authInfo.SysBaseMenus)
}
