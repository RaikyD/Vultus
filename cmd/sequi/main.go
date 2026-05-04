package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sequi/internal/auth"
	"sequi/internal/config"
	"sequi/internal/db"
	dbgen "sequi/internal/db/gen"
	apphttp "sequi/internal/http"
	"sequi/internal/stream"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	// Reset any streams left in "live" state from a previous crashed/restarted session
	if err := dbgen.New(pool).ResetLiveStreams(ctx); err != nil {
		log.Printf("startup: reset live streams: %v", err)
	}

	sessionStore := auth.NewStore(cfg.SessionSecret)

	router := apphttp.NewRouter(pool, sessionStore, apphttp.RouterConfig{
		HLSBaseURL:    cfg.HLSBaseURL,
		HookSecret:    cfg.SRSHookSecret,
		MediaRoot:     cfg.MediaRoot,
		PublicBaseURL: cfg.PublicBaseURL,
	})

	// Start viewer count polling worker
	poller := stream.NewViewerPoller(pool, cfg.SRSAPIUrl)
	go poller.Run(ctx)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")

	cancel() // stop the poller

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
