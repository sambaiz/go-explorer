package db

import (
	"golang.org/x/xerrors"
	"time"
)

// ErrNotFound describes the resource is not found
var ErrNotFound = xerrors.New("resource not found")

// DB is interface to access DB
type DB interface {
	Close()
	FetchRepository(path string) (*Repository, error)
	FetchActiveRepositories(period time.Duration) ([]Repository, error)
	UpsertRepository(repository Repository) error
}

// Repository table
type Repository struct {
	ID              int    `gorm:"primary_key"`
	Path            string `gorm:"unique;not null"`
	Description     string `gorm:"type:text"`
	GoMods          []GoMod
	LastCommittedAt time.Time
	UpdatedAt       time.Time
}

// GoMod table
type GoMod struct {
	ID           int `gorm:"primary_key"`
	RepositoryID int
	GoMod        string `gorm:"type:text"`
}
