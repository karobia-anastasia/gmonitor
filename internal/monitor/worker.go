package monitor

import (
	"context"
	"fmt"
	"gmonitor/internal/models"
	"gmonitor/internal/repository"
	"log"
	"sync"
	"time"
)

// Worker periodically fetches updates for all repositories
type Worker struct {
	Monitor    *Monitor
	Repository repository.RepositoryRepo
}

// NewWorker initializes a new Worker instance
func NewWorker(monitor *Monitor, repo repository.RepositoryRepo) *Worker {
	return &Worker{Monitor: monitor, Repository: repo}
}

// Start runs the monitoring process periodically
func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.Monitor.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker: Shutting down monitoring process...")
			return

		case <-ticker.C:
			log.Println("Starting repository monitoring cycle...")
			if err := w.processRepositories(ctx); err != nil {
				log.Printf("Worker: Error processing repositories: %v", err)
			}
		}
	}
}

// processRepositories handles repository processing with error handling
func (w *Worker) processRepositories(ctx context.Context) error {
	repos, err := w.Repository.GetAllRepositories(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch repositories: %v", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(repos))

	for _, repo := range repos {
		wg.Add(1)
		go func(repo models.Repository) {
			defer wg.Done()
			if err := w.Monitor.FetchNewCommits(repo.Name, ctx); err != nil {
				errChan <- fmt.Errorf("error updating commits for %s: %v", repo.Name, err)
			}
		}(*repo)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Process errors and return first non-nil error
	var firstError error
	for err := range errChan {
		if err != nil {
			if firstError == nil {
				firstError = err
			}
			log.Printf("Worker: %v", err)
		}
	}

	return firstError
}
