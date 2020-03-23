package main

import (
	"context"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sambaiz/go-explorer/pkg/db/sqlite3"
	"github.com/sambaiz/go-explorer/pkg/log"
	"github.com/sambaiz/go-explorer/pkg/repository"
	"github.com/sambaiz/go-explorer/pkg/repository/github"
	"github.com/sambaiz/go-explorer/pkg/usecase"
	"os"
	"time"
)

func main() {
	ctx := context.Background()
	dbClient, err := sqlite3.Open("./data.db")
	if err != nil {
		panic(err)
	}
	defer dbClient.Close()
	if err := usecase.UpdateDB(ctx, dbClient, []repository.Fetcher{
		github.New(ctx),
	}, time.Now()); err != nil {
		log.Logger.Errorf("%+v", err)
		return
	}
	f, err := os.Create("./report.json")
	if err != nil {
		log.Logger.Errorf("%+v", err)
		return
	}
	defer f.Close()
	if err := usecase.MakeReport(f, dbClient, time.Hour*71); err != nil {
		log.Logger.Errorf("%+v", err)
		return
	}
}
