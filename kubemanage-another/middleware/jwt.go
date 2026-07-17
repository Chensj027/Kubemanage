package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/noovertime7/kubemanage/pkg/core/kubemanage/v1"
	"github.com/noovertime7/kubemanage/pkg/globalError"
	"github.com/noovertime7/kubemanage/pkg/utils"
)

// SanitizeWebSocketToken 在日志中间件读取 URL 前处理旧版查询参数令牌。
func SanitizeWebSocketToken() gin.HandlerFunc {
	return func(context *gin.Context) {
		utils.PromoteWebSocketQueryToken(context.Request)
		context.Next()
	}
}

// JWTAuth jwt认证函数
func JWTAuth() gin.HandlerFunc {
	return func(context *gin.Context) {
		if AlwaysAllowPath.Has(context.Request.URL.Path) {
			context.Next()
			return
		}
		// 处理验证逻辑
		claims, err := utils.GetClaims(context)
		if err != nil {
			ResponseError(context, globalError.NewGlobalError(globalError.AuthorizationError, err))
			context.Abort()
			return
		}
		if err := v1.CoreV1.System().User().ValidateClaims(context.Request.Context(), claims); err != nil {
			ResponseError(context, globalError.NewGlobalError(globalError.AuthorizationError, err))
			context.Abort()
			return
		}
		// 继续交由下一个路由处理,并将解析出的信息传递下去
		context.Set("claims", claims)
		utils.RemoveQueryToken(context.Request)
		context.Next()
	}
}
