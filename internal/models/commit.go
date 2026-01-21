package models

import (
	"gorm.io/gorm"
	"time"
)

type Commit struct {
	gorm.Model
	RepoID     uint      `gorm:"not null;index:idx_repo_id_commit_hash;type:INTEGER"`
	CommitHash string    `gorm:"unique;not null;size:40"`
	Author     string    `gorm:"not null;size:255"`
	Message    string    `gorm:"not null;type:TEXT"`
	CommitDate time.Time `gorm:"not null;type:DATETIME DEFAULT CURRENT_TIMESTAMP"`
	CommitURL  string    `gorm:"not null;size:255"`
}
