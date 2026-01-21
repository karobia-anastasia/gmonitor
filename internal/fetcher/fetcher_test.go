package fetcher

import (
	"errors"
	_ "gmonitor/internal/models"
	"io"
	"net/http"
	"strings"
	"testing"
	_ "time"
)

func mockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestFetchRepository_Success(t *testing.T) {
	mockFetcher := &GitHubFetcher{
		Request: func(url, token string) (*http.Response, error) {
			body := `{
				"full_name": "chromium/chromium",
				"description": "desc",
				"html_url": "https://github.com/chromium/chromium",
				"language": "Go",
				"forks_count": 2,
				"stargazers_count": 3,
				"open_issues_count": 1,
				"watchers_count": 4,
				"created_at": "2022-01-01T00:00:00Z",
				"updated_at": "2022-02-01T00:00:00Z"
			}`
			return mockResponse(200, body), nil
		},
	}

	repo, err := mockFetcher.FetchRepository("chromium/chromium", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Name != "chromium/chromium" {
		t.Errorf("unexpected repo name: %s", repo.Name)
	}
}

func TestFetchRepository_JSONError(t *testing.T) {
	mockFetcher := &GitHubFetcher{
		Request: func(url, token string) (*http.Response, error) {
			return mockResponse(200, `{invalid json}`), nil
		},
	}

	_, err := mockFetcher.FetchRepository("chromium/chromium", "")
	if err == nil || !strings.Contains(err.Error(), "error decoding") {
		t.Errorf("expected JSON decoding error, got: %v", err)
	}
}

func TestFetchRepository_RequestError(t *testing.T) {
	mockFetcher := &GitHubFetcher{
		Request: func(url, token string) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}

	_, err := mockFetcher.FetchRepository("chromium/chromium", "")
	if err == nil || !strings.Contains(err.Error(), "error fetching repository") {
		t.Errorf("expected request error, got: %v", err)
	}
}

func TestFetchCommits_Success(t *testing.T) {
	mockFetcher := &GitHubFetcher{
		Request: func(url, token string) (*http.Response, error) {
			body := `[
				{
					"sha": "abc123",
					"commit": {
						"author": { "name": "dev", "date": "2023-01-01T12:00:00Z" },
						"message": "initial commit"
					},
					"html_url": "https://github.com/chromium/chromium/commit/abc123"
				}
			]`
			return mockResponse(200, body), nil
		},
	}

	since := "2023-01-01T00:00:00Z"
	commits, err := mockFetcher.FetchCommits("chromium/chromium", "", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 1 || commits[0].CommitHash != "abc123" {
		t.Errorf("unexpected commits: %+v", commits)
	}
}

func TestFetchCommits_Empty(t *testing.T) {
	mockFetcher := &GitHubFetcher{
		Request: func(url, token string) (*http.Response, error) {
			return mockResponse(200, `[]`), nil
		},
	}

	commits, err := mockFetcher.FetchCommits("chromium/chromium", "", "2023-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 0 {
		t.Errorf("expected empty commits, got: %d", len(commits))
	}
}

func TestFetchCommits_JSONError(t *testing.T) {
	mockFetcher := &GitHubFetcher{
		Request: func(url, token string) (*http.Response, error) {
			return mockResponse(200, `invalid-json`), nil
		},
	}

	_, err := mockFetcher.FetchCommits("chromium/chromium", "", "2023-01-01T00:00:00Z")
	if err == nil || !strings.Contains(err.Error(), "error decoding commits JSON") {
		t.Errorf("expected decoding error, got: %v", err)
	}
}

func TestFetchCommits_RequestError(t *testing.T) {
	mockFetcher := &GitHubFetcher{
		Request: func(url, token string) (*http.Response, error) {
			return nil, errors.New("mock failure")
		},
	}

	_, err := mockFetcher.FetchCommits("chromium/chromium", "", "2023-01-01T00:00:00Z")
	if err == nil || !strings.Contains(err.Error(), "error fetching commits") {
		t.Errorf("expected fetch error, got: %v", err)
	}
}

func TestFetchCommits_InvalidTime(t *testing.T) {
	mockFetcher := &GitHubFetcher{
		Request: func(url, token string) (*http.Response, error) {
			return mockResponse(200, `[]`), nil
		},
	}

	// Should not error, just log internally
	_, err := mockFetcher.FetchCommits("chromium/chromium", "", "invalid-time")
	if err != nil {
		t.Errorf("expected no hard error, got: %v", err)
	}
}
