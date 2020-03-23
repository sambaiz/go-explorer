package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v29/github"
	"github.com/sambaiz/go-explorer/pkg/log"
	"github.com/sambaiz/go-explorer/pkg/repository"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
	"os"
	"strings"
	"time"
)

type gitHub struct {
	client                 *github.Client
	retriedForAPIRateLimit bool
}

// New GitHub fetcher
func New(ctx context.Context) repository.Fetcher {
	return &gitHub{client: newClient(ctx)}
}

func newClient(ctx context.Context) *github.Client {
	var ts oauth2.TokenSource
	if tk := os.Getenv("GITHUB_API_TOKEN"); tk != "" {
		token := &oauth2.Token{AccessToken: tk}
		ts = oauth2.StaticTokenSource(token)
	}
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func available(paths []string) bool {
	return paths[0] == "github.com" && len(paths) >= 3
}

func (g *gitHub) waitIfAPIRateLimitExceeded(ctx context.Context, err error) (bool, error) {
	if err == nil {
		g.retriedForAPIRateLimit = false
		return false, nil
	}
	rerr, ok := err.(*github.RateLimitError)
	if !ok {
		g.retriedForAPIRateLimit = false
		return false, nil
	}
	if g.retriedForAPIRateLimit {
		return false, xerrors.Errorf("API rate limit exceeded: %w", repository.ErrRateLimitExceeded)
	}
	ctx, _ = context.WithTimeout(ctx, rerr.Rate.Reset.Sub(time.Now()))
	select {
	case <-ctx.Done():
	}
	g.retriedForAPIRateLimit = true
	return true, nil
}

// Name of fetcher
func (g *gitHub) Name() string {
	return "GitHub"
}

// FetchRepository fetches a repository by path
func (g *gitHub) FetchRepository(ctx context.Context, path string) (*repository.Repository, error) {
	paths := strings.Split(path, "/")
	if !available(paths) {
		return nil, xerrors.Errorf("%s is not GitHub path: %w", path, repository.ErrUnsupported)
	}
	repo, _, err := g.client.Repositories.Get(ctx, paths[1], paths[2])
	if shouldRetry, err := g.waitIfAPIRateLimitExceeded(ctx, err); err != nil {
		return nil, err
	} else if shouldRetry {
		return g.FetchRepository(ctx, path)
	}
	if err != nil {
		return nil, xerrors.Errorf("failed to get repository: %w", err)
	}
	return &repository.Repository{
		Path:            strings.Replace(repo.GetHTMLURL(), "https://", "", 1),
		Description:     fmt.Sprintf("%s: %s", repo.GetDescription(), strings.Join(repo.Topics, " ")),
		LastCommittedAt: repo.GetPushedAt().Time,
	}, nil
}

// FetchGoRepositories fetches go language repositories up to 1000
func (g *gitHub) FetchGoRepositories(ctx context.Context, day time.Time) ([]repository.Repository, error) {
	ret := []repository.Repository{}
	page := 0
	for {
		page++
		log.Logger.Infof("github.FetchGoRepositories() page %d", page)
		result, resp, err := g.client.Search.Repositories(ctx,
			fmt.Sprintf("language:Go pushed:>=%s", day.Format("2006-01-02")),
			&github.SearchOptions{
				Sort:  "committer-date",
				Order: "desc",
				ListOptions: github.ListOptions{
					Page: page,
					// you can also set a custom page size up to 100 with the ?per_page parameter
					// https://developer.github.com/v3/#pagination
					PerPage: 100,
				},
			})
		if shouldRetry, err := g.waitIfAPIRateLimitExceeded(ctx, err); err != nil {
			return nil, err
		} else if shouldRetry {
			return g.FetchGoRepositories(ctx, day)
		}
		if err != nil {
			return nil, xerrors.Errorf("failed to search repository: %w", err)
		}
		for _, repo := range result.Repositories {
			if repo.GetFork() {
				continue
			}
			ret = append(ret, repository.Repository{
				Path:            strings.Replace(repo.GetHTMLURL(), "https://", "", 1),
				Description:     fmt.Sprintf("%s: %s", repo.GetDescription(), strings.Join(repo.Topics, " ")),
				LastCommittedAt: repo.GetPushedAt().Time,
			})
		}
		if page >= resp.LastPage {
			break
		}
	}
	return ret, nil
}

// FetchGoModFiles fetches all go.mod in repository
func (g *gitHub) FetchGoModFiles(ctx context.Context, repositoryPath string) ([]string, error) {
	paths := strings.Split(repositoryPath, "/")
	if !available(paths) {
		return nil, xerrors.Errorf("%s is not github.com path: %w", repositoryPath, repository.ErrUnsupported)
	}
	res, _, err := g.client.Search.Code(ctx, fmt.Sprintf("repo:%s/%s filename:go.mod", paths[1], paths[2]), nil)
	if shouldRetry, err := g.waitIfAPIRateLimitExceeded(ctx, err); err != nil {
		return nil, err
	} else if shouldRetry {
		return g.FetchGoModFiles(ctx, repositoryPath)
	}
	if err != nil {
		return nil, xerrors.Errorf("failed to search codes: %w", err)
	}
	var goMods []string
	for _, result := range res.CodeResults {
		// remove mod.go and vendor module's go.mod etc.
		if result.GetName() != "go.mod" || strings.Contains(result.GetPath(), "vendor/") {
			continue
		}
		content, _, _, err := g.client.Repositories.GetContents(ctx,
			result.GetRepository().GetOwner().GetLogin(),
			result.GetRepository().GetName(),
			result.GetPath(),
			nil,
		)
		if shouldRetry, err := g.waitIfAPIRateLimitExceeded(ctx, err); err != nil {
			return nil, err
		} else if shouldRetry {
			return g.FetchGoModFiles(ctx, repositoryPath)
		}
		if err != nil {
			return nil, xerrors.Errorf("failed to get repository contents: %w", err)
		}
		goMod, err := content.GetContent()
		if err != nil {
			return nil, xerrors.Errorf("failed to GetContent(): %w", err)
		}
		goMods = append(goMods, goMod)
	}
	return goMods, nil
}
