package sqlite3_test

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sambaiz/go-explorer/pkg/db"
	"github.com/sambaiz/go-explorer/pkg/db/sqlite3"
	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"
	"os"
	"testing"
	"time"
)

func TestSqlite3_FetchRepository(t *testing.T) {
	if err := os.Remove(fmt.Sprintf("./%s.db", t.Name())); err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}
	sqlite3, err := sqlite3.Open(fmt.Sprintf("./%s.db", t.Name()))
	if err != nil {
		t.Error(err)
	}
	_, err = sqlite3.FetchRepository("example.com/repo")
	assert.True(t, xerrors.Is(err, db.ErrNotFound))

	if err = sqlite3.UpsertRepository(db.Repository{
		Path:            "example.com/repo",
		LastCommittedAt: time.Now().Add(time.Hour * -71),
	}); err != nil {
		t.Error(err)
	}
	repo, err := sqlite3.FetchRepository("example.com/repo")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "example.com/repo", repo.Path)
}

func TestSqlite3_FetchActiveRepositories(t *testing.T) {
	if err := os.Remove(fmt.Sprintf("./%s.db", t.Name())); err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}
	sqlite3, err := sqlite3.Open(fmt.Sprintf("./%s.db", t.Name()))
	if err != nil {
		t.Error(err)
	}
	if err = sqlite3.UpsertRepository(db.Repository{
		Path:            "example.com/repo1",
		LastCommittedAt: time.Now().Add(time.Hour * -71),
	}); err != nil {
		t.Error(err)
	}
	if err = sqlite3.UpsertRepository(db.Repository{
		Path:            "example.com/repo2",
		LastCommittedAt: time.Now().Add(time.Hour * -73),
	}); err != nil {
		t.Error(err)
	}
	actual, err := sqlite3.FetchActiveRepositories(time.Hour * 72)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, "example.com/repo1", actual[0].Path)
}

func TestSqlite3_UpsertRepository(t *testing.T) {
	if err := os.Remove(fmt.Sprintf("./%s.db", t.Name())); err != nil && !os.IsNotExist(err) {
		t.Error(err)
	}
	sqlite3, err := sqlite3.Open(fmt.Sprintf("./%s.db", t.Name()))
	if err != nil {
		t.Error(err)
	}
	if err = sqlite3.UpsertRepository(db.Repository{
		Path: "example.com/repo1",
		GoMods: []db.GoMod{
			{
				GoMod: "aaa",
			},
			{
				GoMod: "bbb",
			},
		},
	}); err != nil {
		t.Error(err)
	}
	actual, err := sqlite3.FetchRepository("example.com/repo1")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 2, len(actual.GoMods))

	if err = sqlite3.UpsertRepository(db.Repository{
		Path: "example.com/repo1",
		GoMods: []db.GoMod{
			{
				GoMod: "ccc",
			},
		},
		LastCommittedAt: time.Now().Add(time.Hour * -73),
	}); err != nil {
		t.Error(err)
	}
	actual, err = sqlite3.FetchRepository("example.com/repo1")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, len(actual.GoMods))
	assert.Equal(t, "ccc", actual.GoMods[0].GoMod)
}
