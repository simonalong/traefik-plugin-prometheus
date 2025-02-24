package pluginTraefik_test

import (
	"context"
	pluginTraefik "github.com/traefik/plugindemo"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDemo(t *testing.T) {
	cfg := pluginTraefik.CreateConfig()
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := pluginTraefik.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)
	assertHeader(t, req, "X-Host", "localhost")
	assertHeader(t, req, "X-URL", "http://localhost")
	assertHeader(t, req, "X-Method", "GET")
	assertHeader(t, req, "X-Demo", "test")
}

func assertHeader(t *testing.T, req *http.Request, key, expected string) {
	t.Helper()

	if req.Header.Get(key) != expected {
		t.Errorf("invalid header value: %s", req.Header.Get(key))
	}
}
