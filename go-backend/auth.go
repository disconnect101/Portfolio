package main

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type adminSession struct {
	CSRFToken string `json:"csrfToken"`
	ExpiresAt int64  `json:"expiresAt"`
}

type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]adminSession
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *sessionStore) put(id string, session adminSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UnixMilli()
	for key, value := range s.sessions {
		if value.ExpiresAt <= now {
			delete(s.sessions, key)
		}
	}
	s.sessions[id] = session
}

func (s *sessionStore) get(id string) (adminSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if ok && session.ExpiresAt <= time.Now().UnixMilli() {
		delete(s.sessions, id)
		return adminSession{}, false
	}
	return session, ok
}

func (s *sessionStore) delete(id string) { s.mu.Lock(); defer s.mu.Unlock(); delete(s.sessions, id) }

func (a *app) cookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{Name: "JSESSIONID", Value: value, Path: "/", HttpOnly: true, Secure: a.cfg.SecureCookie, SameSite: http.SameSiteStrictMode, MaxAge: maxAge}
}

func (a *app) session(r *http.Request) (adminSession, bool) {
	cookie, err := r.Cookie("JSESSIONID")
	if err != nil {
		return adminSession{}, false
	}
	return a.sessions.get(cookie.Value)
}

func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback() && strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func (a *app) login(c *gin.Context) {
	w, r := c.Writer, c.Request
	if a.cfg.SecureCookie && !isHTTPS(r) {
		http.Error(w, "Use HTTPS to log in", http.StatusForbidden)
		return
	}
	var body struct {
		Key string `json:"key"`
	}
	if !decode(c, &body) {
		return
	}
	matched, err := a.store.authenticate(r.Context(), body.Key)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Admin account is not configured", http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}
	if !matched {
		http.Error(w, "Invalid admin key", http.StatusUnauthorized)
		return
	}
	id, err := randomToken()
	if err != nil {
		serverError(c, err)
		return
	}
	csrf, err := randomToken()
	if err != nil {
		serverError(c, err)
		return
	}
	if old, err := r.Cookie("JSESSIONID"); err == nil {
		a.sessions.delete(old.Value)
	}
	session := adminSession{CSRFToken: csrf, ExpiresAt: time.Now().Add(a.cfg.SessionTimeout).UnixMilli()}
	a.sessions.put(id, session)
	http.SetCookie(w, a.cookie(id, 0))
	c.JSON(http.StatusOK, session)
}

func (a *app) requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, ok := a.session(c.Request)
		if !ok {
			c.String(http.StatusUnauthorized, "Admin session expired or missing\n")
			c.Abort()
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead && c.Request.Method != http.MethodOptions {
			provided := c.GetHeader("X-CSRF-Token")
			if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(session.CSRFToken)) != 1 {
				c.String(http.StatusForbidden, "Invalid CSRF token\n")
				c.Abort()
				return
			}
		}
		c.Set("adminSession", session)
		c.Next()
	}
}

func (a *app) sessionInfo(c *gin.Context) {
	c.JSON(http.StatusOK, c.MustGet("adminSession"))
}

func (a *app) logout(c *gin.Context) {
	if cookie, err := c.Request.Cookie("JSESSIONID"); err == nil {
		a.sessions.delete(cookie.Value)
	}
	http.SetCookie(c.Writer, a.cookie("", -1))
	c.Status(http.StatusNoContent)
}
