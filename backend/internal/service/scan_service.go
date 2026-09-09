package service

import (
	"context"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/graph"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/ingest/github"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/resolver"
)

type ScanStore interface {
	CreateScan(context.Context, string) (domain.Scan, error)
	GetScan(context.Context, string) (domain.Scan, error)
	SetStage(context.Context, string, string, int) error
	Complete(context.Context, string, string, bool, []string) error
}
type Ingestor interface {
	Fetch(context.Context, string) (github.Files, error)
}
type ScanService struct {
	Findings *FindingService
	Store    ScanStore
	Graph    graph.Repository
	Ingest   Ingestor
	Resolver resolver.Resolver
}

func (s ScanService) CreateScan(ctx context.Context, project string) (domain.Scan, error) {
	return s.Store.CreateScan(ctx, project)
}
func (s ScanService) GetScan(ctx context.Context, id string) (domain.Scan, error) {
	return s.Store.GetScan(ctx, id)
}

// RunScan orchestrates metadata ingestion and resolution through persistence.
// Queue claiming, failure recording, and cancellation belong to jobs.
func (s ScanService) RunScan(ctx context.Context, scan domain.Scan, url string) error {
	if err := s.Store.SetStage(ctx, scan.ID, "FETCHING_REPOSITORY", 15); err != nil {
		return err
	}
	files, err := s.Ingest.Fetch(ctx, url)
	if err != nil {
		return err
	}
	if err = s.Store.SetStage(ctx, scan.ID, "RESOLVING_DEPENDENCIES", 55); err != nil {
		return err
	}
	resolved, err := s.Resolver.Resolve(resolver.Input{Pyproject: files.Pyproject, PoetryLock: files.PoetryLock})
	if err != nil {
		return err
	}
	if err = s.Store.SetStage(ctx, scan.ID, "BUILDING_GRAPH", 80); err != nil {
		return err
	}
	if err = s.Graph.Write(ctx, scan.ID, resolved); err != nil {
		return err
	}
	if s.Findings != nil {
		if err = s.Store.SetStage(ctx, scan.ID, "ENRICHING_VULNERABILITIES", 85); err != nil {
			return err
		}
		if err = s.Findings.Enrich(ctx, scan.ID, resolved.Packages); err != nil {
			return err
		}
		if err = s.Store.SetStage(ctx, scan.ID, "GENERATING_FINDINGS", 95); err != nil {
			return err
		}
	}
	return s.Store.Complete(ctx, scan.ID, files.CommitSHA, resolved.Incomplete, resolved.Limitations)
}
