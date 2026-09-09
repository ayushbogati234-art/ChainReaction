package app

import (
	"context"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/config"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/ingest/github"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/jobs"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/resolver/python"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/service"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/vulnerability/osv"
	"net/http"
)

func RunWorker(ctx context.Context, c config.Config) error {
	store, graph, closeAll, err := dependencies(ctx, c)
	if err != nil {
		return err
	}
	defer closeAll()
	enricher := service.FindingService{Store: store, Scans: store, Graph: graph, Provider: osv.Client{}}
	runner := service.ScanService{Findings: &enricher, Store: store, Graph: graph, Ingest: github.Client{HTTP: &http.Client{Timeout: c.GitHubTimeout}, MaxBytes: c.GitHubMaxBytes}, Resolver: python.Resolver{}}
	worker := jobs.Worker{Job: jobs.ScanJob{Queue: store, Runner: runner}}
	worker.Run(ctx)
	return nil
}
