package utils

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/noovertime7/kubemanage/pkg"
	"github.com/pkg/errors"
)

const legacyTokenHeader = "token"

// GetClaims 从token中，取出claims值
func GetClaims(c *gin.Context) (*pkg.CustomClaims, error) {
	if claims, exists := c.Get("claims"); exists {
		if customClaims, ok := claims.(*pkg.CustomClaims); ok && customClaims != nil {
			return customClaims, nil
		}
	}

	token, err := GetRequestToken(c.Request)
	if err != nil {
		return nil, err
	}

	// 解析token内容
	claims, err := pkg.JWTToken.ParseToken(token)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// GetRequestToken 统一读取 HTTP 与 WebSocket 握手请求中的 JWT。
// 普通 HTTP 请求优先使用 Authorization: Bearer，兼容历史令牌请求头；
// 浏览器 WebSocket 无法设置 Authorization，因此优先使用 kubemanage 子协议携带 JWT，
// 并仅为旧客户端保留查询参数令牌兼容。
func GetRequestToken(r *http.Request) (string, error) {
	if r == nil {
		return "", errors.New("请求为空,无权限访问")
	}

	if authorization := strings.TrimSpace(r.Header.Get("Authorization")); authorization != "" {
		parts := strings.Fields(authorization)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			return "", errors.New("Authorization请求头格式错误")
		}
		return parts[1], nil
	}

	if token := strings.TrimSpace(r.Header.Get(legacyTokenHeader)); token != "" {
		return token, nil
	}

	if isWebSocketUpgrade(r) {
		protocols := strings.Split(r.Header.Get("Sec-WebSocket-Protocol"), ",")
		for index, protocol := range protocols {
			if strings.EqualFold(strings.TrimSpace(protocol), "kubemanage") && index+1 < len(protocols) {
				if token := strings.TrimSpace(protocols[index+1]); token != "" {
					return token, nil
				}
			}
		}
		if token := strings.TrimSpace(r.URL.Query().Get("token")); token != "" {
			return token, nil
		}
		if token := strings.TrimSpace(r.URL.Query().Get("access_token")); token != "" {
			return token, nil
		}
	}

	return "", errors.New("请求未携带token,无权限访问")
}

// PromoteWebSocketQueryToken 将旧客户端的查询参数令牌转为内部请求头并清除 URL。
// 必须放在访问日志中间件之前；反向代理仍可能已记录原始 URL，因此新客户端应使用子协议。
func PromoteWebSocketQueryToken(r *http.Request) {
	if r == nil || r.URL == nil || !isWebSocketUpgrade(r) {
		return
	}
	if r.Header.Get("Authorization") == "" && r.Header.Get(legacyTokenHeader) == "" {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			token = strings.TrimSpace(r.URL.Query().Get("access_token"))
		}
		if token != "" {
			r.Header.Set(legacyTokenHeader, token)
		}
	}
	RemoveQueryToken(r)
}

// RemoveQueryToken 在 WebSocket 完成认证后移除 URL 中的凭证，避免操作日志记录 JWT。
func RemoveQueryToken(r *http.Request) {
	if r == nil || r.URL == nil || !isWebSocketUpgrade(r) {
		return
	}
	query := r.URL.Query()
	query.Del("token")
	query.Del("access_token")
	r.URL.RawQuery = query.Encode()
}

// AllowedCORSOrigins 返回由 KUBEMANAGE_ALLOWED_ORIGINS 配置的跨域来源。
// 多个来源使用逗号分隔；不配置时仅允许同源请求。
func AllowedCORSOrigins() []string {
	values := strings.Split(os.Getenv("KUBEMANAGE_ALLOWED_ORIGINS"), ",")
	origins := make([]string, 0, len(values))
	for _, value := range values {
		origin := normalizeOrigin(value)
		if origin != "" && origin != "*" && isValidOrigin(origin) {
			origins = append(origins, origin)
		}
	}
	return origins
}

// IsCORSOriginAllowed 校验跨域 HTTP 请求是否来自显式配置的来源。
func IsCORSOriginAllowed(origin string) bool {
	origin = normalizeOrigin(origin)
	if !isValidOrigin(origin) {
		return false
	}
	for _, allowed := range AllowedCORSOrigins() {
		if strings.EqualFold(origin, allowed) {
			return true
		}
	}
	return false
}

// IsOriginAllowed 校验 WebSocket 请求来源，同源默认允许，额外来源由环境变量显式配置。
func IsOriginAllowed(r *http.Request) bool {
	if r == nil {
		return false
	}
	origin := normalizeOrigin(r.Header.Get("Origin"))
	if origin == "" {
		// 非浏览器客户端通常不携带来源请求头。
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || !isValidOrigin(origin) {
		return false
	}
	if strings.EqualFold(parsed.Host, r.Host) {
		return true
	}
	if IsCORSOriginAllowed(origin) {
		return true
	}
	return false
}

func isWebSocketUpgrade(r *http.Request) bool {
	return r != nil && strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket")
}

func normalizeOrigin(origin string) string {
	return strings.TrimSuffix(strings.TrimSpace(origin), "/")
}

func isValidOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	return err == nil && parsed.User == nil && parsed.Host != "" &&
		(parsed.Scheme == "http" || parsed.Scheme == "https") &&
		parsed.Path == "" && parsed.RawQuery == "" && parsed.Fragment == ""
}

// GetUserAuthorityId 从Gin的Context中获取从jwt解析出来的用户角色id
func GetUserAuthorityId(c *gin.Context) (uint, error) {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c); err != nil {
			return 0, err
		} else {
			return cl.AuthorityId, nil
		}
	} else {
		waitUse := claims.(*pkg.CustomClaims)
		return waitUse.AuthorityId, nil
	}
}

// GetUserInfo 从Gin的Context中获取从jwt解析出来的用户角色id
func GetUserInfo(c *gin.Context) *pkg.CustomClaims {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c); err != nil {
			return nil
		} else {
			return cl
		}
	} else {
		waitUse := claims.(*pkg.CustomClaims)
		return waitUse
	}
}
