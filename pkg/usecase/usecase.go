package usecase

import (
	"context"
	"github.com/sambaiz/go-explorer/pkg/db"
	"github.com/sambaiz/go-explorer/pkg/gomod"
	"github.com/sambaiz/go-explorer/pkg/log"
	"github.com/sambaiz/go-explorer/pkg/report"
	"github.com/sambaiz/go-explorer/pkg/repository"
	"golang.org/x/xerrors"
	"io"
	"time"
)

// UpdateDB updates repository database
func UpdateDB(ctx context.Context, dbClient db.DB, fetchers []repository.Fetcher, now time.Time) error {
	for _, fetcher := range fetchers {
		log.Logger.Infof("UpdateDB() starts with fetcher %s", fetcher.Name())
		repos, err := fetcher.FetchGoRepositories(ctx, now)
		if err != nil {
			return err
		}
		for i, repo := range repos {
			log.Logger.Infof("UpdateDB() %d/%d repositories", i+1, len(repos))
			goModFiles, err := FetchGoModFiles(ctx, repo.Path, fetchers)
			if err != nil {
				return err
			}
			var goMods []db.GoMod
			for _, goMod := range goModFiles {
				goMods = append(goMods, db.GoMod{
					GoMod: goMod,
				})
				f, err := gomod.Parse([]byte(goMod))
				if err != nil {
					log.Logger.Warnf("%+v", err)
					continue
				}
				for _, require := range f.Require {
					if _, err := FetchAndUpsertRepository(ctx, dbClient, fetchers, require.Mod.Path); err != nil {
						return err
					}
				}
			}
			if err := dbClient.UpsertRepository(db.Repository{
				Path:            repo.Path,
				Description:     repo.Description,
				GoMods:          goMods,
				LastCommittedAt: repo.LastCommittedAt,
			}); err != nil {
				log.Logger.Warnf("%+v", err)
				continue
			}
		}
	}
	return nil
}

// FetchGoModFiles fetches all go.mod files by any fetchers
func FetchGoModFiles(ctx context.Context, repositoryPath string, fetchers []repository.Fetcher) ([]string, error) {
	var (
		goModFiles []string
		err        error
	)
	for _, fetcher := range fetchers {
		goModFiles, err = fetcher.FetchGoModFiles(ctx, repositoryPath)
		if err != nil {
			if xerrors.Is(err, repository.ErrRateLimitExceeded) {
				return nil, err
			}
			if !xerrors.Is(err, repository.ErrUnsupported) {
				log.Logger.Warnf("%+v", err)
				break
			}
		} else {
			break
		}
	}
	return goModFiles, nil
}

// FetchAndUpsertRepository fetches repository from DB. If not exists, fetches original by any fetchers and saves to DB.
func FetchAndUpsertRepository(ctx context.Context, dbClient db.DB, fetchers []repository.Fetcher, path string) (*repository.Repository, error) {
	dbRepo, err := dbClient.FetchRepository(path)
	if err != nil {
		if !xerrors.Is(err, db.ErrNotFound) {
			return nil, err
		}
	} else {
		return &repository.Repository{
			Path:            dbRepo.Path,
			Description:     dbRepo.Description,
			LastCommittedAt: dbRepo.LastCommittedAt,
		}, nil
	}
	var repo *repository.Repository
	for _, fetcher := range fetchers {
		repo, err = fetcher.FetchRepository(ctx, path)
		if err != nil {
			if !xerrors.Is(err, repository.ErrUnsupported) {
				return nil, err
			}
		}
	}
	if repo == nil {
		// upsert only path
		repo = &repository.Repository{
			Path: path,
		}
	}
	if err := dbClient.UpsertRepository(db.Repository{
		Path:            repo.Path,
		Description:     repo.Description,
		GoMods:          nil, // does not fetch here to suppress request
		LastCommittedAt: repo.LastCommittedAt,
	}); err != nil {
		return nil, err
	}
	return repo, nil
}

// MakeReport makes report and writes it
func MakeReport(writer io.Writer, dbClient db.DB, activePeriod time.Duration) error {
	repos, err := dbClient.FetchActiveRepositories(activePeriod)
	if err != nil {
		return err
	}
	depCountByPath := map[string]int{}
	for i, repo := range repos {
		log.Logger.Infof("MakeReport() %d/%d repositories", i+1, len(repos))
		for _, goMod := range repo.GoMods {
			f, err := gomod.Parse([]byte(goMod.GoMod))
			if err != nil {
				return err
			}
			for _, require := range f.Require {
				depCountByPath[require.Mod.Path]++
			}
		}
	}
	rep := report.Report{}
	rep.Summary = report.Summary{
		UpdatedAt:           time.Now().Unix(),
		ActiveRepositoryNum: len(repos),
	}
	for path, count := range depCountByPath {
		repo, err := dbClient.FetchRepository(path)
		if err != nil {
			if !xerrors.Is(err, db.ErrNotFound) {
				log.Logger.Warnf("%+v", err)
			}
			rep.Modules = append(rep.Modules, report.Module{
				Path:           path,
				ActiveDepCount: count,
			})
		} else {
			rep.Modules = append(rep.Modules, report.Module{
				Path:           repo.Path,
				Description:    repo.Description,
				ActiveDepCount: count,
			})
		}
	}
	if err := rep.Write(writer); err != nil {
		return err
	}
	return nil
}
