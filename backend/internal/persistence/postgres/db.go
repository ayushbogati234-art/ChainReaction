package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schema = `CREATE TABLE IF NOT EXISTS projects (id UUID PRIMARY KEY, repository_url TEXT NOT NULL UNIQUE, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS scans (id UUID PRIMARY KEY, project_id UUID NOT NULL REFERENCES projects(id), commit_sha TEXT, status TEXT NOT NULL, stage TEXT NOT NULL, progress INTEGER NOT NULL DEFAULT 0, incomplete BOOLEAN NOT NULL DEFAULT false, limitations JSONB NOT NULL DEFAULT '[]', error_code TEXT, error_message TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), started_at TIMESTAMPTZ, completed_at TIMESTAMPTZ);
CREATE TABLE IF NOT EXISTS scan_jobs (id BIGSERIAL PRIMARY KEY, scan_id UUID NOT NULL UNIQUE REFERENCES scans(id), status TEXT NOT NULL DEFAULT 'QUEUED', attempts INTEGER NOT NULL DEFAULT 0, available_at TIMESTAMPTZ NOT NULL DEFAULT now(), locked_at TIMESTAMPTZ, completed_at TIMESTAMPTZ, last_error TEXT);`

type Store struct{ DB *pgxpool.Pool }

func (s Store) Migrate(ctx context.Context) error { _, e := s.DB.Exec(ctx, schema); return e }
func (s Store) CreateProject(ctx context.Context, url string) (domain.Project, error) {
	p := domain.Project{ID: uuid.NewString(), RepositoryURL: url}
	_, e := s.DB.Exec(ctx, "INSERT INTO projects (id,repository_url) VALUES ($1,$2) ON CONFLICT (repository_url) DO NOTHING", p.ID, p.RepositoryURL)
	if e != nil {
		return p, e
	}
	var existing domain.Project
	e = s.DB.QueryRow(ctx, "SELECT id,repository_url,created_at FROM projects WHERE repository_url=$1", url).Scan(&existing.ID, &existing.RepositoryURL, &existing.CreatedAt)
	return existing, e
}
func (s Store) CreateScan(ctx context.Context, project string) (domain.Scan, error) {
	x := domain.Scan{ID: uuid.NewString(), ProjectID: project, Status: "QUEUED", Stage: "QUEUED"}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return x, e
	}
	defer tx.Rollback(ctx)
	var exists bool
	if e = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE id=$1)", project).Scan(&exists); e != nil {
		return x, translate(e)
	}
	if !exists {
		return x, domain.ErrNotFound
	}
	_, e = tx.Exec(ctx, "INSERT INTO scans (id,project_id,status,stage) VALUES ($1,$2,$3,$4)", x.ID, x.ProjectID, x.Status, x.Stage)
	if e == nil {
		_, e = tx.Exec(ctx, "INSERT INTO scan_jobs (scan_id) VALUES ($1)", x.ID)
	}
	if e != nil {
		return x, e
	}
	e = tx.Commit(ctx)
	if e != nil {
		return x, e
	}
	return s.GetScan(ctx, x.ID)
}
func (s Store) Claim(ctx context.Context) (domain.Scan, string, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return domain.Scan{}, "", translate(e)
	}
	defer tx.Rollback(ctx)
	// A scan has a two-minute execution timeout. Recover jobs abandoned by a
	// crashed worker after three minutes as explicit terminal failures.
	_, e = tx.Exec(ctx, `WITH abandoned AS (
 UPDATE scan_jobs SET status='FAILED',completed_at=now(),last_error='Worker lease expired'
 WHERE status='RUNNING' AND locked_at < now()-interval '3 minutes' RETURNING scan_id)
 UPDATE scans SET status='FAILED',stage='FAILED',error_code='WORKER_INTERRUPTED',
 error_message='Worker lease expired',completed_at=now()
 WHERE id IN (SELECT scan_id FROM abandoned)`)
	if e != nil {
		return domain.Scan{}, "", translate(e)
	}
	var job int64
	var x domain.Scan
	var url string
	e = tx.QueryRow(ctx, `WITH next AS (SELECT id FROM scan_jobs WHERE status='QUEUED' AND available_at<=now() ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 1) UPDATE scan_jobs j SET status='RUNNING',attempts=attempts+1,locked_at=now() FROM next WHERE j.id=next.id RETURNING j.id,j.scan_id`).Scan(&job, &x.ID)
	if errors.Is(e, pgx.ErrNoRows) {
		if err := tx.Commit(ctx); err != nil {
			return domain.Scan{}, "", err
		}
		return domain.Scan{}, "", domain.ErrNotFound
	}
	if e != nil {
		return domain.Scan{}, "", translate(e)
	}
	e = tx.QueryRow(ctx, "SELECT s.project_id,p.repository_url FROM scans s JOIN projects p ON p.id=s.project_id WHERE s.id=$1", x.ID).Scan(&x.ProjectID, &url)
	if e != nil {
		return domain.Scan{}, "", translate(e)
	}
	if e = tx.Commit(ctx); e != nil {
		return domain.Scan{}, "", translate(e)
	}
	return x, url, nil
}
func (s Store) SetStage(ctx context.Context, id, stage string, progress int) error {
	_, e := s.DB.Exec(ctx, "UPDATE scans SET status=$2,stage=$2,progress=$3,started_at=COALESCE(started_at,now()) WHERE id=$1", id, stage, progress)
	return e
}
func (s Store) Complete(ctx context.Context, id, sha string, incomplete bool, limits []string) error {
	if limits == nil {
		limits = []string{}
	}
	raw, err := json.Marshal(limits)
	if err != nil {
		return err
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, "UPDATE scans SET status='COMPLETED',stage='COMPLETED',progress=100,commit_sha=$2,incomplete=$3,limitations=$4,completed_at=now() WHERE id=$1", id, sha, incomplete, raw)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "UPDATE scan_jobs SET status='COMPLETED',completed_at=now() WHERE scan_id=$1", id)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s Store) Fail(ctx context.Context, id, code, msg string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, "UPDATE scans SET status='FAILED',stage='FAILED',error_code=$2,error_message=$3,completed_at=now() WHERE id=$1", id, code, msg)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "UPDATE scan_jobs SET status='FAILED',last_error=$2,completed_at=now() WHERE scan_id=$1", id, msg)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s Store) GetScan(ctx context.Context, id string) (domain.Scan, error) {
	var x domain.Scan
	var raw []byte
	e := s.DB.QueryRow(ctx, "SELECT id,project_id,COALESCE(commit_sha,''),status,stage,progress,incomplete,limitations,error_code,error_message,created_at,started_at,completed_at FROM scans WHERE id=$1", id).Scan(&x.ID, &x.ProjectID, &x.CommitSHA, &x.Status, &x.Stage, &x.Progress, &x.Incomplete, &raw, &x.ErrorCode, &x.ErrorMessage, &x.CreatedAt, &x.StartedAt, &x.CompletedAt)
	if e == nil {
		_ = json.Unmarshal(raw, &x.Limitations)
	}
	return x, translate(e)
}
func (s Store) GetProject(ctx context.Context, id string) (domain.Project, error) {
	var p domain.Project
	e := s.DB.QueryRow(ctx, "SELECT id,repository_url,created_at FROM projects WHERE id=$1", id).Scan(&p.ID, &p.RepositoryURL, &p.CreatedAt)
	return p, translate(e)
}
func translate(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) && pgerr.Code == "22P02" {
		return domain.ErrNotFound
	}
	return err
}
