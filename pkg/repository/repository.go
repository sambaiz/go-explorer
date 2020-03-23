package repository

import (
	"context"
	"golang.org/x/xerrors"
	"time"
)

var (
	// ErrUnsupported describes the fetcher does not support the request
	ErrUnsupported = xerrors.New("fetcher does not support")
	// ErrRateLimitExceeded describes some API rate limit exceeded
	ErrRateLimitExceeded = xerrors.New("rate limit exceeded")
)

// Repository to fetch
type Repository struct {
	Path            string
	Description     string
	LastCommittedAt time.Time
}

// Fetcher of repository
type Fetcher interface {
	Name() string
	FetchRepository(ctx context.Context, path string) (*Repository, error)
	FetchGoRepositories(ctx context.Context, day time.Time) ([]Repository, error)
	FetchGoModFiles(ctx context.Context, repositoryPath string) ([]string, error)
}
