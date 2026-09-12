# Phase 2 — vulnerability enrichment and evidence-backed findings

Status: closed out and verified for the supported matching subset below. The
inspector consistency follow-up and regression are included in user commit
`a3c6798`. The full project is not complete. Full scope authority is the
[blueprint](../../ChainReaction-Architecture_Blueprint.md); coverage and remaining
product work are in [FEATURE_MATRIX.md](../FEATURE_MATRIX.md). Preserve the
[Phase 1 closeout](../ARCHITECTURE.md). See [HANDOFF.md](../HANDOFF.md) for current state.

Approved: Python exact-version OSV evidence, provider abstraction and bounded
failure handling; scan-scoped snapshots/findings and vulnerability graph links;
alias deduplication and repeat-safe processing; findings APIs/UI/path integration;
separate confidence and coverage; reproducible provider/database/browser tests.
Excluded: new input formats, tox parsing, scoring, simulations, ranking and RRP.

Architecture mapping: blueprint §§9–13, 21, 24–31, 34–35, 38. New implementations
belong in domain/{vulnerability,finding}.go, vulnerability/provider.go,
vulnerability/osv/{client,mapper}.go, vulnerability/match.go,
analysis/confidence/calculator.go, service/finding_service.go,
persistence/postgres/finding_repo.go, transport/http/handler/finding_handler.go,
graph/neo4j, backend/migrations and tests; frontend features/findings and
components/findings with shared types. Startup remains in internal/app.

## Implemented behavior and data

The worker writes dependency topology first, then runs bounded OSV enrichment,
generates evidence-backed findings and completes the scan. Provider failures do
not discard successful dependency analysis. Scan completion, dependency
`incomplete` and `enrichment_status` are distinct. PostgreSQL persistence failure
still fails the scan; this is not disguised as a provider outage.

OSV queries use validated PyPI name + exact numeric version, not simultaneous
versioned purl and version fields. Input PURLs must agree with the normalized
name/version. Only default PyPI lock sources are queried; custom-index/source
entries, unresolved versions, prerelease/post/dev/local/epoch versions are
SKIPPED with an explanation. Supporting these other exact version forms remains
partial work; it is not an expansion of repository input formats.

Matching verifies the package ecosystem/name (and purl when supplied), then
requires an exact enumerated affected version or a verifiable numeric release
interval. Fixed/limit boundaries are exclusive; last_affected is inclusive.
Only fixed boundaries of intervals applicable to the queried version are exposed
as fixed-version evidence. Unknown range syntax does not become a fabricated
match. Withdrawn advisories are excluded. Advisory records without verifiable
matching evidence make package coverage partial, preserving valid matches from
the same response. Missing severity or fixed versions do not become guessed data.

Provider/source IDs, alias groups, references, summaries/descriptions, raw OSV
records, affected entries, supplied CVSS vectors and severity labels are retained.
Alias connected components deduplicate per package; stable scan/package/advisory
IDs and transactional snapshot replacement make processing safe to repeat.
Distinct raw records sharing an advisory ID are preserved; identical duplicates
collapse. No cross-package/version identity merging occurs.

Requests default to 8 seconds each (hard upper bound 15), 4 MiB response bodies,
200 ms spacing, five pages per package and a 60-second enrichment budget within
the existing two-minute job. Pagination tokens are followed and repeated/capped
tokens report partial coverage. HTTP 429/503 records Retry-After and stops further
queries for that scan; there is no automatic retry, so it never retries earlier
than the provider asks. A new scan retries. Other failed requests and budget
exhaustion retain explicit coverage results. Snapshots are not a cross-scan cache.

Coverage states:

|Level|States and meaning|
|---|---|
|Package|CHECKED: all returned pages/records handled; PARTIAL: evidence/page gaps; FAILED: attempted provider request failed; SKIPPED: unsupported identity/source; UNQUERIED: no request due to budget/backoff/missing result|
|Scan enrichment|COMPLETE: all retained packages checked; PARTIAL: some coverage missing or vulnerability graph degraded; DEGRADED: no package coverage succeeded; NOT_RUN: legacy scan/no enrichment snapshot|
|Vulnerability graph|COMPLETE, DEGRADED or NOT_RUN, reported separately|

`package_snapshots`, `vulnerability_snapshots`, `findings` and `scan_enrichments`
are added by migrations/002_findings.sql. Migrations are embedded from the
prescribed directory and run under a transaction/advisory lock. Composite
scan-keyed foreign keys prevent cross-scan findings. All one-scan PostgreSQL
snapshots replace atomically. Topology and advisory writes are separate Neo4j
transactions; there is no distributed transaction with PostgreSQL. A vulnerability
graph failure is reported as degraded graph status with PostgreSQL findings still
available. A failed final PostgreSQL write can leave graph records for a failed
scan, which remain inaccessible through completed-scan APIs.

Neo4j stores scan-scoped `Vulnerability` nodes and `AFFECTED_BY`, including aliases,
supplied severity/CVSS and finding IDs. The dependency graph endpoint explicitly
returns application/package nodes and DEPENDS_ON only; all original dependency
nodes and edges remain. A full Vulnerability View mode is still deferred.

