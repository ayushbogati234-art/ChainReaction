package service

import (
	"context"
	"fmt"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/analysis/blast"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/analysis/criticality"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/graph"
	"time"
)

type GraphService struct {
	Normalization *blast.Cache
	Criticality   *criticality.Cache
	Graph         graph.Repository
	Scans         interface {
		GetScan(context.Context, string) (domain.Scan, error)
	}
}

var defaultNormalization blast.Cache
var defaultCriticality criticality.Cache

func (s GraphService) ready(ctx context.Context, id string) (domain.Scan, error) {
	scan, err := s.Scans.GetScan(ctx, id)
	if err != nil {
		return scan, err
	}
	if scan.Status != "COMPLETED" {
		return scan, fmt.Errorf("SCAN_NOT_READY: scan is %s", scan.Status)
	}
	return scan, nil
}
func (s GraphService) GetGraph(ctx context.Context, id string) (map[string]any, error) {
	if _, err := s.ready(ctx, id); err != nil {
		return nil, err
	}
	return s.Graph.Graph(ctx, id)
}
func (s GraphService) GetPackage(ctx context.Context, scan, id string) (map[string]any, error) {
	if _, err := s.ready(ctx, scan); err != nil {
		return nil, err
	}
	return s.Graph.Package(ctx, scan, id)
}
func (s GraphService) GetImpact(ctx context.Context, scanID, packageID string, options domain.PathOptions) (domain.ImpactAnalysis, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	scan, err := s.ready(ctx, scanID)
	if err != nil {
		return domain.ImpactAnalysis{}, err
	}
	snapshot, err := s.Graph.Snapshot(ctx, scanID)
	if err != nil {
		return domain.ImpactAnalysis{}, err
	}
	result, err := blast.Analyze(snapshot, packageID, options)
	if err != nil {
		return domain.ImpactAnalysis{}, err
	}
	cache := s.Normalization
	if cache == nil {
		cache = &defaultNormalization
	}
	result.BlastRadius = cache.Score(ctx, scanID, snapshot, result, scan.Incomplete)
	criticalityCache := s.Criticality
	if criticalityCache == nil {
		criticalityCache = &defaultCriticality
	}
	result.StructuralCriticality = criticalityCache.Score(ctx, scanID, snapshot, result, scan.Incomplete)
	if scan.Incomplete {
		result.EvidenceComplete = false
		result.Paths.EvidenceComplete = false
		result.Limitations = append(result.Limitations, "Dependency evidence is incomplete; structural counts are exact only for the persisted scan snapshot.")
		result.Limitations = append(result.Limitations, scan.Limitations...)
		result.Paths.Limitations = append(result.Paths.Limitations, "Dependency evidence is incomplete; paths reflect only the persisted scan snapshot.")
	}
	return result, nil
}

func (s GraphService) GetPaths(ctx context.Context, scanID, packageID string, options domain.PathOptions) (domain.PathResult, error) {
	result, err := s.GetImpact(ctx, scanID, packageID, options)
	if err != nil {
		return domain.PathResult{}, err
	}
	return result.Paths, nil
}
