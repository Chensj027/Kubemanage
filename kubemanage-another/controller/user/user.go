package user

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/noovertime7/kubemanage/pkg/core/kubemanage/v1"

	"github.com/noovertime7/kubemanage/dto"
	"github.com/noovertime7/kubemanage/middleware"
	"github.com/noovertime7/kubemanage/pkg"
	"github.com/noovertime7/kubemanage/pkg/globalError"
	"github.com/noovertime7/kubemanage/pkg/utils"
)

func (u *userController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	out, e := v1.CoreV1.System().User().List(c, page, size, c.Query("keyword"))
	if e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ServerError, e))
		return
	}
	middleware.ResponseSuccess(c, out)
}
func (u *userController) Create(c *gin.Context) {
	var in dto.UserCreateInput
	if e := c.ShouldBindJSON(&in); e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ParamBindError, e))
		return
	}
	out, e := v1.CoreV1.System().User().Create(c, &in)
	if e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ServerError, e))
		return
	}
	middleware.ResponseSuccess(c, out)
}
func (u *userController) Update(c *gin.Context) {
	id, e := strconv.Atoi(c.Param("id"))
	if e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ParamBindError, e))
		return
	}
	var in dto.UserUpdateInput
	if e = c.ShouldBindJSON(&in); e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ParamBindError, e))
		return
	}
	claims := utils.GetUserInfo(c)
	if claims != nil && claims.ID == id {
		if (in.Enable != nil && *in.Enable != 1) || (in.AuthorityId != 0 && in.AuthorityId != claims.AuthorityId) {
			middleware.ResponseError(c, globalError.NewGlobalError(globalError.AuthErr, fmt.Errorf("不能冻结当前用户或修改当前用户角色")))
			return
		}
	}
	if e = v1.CoreV1.System().User().Update(c, id, &in); e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ServerError, e))
		return
	}
	middleware.ResponseSuccess(c, "操作成功")
}
func (u *userController) Enable(c *gin.Context) {
	id, e := strconv.Atoi(c.Param("id"))
	if e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ParamBindError, e))
		return
	}
	var in struct {
		Enable int `json:"enable"`
	}
	if e = c.ShouldBindJSON(&in); e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ParamBindError, e))
		return
	}
	if claims := utils.GetUserInfo(c); claims != nil && claims.ID == id && in.Enable != 1 {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.AuthErr, fmt.Errorf("不能冻结当前登录用户")))
		return
	}
	if e = v1.CoreV1.System().User().SetEnable(c, id, in.Enable); e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ServerError, e))
		return
	}
	middleware.ResponseSuccess(c, "操作成功")
}

// Login godoc
// @Summary 管理员登录
// @Description 管理员登录
// @Tags 管理员接口
// @ID /user/login
// @Accept  json
// @Produce  json
// @Param polygon body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOut} "success"
// @Router /api/user/login [post]
func (u *userController) Login(ctx *gin.Context) {
	// 创建一个 request空对象，用来接收 接口请求数据
	params := &dto.AdminLoginInput{}
	// 将ctx中的数据与params对象进行绑定，同时对数据进行校验
	if err := params.BindingValidParams(ctx); err != nil {
		// 校验或绑定失败，日志记录错误信息
		v1.Log.ErrorWithCode(globalError.ParamBindError, err)
		// 响应失败的公共处理，同时就会封装好响应体，设置给context
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		// 响应体已经设置好了，因此可以直接返回了
		return
	}
	// 使用全局的v1.CoreV1对象，获取一个SystemInterface接口的对象，然后获取一个UserService对象，调用Login方法
	token, err := v1.CoreV1.System().User().Login(ctx, params)
	if err != nil {
		// 登陆失败，日志记录错误信息
		v1.Log.ErrorWithCode(globalError.LoginErr, err)
		// 响应失败的公共处理，同时就会封装好响应体，设置给context
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.LoginErr, err))
		// 响应体已经设置好了，因此可以直接返回了
		return
	}
	// 响应成功的公共处理，同时就会封装好响应体，设置给context
	middleware.ResponseSuccess(ctx, &dto.AdminLoginOut{Token: token})
}

// LoginOut godoc
// @Summary 管理员退出登录
// @Description 管理员登录
// @Tags 管理员接口
// @ID /user/loginout
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOut} "success"
// @Router /api/user/loginout [get]
func (u *userController) LoginOut(ctx *gin.Context) {
	// 从context中，取出claims的值。如果不存在，则说明没有这个用户的登录信息，报错内部错误
	claims, exists := ctx.Get("claims")
	if !exists {
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.AuthorizationError, fmt.Errorf("登录信息不存在")))
		return
	}
	// 断言claims是否为CustomClaims类型，转成CustomClaims类型
	cla, ok := claims.(*pkg.CustomClaims)
	if !ok || cla == nil {
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.AuthorizationError, fmt.Errorf("登录信息无效")))
		return
	}
	// 使用全局的v1.CoreV1对象，获取一个SystemInterface接口的对象，然后获取一个UserService对象，调用LoginOut方法
	if err := v1.CoreV1.System().User().LoginOut(ctx, cla.ID); err != nil {
		// 登出失败，日志记录错误信息
		v1.Log.ErrorWithCode(globalError.LogoutErr, err)
		// 响应失败的公共处理，同时就会封装好响应体，设置给context
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		// 响应体已经设置好了，因此可以直接返回了
		return
	}
	// 响应成功的公共处理，同时就会封装好响应体，设置给context
	middleware.ResponseSuccess(ctx, "退出成功")
}

