package server

import (
	"context"
	"errors"
	"fmt"
	"gmonitor/config"
	"gmonitor/internal/fetcher"
	"gmonitor/internal/repository"
	"gmonitor/pkg/cache"
	"log"
	"net/http"
)

// StartServer initializes and starts the HTTP server
func StartServer(ctx context.Context, cfg config.Config, repoRepo *repository.RepositoryRepo, commitRepo *repository.CommitRepo, fetcher fetcher.GitHubFetcher, cache *cache.Cache,
) {
	mux := http.NewServeMux()

	// Register handlers
	RegisterHandlers(mux, repoRepo, commitRepo, fetcher, ctx, cfg.GitHubToken, cache)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.PORT),
		Handler: mux,
	}

	// Run the server
	go func() {
		log.Println("Starting HTTP server on port", cfg.PORT)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Listen for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down HTTP server...")
	err := server.Shutdown(ctx)
	if err != nil {
		return
	}
}
