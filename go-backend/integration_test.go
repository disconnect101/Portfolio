package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

// TEST_DATABASE_URL must point to a disposable database whose name ends in _test.
func TestPostgresAPI(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a disposable PostgreSQL database ending in _test")
	}
	u, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must use a database name ending in _test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	schema, err := files.ReadFile("schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, string(schema)); err != nil {
		t.Fatal(err)
	}
	store := &Store{db: db}
	if err = store.seed(ctx); err != nil {
		t.Fatal(err)
	}
	key := "integration-test-key-with-over-32-characters"
	if err = store.bootstrapAdmin(ctx, key); err != nil {
		t.Fatal(err)
	}
	if err = store.bootstrapAdmin(ctx, "a-different-bootstrap-key-long-enough-to-be-valid"); err != nil {
		t.Fatal(err)
	}
	matched, err := store.authenticate(ctx, key)
	if err != nil || !matched {
		t.Fatalf("admin credential compatibility: matched=%v error=%v", matched, err)
	}
	var stored string
	if err = db.QueryRow(`SELECT password_hash FROM admin_credential WHERE id=1`).Scan(&stored); err != nil || stored == key {
		t.Fatal("expected a stored password hash")
	}

	app := newApp(store, config{SessionTimeout: 30 * time.Minute, AllowedOrigins: []string{"http://localhost:5173"}})
	handler := app.routes()
	request := func(method, path, body, csrf string, cookie *http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if cookie != nil {
			r.AddCookie(cookie)
		}
		if csrf != "" {
			r.Header.Set("X-CSRF-Token", csrf)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	check := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("expected status %d, got %d: %s", status, w.Code, w.Body.String())
		}
	}
	for _, path := range []string{"/api/portfolio", "/api/profile", "/api/projects", "/api/education", "/api/experience"} {
		check(request("GET", path, "", "", nil), http.StatusOK)
	}
	check(request("GET", "/api/admin/session", "", "", nil), http.StatusUnauthorized)
	check(request("POST", "/api/admin/login", `{"key":"wrong"}`, "", nil), http.StatusUnauthorized)
	login := request("POST", "/api/admin/login", `{"key":"`+key+`"}`, "", nil)
	check(login, http.StatusOK)
	var session adminSession
	if err = json.Unmarshal(login.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	cookies := login.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatal("expected an HttpOnly SameSite=Strict session cookie")
	}
	secureApp := newApp(store, config{SessionTimeout: 30 * time.Minute, SecureCookie: true})
	secureRequest := httptest.NewRequest("POST", "/api/admin/login", strings.NewReader(`{"key":"`+key+`"}`))
	secureRequest.RemoteAddr = "127.0.0.1:12345"
	secureResult := httptest.NewRecorder()
	secureApp.routes().ServeHTTP(secureResult, secureRequest)
	check(secureResult, http.StatusForbidden)
	secureRequest = httptest.NewRequest("POST", "/api/admin/login", strings.NewReader(`{"key":"`+key+`"}`))
	secureRequest.RemoteAddr = "127.0.0.1:12345"
	secureRequest.Header.Set("X-Forwarded-Proto", "https")
	secureResult = httptest.NewRecorder()
	secureApp.routes().ServeHTTP(secureResult, secureRequest)
	check(secureResult, http.StatusOK)
	if len(secureResult.Result().Cookies()) != 1 || !secureResult.Result().Cookies()[0].Secure {
		t.Fatal("HTTPS login should set a Secure cookie")
	}
	cookie := cookies[0]
	check(request("GET", "/api/admin/session", "", "", cookie), http.StatusOK)
	check(request("POST", "/api/admin/projects", `{"title":"forbidden"}`, "", cookie), http.StatusForbidden)

	for _, tc := range []struct{ kind, create, update, field, want string }{
		{"projects", `{"title":"integration project","technologies":["Go","SQL"]}`, `{"title":"updated project","technologies":["React"]}`, "title", "updated project"},
		{"education", `{"institution":"integration school"}`, `{"institution":"updated school"}`, "institution", "updated school"},
		{"experience", `{"company":"integration company","highlights":["one","two"]}`, `{"company":"updated company","highlights":["three"]}`, "company", "updated company"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			created := request("POST", "/api/admin/"+tc.kind, tc.create, session.CSRFToken, cookie)
			check(created, http.StatusCreated)
			var row struct {
				ID json.Number `json:"id"`
			}
			if err := json.Unmarshal(created.Body.Bytes(), &row); err != nil || row.ID == "" {
				t.Fatalf("invalid create response: %s", created.Body.String())
			}
			path := "/api/admin/" + tc.kind + "/" + row.ID.String()
			check(request("PUT", path, tc.update, session.CSRFToken, cookie), http.StatusOK)
			list := request("GET", "/api/"+tc.kind, "", "", nil)
			check(list, http.StatusOK)
			var items []map[string]any
			if err := json.Unmarshal(list.Body.Bytes(), &items); err != nil {
				t.Fatal(err)
			}
			found := false
			for _, item := range items {
				if row.ID.String() == formatID(item["id"]) {
					found = true
					if item[tc.field] != tc.want {
						t.Fatalf("field %s did not update: %v", tc.field, item[tc.field])
					}
					if tc.kind == "projects" && formatID(item["technologies"]) != `["React"]` {
						t.Fatal("project technologies were not updated")
					}
					if tc.kind == "experience" && formatID(item["highlights"]) != `["three"]` {
						t.Fatal("experience highlights were not updated")
					}
				}
			}
			if !found {
				t.Fatal("created item missing from public API")
			}
			check(request("DELETE", path, "", session.CSRFToken, cookie), http.StatusNoContent)
			check(request("DELETE", path, "", session.CSRFToken, cookie), http.StatusNotFound)
		})
	}
	profile := request("GET", "/api/profile", "", "", nil)
	check(profile, http.StatusOK)
	original := profile.Body.String()
	var changed map[string]any
	if err := json.Unmarshal([]byte(original), &changed); err != nil {
		t.Fatal(err)
	}
	changed["title"] = "Go integration test"
	newProfile, err := json.Marshal(changed)
	if err != nil {
		t.Fatal(err)
	}
	check(request("PUT", "/api/admin/profile", string(newProfile), session.CSRFToken, cookie), http.StatusOK)
	profile = request("GET", "/api/profile", "", "", nil)
	check(profile, http.StatusOK)
	var result Profile
	if err := json.Unmarshal(profile.Body.Bytes(), &result); err != nil || result.Title == nil || *result.Title != "Go integration test" {
		t.Fatal("profile update was not persisted")
	}
	check(request("PUT", "/api/admin/profile", original, session.CSRFToken, cookie), http.StatusOK)
	check(request("POST", "/api/admin/logout", "", session.CSRFToken, cookie), http.StatusNoContent)
	check(request("GET", "/api/admin/session", "", "", cookie), http.StatusUnauthorized)

	login = request("POST", "/api/admin/login", `{"key":"`+key+`"}`, "", nil)
	check(login, http.StatusOK)
	cookie = login.Result().Cookies()[0]
	app.sessions.mu.Lock()
	expired := app.sessions.sessions[cookie.Value]
	expired.ExpiresAt = time.Now().Add(-time.Second).UnixMilli()
	app.sessions.sessions[cookie.Value] = expired
	app.sessions.mu.Unlock()
	check(request("GET", "/api/admin/session", "", "", cookie), http.StatusUnauthorized)
}

func formatID(id any) string {
	value, _ := json.Marshal(id)
	return string(value)
}