// GetUserInfo
// @Tags      SysUser
// @Summary   获取用户信息
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  middleware.Response{data=model.SysUser,msg=string}  "获取用户信息"
// @Router    /api/user/getinfo [get]
func (u *userController) GetUserInfo(ctx *gin.Context) {
	// 从token中，取出claims值
	clalms, err := utils.GetClaims(ctx)
	if err != nil {
		// 发生错误，说明参数有误，记录错误日志，然后返回错误
		v1.Log.ErrorWithCode(globalError.ParamBindError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		return
	}
	// 使用全局的v1.CoreV1对象，获取一个SystemInterface接口的对象，然后获取一个UserService对象，调用GetUserInfo方法
	userInfo, err := v1.CoreV1.System().User().GetUserInfo(ctx, clalms.ID, clalms.AuthorityId)
	if err != nil {
		v1.Log.ErrorWithCode(globalError.ParamBindError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		return
	}
	middleware.ResponseSuccess(ctx, userInfo)
}

// SetUserAuthority
// @Tags      SysUser
// @Summary   更改用户权限
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      dto.SetUserAuth          true  "角色ID"
// @Success   200   {object}  middleware.Response{msg=string}  "设置用户权限"
// @Router    /api/user/{id}/set_auth [put]
func (u *userController) SetUserAuthority(ctx *gin.Context) {
	uid, err := utils.ParseInt(ctx.Param("id"))
	if err != nil {
		v1.Log.ErrorWithCode(globalError.ParamBindError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		return
	}
	params := &dto.SetUserAuth{}
	if err := params.BindingValidParams(ctx); err != nil {
		v1.Log.ErrorWithCode(globalError.ParamBindError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		return
	}
	if claims := utils.GetUserInfo(ctx); claims != nil && claims.ID == uid && claims.AuthorityId != params.AuthorityId {
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.AuthErr, fmt.Errorf("不能修改当前登录用户的角色")))
		return
	}
	if err := v1.CoreV1.System().User().SetUserAuth(ctx, uid, params.AuthorityId); err != nil {
		v1.Log.ErrorWithCode(globalError.ParamBindError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		return
	}
	middleware.ResponseSuccess(ctx, "操作成功")
}

// DeleteUser
// @Tags      SysUser
// @Summary   删除用户
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200   {object}  middleware.Response{msg=string}  "删除用户"
// @Router    /api/user/{id}/delete_user [delete]
func (u *userController) DeleteUser(ctx *gin.Context) {
	uid, err := utils.ParseInt(ctx.Param("id"))
	if err != nil {
		v1.Log.ErrorWithCode(globalError.ParamBindError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		return
	}
	if claims := utils.GetUserInfo(ctx); claims != nil && claims.ID == uid {
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.AuthErr, fmt.Errorf("不能删除当前登录用户")))
		return
	}
	if err := v1.CoreV1.System().User().DeleteUser(ctx, uid); err != nil {
		v1.Log.ErrorWithCode(globalError.DeleteError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.DeleteError, err))
		return
	}
	middleware.ResponseSuccess(ctx, "操作成功")
}

// ChangePassword
// @Tags      SysUser
// @Summary   用户修改密码
// @Security  ApiKeyAuth
// @Produce  application/json
// @Param     data  body      dto.ChangeUserPwdInput    true  "用户ID, 原密码, 新密码"
// @Success   200   {object}  middleware.Response{msg=string}  "用户修改密码"
// @Router    /api/user/{id}/change_pwd [post]
func (u *userController) ChangePassword(ctx *gin.Context) {
	uid, err := utils.ParseInt(ctx.Param("id"))
	if err != nil {
		v1.Log.ErrorWithCode(globalError.ParamBindError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		return
	}
	claims := utils.GetUserInfo(ctx)
	if claims == nil || (claims.ID != uid && claims.AuthorityId != pkg.AdminDefaultAuth) {
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.AuthErr, fmt.Errorf("只能修改自己的密码")))
		return
	}
	params := &dto.ChangeUserPwdInput{}
	if err := params.BindingValidParams(ctx); err != nil {
		v1.Log.ErrorWithCode(globalError.ParamBindError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		return
	}
	if err := v1.CoreV1.System().User().ChangePassword(ctx, uid, params); err != nil {
		v1.Log.ErrorWithCode(globalError.ServerError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ServerError, err))
		return
	}
	middleware.ResponseSuccess(ctx, "")
}

// ResetPassword
// @Tags      SysUser
// @Summary   重置用户密码
// @Security  ApiKeyAuth
// @Produce  application/json
// @Success   200   {object}  middleware.Response{msg=string}  "重置用户密码"
// @Router    /api/user/{id}/reset_pwd [put]
func (u *userController) ResetPassword(ctx *gin.Context) {
	uid, err := utils.ParseInt(ctx.Param("id"))
	if err != nil {
		v1.Log.ErrorWithCode(globalError.ParamBindError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		return
	}
	params := &dto.ResetUserPwdInput{}
	if err := params.BindingValidParams(ctx); err != nil {
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		return
	}
	if err := v1.CoreV1.System().User().ResetPassword(ctx, uid, params.NewPwd); err != nil {
		v1.Log.ErrorWithCode(globalError.ServerError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ServerError, err))
		return
	}
	middleware.ResponseSuccess(ctx, "操作成功")
}
