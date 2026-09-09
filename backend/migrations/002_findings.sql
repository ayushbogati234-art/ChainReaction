CREATE TABLE IF NOT EXISTS scan_enrichments (
 scan_id UUID PRIMARY KEY REFERENCES scans(id), status TEXT NOT NULL,
 data JSONB NOT NULL, updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS package_snapshots (
 scan_id UUID NOT NULL REFERENCES scans(id), id TEXT NOT NULL, purl TEXT NOT NULL,
 data JSONB NOT NULL, PRIMARY KEY(scan_id,id)
);
CREATE TABLE IF NOT EXISTS vulnerability_snapshots (
 scan_id UUID NOT NULL REFERENCES scans(id), id UUID NOT NULL,
 advisory_id TEXT NOT NULL, source TEXT NOT NULL, raw_metadata JSONB NOT NULL,
 PRIMARY KEY(scan_id,id)
);
CREATE TABLE IF NOT EXISTS findings (
 scan_id UUID NOT NULL REFERENCES scans(id), id UUID NOT NULL, package_snapshot_id TEXT NOT NULL,
 vulnerability_snapshot_id UUID NOT NULL, type TEXT NOT NULL, data JSONB NOT NULL,
 PRIMARY KEY(scan_id,id),
 FOREIGN KEY(scan_id,package_snapshot_id) REFERENCES package_snapshots(scan_id,id),
 FOREIGN KEY(scan_id,vulnerability_snapshot_id) REFERENCES vulnerability_snapshots(scan_id,id)
);
