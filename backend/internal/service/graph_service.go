package service

import (
	"context"
	"fmt"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/graph"
)

type GraphService struct {
	Graph graph.Repository
	Scans interface {
		GetScan(context.Context, string) (domain.Scan, error)
	}
}

func (s GraphService) ready(ctx context.Context, id string) error {
	scan, err := s.Scans.GetScan(ctx, id)
	if err != nil {
		return err
	}
	if scan.Status != "COMPLETED" {
		return fmt.Errorf("SCAN_NOT_READY: scan is %s", scan.Status)
	}
	return nil
}
func (s GraphService) GetGraph(ctx context.Context, id string) (map[string]any, error) {
	if err := s.ready(ctx, id); err != nil {
		return nil, err
	}
	return s.Graph.Graph(ctx, id)
}
func (s GraphService) GetPackage(ctx context.Context, scan, id string) (map[string]any, error) {
	if err := s.ready(ctx, scan); err != nil {
		return nil, err
	}
	return s.Graph.Package(ctx, scan, id)
}
func (s GraphService) GetPaths(ctx context.Context, scan, id string, limit int) (domain.PathResult, error) {
	if err := s.ready(ctx, scan); err != nil {
		return domain.PathResult{}, err
	}
	if limit < 1 || limit > 100 {
		return domain.PathResult{}, fmt.Errorf("INVALID_REQUEST: path limit must be between 1 and 100")
	}
	if _, err := s.Graph.Package(ctx, scan, id); err != nil {
		return domain.PathResult{}, err
	}
	return s.Graph.Paths(ctx, scan, id, limit)
}
