package repository

import (
	"context"
	"fmt"
	"gmonitor/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

// CommitRepo provides database operations for commits
type CommitRepo struct {
	db *gorm.DB
}

// NewCommitRepo creates a new repository instance
func NewCommitRepo(db *gorm.DB) *CommitRepo {
	return &CommitRepo{
		db: db,
	}
}

// SaveCommit saves a commit to the database if it doesn't already exist
func (r *CommitRepo) SaveCommit(ctx context.Context, commit *models.Commit) error {
	// Attempt to insert, skip if conflict on commit_hash
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(commit).Error

	if err != nil {
		return fmt.Errorf("failed to save commit: %w", err)
	}
	return nil
}

// SaveCommits saves multiple commits to the database
func (r *CommitRepo) SaveCommits(ctx context.Context, repoID uint, commits []models.Commit) error {
	if len(commits) == 0 {
		return nil
	}

	// Set RepoID for all commits
	for i := range commits {
		commits[i].RepoID = repoID
	}

	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(&commits).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save commits: %w", err)
	}

	return tx.Commit().Error
}

// GetTopCommitAuthors retrieves the top N commit authors by commit count
func (r *CommitRepo) GetTopCommitAuthors(ctx context.Context, limit int) ([]struct {
	Author string
	Count  int
}, error) {
	var results []struct {
		Author string
		Count  int
	}

	err := r.db.WithContext(ctx).
		Model(&models.Commit{}).
		Select("author, COUNT(*) AS count").
		Group("author").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get top commit authors: %w", err)
	}

	return results, nil
}

// GetLatestCommitDate retrieves only the most recent commit date
func (r *CommitRepo) GetLatestCommitDate(ctx context.Context) (time.Time, error) {
	var latest string

	// Fetch the latest commit date as a string
	err := r.db.WithContext(ctx).
		Model(&models.Commit{}).
		Select("MAX(commit_date)").
		Scan(&latest).Error

	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get latest commit date: %w", err)
	}

	// Replace space with 'T' for correct RFC3339 parsing
	latest = strings.Replace(latest, " ", "T", 1)

	// Parse the string into time.Time using RFC3339 format
	latestTime, err := time.Parse(time.RFC3339, latest)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse latest commit date: %w", err)
	}

	return latestTime, nil
}

// GetCommitsByRepository retrieves paginated commits for a given repository by name
func (r *CommitRepo) GetCommitsByRepository(ctx context.Context, repoName string, limit, offset int) ([]*models.Commit, error) {
	var commits []*models.Commit

	err := r.db.WithContext(ctx).
		Model(&models.Commit{}).
		Joins("JOIN repositories ON commits.repo_id = repositories.id").
		Where("repositories.name = ?", repoName).
		Limit(limit).
		Offset(offset).
		Order("commit_date DESC").
		Find(&commits).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get commits for repository %q: %w", repoName, err)
	}

	return commits, nil
}
