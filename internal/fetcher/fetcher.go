package fetcher

import (
	"encoding/json"
	"fmt"
	"gmonitor/internal/models"
	"io"
	"log"
	"net/http"
	"time"
)

type HTTPFetcher func(url, token string) (*http.Response, error)

type GitHubFetcher struct {
	Request HTTPFetcher
}

func NewGitHubFetcher() *GitHubFetcher {
	return &GitHubFetcher{Request: makeGitHubRequest}
}

func (f *GitHubFetcher) FetchRepository(repoName, token string) (*models.Repository, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s", repoName)

	resp, err := f.Request(url, token)
	if err != nil {
		return nil, fmt.Errorf("error fetching repository: %v", err)
	}
	defer resp.Body.Close()

	var repo GitHubRepositoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("error decoding repository JSON: %v", err)
	}

	return &models.Repository{
		Name:            repo.FullName,
		Description:     repo.Description,
		URL:             repo.Url,
		Language:        repo.Language,
		ForksCount:      repo.ForksCount,
		StarsCount:      repo.StargazersCount,
		OpenIssuesCount: repo.OpenIssuesCount,
		WatchersCount:   repo.WatchersCount,
		CreatedAt:       repo.CreatedAt,
		UpdatedAt:       repo.UpdatedAt,
	}, nil
}

func (f *GitHubFetcher) FetchCommits(repoName, token, since string) ([]models.Commit, error) {
	until, err := time.Parse(time.RFC3339, since)
	if err != nil {
		log.Printf("Failed to parse time: %v", err)
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits?since=%s&until=%s&per_page=100", repoName, since, until.Add(time.Hour).Format(time.RFC3339))

	resp, err := f.Request(url, token)
	if err != nil {
		return nil, fmt.Errorf("error fetching commits: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}(resp.Body)

	var commits []GitHubCommitResponse
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, fmt.Errorf("error decoding commits JSON: %v", err)
	}

	if len(commits) == 0 {
		log.Printf("No new commits found for repository: %s since %s", repoName, since)
		return []models.Commit{}, nil
	}

	commitRecords := make([]models.Commit, 0, len(commits))
	for _, commit := range commits {
		commitRecords = append(commitRecords, models.Commit{
			CommitHash: commit.SHA,
			Author:     commit.Commit.Author.Name,
			Message:    commit.Commit.Message,
			CommitURL:  commit.HTMLURL,
			CommitDate: commit.Commit.Author.Date,
		})
	}

	return commitRecords, nil
}
