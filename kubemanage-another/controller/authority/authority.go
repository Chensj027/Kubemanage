package authority

import (
	"github.com/gin-gonic/gin"

	"github.com/noovertime7/kubemanage/dto"
	"github.com/noovertime7/kubemanage/middleware"
	v1 "github.com/noovertime7/kubemanage/pkg/core/kubemanage/v1"
	"github.com/noovertime7/kubemanage/pkg/globalError"
	"strconv"
)

func (a *authorityController) Create(c *gin.Context) {
	var in dto.AuthorityInput
	if e := c.ShouldBindJSON(&in); e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ParamBindError, e))
		return
	}
	if e := v1.CoreV1.System().Authority().Create(c, &in); e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ServerError, e))
		return
	}
	middleware.ResponseSuccess(c, "操作成功")
}
func (a *authorityController) Update(c *gin.Context) {
	id, e := strconv.ParseUint(c.Param("authorityId"), 10, 64)
	if e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ParamBindError, e))
		return
	}
	var in dto.AuthorityInput
	if e = c.ShouldBindJSON(&in); e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ParamBindError, e))
		return
	}
	if e = v1.CoreV1.System().Authority().Update(c, uint(id), &in); e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ServerError, e))
		return
	}
	middleware.ResponseSuccess(c, "操作成功")
}
func (a *authorityController) Delete(c *gin.Context) {
	id, e := strconv.ParseUint(c.Param("authorityId"), 10, 64)
	if e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.ParamBindError, e))
		return
	}
	if e = v1.CoreV1.System().Authority().Delete(c, uint(id)); e != nil {
		middleware.ResponseError(c, globalError.NewGlobalError(globalError.DeleteError, e))
		return
	}
	middleware.ResponseSuccess(c, "操作成功")
}

func (a *authorityController) GetAuthorityList(ctx *gin.Context) {
	params := &dto.PageInfo{}
	if err := params.BindingValidParams(ctx); err != nil {
		v1.Log.ErrorWithCode(globalError.ParamBindError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.ParamBindError, err))
		return
	}
	data, err := v1.CoreV1.System().Authority().GetAuthorityList(ctx, *params)
	if err != nil {
		v1.Log.ErrorWithCode(globalError.GetError, err)
		middleware.ResponseError(ctx, globalError.NewGlobalError(globalError.GetError, err))
		return
	}
	middleware.ResponseSuccess(ctx, data)
}
