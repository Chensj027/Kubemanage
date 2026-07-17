package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/noovertime7/kubemanage/pkg"
	"k8s.io/apimachinery/pkg/util/sets"
)

var AlwaysAllowPath sets.String

// InstallMiddlewares 安装需要的中间件
func InstallMiddlewares(ginEngine *gin.RouterGroup) {
	// 初始化可忽略的请求路径
	// 只有登录接口允许匿名访问；退出和 WebShell 都必须通过 JWT 与 Casbin。
	AlwaysAllowPath = sets.NewString(pkg.LoginURL)
	ginEngine.Use(SanitizeWebSocketToken(), Logger(), Cores(), Limiter(), Recovery(true), TranslationMiddleware(), JWTAuth(), CasbinHandler())
}
