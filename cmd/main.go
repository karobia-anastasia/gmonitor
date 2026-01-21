package main

import (
	"context"
	_ "fmt"
	"github.com/joho/godotenv"
	"gmonitor/config"
	"gmonitor/internal/db"
	"gmonitor/internal/fetcher"
	"gmonitor/internal/monitor"
	"gmonitor/internal/repository"
	"gmonitor/internal/server"
	"gmonitor/pkg/cache"
	"log"
	"os"
	"os/signal"
	"syscall"
	_ "time"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: failed to load .env file: %v", err)
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize database
	dbConf := db.NewConfig(cfg.DatabaseURL)
	database, err := db.Connect(dbConf)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	err = db.Migrate(database)
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}
	defer db.Close(database)

	// Initialize repositories
	repoRepo := repository.NewRepositoryRepo(database)
	commitRepo := repository.NewCommitRepo(database)

	// Initialize fetcher
	fetch := fetcher.NewGitHubFetcher()

	//Initialize Cache
	newCache := cache.NewCache(ctx, cfg.RedisHost, cfg.RedisPassword)

	// Start HTTP server
	go server.StartServer(ctx, *cfg, repoRepo, commitRepo, *fetch, newCache)

	// Start monitoring worker
	mon := monitor.NewMonitor(database, cfg.GitHubToken, cfg.PollInterval, *repoRepo, *commitRepo, *fetch)
	scheduler := monitor.NewWorker(mon, *repoRepo)
	go scheduler.Start(ctx)

	// Handle shutdown signals
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Wait for shutdown
	<-ctx.Done()
	log.Println("Service shutting down...")
}
