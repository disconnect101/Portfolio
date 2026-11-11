package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type config struct {
	Address        string
	DatabaseURL    string
	UploadDir      string
	BootstrapKey   string
	SessionTimeout time.Duration
	SecureCookie   bool
	AllowedOrigins []string
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func loadConfig() (config, error) {
	dbURL := strings.TrimPrefix(env("DB_URL", "postgresql://localhost:5432/portfolio?sslmode=disable"), "jdbc:")
	u, err := url.Parse(dbURL)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" {
		return config{}, fmt.Errorf("DB_URL must be a PostgreSQL URL")
	}
	username, password := "portfolio", "portfolio"
	if u.User != nil {
		username = u.User.Username()
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}
	if v, ok := os.LookupEnv("DB_USER"); ok {
		username = v
	}
	if v, ok := os.LookupEnv("DB_PASSWORD"); ok {
		password = v
	}
	u.User = url.UserPassword(username, password)
	query := u.Query()
	host := u.Hostname()
	ip := net.ParseIP(host)
	if query.Get("sslmode") == "" && (strings.EqualFold(host, "localhost") || (ip != nil && ip.IsLoopback())) {
		query.Set("sslmode", "disable")
		u.RawQuery = query.Encode()
	}

	minutes, err := strconv.Atoi(env("ADMIN_SESSION_MINUTES", "30"))
	if err != nil || minutes < 1 || minutes > 1440 {
		return config{}, fmt.Errorf("ADMIN_SESSION_MINUTES must be between 1 and 1440")
	}
	secure, err := strconv.ParseBool(env("SESSION_COOKIE_SECURE", "false"))
	if err != nil {
		return config{}, fmt.Errorf("SESSION_COOKIE_SECURE must be true or false")
	}
	return config{
		Address:     net.JoinHostPort(env("SERVER_ADDRESS", "127.0.0.1"), env("SERVER_PORT", "8080")),
		UploadDir:   env("UPLOAD_DIR", "uploads"),
		DatabaseURL: u.String(), BootstrapKey: os.Getenv("ADMIN_BOOTSTRAP_KEY"),
		SessionTimeout: time.Duration(minutes) * time.Minute, SecureCookie: secure,
		AllowedOrigins: strings.Split(env("CORS_ORIGINS", "http://localhost:5173"), ","),
	}, nil
}
