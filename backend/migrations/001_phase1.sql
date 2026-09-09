CREATE TABLE IF NOT EXISTS projects (
  id UUID PRIMARY KEY, repository_url TEXT NOT NULL UNIQUE, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS scans (
  id UUID PRIMARY KEY, project_id UUID NOT NULL REFERENCES projects(id), commit_sha TEXT,
  status TEXT NOT NULL, stage TEXT NOT NULL, progress INTEGER NOT NULL DEFAULT 0,
  incomplete BOOLEAN NOT NULL DEFAULT false, limitations JSONB NOT NULL DEFAULT '[]',
  error_code TEXT, error_message TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  started_at TIMESTAMPTZ, completed_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS scan_jobs (
  id BIGSERIAL PRIMARY KEY, scan_id UUID NOT NULL UNIQUE REFERENCES scans(id), status TEXT NOT NULL DEFAULT 'QUEUED',
  attempts INTEGER NOT NULL DEFAULT 0, available_at TIMESTAMPTZ NOT NULL DEFAULT now(), locked_at TIMESTAMPTZ, completed_at TIMESTAMPTZ, last_error TEXT
);
