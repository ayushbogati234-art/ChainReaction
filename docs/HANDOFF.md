# ChainReaction project handoff

Updated 2026-09-09 — Phase 2 authorized and starting.

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
is preserved in ARCHITECTURE.md. Phase 2 implementation and verification are
unfinished. Initial next step: document full blueprint coverage, then implement
provider/domain evidence types and deterministic matching fixtures.

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

## Git and files

Starting branch: main. HEAD: 0b16c95 (repository link injestion completed).
Working tree was clean before this handoff. This milestone adds HANDOFF.md;
FEATURE_MATRIX.md and phases/PHASE_2.md are next. Preserve all user commits/work.

## Verification and gaps

Previous Phase 1 results (not rerun yet this phase): backend tests, real database
workflow with simulated GitHub archives, Docker build, Chromium graph/provenance
checks and live public GitHub ingestion passed; exact commands/scan IDs in
ARCHITECTURE.md. Last live scan b1a39eb6-6948-4ef5-8a35-04476ef1a4a6 had 48 nodes,
75 edges, 123 limitations. Phase 2 has no test results yet. Full product demo
including simulations/ranking remains unimplemented.

## Limitations, decisions, runtime

No Python environment selection or full version-range/freshness validation;
older scans lack newly recorded provenance. Deferred formulas, exploitability
source, remediation model and deployment target require later decisions.
Local UI http://localhost:3000, API :8080, PostgreSQL :5432, Neo4j :7474/:7687
were running previously; verify with `docker compose ps`. Start/build with
`docker compose up -d --build` from repository root. Use local `.env`; never put
credentials here. Host database authentication differed from Compose previously;
run integration tests on Compose network with `.env`. Temporary browser tooling
was under /tmp/chainreaction-graph-browser-0TMQ3w and may need restoration.

Update this file at meaningful milestones and before context resets/handoffs,
including exact verification outcomes, current Git state and next concrete step.
