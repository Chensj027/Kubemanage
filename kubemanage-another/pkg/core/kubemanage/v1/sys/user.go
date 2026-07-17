package sys

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"

	"github.com/noovertime7/kubemanage/dao"
	"github.com/noovertime7/kubemanage/dao/model"
	"github.com/noovertime7/kubemanage/dto"
	"github.com/noovertime7/kubemanage/pkg"
	"github.com/pkg/errors"
)

// UserServiceGetter UserService对象获取器
type UserServiceGetter interface {
	// User 获取一个新的UserService对象
	User() UserService
}

// UserService User相关api操作的Service方法
type UserService interface {
	// Login 用户登陆
	Login(ctx *gin.Context, userInfo *dto.AdminLoginInput) (string, error)
	// LoginOut 用户登出
	LoginOut(ctx *gin.Context, uid int) error
	// GetUserInfo 获取用户信息
	GetUserInfo(ctx *gin.Context, uid int, aid uint) (*dto.UserInfoOut, error)
	// SetUserAuth 更改用户权限
	SetUserAuth(ctx *gin.Context, uid int, aid uint) error
	// DeleteUser 删除用户
	DeleteUser(ctx *gin.Context, uid int) error
	// ChangePassword 更改用户密码
	ChangePassword(ctx *gin.Context, uid int, info *dto.ChangeUserPwdInput) error
	// ResetPassword 重置用户密码
	ResetPassword(ctx *gin.Context, uid int, newPassword string) error
	List(ctx context.Context, page, pageSize int, keyword string) (*dto.UserListOut, error)
	Create(ctx context.Context, in *dto.UserCreateInput) (*model.SysUser, error)
	Update(ctx context.Context, uid int, in *dto.UserUpdateInput) error
	SetEnable(ctx context.Context, uid int, enable int) error
	ValidateClaims(ctx context.Context, claims *pkg.CustomClaims) error
}

// userService UserService接口的实现类（在user操作相关的接口中，需要用到哪个service，就在这里内置什么Service）
type userService struct {
	// Menu userService中，需要内置一个MenuService对象，用于操作Menu
	Menu MenuService
	// Casbin userService中，需要内置一个CasbinService对象，用于操作Casbin
	Casbin CasbinService
	// factory userService中，需要内置一个db工厂对象，用于操作数据库
	factory dao.ShareDaoFactory
}

// NewUserService 新建一个 *userService 对象
func NewUserService(factory dao.ShareDaoFactory) *userService {
	return &userService{
		// factory
		factory: factory,
		//
		Menu:   NewMenuService(factory),
		Casbin: NewCasbinService(factory),
	}
}

var _ UserService = &userService{}

