package models

import (
	"gorm.io/gorm"
	"time"
)

// Repository represents a GitHub repository
type Repository struct {
	gorm.Model
	Name            string         `gorm:"unique;not null;size:255"`
	Description     string         `gorm:"type:TEXT"`
	URL             string         `gorm:"not null;size:255"`
	Language        string         `gorm:"size:50"`
	ForksCount      int            `gorm:"default:0"`
	StarsCount      int            `gorm:"default:0"`
	OpenIssuesCount int            `gorm:"default:0"`
	WatchersCount   int            `gorm:"default:0"`
	CreatedAt       time.Time      `gorm:"not null;type:DATETIME DEFAULT CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time      `gorm:"not null;type:DATETIME DEFAULT CURRENT_TIMESTAMP"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
	Commits         []Commit       `gorm:"foreignKey:RepoID;references:ID"`
}
