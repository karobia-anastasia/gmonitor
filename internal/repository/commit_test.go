package repository_test

import (
	"context"
	"github.com/glebarez/sqlite"
	"gmonitor/internal/models"
	"gmonitor/internal/repository"
	"gorm.io/gorm"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect test db: %v", err)
	}

	if err := db.AutoMigrate(&models.Repository{}, &models.Commit{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	return db
}

func TestSaveCommit_NewCommit(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCommitRepo(db)

	commit := &models.Commit{
		CommitHash: "hash123",
		Author:     "Alice",
		Message:    "Initial commit",
		CommitDate: time.Now(),
	}

	err := repo.SaveCommit(context.Background(), commit)
	if err != nil {
		t.Errorf("failed to save new commit: %v", err)
	}
}

func TestSaveCommit_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCommitRepo(db)

	commit := &models.Commit{
		CommitHash: "dupHash",
		Author:     "Bob",
		Message:    "First",
		CommitDate: time.Now(),
	}
	_ = repo.SaveCommit(context.Background(), commit)

	// Attempt to save the same commit again (should not result in error)
	err := repo.SaveCommit(context.Background(), commit)
	if err != nil {
		t.Errorf("expected no error for duplicate commit, got: %v", err)
	}
}

func TestGetCommitsByRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCommitRepo(db)

	r := models.Repository{Name: "test-repo"}
	db.Create(&r)

	commits := []models.Commit{
		{CommitHash: "1", Author: "Alice", Message: "m1", RepoID: r.ID, CommitDate: time.Now()},
		{CommitHash: "2", Author: "Bob", Message: "m2", RepoID: r.ID, CommitDate: time.Now()},
	}
	db.Create(&commits)

	found, err := repo.GetCommitsByRepository(context.Background(), "test-repo", 20, 1)
	if err != nil {
		t.Fatalf("failed to get commits: %v", err)
	}
	if len(found) != 2 {
		t.Errorf("expected 2 commits, got: %d", len(found))
	}
}

func TestSaveCommits(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCommitRepo(db)

	r := models.Repository{Name: "bulk-repo"}
	db.Create(&r)

	commits := []models.Commit{
		{CommitHash: "h1", Author: "A", Message: "m1", CommitDate: time.Now()},
		{CommitHash: "h2", Author: "B", Message: "m2", CommitDate: time.Now()},
	}

	err := repo.SaveCommits(context.Background(), r.ID, commits)
	if err != nil {
		t.Fatalf("failed to save commits: %v", err)
	}

	var count int64
	db.Model(&models.Commit{}).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 commits, got: %d", count)
	}
}

func TestGetTopCommitAuthors(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCommitRepo(db)

	r := models.Repository{Name: "stats-repo"}
	db.Create(&r)

	commits := []models.Commit{
		{CommitHash: "a1", Author: "Dev1", RepoID: r.ID, CommitDate: time.Now()},
		{CommitHash: "a2", Author: "Dev1", RepoID: r.ID, CommitDate: time.Now()},
		{CommitHash: "a3", Author: "Dev2", RepoID: r.ID, CommitDate: time.Now()},
	}
	db.Create(&commits)

	top, err := repo.GetTopCommitAuthors(context.Background(), 2)
	if err != nil {
		t.Fatalf("failed to get top authors: %v", err)
	}
	if len(top) != 2 || top[0].Author != "Dev1" || top[0].Count != 2 {
		t.Errorf("unexpected top authors: %+v", top)
	}
}

func TestGetLatestCommitDate(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCommitRepo(db)

	now := time.Now().UTC()
	commits := []models.Commit{
		{CommitHash: "c1", Author: "X", CommitDate: now.Add(-2 * time.Hour)},
		{CommitHash: "c2", Author: "Y", CommitDate: now},
	}
	db.Create(&commits)

	latest, err := repo.GetLatestCommitDate(context.Background())
	if err != nil {
		t.Fatalf("failed to get latest commit date: %v", err)
	}
	if !latest.Equal(now) {
		t.Errorf("expected latest date %v, got %v", now, latest)
	}
}
