package server

import (
	"context"
	"encoding/json"
	"fmt"
	"gmonitor/internal/fetcher"
	"gmonitor/internal/repository"
	"gmonitor/pkg/cache"
	"log"
	"net/http"
	"strconv"
	"time"
)

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func RegisterHandlers(
	mux *http.ServeMux,
	repoRepo *repository.RepositoryRepo,
	commitRepo *repository.CommitRepo,
	fetcher fetcher.GitHubFetcher,
	ctx context.Context,
	token string,
	cache *cache.Cache,
) {
	mux.HandleFunc("POST /api/v1/repos", func(w http.ResponseWriter, r *http.Request) {
		handleAddRepo(w, r, repoRepo, commitRepo, fetcher, ctx, token, cache)
	})
	mux.HandleFunc("GET /api/v1/repos", func(w http.ResponseWriter, r *http.Request) {
		handleGetRepo(w, r, repoRepo, ctx, cache)
	})
	mux.HandleFunc("GET /api/v1/repos/commit-authors", func(w http.ResponseWriter, r *http.Request) {
		handleGetCommitAuthors(w, r, commitRepo, ctx, cache)
	})
	mux.HandleFunc("GET /api/v1/repos/commits", func(w http.ResponseWriter, r *http.Request) {
		handleGetRepoCommit(w, r, commitRepo, ctx, cache)
	})
}

func jsonResponse(w http.ResponseWriter, status int, success bool, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(JSONResponse{Success: success, Message: msg, Data: data})
}

func getFromCache(cache *cache.Cache, key string) (interface{}, bool) {
	val, found, err := cache.Get(key)
	if err != nil {
		log.Printf("cache get error for key '%s': %v", key, err)
	}
	return val, found
}

func setToCache(cache *cache.Cache, key string, val interface{}) {
	if err := cache.Set(key, val, 5*time.Minute); err != nil {
		log.Printf("cache set error for key '%s': %v", key, err)
	}
}

func handleAddRepo(
	w http.ResponseWriter,
	r *http.Request,
	repoRepo *repository.RepositoryRepo,
	commitRepo *repository.CommitRepo,
	fetcher fetcher.GitHubFetcher,
	ctx context.Context,
	token string,
	cache *cache.Cache,
) {
	var req struct {
		Repo  string `json:"repo"`
		Owner string `json:"owner"`
		Date  string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, false, "Invalid request payload", nil)
		return
	}

	repoName := fmt.Sprintf("%s/%s", req.Owner, req.Repo)

	if cached, found := getFromCache(cache, repoName); found {
		jsonResponse(w, http.StatusOK, true, "Repository found in cache", cached)
		return
	}

	repo, err := fetcher.FetchRepository(repoName, token)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, false, "Failed to fetch repository", nil)
		return
	}

	setToCache(cache, repoName, repo)

	if err := repoRepo.SaveRepository(ctx, repo); err != nil {
		jsonResponse(w, http.StatusInternalServerError, false, "Failed to save repository", nil)
		return
	}

	since, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		log.Printf("invalid date format: %v", err)
	}

	log.Printf("Pulling commits for %s since %s", repoName, since.Format(time.RFC850))

	commits, err := fetcher.FetchCommits(repoName, token, req.Date)
	if err != nil {
		log.Printf("Failed to fetch commits: %v", err)
	}

	if err := commitRepo.SaveCommits(ctx, repo.ID, commits); err != nil {
		log.Printf("Failed to save commits: %v", err)
	}

	jsonResponse(w, http.StatusCreated, true, "Repository added successfully", repo)
}

func handleGetRepo(w http.ResponseWriter, r *http.Request, repoRepo *repository.RepositoryRepo, ctx context.Context, cache *cache.Cache) {
	repoName := r.URL.Query().Get("repo")

	if cached, found := getFromCache(cache, repoName); found {
		jsonResponse(w, http.StatusOK, true, "Repository found in cache", cached)
		return
	}

	repo, err := repoRepo.GetRepository(ctx, repoName)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, false, "Repository not found", nil)
		return
	}

	setToCache(cache, repoName, repo)
	jsonResponse(w, http.StatusOK, true, "Repository found", repo)
}

func handleGetCommitAuthors(w http.ResponseWriter, r *http.Request, commitRepo *repository.CommitRepo, ctx context.Context, cache *cache.Cache) {
	repoName := r.URL.Query().Get("repo")
	if repoName == "" {
		jsonResponse(w, http.StatusBadRequest, false, "Repository name required", nil)
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		jsonResponse(w, http.StatusBadRequest, false, "Invalid limit value", nil)
		return
	}

	cacheKey := fmt.Sprintf("%s_authors_%d", repoName, limit)
	if cached, found := getFromCache(cache, cacheKey); found {
		jsonResponse(w, http.StatusOK, true, "Commit authors found in cache", cached)
		return
	}

	authors, err := commitRepo.GetTopCommitAuthors(ctx, limit)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, false, "Failed to fetch commit authors", nil)
		return
	}

	setToCache(cache, cacheKey, authors)
	jsonResponse(w, http.StatusOK, true, "Commit authors retrieved", authors)
}

func handleGetRepoCommit(w http.ResponseWriter, r *http.Request, commitRepo *repository.CommitRepo, ctx context.Context, cache *cache.Cache) {
	repo := r.URL.Query().Get("repo")
	if repo == "" {
		jsonResponse(w, http.StatusBadRequest, false, "Missing repository name", nil)
		return
	}

	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size <= 0 {
		size = 20
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * size
	cacheKey := fmt.Sprintf("%s_commits_%d_%d", repo, size, page)

	if cached, found := getFromCache(cache, cacheKey); found {
		jsonResponse(w, http.StatusOK, true, "Commits found in cache", cached)
		return
	}

	commits, err := commitRepo.GetCommitsByRepository(ctx, repo, size, offset)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, false, "Failed to fetch commits", nil)
		return
	}

	setToCache(cache, cacheKey, commits)
	jsonResponse(w, http.StatusOK, true, "Commits retrieved", commits)
}
