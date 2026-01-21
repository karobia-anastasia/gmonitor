package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gmonitor/internal/models"
	"gorm.io/gorm"
)

// RepositoryRepo provides database operations for repositories
type RepositoryRepo struct {
	db *gorm.DB
}

// NewRepositoryRepo creates a new repository instance
func NewRepositoryRepo(db *gorm.DB) *RepositoryRepo {
	return &RepositoryRepo{
		db: db,
	}
}

// SaveRepository stores a repository in the database
func (r *RepositoryRepo) SaveRepository(ctx context.Context, repo *models.Repository) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(repo).Error; err != nil {
			return fmt.Errorf("failed to save repository: %w", err)
		}
		return nil
	})
}

// GetRepository retrieves a repository by name
func (r *RepositoryRepo) GetRepository(ctx context.Context, name string) (*models.Repository, error) {
	var repo models.Repository

	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&repo).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	return &repo, nil
}

// GetAllRepositories retrieves all repositories from the database
func (r *RepositoryRepo) GetAllRepositories(ctx context.Context) ([]*models.Repository, error) {
	var repositories []*models.Repository

	err := r.db.WithContext(ctx).Find(&repositories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get repositories: %w", err)
	}

	return repositories, nil
}
