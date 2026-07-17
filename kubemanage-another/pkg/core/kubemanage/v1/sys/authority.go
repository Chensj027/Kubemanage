package sys

import (
	"context"
	"strings"

	"github.com/noovertime7/kubemanage/dto"
	"github.com/pkg/errors"

	"github.com/noovertime7/kubemanage/dao"
	"github.com/noovertime7/kubemanage/dao/model"
	"github.com/noovertime7/kubemanage/pkg"
)

type AuthorityGetter interface {
	Authority() Authority
}

type Authority interface {
	SetMenuAuthority(ctx context.Context, auth *model.SysAuthority) error
	GetAuthorityList(ctx context.Context, pageInfo dto.PageInfo) (*dto.AuthorityList, error)
	Create(ctx context.Context, in *dto.AuthorityInput) error
	Update(ctx context.Context, id uint, in *dto.AuthorityInput) error
	Delete(ctx context.Context, id uint) error
}

func (a *authority) Create(ctx context.Context, in *dto.AuthorityInput) error {
	if in.AuthorityId == 0 || strings.TrimSpace(in.AuthorityName) == "" {
		return errors.New("角色ID和角色名称不能为空")
	}
	if _, err := a.factory.Authority().Find(ctx, &model.SysAuthority{AuthorityId: in.AuthorityId}); err == nil {
		return errors.New("角色ID已存在")
	}
	if err := a.validateParent(ctx, in.AuthorityId, in.ParentId); err != nil {
		return err
	}
	if in.DefaultRouter == "" {
		in.DefaultRouter = "/home"
	}
	if err := a.factory.Authority().Save(ctx, &model.SysAuthority{AuthorityId: in.AuthorityId, AuthorityName: strings.TrimSpace(in.AuthorityName), ParentId: in.ParentId, DefaultRouter: in.DefaultRouter}); err != nil {
		return err
	}
	return NewCasbinService(a.factory).SyncRoleParent(in.AuthorityId, in.ParentId)
}
func (a *authority) Update(ctx context.Context, id uint, in *dto.AuthorityInput) error {
	if id == pkg.AdminDefaultAuth && in.ParentId != 0 {
		return errors.New("管理员角色不能继承其他角色")
	}
	if strings.TrimSpace(in.AuthorityName) == "" {
		return errors.New("角色名称不能为空")
	}
	if _, err := a.factory.Authority().Find(ctx, &model.SysAuthority{AuthorityId: id}); err != nil {
		return errors.New("角色不存在")
	}
	if err := a.validateParent(ctx, id, in.ParentId); err != nil {
		return err
	}
	if in.DefaultRouter == "" {
		in.DefaultRouter = "/home"
	}
	if err := a.factory.Authority().Updates(ctx, &model.SysAuthority{AuthorityId: id, AuthorityName: strings.TrimSpace(in.AuthorityName), ParentId: in.ParentId, DefaultRouter: in.DefaultRouter}); err != nil {
		return err
	}
	return NewCasbinService(a.factory).SyncRoleParent(id, in.ParentId)
}
func (a *authority) Delete(ctx context.Context, id uint) error {
	if id == pkg.AdminDefaultAuth {
		return errors.New("管理员角色不能删除")
	}
	if err := a.factory.Authority().Delete(ctx, id); err != nil {
		return err
	}
	return NewCasbinService(a.factory).RemoveRole(id)
}

func (a *authority) validateParent(ctx context.Context, authorityID, parentID uint) error {
	if parentID == 0 {
		return nil
	}
	if parentID == authorityID {
		return errors.New("角色不能继承自己")
	}
	visited := map[uint]struct{}{authorityID: {}}
	current := parentID
	for current != 0 {
		if _, exists := visited[current]; exists {
			return errors.New("角色继承关系不能形成循环")
		}
		visited[current] = struct{}{}
		var role model.SysAuthority
		if err := a.factory.GetDB().WithContext(ctx).Select("authority_id", "parent_id").Where("authority_id = ?", current).First(&role).Error; err != nil {
			return errors.New("父角色不存在")
		}
		current = role.ParentId
	}
	return nil
}

type authority struct {
	factory dao.ShareDaoFactory
}

func NewAuthority(factory dao.ShareDaoFactory) *authority {
	return &authority{factory: factory}
}

func (a *authority) SetMenuAuthority(ctx context.Context, auth *model.SysAuthority) error {
	return a.factory.Authority().SetMenuAuthority(ctx, auth)
}

func (a *authority) GetAuthorityList(ctx context.Context, pageInfo dto.PageInfo) (*dto.AuthorityList, error) {
	list, total, err := a.factory.Authority().PageList(ctx, pageInfo)
	if err != nil {
		return nil, err
	}
	return &dto.AuthorityList{
		PageInfo:          pageInfo,
		Total:             total,
		AuthorityListItem: list,
	}, nil
}
