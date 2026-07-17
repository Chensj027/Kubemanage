package utils

import (
	"net/http/httptest"
	"testing"
)

func TestGetRequestToken(t *testing.T) {
	tests := []struct {
		name      string
		authority string
		legacy    string
		upgrade   bool
		url       string
		want      string
		wantErr   bool
	}{
		{name: "bearer", authority: "Bearer bearer-token", legacy: "legacy-token", url: "http://example.test/api", want: "bearer-token"},
		{name: "case insensitive bearer", authority: "bearer bearer-token", url: "http://example.test/api", want: "bearer-token"},
		{name: "legacy header", legacy: "legacy-token", url: "http://example.test/api", want: "legacy-token"},
		{name: "websocket subprotocol", upgrade: true, authority: "", url: "http://example.test/api", want: "protocol-token"},
		{name: "websocket query", upgrade: true, url: "http://example.test/api?token=query-token", want: "query-token"},
		{name: "http query rejected", url: "http://example.test/api?token=query-token", wantErr: true},
		{name: "malformed authorization", authority: "bearer-token", legacy: "legacy-token", url: "http://example.test/api", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			req.Header.Set("Authorization", tt.authority)
			req.Header.Set("token", tt.legacy)
			if tt.upgrade {
				req.Header.Set("Upgrade", "websocket")
			}
			if tt.name == "websocket subprotocol" {
				req.Header.Set("Sec-WebSocket-Protocol", "kubemanage, protocol-token")
			}
			got, err := GetRequestToken(req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetRequestToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("GetRequestToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPromoteWebSocketQueryToken(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.test/api?namespace=default&token=secret", nil)
	req.Header.Set("Upgrade", "websocket")
	PromoteWebSocketQueryToken(req)

	if req.Header.Get("token") != "secret" {
		t.Fatalf("PromoteWebSocketQueryToken() token header = %q", req.Header.Get("token"))
	}
	if req.URL.Query().Get("token") != "" || req.URL.Query().Get("namespace") != "default" {
		t.Fatalf("PromoteWebSocketQueryToken() query = %q", req.URL.RawQuery)
	}
}

func TestRemoveQueryToken(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.test/api?namespace=default&token=secret&access_token=other", nil)
	req.Header.Set("Upgrade", "websocket")
	RemoveQueryToken(req)

	query := req.URL.Query()
	if query.Get("token") != "" || query.Get("access_token") != "" {
		t.Fatalf("RemoveQueryToken() left credentials in query: %q", req.URL.RawQuery)
	}
	if query.Get("namespace") != "default" {
		t.Fatalf("RemoveQueryToken() removed unrelated query: %q", req.URL.RawQuery)
	}
}

func TestIsOriginAllowed(t *testing.T) {
	t.Setenv("KUBEMANAGE_ALLOWED_ORIGINS", "https://console.example.test")

	tests := []struct {
		name   string
		host   string
		origin string
		want   bool
	}{
		{name: "same origin", host: "api.example.test", origin: "https://api.example.test", want: true},
		{name: "configured origin", host: "api.example.test", origin: "https://console.example.test", want: true},
		{name: "foreign origin", host: "api.example.test", origin: "https://evil.example.test", want: false},
		{name: "invalid scheme", host: "api.example.test", origin: "file://api.example.test", want: false},
		{name: "origin with path", host: "api.example.test", origin: "https://api.example.test/path", want: false},
		{name: "non browser client", host: "api.example.test", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://"+tt.host+"/api", nil)
			req.Host = tt.host
			req.Header.Set("Origin", tt.origin)
			if got := IsOriginAllowed(req); got != tt.want {
				t.Fatalf("IsOriginAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
