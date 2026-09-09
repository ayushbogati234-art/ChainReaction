package service

import (
	"context"
	"fmt"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/vulnerability"
	"time"
)

type FindingStore interface {
	SaveEnrichment(context.Context, string, []domain.Package, domain.Enrichment) error
	GetEnrichment(context.Context, string) (domain.Enrichment, error)
	GetFinding(context.Context, string, string) (domain.Finding, error)
}
type VulnerabilityGraph interface {
	WriteVulnerabilities(context.Context, string, []domain.Finding) error
}
type FindingService struct {
	Store FindingStore
	Scans interface {
		GetScan(context.Context, string) (domain.Scan, error)
	}
	Provider vulnerability.Provider
	Graph    VulnerabilityGraph
}

func (s FindingService) Enrich(ctx context.Context, scan string, packages []domain.Package) error {
	queryCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	checks := []domain.PackageCheck{}
	if s.Provider != nil {
		checks = s.Provider.Query(queryCtx, packages)
	}
	// A missing result is unqueried, never a successful empty check.
	byID := map[string]domain.PackageCheck{}
	for _, c := range checks {
		byID[c.PackageID] = c
	}
	checks = []domain.PackageCheck{}
	checked := 0
	covered := 0
	for _, p := range packages {
		c, ok := byID[p.ID]
		if !ok {
			c = domain.PackageCheck{PackageID: p.ID, PURL: p.PURL, Status: "UNQUERIED", Reason: "provider returned no package result", Advisories: []domain.Advisory{}}
		}
		checks = append(checks, c)
		if c.Status == "CHECKED" {
			checked++
		}
		if c.Status == "CHECKED" || len(c.Advisories) > 0 {
			covered++
		}
	}
	state := "COMPLETE"
	if checked != len(packages) {
		state = "PARTIAL"
	}
	if covered == 0 && len(packages) > 0 {
		state = "DEGRADED"
	}
	result := domain.Enrichment{Status: state, Provider: "OSV", GraphStatus: "COMPLETE", Checks: checks, Findings: vulnerability.Findings(scan, packages, checks)}
	if s.Graph == nil {
		result.GraphStatus = "NOT_RUN"
	} else if err := s.Graph.WriteVulnerabilities(ctx, scan, result.Findings); err != nil {
		result.GraphStatus = "DEGRADED"
		if result.Status == "COMPLETE" {
			result.Status = "PARTIAL"
		}
		result.Error = "Vulnerability graph persistence failed; PostgreSQL findings remain available."
	}
	return s.Store.SaveEnrichment(ctx, scan, packages, result)
}
func (s FindingService) ready(ctx context.Context, scan string) error {
	x, err := s.Scans.GetScan(ctx, scan)
	if err != nil {
		return err
	}
	if x.Status != "COMPLETED" {
		return fmt.Errorf("SCAN_NOT_READY: scan is %s", x.Status)
	}
	return nil
}
func (s FindingService) List(ctx context.Context, scan string) (domain.Enrichment, error) {
	if err := s.ready(ctx, scan); err != nil {
		return domain.Enrichment{}, err
	}
	return s.Store.GetEnrichment(ctx, scan)
}
func (s FindingService) Detail(ctx context.Context, scan, id string) (domain.Finding, error) {
	if err := s.ready(ctx, scan); err != nil {
		return domain.Finding{}, err
	}
	return s.Store.GetFinding(ctx, scan, id)
}
