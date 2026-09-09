package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

// Replace one scan's snapshot atomically. Stable IDs and scan-keyed foreign keys
// prevent repeat-processing duplicates and cross-scan finding references.
func (s Store) SaveEnrichment(ctx context.Context, scan string, packages []domain.Package, e domain.Enrichment) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var id string
	if err = tx.QueryRow(ctx, "SELECT id FROM scans WHERE id=$1 FOR UPDATE", scan).Scan(&id); err != nil {
		return translate(err)
	}
	for _, table := range []string{"findings", "vulnerability_snapshots", "package_snapshots"} {
		if _, err = tx.Exec(ctx, "DELETE FROM "+table+" WHERE scan_id=$1", scan); err != nil {
			return err
		}
	}
	for _, p := range packages {
		data, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO package_snapshots(scan_id,id,purl,data) VALUES($1,$2,$3,$4)", scan, p.ID, p.PURL, data); err != nil {
			return err
		}
	}
	for _, f := range e.Findings {
		if f.ScanID != scan {
			return errors.New("finding scan mismatch")
		}
		raw, err := json.Marshal(f.Advisories)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO vulnerability_snapshots(scan_id,id,advisory_id,source,raw_metadata) VALUES($1,$2,$3,'OSV',$4)", scan, f.ID, f.AdvisoryID, raw); err != nil {
			return err
		}
		data, err := json.Marshal(f)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO findings(scan_id,id,package_snapshot_id,vulnerability_snapshot_id,type,data) VALUES($1,$2,$3,$2,$4,$5)", scan, f.ID, f.PackageID, f.Type, data); err != nil {
			return err
		}
	}
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO scan_enrichments(scan_id,status,data) VALUES($1,$2,$3) ON CONFLICT(scan_id) DO UPDATE SET status=excluded.status,data=excluded.data,updated_at=now()", scan, e.Status, data); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s Store) GetEnrichment(ctx context.Context, scan string) (domain.Enrichment, error) {
	e := domain.Enrichment{Status: "NOT_RUN", Provider: "OSV", GraphStatus: "NOT_RUN", Checks: []domain.PackageCheck{}, Findings: []domain.Finding{}}
	var raw []byte
	err := s.DB.QueryRow(ctx, "SELECT data FROM scan_enrichments WHERE scan_id=$1", scan).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return e, nil
	}
	if err != nil {
		return e, translate(err)
	}
	err = json.Unmarshal(raw, &e)
	return e, err
}
func (s Store) GetFinding(ctx context.Context, scan, id string) (domain.Finding, error) {
	var f domain.Finding
	var raw []byte
	err := s.DB.QueryRow(ctx, "SELECT data FROM findings WHERE scan_id=$1 AND id=$2", scan, id).Scan(&raw)
	if err != nil {
		return f, translate(err)
	}
	err = json.Unmarshal(raw, &f)
	return f, err
}