// Login 用户登陆
func (u *userService) Login(ctx *gin.Context, userInfo *dto.AdminLoginInput) (string, error) {
	// 从userService中，取出内置的db工厂对象，先User()获取一个user.User对象，然后调用Find查询username符合条件的User
	user, err := u.factory.User().Find(ctx, &model.SysUser{UserName: userInfo.UserName})
	if err != nil {
		// 查询出错了，返回错误
		return "", err
	}

	// 查询到了user，进行密码比对
	if !pkg.CheckPassword(userInfo.Password, user.Password) {
		return "", errors.New("密码错误,请重新输入")
	}
	if user.Enable != 1 {
		return "", errors.New("用户已被冻结")
	}
	_ = u.factory.GetDB().WithContext(ctx).Model(&model.SysUser{}).Where("id = ?", user.ID).Update("status", sql.NullInt64{Int64: 1, Valid: true}).Error

	// 使用JWT生成token
	token, err := pkg.JWTToken.GenerateToken(pkg.BaseClaims{
		UUID:         user.UUID,
		ID:           user.ID,
		Username:     user.UserName,
		NickName:     user.NickName,
		AuthorityId:  user.AuthorityId,
		TokenVersion: user.TokenVersion,
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

func (u *userService) LoginOut(ctx *gin.Context, uid int) error {
	// 创建一个SysUser对象，设置 id + status=0（TODO sql.NullInt64方法仔细看看）
	var current model.SysUser
	if err := u.factory.GetDB().WithContext(ctx).First(&current, uid).Error; err != nil {
		return err
	}
	user := &model.SysUser{ID: uid, Status: sql.NullInt64{Int64: 0, Valid: true}, TokenVersion: current.TokenVersion + 1}
	// 从userService中取出db factory，从中创建一个user.User，调用User表的Updates方法，完成更新
	return u.factory.User().Updates(ctx, user)
}

func (u *userService) GetUserInfo(ctx *gin.Context, uid int, aid uint) (*dto.UserInfoOut, error) {
	// 从userService中取出db factory，从中创建一个user.User，然后根据userId查询用户
	user, err := u.factory.User().Find(ctx, &model.SysUser{ID: uid})
	if err != nil {
		return nil, err
	}
	// 从userService中取出内置的MenuService对象，根据user的authorityId，查询用户的menu列表
	menus, err := u.Menu.GetMenuByAuthorityID(ctx, aid)
	if err != nil {
		return nil, err
	}
	var outRules []string
	rules := u.Casbin.GetImplicitPolicyPathByAuthorityId(aid)
	for _, rule := range rules {
		item := rule.Path + "," + rule.Method
		outRules = append(outRules, item)
	}
	return &dto.UserInfoOut{
		User:      *user,
		Menus:     menus,
		RuleNames: outRules,
	}, nil
}

func (u *userService) SetUserAuth(ctx *gin.Context, uid int, aid uint) error {
	if uid == 1 && aid != pkg.AdminDefaultAuth {
		return errors.New("不能降级管理员")
	}
	var role model.SysAuthority
	if e := u.factory.GetDB().WithContext(ctx).Where("authority_id = ?", aid).First(&role).Error; e != nil {
		return errors.New("角色不存在")
	}
	user := &model.SysUser{ID: uid, AuthorityId: aid, TokenVersion: 1}
	var old model.SysUser
	if err := u.factory.GetDB().WithContext(ctx).First(&old, uid).Error; err != nil {
		return err
	}
	user.TokenVersion = old.TokenVersion + 1
	return u.factory.User().Updates(ctx, user)
}

func (u *userService) DeleteUser(ctx *gin.Context, uid int) error {
	if uid == 1 {
		return errors.New("不能删除管理员")
	}
	user := &model.SysUser{ID: uid}
	return u.factory.User().Delete(ctx, user)
}

func (u *userService) List(ctx context.Context, page, pageSize int, keyword string) (*dto.UserListOut, error) {
	var users []model.SysUser
	var total int64
	q := u.factory.GetDB().WithContext(ctx).Model(&model.SysUser{})
	if keyword != "" {
		q = q.Where("user_name LIKE ? OR nick_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	q.Count(&total)
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	err := q.Preload("Authority").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return &dto.UserListOut{List: users, Total: total}, err
}
func (u *userService) Create(ctx context.Context, in *dto.UserCreateInput) (*model.SysUser, error) {
	in.UserName = strings.TrimSpace(in.UserName)
	if in.UserName == "" {
		return nil, errors.New("用户名不能为空")
	}
	if len(in.Password) < 6 || len(in.Password) > 72 {
		return nil, errors.New("密码长度须为6-72位")
	}
	var dup model.SysUser
	if e := u.factory.GetDB().WithContext(ctx).Where("user_name = ?", in.UserName).First(&dup).Error; e == nil {
		return nil, errors.New("用户名已存在")
	} else if !errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, e
	}
	p, e := pkg.GenSaltPassword(in.Password)
	if e != nil {
		return nil, e
	}
	aid := in.AuthorityId
	if aid == 0 {
		aid = 222
	}
	var role model.SysAuthority
	if e := u.factory.GetDB().WithContext(ctx).Where("authority_id = ?", aid).First(&role).Error; e != nil {
		return nil, errors.New("角色不存在")
	}
	en := in.Enable
	if en == 0 {
		en = 1
	}
	if en != 1 && en != 2 {
		return nil, errors.New("用户状态只能是1或2")
	}
	if strings.TrimSpace(in.NickName) == "" {
		in.NickName = in.UserName
	}
	user := &model.SysUser{UUID: uuid.NewV4(), UserName: in.UserName, Password: p, NickName: strings.TrimSpace(in.NickName), Phone: strings.TrimSpace(in.Phone), Email: strings.TrimSpace(in.Email), AuthorityId: aid, Enable: en, TokenVersion: 1}
	return user, u.factory.GetDB().WithContext(ctx).Create(user).Error
}
func (u *userService) Update(ctx context.Context, uid int, in *dto.UserUpdateInput) error {
	var user model.SysUser
	if e := u.factory.GetDB().WithContext(ctx).First(&user, uid).Error; e != nil {
		return e
	}
	if in.NickName != "" {
		user.NickName = in.NickName
	}
	user.Phone = in.Phone
	user.Email = in.Email
	if in.AuthorityId != 0 && in.AuthorityId != user.AuthorityId {
		if uid == 1 {
			return errors.New("不能降级管理员")
		}
		var role model.SysAuthority
		if e := u.factory.GetDB().WithContext(ctx).Where("authority_id = ?", in.AuthorityId).First(&role).Error; e != nil {
			return errors.New("角色不存在")
		}
		user.AuthorityId = in.AuthorityId
		user.TokenVersion++
	}
	if in.Enable != nil && *in.Enable != user.Enable {
		if *in.Enable != 1 && *in.Enable != 2 {
			return errors.New("用户状态只能是1或2")
		}
		if uid == 1 && *in.Enable != 1 {
			return errors.New("不能冻结管理员")
		}
		user.Enable = *in.Enable
		user.TokenVersion++
	}
	return u.factory.GetDB().WithContext(ctx).Save(&user).Error
}
func (u *userService) SetEnable(ctx context.Context, uid int, enable int) error {
	if enable != 1 && enable != 2 {
		return errors.New("用户状态只能是1或2")
	}
	if uid == 1 && enable != 1 {
		return errors.New("不能冻结管理员")
	}
	return u.Update(ctx, uid, &dto.UserUpdateInput{Enable: &enable})
}
func (u *userService) ValidateClaims(ctx context.Context, claims *pkg.CustomClaims) error {
	var user model.SysUser
	if err := u.factory.GetDB().WithContext(ctx).First(&user, claims.ID).Error; err != nil {
		return errors.New("用户不存在或已删除")
	}
	if user.Enable != 1 {
		return errors.New("用户已被冻结")
	}
	if user.UUID != claims.UUID || user.AuthorityId != claims.AuthorityId || user.TokenVersion != claims.TokenVersion {
		return errors.New("登录状态已失效，请重新登录")
	}
	return nil
}

func (u *userService) ChangePassword(ctx *gin.Context, uid int, info *dto.ChangeUserPwdInput) error {
	if len(info.NewPwd) < 6 || len(info.NewPwd) > 72 {
		return errors.New("密码长度须为6-72位")
	}
	userDB := &model.SysUser{ID: uid}
	user, err := u.factory.User().Find(ctx, userDB)
	if err != nil {
		return err
	}

	if !pkg.CheckPassword(info.OldPwd, user.Password) {
		return errors.New("原密码错误,请重新输入")
	}

	//生成新密码
	user.Password, err = pkg.GenSaltPassword(info.NewPwd)
	if err != nil {
		return err
	}
	user.TokenVersion++
	return u.factory.User().Updates(ctx, user)
}

func (u *userService) ResetPassword(ctx *gin.Context, uid int, newPassword string) error {
	if len(newPassword) < 6 || len(newPassword) > 72 {
		return errors.New("密码长度须为6-72位")
	}
	newPwd, err := pkg.GenSaltPassword(newPassword)
	if err != nil {
		return err
	}
	var current model.SysUser
	if err := u.factory.GetDB().WithContext(ctx).First(&current, uid).Error; err != nil {
		return err
	}
	user := &model.SysUser{ID: uid, Password: newPwd, TokenVersion: current.TokenVersion + 1}
	return u.factory.User().Updates(ctx, user)
}