## API, UI and confidence

GET `/api/v1/scans/{id}/findings` returns the persisted enrichment report with
package coverage and findings. GET `/api/v1/scans/{id}/findings/{finding_id}`
returns the scoped detail. Reads before completion return 409; missing/cross-scan
finding IDs return 404. Legacy scans return NOT_RUN rather than a clean report.
See [API](../API.md) and [OpenAPI](../openapi.json).

The existing inspector now lists findings, coverage and source evidence. Selecting
a finding loads its detail and package; Show dependency paths uses the existing
path API and highlighting. It shows supplied severity/CVSS and applicable fixes,
explicit unavailable labels, direct declarations and dependency limitations.
References render as text or safe HTTP(S) links; provider text is not injected HTML.
The language is potential dependency exposure, not confirmed exploitation.
Empty partial/failed/unqueried coverage never becomes “no vulnerabilities”.
No working-looking risk/simulation controls or fabricated scores were added.

Dependency confidence remains the Phase 1 metadata confidence. Vulnerability
confidence is HIGH for an exact enumerated version or validated numeric affected
range with matching PyPI identity; MEDIUM if additional unsupported ranges leave
uncertainty alongside established matching evidence. Unverifiable matches are
not findings. Combined confidence is the conservative minimum, implemented in
analysis/confidence. None of these levels means safety, installation, environment
activation or exploitability. Wider Python confidence rules remain pending.

## Files and architectural boundaries

New and changed implementation locations (all under the existing repository):

```text
backend/
  internal/
    domain/{vulnerability,finding,scan}.go
    vulnerability/provider.go
    vulnerability/{match,match_test}.go
    vulnerability/osv/{client,mapper,client_test}.go
    analysis/confidence/calculator.go
    service/{finding_service,finding_service_test,scan_service}.go
    persistence/postgres/{db,finding_repo}.go
    graph/neo4j/{repository,writer}.go
    transport/http/router.go
    transport/http/handler/finding_handler.go
    app/{api,worker}.go
  migrations/{embed.go,002_findings.sql}
  tests/integration_test.go
frontend/
  src/components/findings/FindingDetail.tsx
  src/features/findings/{FindingsPanel.tsx,FindingsPanel.module.css}
  src/features/scan/ScanWorkspace.tsx
  src/types/{finding,scan}.ts
  tests/phase2-workflow.cjs
fixtures/python/vulnerabilities/affected.json
docs/{HANDOFF.md,FEATURE_MATRIX.md,API.md,openapi.json,ARCHITECTURE.md}
docs/phases/PHASE_2.md
README.md
```

No new dependencies, nested project, blueprint edits, deferred stubs or substantial
reorganization. Existing Phase 1 consolidation remains accepted. Domain is
client-independent; provider mapping/identity algorithms stay in vulnerability,
confidence calculation in analysis/confidence, orchestration in service, SQL and
Cypher behind interfaces, routes delegate, and app performs wiring only.
The new embedded-migration helper and phase/coverage documentation are routine
additions within the requested directories.

## Validation actually performed (2026-09-09)

From backend:

```sh
go test ./...
go test -race ./...
go vet ./...
PHASE2_LIVE_OSV=1 go test ./internal/vulnerability/osv -run TestLiveProvider -v -count=1
```

All passed. Initial fixture path failure was fixed. The first race run detected a
shared counter in the timeout test server; it was made atomic and the rerun passed.
A final focused `go test ./internal/service` passed for simultaneous provider/graph
degradation, preserving DEGRADED rather than obscuring entirely unqueried coverage.
Deterministic cases cover affected/unaffected releases, wrong ecosystem/package,
fixed boundary, withdrawn advisories, connected alias groups, distinct source
snapshots, missing severity/fix, unsupported/unresolved inputs, timeout/429/503,
invalid JSON/error responses, missing package results, capped/repeated pages,
partial-page evidence retention, graph failures and confidence minimum.

From repository root, real databases with deterministic GitHub and OSV responses:

```sh
docker run --rm --network chainreaction_default --env-file .env \
  -e PHASE1_INTEGRATION=1 \
  --mount type=bind,source=/home/anish/ChainReaction,target=/src,readonly \
  -w /src/backend golang:1.26-alpine \
  go test ./tests -run TestDatabaseWorkflow -v -count=1
```

Passed, including `Phase2EvidencePersistence`: alias/source snapshot persistence,
findings list/detail endpoints, invalid/cross-scan IDs, repeat writes without
duplicate rows/relationships, three-node/two-edge fixture topology preserved,
paths, provider 429 -> completed dependency scan + DEGRADED enrichment with
FAILED/UNQUERIED package states. Existing P1 queue, isolation, unsupported inputs,
provenance-change fixture, conditions and worker recovery also passed. Unique
test schemas/scan IDs are cleaned up; existing user scan records remain.

Frontend `npx tsc --noEmit --incremental false` passed. Docker
`docker compose up -d --build` passed, including Next.js production compilation
and type checks. Final backend containers were refreshed after the last
degraded-state fix. OpenAPI JSON parses and all 55 schema references resolve;
this is not a complete external OpenAPI conformance audit. `git diff --check`
passed; the architectural blueprint has no diff.

