package github

import (
	"context"
	"github.com/sambaiz/go-explorer/pkg/gomod"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGitHub_FetchRepository(t *testing.T) {
	ctx := context.Background()
	repo, err := New(ctx).FetchRepository(ctx, "github.com/golang/go")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "github.com/golang/go", repo.Path)
}

func TestGitHub_FetchGoRepositories(t *testing.T) {
	ctx := context.Background()
	_, err := New(ctx).FetchGoRepositories(ctx, time.Now())
	if err != nil {
		t.Error(err)
	}
}

func TestGitHub_FetchGoModFiles(t *testing.T) {
	ctx := context.Background()
	gomods, err := New(ctx).FetchGoModFiles(ctx, "github.com/golang/go")
	if err != nil {
		t.Error(err)
	}
	for _, mod := range gomods {
		if _, err := gomod.Parse([]byte(mod)); err != nil {
			t.Error(err)
		}
	}
	assert.NotEmpty(t, gomods)
}
