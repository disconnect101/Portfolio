package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAdminRoutesRequireSessionAndCSRF(t *testing.T) {
	app := newApp(nil, config{AllowedOrigins: []string{"http://localhost:5173"}})
	handler := app.routes()
	app.sessions.put("valid-session", adminSession{CSRFToken: "valid-csrf", ExpiresAt: time.Now().Add(time.Minute).UnixMilli()})
	routes := []struct{ method, path string }{
		{"GET", "/api/admin/session"}, {"POST", "/api/admin/logout"},
		{"PUT", "/api/admin/profile"},
		{"PUT", "/api/admin/education-photo"}, {"PUT", "/api/admin/experience/1/logo"},
		{"POST", "/api/admin/projects"}, {"PUT", "/api/admin/projects/1"}, {"DELETE", "/api/admin/projects/1"},
		{"POST", "/api/admin/education"}, {"PUT", "/api/admin/education/1"}, {"DELETE", "/api/admin/education/1"},
		{"POST", "/api/admin/experience"}, {"PUT", "/api/admin/experience/1"}, {"DELETE", "/api/admin/experience/1"},
	}
	for _, route := range routes {
		t.Run(route.method+route.path, func(t *testing.T) {
			request := httptest.NewRequest(route.method, route.path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("without session got %d", response.Code)
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("admin responses must not be cached")
			}
			if route.method == "GET" {
				return
			}
			request = httptest.NewRequest(route.method, route.path, nil)
			request.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: "valid-session"})
			response = httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusForbidden {
				t.Fatalf("without CSRF token got %d", response.Code)
			}
		})
	}
}

func TestCORSPublicGETAndJSONValidation(t *testing.T) {
	app := newApp(nil, config{AllowedOrigins: []string{"http://localhost:5173"}})
	handler := app.routes()
	request := httptest.NewRequest(http.MethodOptions, "/api/projects", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", "GET")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("unexpected preflight response: %d %v", response.Code, response.Header())
	}

	app.sessions.put("valid-session", adminSession{CSRFToken: "valid-csrf", ExpiresAt: time.Now().Add(time.Minute).UnixMilli()})
	for _, body := range []string{`{"title":"x"} {}`, `{"title":`} {
		request = httptest.NewRequest(http.MethodPost, "/api/admin/projects", strings.NewReader(body))
		request.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: "valid-session"})
		request.Header.Set("X-CSRF-Token", "valid-csrf")
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("invalid JSON returned %d", response.Code)
		}
	}
}

func TestRecoveryReturnsServerError(t *testing.T) {
	app := newApp(nil, config{})
	response := httptest.NewRecorder()
	app.routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/portfolio", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected response after panic: %d", response.Code)
	}
}
