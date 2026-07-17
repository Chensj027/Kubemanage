package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/noovertime7/kubemanage/pkg/utils"
)

// Cores 处理跨域请求，支持options访问
func Cores() gin.HandlerFunc {
	c := cors.Config{
		AllowOriginFunc: utils.IsCORSOriginAllowed,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:    []string{"Content-Type", "Access-Token", "Authorization", "Token"},
		ExposeHeaders:   []string{"New-Token"},
		MaxAge:          6 * time.Hour,
	}

	return cors.New(c)
}
