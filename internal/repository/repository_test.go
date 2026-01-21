package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/glebarez/sqlite"
	"gmonitor/internal/models"
	"gmonitor/internal/repository"
	"gorm.io/gorm"
	"testing"
	_ "time"
)

func setupRepoTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test DB: %v", err)
	}

	if err := db.AutoMigrate(&models.Repository{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	return db
}

func TestSaveRepository_Success(t *testing.T) {
	db := setupRepoTestDB(t)
	repoStore := repository.NewRepositoryRepo(db)

	repo := &models.Repository{
		Name:        "gmonitor",
		Description: "GitHub monitor service",
		URL:         "https://github.com/example/gmonitor",
		Language:    "Go",
	}

	err := repoStore.SaveRepository(context.Background(), repo)
	if err != nil {
		t.Errorf("expected save to succeed, got error: %v", err)
	}
}

func TestSaveRepository_ContextCancelled(t *testing.T) {
	db := setupRepoTestDB(t)
	repoStore := repository.NewRepositoryRepo(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := &models.Repository{Name: "canceled"}

	err := repoStore.SaveRepository(ctx, repo)

	if err == nil {
		t.Fatal("expected an error due to cancelled context, got nil")
	}
}

func TestGetRepository_Found(t *testing.T) {
	db := setupRepoTestDB(t)
	repoStore := repository.NewRepositoryRepo(db)

	expected := &models.Repository{
		Name:        "test-repo",
		Description: "Test Repo",
	}
	db.Create(expected)

	repo, err := repoStore.GetRepository(context.Background(), "test-repo")
	if err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}
	if repo.Name != expected.Name {
		t.Errorf("expected %s, got %s", expected.Name, repo.Name)
	}
}

func TestGetRepository_NotFound(t *testing.T) {
	db := setupRepoTestDB(t)
	repoStore := repository.NewRepositoryRepo(db)

	_, err := repoStore.GetRepository(context.Background(), "missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got: %v", err)
	}
}

func TestGetAllRepositories_WithData(t *testing.T) {
	db := setupRepoTestDB(t)
	repoStore := repository.NewRepositoryRepo(db)

	repos := []*models.Repository{
		{Name: "repo1"},
		{Name: "repo2"},
	}
	for _, r := range repos {
		db.Create(r)
	}

	all, err := repoStore.GetAllRepositories(context.Background())
	if err != nil {
		t.Fatalf("failed to get all repositories: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 repositories, got %d", len(all))
	}
}

func TestGetAllRepositories_Empty(t *testing.T) {
	db := setupRepoTestDB(t)
	repoStore := repository.NewRepositoryRepo(db)

	all, err := repoStore.GetAllRepositories(context.Background())
	if err != nil {
		t.Fatalf("failed to get repositories: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 repositories, got %d", len(all))
	}
}
