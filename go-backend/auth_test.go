package main

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPSForwardedOnlyFromLoopback(t *testing.T) {
	r := httptest.NewRequest("POST", "http://example.com/api/admin/login", nil)
	r.RemoteAddr = "127.0.0.1:12345"
	r.Header.Set("X-Forwarded-Proto", "https")
	if !isHTTPS(r) {
		t.Fatal("expected HTTPS forwarded by local proxy")
	}
	r.RemoteAddr = "203.0.113.1:12345"
	if isHTTPS(r) {
		t.Fatal("must not trust a remote forwarded-proto header")
	}
}

func TestExpiredSessionCannotBeUsed(t *testing.T) {
	s := &sessionStore{sessions: make(map[string]adminSession)}
	s.put("expired", adminSession{ExpiresAt: time.Now().Add(-time.Second).UnixMilli()})
	if _, ok := s.get("expired"); ok {
		t.Fatal("expired session must be rejected")
	}
	s.put("active", adminSession{ExpiresAt: time.Now().Add(time.Minute).UnixMilli()})
	if _, ok := s.get("active"); !ok {
		t.Fatal("unexpired session should be available")
	}
	s.delete("active")
	if _, ok := s.get("active"); ok {
		t.Fatal("deleted session must be rejected")
	}
}
