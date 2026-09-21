package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"eedg-shortlink/config"
	"eedg-shortlink/handlers"
	"eedg-shortlink/middleware"
	"eedg-shortlink/storage"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		log.Fatalf("mkdir: %v", err)
	}

	store, err := storage.NewStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer store.Close()

	app := &handlers.App{
		Store:        store,
		DefaultTTL:   cfg.DefaultTTL,
		BaseShortURL: cfg.BaseShortURL,
		FetchTimeout: cfg.FetchTimeout,
		MaxContentMB: cfg.MaxContentMB,
	}

	// Фоновая чистка просроченного кеша.
	stopCleanup := startCleanup(store, cfg.CleanupPeriod)
	defer stopCleanup()

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))

	// Rate limit только на API создания/обновления.
	apiLimiter := httprate.LimitByIP(cfg.RateLimitRPM, time.Minute)
	apiAuth := middleware.APIKey(cfg.APIKey)

	r.Route("/api", func(api chi.Router) {
		api.Use(apiLimiter)
		api.Use(apiAuth)

		api.Post("/links", app.HandleCreate)
		api.Get("/links/{code}", app.HandleInfo)
		api.Post("/links/{code}/refresh", app.HandleRefreshCache)
	})

	// Публичный редирект — без авторизации и без rate-limit.
	r.Get("/{code}", app.HandleRedirect)

	r.Get("/v1/health", app.HealthCheck)

	httpPort := getEnv("HTTP_PORT", "8443")

	srv := &http.Server{
		Addr:              ":" + httpPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Graceful shutdown.
	go func() {
		log.Printf("listening on http://%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// startCleanup запускает горутину периодической чистки кеша.
func startCleanup(store *storage.Store, period time.Duration) func() {
	t := time.NewTicker(period)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-t.C:
				if err := store.DeleteExpired(); err != nil {
					log.Printf("cleanup: %v", err)
				}
			case <-done:
				t.Stop()
				return
			}
		}
	}()
	return func() { close(done) }
}
