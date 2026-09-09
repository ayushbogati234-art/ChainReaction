# ChainReaction project handoff

Updated 2026-09-09 — Phase 2 implemented and verified for its documented supported
matching subset; awaiting user review. Full project completion is not claimed.

## Authority and current scope

The [architecture blueprint](../ChainReaction-Architecture_Blueprint.md) remains
the full product specification. Track coverage in [FEATURE_MATRIX.md](FEATURE_MATRIX.md),
Phase 1 evidence in [ARCHITECTURE.md](ARCHITECTURE.md), and current work in
[phases/PHASE_2.md](phases/PHASE_2.md). Do not edit the blueprint into a progress report.

The user explicitly authorized Phase 2: OSV vulnerability provider abstraction,
exact validated Python package/version queries, preserved advisory evidence and
alias deduplication, bounded requests, scan-scoped PostgreSQL snapshots/findings,
Neo4j vulnerability relationships, repeat-safe processing, partial/degraded
coverage, findings list/detail APIs and UI connected to package paths, separate
dependency and vulnerability confidence, and deterministic/database/live tests.
No scoring formulas, structural-criticality formulas, simulations, mitigation
ranking, RRP, extra dependency formats or tox.ini parsing are authorized. Do not
start Phase 3 automatically. Do not stage or commit without user permission.

## Completed and unfinished

Phase 1 graph/ingestion/provenance work is committed by the user and its evidence
is preserved in ARCHITECTURE.md. Phase 2 code includes provider/domain/confidence
modules, numeric-release OSV matching, alias groups, per-package coverage,
migrations and snapshot/findings repositories, vulnerability graph writes,
service/worker wiring, findings APIs and UI. Full blueprint matrix exists.
Deterministic provider, database, live OSV/browser and graph regression checks
passed; exact commands, scan IDs and gaps are in PHASE_2.md. API/setup docs and
the feature matrix are updated. Next concrete implementation step: address the
user's Phase 2 review findings; no further phase is authorized. Full PEP 440/custom
sources, caching, full observability, all scoring/simulations/ranking and the full
demo remain partial/deferred as recorded in FEATURE_MATRIX.md.

## Architecture decisions and corrections

Python-only analysis; Go/Next.js are implementation technologies. One root
application with paired Poetry pyproject.toml + poetry.lock. Manifest supplies
identity/direct declarations, lock supplies versions/package dependency evidence.
Metadata-only scans never install or execute scanned code. Preserve all final
blueprint capabilities even when deferred. Use the existing repository and
prescribed module/file locations; consolidated Phase 1 files are accepted when
boundaries remain correct. Domain is client-independent; services orchestrate;
jobs own queue workflow; persistence and graph queries stay behind interfaces.
Keep Cytoscape, dependent -> dependency arrows, stable inspection and dark UI.
Phase 2 inspector correction: package selection owns the inspector; selecting a
different package clears a finding from another package, while selecting a finding
selects/preserves its own package. Finding requests are aborted and identity-checked
against scan/finding/package so late details cannot restore stale content.

## Git and files

Branch: main. Starting HEAD was 0b16c95. Current observed HEAD is the user's
03fa2de (phases doumentations files add for context); preserve this new commit.
Working tree was clean before this handoff. Uncommitted work includes the three
progress documents; backend domain/vulnerability/confidence/service/handler/
Postgres/Neo4j/migration additions, app/scan/router wiring; provider fixtures/tests
and database regression extensions; frontend finding types/components/features
and scan integration, README/API/architecture links and docs/openapi.json.
The exact changed file inventory is in PHASE_2.md under Files and architectural
boundaries. Git status is dirty (modified tracked files and new untracked Phase 2
files), with an empty staged diff. Nothing staged/committed by this session.
The blueprint is unchanged. Recheck `git status --short` before further work.

## Verification and gaps

Previous Phase 1 evidence: backend tests, real database
workflow with simulated GitHub archives, Docker build, Chromium graph/provenance
checks and live public GitHub ingestion passed; exact commands/scan IDs in
ARCHITECTURE.md. Last live scan b1a39eb6-6948-4ef5-8a35-04476ef1a4a6 had 48 nodes,
75 edges, 123 limitations. Phase 2 `go test ./...`, `go test -race ./...`, `go vet
./...`, frontend `npx tsc --noEmit --incremental false`, Docker build and real
database integration passed. Provider fixture path and timeout-test counter race
were fixed; reruns passed. Final focused service test passed for simultaneous
provider/graph degradation. Live OSV requests@2.19.0 returned 10 matched records.
Live browser scan 579acc06-c7c0-45c1-8bba-a5eb0002d2db of the same public Poetry
example completed: 47/47 packages checked, 52 findings. Fixture browser checked
missing severity/fixes and partial/degraded states; graph regression retained
48 nodes/75 edges and stable interactions at narrow widths. OpenAPI JSON parses
and 55 schema references resolve (not full external conformance verification).
All exact commands and evidence are in PHASE_2.md. Full product demo
including simulations/ranking remains unimplemented.

## Limitations, decisions, runtime

No current blocker for the tested local workflow. No Python environment selection
or full dependency version-range/freshness validation;
older scans lack newly recorded provenance. Deferred formulas, exploitability
source, remediation model and deployment target require later decisions.
Local UI http://localhost:3000, API :8080, PostgreSQL :5432, Neo4j :7474/:7687
are running; PostgreSQL/Neo4j health checks passed. Verify with `docker compose ps`. Start/build with
`docker compose up -d --build` from repository root. Use local `.env`; never put
credentials here. Host database authentication differed from Compose previously;
run integration tests on Compose network with `.env`. Temporary browser tooling
is under /tmp/chainreaction-graph-browser-0TMQ3w and may need restoration after a
restart. See PHASE_2.md for browser environment variables and database test command.
Phase 2 uses public OSV without an API key. Matching is limited to numeric exact
releases/default PyPI; skipped/failed/unqueried packages stay explicit. Provider
pagination/outages and graph degradation are fixture-tested, not live outages.
Large-scale behavior, distributed reconciliation and external OpenAPI conformance
remain unverified. No automatic phase advance or publication is authorized.

The Phase 2 inspector consistency regression passed after rebuilding frontend:
deterministic `frontend/tests/finding-selection.cjs` uses delayed black finding
detail data and rapid black/attrs changes; persisted-scan
`frontend/tests/finding-selection-live.cjs` passed the exact
black@23.3.0/CVE-2024-21503 -> attrs@22.2.0 -> black -> paths sequence. Both show
no stale detail/browser errors. TypeScript, diff check and frontend Docker build
also passed. Current next step remains user review/follow-up fixes within Phase 2.

Update this file at meaningful milestones and before context resets/handoffs,
including exact verification outcomes, current Git state and next concrete step.
