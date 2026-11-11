package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

//go:embed schema.sql
var files embed.FS

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	schema, err := files.ReadFile("schema.sql")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(schema)); err != nil {
		log.Fatalf("schema initialization failed: %v", err)
	}
	store := &Store{db: db}
	if err := store.seed(ctx); err != nil {
		log.Fatalf("portfolio seed failed: %v", err)
	}
	if err := store.bootstrapAdmin(ctx, cfg.BootstrapKey); err != nil {
		log.Fatalf("admin bootstrap failed: %v", err)
	}
	cfg.BootstrapKey = ""
	app := newApp(store, cfg)
	server := &http.Server{Addr: cfg.Address, Handler: app.routes(), ReadHeaderTimeout: 5 * time.Second}
	stop, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()
	go func() {
		<-stop.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()
	log.Printf("portfolio API listening on %s", cfg.Address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
