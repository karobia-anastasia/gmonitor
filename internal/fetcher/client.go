package fetcher

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// makeGitHubRequest sends an authenticated request to the GitHub API
func makeGitHubRequest(url, token string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("error closing body: %v", err)
			}
		}(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}
	return resp, nil
}
