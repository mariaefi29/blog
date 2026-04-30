package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	const timeout = 5 * time.Second

	srv := New(Params{
		Port:      8080,
		Timeout:   timeout,
		StaticDir: "../../public",
	})

	if srv.Addr != ":8080" {
		t.Fatalf("server address = %q, want %q", srv.Addr, ":8080")
	}
	if srv.ReadHeaderTimeout != timeout {
		t.Fatalf("read header timeout = %s, want %s", srv.ReadHeaderTimeout, timeout)
	}
	if srv.ReadTimeout != 0 {
		t.Fatalf("read timeout = %s, want 0", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 0 {
		t.Fatalf("write timeout = %s, want 0", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 0 {
		t.Fatalf("idle timeout = %s, want 0", srv.IdleTimeout)
	}

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()

	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d", rec.Code, http.StatusOK)
	}
}
