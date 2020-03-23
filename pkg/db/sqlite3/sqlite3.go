package sqlite3

import (
	"github.com/jinzhu/gorm"
	"github.com/sambaiz/go-explorer/pkg/db"
	"golang.org/x/xerrors"
	"time"
)

type sqlite3 struct {
	client *gorm.DB
}

// Open DB file
func Open(fileName string) (db.DB, error) {
	client, err := gorm.Open("sqlite3", fileName)
	if err != nil {
		return nil, xerrors.Errorf("failed to open SQLite file: %w", err)
	}
	if err := client.AutoMigrate(&db.Repository{}, &db.GoMod{}).Error; err != nil {
		return nil, xerrors.Errorf("failed to migrate: %w", err)
	}
	return &sqlite3{client: client}, nil
}

// Close DB
func (s *sqlite3) Close() {
	s.client.Close()
}

// FetchRepository fetches a repository by path
func (s *sqlite3) FetchRepository(path string) (*db.Repository, error) {
	var repository db.Repository
	if err := s.client.Preload("GoMods").Where("path = ?", path).First(&repository).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, xerrors.Errorf("%s is not found: %w", path, db.ErrNotFound)
		}
		return nil, xerrors.Errorf("failed to fetch %s: %w", path, err)
	}
	return &repository, nil
}

// FetchActiveRepositories fetches repositories updated in period
func (s *sqlite3) FetchActiveRepositories(period time.Duration) ([]db.Repository, error) {
	var repositories []db.Repository
	if err := s.client.Preload("GoMods").Where("last_committed_at >= ?", time.Now().Add(-1*period)).
		Find(&repositories).Error; err != nil {
		return nil, xerrors.Errorf("failed to fetch active repositories: %w", err)
	}
	return repositories, nil
}

// UpsertRepository inserts or updates repository
func (s *sqlite3) UpsertRepository(repository db.Repository) error {
	var old db.Repository
	if s.client.Where("path = ?", repository.Path).First(&old).RecordNotFound() {
		if err := s.client.Create(&repository).Error; err != nil {
			return xerrors.Errorf("failed to create repository %s: %w", repository.Path, err)
		}
	} else {
		if err := s.client.Model(&old).Association("GoMods").Clear().Error; err != nil {
			return xerrors.Errorf("failed to clear old gomods: %w", err)
		}
		repository.ID = old.ID
		if err := s.client.Save(&repository).Error; err != nil {
			return xerrors.Errorf("failed to update repository %s: %w", repository.Path, err)
		}
	}
	return nil
}
