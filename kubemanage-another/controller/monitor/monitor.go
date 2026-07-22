package monitor

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/noovertime7/kubemanage/cmd/app/config"
	"github.com/noovertime7/kubemanage/middleware"
	"github.com/noovertime7/kubemanage/pkg/utils"
)

const (
	grafanaSessionCookie = "kubemanage_grafana_session"
	grafanaCookiePath    = "/grafana"
)

// --- 一次性票据存储（内存实现；后端单副本足够，多副本需改 DB/redis）---

type ticketEntry struct {
	username string
	role     string
	expireAt time.Time
}

type ticketStore struct {
	mu      sync.Mutex
	entries map[string]ticketEntry
}

func newTicketStore() *ticketStore {
	return &ticketStore{entries: make(map[string]ticketEntry)}
}

func (s *ticketStore) issue(username, role string, ttl time.Duration) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.entries { // 顺带清理过期票据
		if now.After(v.expireAt) {
			delete(s.entries, k)
		}
	}
	s.entries[token] = ticketEntry{username: username, role: role, expireAt: now.Add(ttl)}
	return token, nil
}

// consume 取出并删除票据（一次性）；过期或不存在返回 false。
func (s *ticketStore) consume(token string) (ticketEntry, bool) {
	if token == "" {
		return ticketEntry{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[token]
	if ok {
		delete(s.entries, token)
	}
	if !ok || time.Now().After(e.expireAt) {
		return ticketEntry{}, false
	}
	return e, true
}

// --- 控制器 ---

// Controller 负责 Grafana 单点登录与反向代理。
type Controller struct {
	tickets *ticketStore
	proxy   *httputil.ReverseProxy
}

// NewMonitorRouter 注册监控相关路由：
//   - POST /api/monitor/grafana/ticket ：登录用户换取一次性票据（手动校验 JWT，不经 Casbin）
//   - ANY  /grafana/*path              ：Grafana 反向代理（/sso 交接 + 逐请求注入身份头）
func NewMonitorRouter(engine *gin.Engine) {
	c := &Controller{tickets: newTicketStore()}
	c.proxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			target, err := url.Parse(config.SysConfig.Grafana.Upstream)
			if err != nil {
				return
			}
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
			// 路径保持不变（含 /grafana 前缀）：Grafana 以子路径提供服务
		},
	}
	engine.POST("/api/monitor/grafana/ticket", c.Ticket)
	engine.Any("/grafana/*path", c.Grafana)
}

// Ticket 换取一次性票据。前端持 Bearer JWT 调用。
func (c *Controller) Ticket(ctx *gin.Context) {
	claims, err := utils.GetClaims(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "未登录或登录已过期"})
		return
	}
	role, ok := mapRole(claims.AuthorityId)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "当前角色无监控访问权限"})
		return
	}
	ttl := time.Duration(intOrDefault(config.SysConfig.Grafana.TicketTTL, 30)) * time.Second
	token, err := c.tickets.issue(claims.Username, role, ttl)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "生成票据失败"})
		return
	}
	middleware.ResponseSuccess(ctx, gin.H{"ticket": token})
}

// Grafana 反代入口，按子路径分派。
func (c *Controller) Grafana(ctx *gin.Context) {
	if ctx.Param("path") == "/sso" {
		c.sso(ctx)
		return
	}
	c.forward(ctx)
}

// sso 验票 → 下发签名会话 Cookie → 302 进入 Grafana。
func (c *Controller) sso(ctx *gin.Context) {
	entry, ok := c.tickets.consume(ctx.Query("ticket"))
	if !ok {
		// 票据无效/过期：回到 Kubemanage，让用户重新从菜单发起
		ctx.Redirect(http.StatusFound, "/")
		return
	}
	ttl := intOrDefault(config.SysConfig.Grafana.SessionTTL, 28800)
	value := signSession(sessionPayload{
		User: entry.username,
		Role: entry.role,
		Exp:  time.Now().Add(time.Duration(ttl) * time.Second).Unix(),
	})
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     grafanaSessionCookie,
		Value:    value,
		Path:     grafanaCookiePath,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   ttl,
		// 生产为 http(NodePort)，不设置 Secure；Cookie 由 HMAC 保证完整性
	})
	ctx.Redirect(http.StatusFound, "/grafana/")
}

// forward 校验会话 Cookie → 注入身份头 → 反代到 Grafana。
func (c *Controller) forward(ctx *gin.Context) {
	cookie, err := ctx.Request.Cookie(grafanaSessionCookie)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "监控会话不存在或已过期，请从 Kubemanage 菜单重新进入"})
		return
	}
	sess, ok := verifySession(cookie.Value)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "监控会话无效，请从 Kubemanage 菜单重新进入"})
		return
	}
	req := ctx.Request
	// 安全：剔除客户端可能伪造的 X-WEBAUTH-* 头，只使用服务端注入值
	for k := range req.Header {
		if strings.HasPrefix(http.CanonicalHeaderKey(k), "X-Webauth-") {
			req.Header.Del(k)
		}
	}
	req.Header.Set("X-WEBAUTH-USER", sess.User)
	req.Header.Set("X-WEBAUTH-ROLE", sess.Role)
	if req.Header.Get("X-Forwarded-Host") == "" {
		req.Header.Set("X-Forwarded-Host", req.Host)
	}
	c.proxy.ServeHTTP(ctx.Writer, req)
}

// --- 角色映射：Kubemanage AuthorityId -> Grafana Org 角色 ---

func mapRole(authorityID uint) (string, bool) {
	g := config.SysConfig.Grafana
	if role, ok := g.RoleMapping[strconv.Itoa(int(authorityID))]; ok && role != "" {
		return role, true
	}
	if g.DefaultRole != "" {
		return g.DefaultRole, true
	}
	return "", false
}

// --- 会话 Cookie 签名（HMAC-SHA256，密钥复用 JWTSecret）---

type sessionPayload struct {
	User string `json:"u"`
	Role string `json:"r"`
	Exp  int64  `json:"e"`
}

func signSession(p sessionPayload) string {
	raw, _ := json.Marshal(p)
	body := base64.RawURLEncoding.EncodeToString(raw)
	return body + "." + sign(body)
}

func verifySession(value string) (sessionPayload, bool) {
	parts := strings.SplitN(value, ".", 2)
	if len(parts) != 2 {
		return sessionPayload{}, false
	}
	if !hmac.Equal([]byte(sign(parts[0])), []byte(parts[1])) {
		return sessionPayload{}, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return sessionPayload{}, false
	}
	var p sessionPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return sessionPayload{}, false
	}
	if time.Now().Unix() > p.Exp {
		return sessionPayload{}, false
	}
	return p, true
}

func sign(body string) string {
	mac := hmac.New(sha256.New, []byte(config.SysConfig.Default.JWTSecret))
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func intOrDefault(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
