package middleware

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/noovertime7/kubemanage/pkg/core/kubemanage/v1"
	"github.com/noovertime7/kubemanage/pkg/globalError"
	"github.com/noovertime7/kubemanage/pkg/utils"
)

// CasbinHandler 拦截器
func CasbinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if AlwaysAllowPath.Has(c.Request.URL.Path) {
			c.Next()
			return
		}
		waitUse, err := utils.GetClaims(c)
		if err != nil {
			ResponseError(c, globalError.NewGlobalError(globalError.ServerError, err))
			c.Abort()
			return
		}
		// 使用 Gin 匹配到的规范路由模板，避免重复斜杠、编码差异或路径参数
		// 导致策略绕过或误拒绝。找不到模板时再回退到清理后的 URL 路径。
		obj := c.FullPath()
		if obj == "" {
			obj = path.Clean("/" + strings.TrimPrefix(c.Request.URL.Path, "/"))
		}
		// 获取请求方法
		act := strings.ToUpper(c.Request.Method)
		// 获取用户的角色
		sub := strconv.Itoa(int(waitUse.AuthorityId))
		e := v1.CoreV1.System().CasbinService().Casbin() // 判断策略中是否存在
		success, err := e.Enforce(sub, obj, act)
		if err != nil {
			ResponseError(c, globalError.NewGlobalError(globalError.ServerError, err))
			c.Abort()
			return
		}
		if success {
			c.Next()
		} else {
			ResponseError(c, globalError.NewGlobalError(globalError.AuthErr, fmt.Errorf("角色ID %d 请求 %s %s 无权限", waitUse.AuthorityId, act, obj)))
			c.Abort()
			return
		}
	}
}