Browser tooling was external to the repository, under the prior temporary
Playwright installation. Commands from repository root (installation paths are
environment-specific, not runtime dependencies):

```sh
LD_LIBRARY_PATH=/tmp/chainreaction-graph-browser-0TMQ3w/libs/usr/lib/x86_64-linux-gnu \
PLAYWRIGHT_MODULE=/tmp/chainreaction-graph-browser-0TMQ3w/node_modules/playwright \
PLAYWRIGHT_BROWSERS_PATH=/tmp/chainreaction-graph-browser-0TMQ3w/browsers \
node frontend/tests/phase2-workflow.cjs
```

Passed deterministic API-fixture UI: missing severity/fix, retained aliases,
finding details, package/paths, stable graph viewport/positions, 390px inspector,
partial/degraded warnings, and no false clean results/browser errors.

The same command with `TEST_LIVE=1` and
`TEST_SCREENSHOT=/tmp/chainreaction-phase2-live.png` passed with **no mocked
requests**: real GitHub URL submission, persisted queue/worker, OSV, both databases,
finding detail, package paths and narrow browser layout. Screenshot reviewed.

* Repository: https://github.com/Hasenpfote/python-poetry-example
* Scan: `579acc06-c7c0-45c1-8bba-a5eb0002d2db`
* Commit: `f0e6e6b09c0090728fd1edbb09ee3f39fd1e3b9b`
* Enrichment COMPLETE: 47/47 packages CHECKED, 52 alias-deduplicated findings.
* Independent live OSV smoke: requests@2.19.0 CHECKED, 10 matched advisory records.
* Dependency completeness remains independent; the Poetry graph retains its
  environment/constraint limitations. No exploitation or upgrade was tested.

Using the same browser environment and `TEST_SCAN_ID` above,
`node frontend/tests/graph-readability.cjs` passed: all 48 nodes/75 edges retained,
zero node/label overlaps at normal zoom, zoom/wheel/pan/fit/reset, neighborhood and
full directed path highlighting, stable selection and 1000/760/390px inspector.
This regression replays only creation responses to open the saved scan; status,
graph, package, path and finding reads are real.

## Remaining limitations and next work

### Inspector selection consistency follow-up — 2026-09-09

Reproduced the Phase 2 inspector issue on the persisted Poetry example scan:
selecting `black@23.3.0` / `CVE-2024-21503`, then selecting `attrs@22.2.0`,
previously left the black finding detail beneath the attrs package evidence.
Finding selection now belongs to the selected package. A package change clears a
finding only when it belongs to another package; selecting a finding still selects
and preserves its own package. The detail request is tied to scan, finding and
package identity, is aborted when any change makes it stale, and validates all
three identities before rendering. A late response therefore cannot restore a
detail for the prior package.

Added `frontend/tests/finding-selection.cjs`, a deterministic Chromium regression
with a delayed black detail response. It verifies black finding -> attrs clears the
detail, rapid black -> attrs -> black changes settle on the matching detail, a
finding selects its package, and Show dependency paths still works. Passed against
the rebuilt frontend. Added `frontend/tests/finding-selection-live.cjs`, which
replays only POST creation responses to open scan
`579acc06-c7c0-45c1-8bba-a5eb0002d2db`; its graph/findings/package/path reads are
real. It passed the requested black CVE -> attrs -> black -> paths sequence with
no browser errors. `npx tsc --noEmit --incremental false`, `git diff --check`, and
the Docker frontend production build passed. No Phase 3 work was started.

No blockers remain for the tested local Phase 2 workflow. Full PEP 440 matching,
custom-index identity policy, large-scan/report pagination/load tests, caching,
complete observability/request IDs and full OpenAPI conformance remain partial or
unimplemented. Provider pagination is fixture-tested, not triggered in the live
smoke; provider/graph outage behavior was injected, not forced against live OSV.
There is no distributed transaction or background enrichment retry/reconciliation.
Live advisory results can change; snapshots preserve what the scan observed.

The complete remaining product is retained in FEATURE_MATRIX: Blast Radius,
structural criticality, explainable overall Risk Score, full Propagation Path
Explorer aggregates/filters, compromise/remediation simulations, Mitigation
Priority, RRP, full Finding Details, search/filters/graph modes, Systemic Dependency,
broader confidence/input handling, caches, observability and the Butterfly Effect
end-to-end demo. Formulas, exploitability source, remediation strategy and
deployment remain architect/user decisions. tox.ini and new input formats are
unchanged and unapproved. No Phase 3 work was started.

Phase 2 was closed out before authorized Phase 3 work began. Its supported
matching scope and limitations remain unchanged. Phase 3 progress is recorded in
[PHASE_3.md](PHASE_3.md); this historical evidence is not rewritten as Phase 3.

Provider references consulted: [OSV query API](https://github.com/google/osv.dev/blob/master/docs/api/post-v1-query.md)
and [OSV schema](https://ossf.github.io/osv-schema/). OSV query responses can be
paginated; the adapter independently verifies supported affected-version evidence
instead of relying solely on the version-query response.
