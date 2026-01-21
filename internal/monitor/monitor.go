package monitor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gmonitor/internal/fetcher"
	"gmonitor/internal/repository"
	"gorm.io/gorm"
	"log"
	"time"
)

// Monitor is responsible for tracking repositories and fetching new commits
type Monitor struct {
	DB             *gorm.DB
	GitHubToken    string
	Interval       time.Duration
	RepositoryRepo repository.RepositoryRepo
	CommitRepo     repository.CommitRepo
	Fetcher        fetcher.GitHubFetcher
}

// NewMonitor initializes a new Monitor instance
func NewMonitor(db *gorm.DB, githubToken string, interval time.Duration, repo repository.RepositoryRepo, commitRepo repository.CommitRepo, fetcher fetcher.GitHubFetcher) *Monitor {
	return &Monitor{
		DB:             db,
		GitHubToken:    githubToken,
		Interval:       interval,
		RepositoryRepo: repo,
		CommitRepo:     commitRepo,
		Fetcher:        fetcher,
	}
}

// FetchNewCommits retrieves new commits for a given repository and updates the database
func (m *Monitor) FetchNewCommits(repoName string, ctx context.Context) error {
	log.Printf("Checking for new commits in repository: %s", repoName)

	// Create a context with timeout for GitHub API calls
	githubCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Get repository details
	repo, err := m.RepositoryRepo.GetRepository(githubCtx, repoName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("RepositoryRepo %s not found in the database. Skipping.", repoName)
			return nil
		}
		return fmt.Errorf("failed to get repository: %v\n", err)
	}

	//Get the last fetched commit
	commitDate, err := m.CommitRepo.GetLatestCommitDate(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest commit: %v\n", err)
	}

	// Fetch latest commits from GitHub API and ensure to fetch commits from the next second to avoid same fetching commits again
	commits, err := m.Fetcher.FetchCommits(repoName, m.GitHubToken, commitDate.Add(time.Hour).Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to fetch commits: %v\n", err)
	}

	// Save new commits to database
	if len(commits) > 0 {
		err := m.CommitRepo.SaveCommits(ctx, repo.ID, commits)
		if err != nil {
			return fmt.Errorf("failed to save commits: %v\n\n", err)
		}
		log.Printf("Added %d new commits for repository %s\n\n", len(commits), repoName)
	} else {
		log.Printf("No new commits found for repository %s\n\n", repoName)
	}

	return nil
}
