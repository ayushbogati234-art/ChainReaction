package service

import (
	"context"
	"errors"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"testing"
)

type evidenceStore struct{ result domain.Enrichment }

func (s *evidenceStore) SaveEnrichment(_ context.Context, _ string, _ []domain.Package, e domain.Enrichment) error {
	s.result = e
	return nil
}
func (s *evidenceStore) GetEnrichment(context.Context, string) (domain.Enrichment, error) {
	return s.result, nil
}
func (s *evidenceStore) GetFinding(context.Context, string, string) (domain.Finding, error) {
	return domain.Finding{}, domain.ErrNotFound
}

type partialProvider struct{}

func (partialProvider) Query(context.Context, []domain.Package) []domain.PackageCheck {
	return []domain.PackageCheck{{PackageID: "a", Status: "CHECKED", Advisories: []domain.Advisory{}}}
}

type unavailableGraph struct{}

func (unavailableGraph) WriteVulnerabilities(context.Context, string, []domain.Finding) error {
	return errors.New("injected graph outage")
}
func TestMissingProviderResultsAndGraphFailure(t *testing.T) {
	store := &evidenceStore{}
	s := FindingService{Store: store, Provider: partialProvider{}, Graph: unavailableGraph{}}
	if err := s.Enrich(context.Background(), "scan", []domain.Package{{ID: "a"}, {ID: "b"}}); err != nil {
		t.Fatal(err)
	}
	e := store.result
	if e.Status != "PARTIAL" || e.GraphStatus != "DEGRADED" || e.Checks[0].Status != "CHECKED" || e.Checks[1].Status != "UNQUERIED" || e.Error == "" {
		t.Fatalf("lost degraded coverage: %+v", e)
	}
	s.Provider = nil
	if err := s.Enrich(context.Background(), "scan", []domain.Package{{ID: "a"}}); err != nil {
		t.Fatal(err)
	}
	if store.result.Status != "DEGRADED" {
		t.Fatal("graph failure obscured entirely unqueried coverage")
	}
}
