package monitor

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/noovertime7/kubemanage/cmd/app/config"
	"github.com/noovertime7/kubemanage/pkg"
)

func setupConfig(upstream string) {
	pkg.RegisterJwt("testsecret")
	config.SysConfig = &config.Config{}
	config.SysConfig.Default.JWTSecret = "testsecret"
	config.SysConfig.Grafana = config.GrafanaOptions{
		Upstream:    upstream,
		DefaultRole: "Viewer",
		RoleMapping: map[string]string{"111": "Admin", "222": "Editor", "2221": "Viewer"},
		TicketTTL:   30,
		SessionTTL:  28800,
	}
}

// crNotifier 让 httptest 记录器实现 http.CloseNotifier，
// 否则 httputil.ReverseProxy 经由 gin 的 responseWriter 调用 CloseNotify 会 panic（仅测试环境问题）。
type crNotifier struct {
	*httptest.ResponseRecorder
	ch chan bool
}

func (c *crNotifier) CloseNotify() <-chan bool { return c.ch }

func serve(engine *gin.Engine, req *http.Request) *crNotifier {
	rec := &crNotifier{httptest.NewRecorder(), make(chan bool, 1)}
	engine.ServeHTTP(rec, req)
	return rec
}

// TestGrafanaSSOFlow 端到端验证 ticket -> sso -> proxy 全链路及安全性。
func TestGrafanaSSOFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotUser, gotRole string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = r.Header.Get("X-WEBAUTH-USER")
		gotRole = r.Header.Get("X-WEBAUTH-ROLE")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "grafana-ok "+r.URL.Path)
	}))
	defer upstream.Close()

	setupConfig(upstream.URL)
	engine := gin.New()
	NewMonitorRouter(engine)

	// 1) 合法 JWT(AuthorityId 222 -> Editor) 换票
	token, err := pkg.JWTToken.GenerateToken(pkg.BaseClaims{Username: "alice", AuthorityId: 222})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/monitor/grafana/ticket", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := serve(engine, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("ticket status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			Ticket string `json:"ticket"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode ticket resp: %v", err)
	}
	ticket := resp.Data.Ticket
	if ticket == "" {
		t.Fatal("empty ticket")
	}

	// 1b) 未携带 JWT -> 401
	if rec := serve(engine, httptest.NewRequest(http.MethodPost, "/api/monitor/grafana/ticket", nil)); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no-jwt ticket expected 401, got %d", rec.Code)
	}

	// 2) 验票 -> 302 + 会话 Cookie
	rec = serve(engine, httptest.NewRequest(http.MethodGet, "/grafana/sso?ticket="+ticket, nil))
	if rec.Code != http.StatusFound {
		t.Fatalf("sso status=%d", rec.Code)
	}
	var session *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == grafanaSessionCookie {
			session = c
		}
	}
	if session == nil {
		t.Fatal("missing session cookie")
	}

	// 2b) 票据一次性：重复使用 -> 视为无效，重定向回 "/"
	rec = serve(engine, httptest.NewRequest(http.MethodGet, "/grafana/sso?ticket="+ticket, nil))
	if loc := rec.Header().Get("Location"); loc != "/" {
		t.Fatalf("reused ticket should redirect to /, got %q", loc)
	}

	// 3) 带 Cookie + 伪造头 访问代理：上游应看到 Cookie 身份，伪造头被剔除
	req = httptest.NewRequest(http.MethodGet, "/grafana/api/dashboards", nil)
	req.AddCookie(session)
	req.Header.Set("X-WEBAUTH-USER", "attacker") // 伪造
	req.Header.Set("X-WEBAUTH-ROLE", "Admin")    // 伪造
	rec = serve(engine, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("proxy status=%d body=%s", rec.Code, rec.Body.String())
	}
	if gotUser != "alice" {
		t.Fatalf("upstream user=%q want alice (伪造头未被剔除?)", gotUser)
	}
	if gotRole != "Editor" {
		t.Fatalf("upstream role=%q want Editor", gotRole)
	}

	// 3b) 无 Cookie 访问代理 -> 401（不落到 Grafana）
	if rec := serve(engine, httptest.NewRequest(http.MethodGet, "/grafana/api/dashboards", nil)); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no-cookie proxy expected 401, got %d", rec.Code)
	}
}

func TestMapRole(t *testing.T) {
	config.SysConfig = &config.Config{}
	config.SysConfig.Grafana = config.GrafanaOptions{
		RoleMapping: map[string]string{"111": "Admin", "222": "Editor"},
		DefaultRole: "",
	}
	if r, ok := mapRole(111); !ok || r != "Admin" {
		t.Fatalf("111 -> %q,%v want Admin", r, ok)
	}
	if _, ok := mapRole(999); ok {
		t.Fatal("未映射且默认为空时应拒绝")
	}
	config.SysConfig.Grafana.DefaultRole = "Viewer"
	if r, ok := mapRole(999); !ok || r != "Viewer" {
		t.Fatalf("999 -> %q,%v want Viewer(default)", r, ok)
	}
}

func TestSessionSignVerify(t *testing.T) {
	config.SysConfig = &config.Config{}
	config.SysConfig.Default.JWTSecret = "s3cr3t"
	valid := signSession(sessionPayload{User: "u", Role: "Viewer", Exp: time.Now().Add(time.Hour).Unix()})
	if _, ok := verifySession(valid); !ok {
		t.Fatal("有效会话应通过校验")
	}
	if _, ok := verifySession(valid + "tamper"); ok {
		t.Fatal("被篡改签名应校验失败")
	}
	expired := signSession(sessionPayload{User: "u", Role: "Viewer", Exp: time.Now().Add(-time.Hour).Unix()})
	if _, ok := verifySession(expired); ok {
		t.Fatal("过期会话应校验失败")
	}
}
