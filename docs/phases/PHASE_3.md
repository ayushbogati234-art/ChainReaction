# Phase 3 — graph impact analysis

Status: Phase 3 is complete for its approved prototype scope, including graph
metrics, bounded paths, numeric Blast Radius, and API/UI integration. The
numeric Blast Radius formula below was explicitly approved by the user and is
implemented in analysis/blast/normalize.go. The full specification remains the
[architecture blueprint](../../ChainReaction-Architecture_Blueprint.md); see the
[feature matrix](../FEATURE_MATRIX.md), [API](../API.md), and
[handoff](../HANDOFF.md). Phase 2 is closed out in [PHASE_2.md](PHASE_2.md).

## Approved scope and architecture mapping

This phase implements blueprint §§3.7–3.14, 4.2–4.5, 10.3–10.5, 10.9–10.10,
14–15, 17, 21, 24, 26–28 and 35 only to the extent described here, including
the approved score in §§3.13–3.14/14. It does not implement overall Risk Score, structural
criticality, simulations, mitigation/RRP, Systemic Dependency detection, new
input formats or multi-application repository discovery.

Neo4j exposes a typed, scan-filtered topology snapshot through the graph
repository. `internal/analysis/blast` owns reverse reachability, unique-set and
simple-path algorithms. `internal/service` verifies completed-scan state and adds
dependency-completeness evidence. HTTP parses filters and delegates. The frontend
only renders returned evidence and graph overlays; it does not calculate impact.
Domain types have no database, HTTP or driver dependencies. Startup remains in
`internal/app`. The blueprint is unchanged, no nested project or deferred stubs
were created, and the Phase 1/2 consolidated files remain accepted.

## Implemented behavior

For a selected package-version node, the analysis returns unique immediate and
transitive dependent package IDs, their unique combined count, unique exposed
application roots, and the affected node region. The selected node is excluded
from the downstream count. Application roots are counted separately. Package
versions remain distinct identities. Reachability is exact for the persisted,
scan-isolated topology snapshot. `evidence_complete=false` and limitations make
clear when incomplete dependency discovery means the snapshot itself may omit
relationships.

Path enumeration accepts `application_id`, `max_paths` (1–100, default 20) and
`max_depth` (1–100, default 100). The legacy `limit` query alias remains accepted.
It counts simple paths only, guards cycles, has graph-size limits (10,000 nodes /
100,000 edges), separates displayed paths from aggregate count, and stops an
aggregate count at 1,000. `count_exact`, `count_capped`, `depth_limited` and
`truncated` distinguish the cases. A capped `total_paths` is a lower bound shown
as `1000+`; it is never described as exact. Every displayed route has an explicit
edge length. `maximum_depth_exact=false` labels depth as the maximum observed
within bounded results. Invalid bounds/application IDs return a clear 400.

The package inspector shows exposed applications, downstream immediate/transitive
packages, count semantics, confidence, completeness, paths and depth. Findings
show the same package impact summary without conflating dependency confidence,
vulnerability-match confidence or exploitability. Users can filter routes by
application and highlight a full directed path. Package inspection highlights
the reverse-reachable affected region and immediate dependents. Unrelated graph
elements dim, while Cytoscape positions and viewport remain stable. Existing
dependent → dependency arrows, layout controls, package/finding selection,
coverage and narrow inspector behavior remain.

## Files changed

```text
backend/internal/domain/{model,exposure}.go
backend/internal/graph/{repository,model}.go
backend/internal/graph/neo4j/repository.go
backend/internal/analysis/blast/{calculator,calculator_test,normalize,normalize_test}.go
backend/internal/service/graph_service.go
backend/internal/transport/http/router.go
backend/tests/integration_test.go
frontend/src/types/scan.ts
frontend/src/components/graph/{DependencyGraph,ImpactEvidence}.tsx
frontend/src/components/graph/GraphWorkspace.module.css
frontend/src/components/findings/FindingDetail.tsx
frontend/src/features/{scan/ScanWorkspace,findings/FindingsPanel}.tsx
frontend/src/lib/graph/{focus,style}.ts
frontend/tests/{phase3-impact,finding-selection,phase2-workflow,graph-readability}.cjs
README.md
docs/{API.md,ARCHITECTURE.md,openapi.json,FEATURE_MATRIX.md,HANDOFF.md}
docs/phases/{PHASE_2,PHASE_3}.md
```

## Validation performed

### Independent worktree audit — 2026-09-12

Phase 3 remains uncommitted relative to `a3c6798` (`phase 2 completed`). The
current staged change set was compared with the blueprint, Phase 2 closeout,
API/OpenAPI documentation, feature matrix and handoff. The implementation and
this closeout agree on the delivered boundary: graph impact and bounded path
analysis, the approved scan-relative Blast Radius model, API/UI evidence, and
regressions for existing finding/graph interactions. No Phase 4 criticality,
overall Risk Score or systemic-dependency implementation is present.

Fresh checks from the current worktree:

- `go test ./...` passed with a writable temporary Go build cache and localhost
  test-server access. The initial sandboxed attempts were environmental failures
  (read-only default cache, then denied localhost sockets), not test failures.
- `go test ./internal/analysis/blast -count=1` passed without test-cache reuse.
- Frontend `npm exec tsc -- --noEmit --incremental false` passed.
- `docs/openapi.json` parsed successfully.
- `git diff --cached --check` and `git diff --check` passed.
- Docker was unavailable in the current WSL environment, so Compose, real-database
  integration and Chromium workflows were not rerun on 2026-09-12. Their last
  recorded successful runs are the dated 2026-09-10 evidence below.

No commit was created by this audit.

### Phase completion evidence — 2026-09-10

On 2026-09-10:

- Focused and complete Go unit suites passed. `analysis/blast` covers chain,
  diamond/multiple paths, two synthetic application roots, application filter,
  cycle, disconnected package, multiple versions, representative/count caps,
  maximum-depth bound, not-found identity and weakest-evidence confidence.
- `go test ./... -count=1` passed inside Docker with real Compose PostgreSQL and
  Neo4j and `PHASE1_INTEGRATION=1`. The integration adds impact/path response
  checks, incomplete evidence, invalid filters/bounds and cross-scan package
  isolation, while preserving ingestion, failures, findings and repeat safety.
- `go test -race ./... -count=1` and `go vet ./...` passed. Frontend
  `npm exec tsc -- --noEmit --incremental false` passed.
- `docker compose up -d --build` passed, including Go binaries and the Next.js
  production build. UI/API remain at `http://localhost:3000` and `:8080`.
- Deterministic Chromium `frontend/tests/phase3-impact.cjs` passed with a
  two-application diamond: exact counts/path lengths/depth, application filter,
  affected-region/path highlighting, stable positions/viewport, finding linkage,
  responsive inspector and no browser errors.
- Phase 2 deterministic regressions (`finding-selection.cjs` and
  `phase2-workflow.cjs`) passed after the rebuild, including delayed/rapid
  selection changes and dependency paths.
- Saved real scan `579acc06-c7c0-45c1-8bba-a5eb0002d2db` of
  `https://github.com/Hasenpfote/python-poetry-example` returned live impact for
  `black@23.3.0`: 0 downstream packages, 1 exposed application, 1 exact path of
  depth 1, HIGH snapshot evidence and incomplete dependency evidence. The live
  finding regression passed black/CVE-2024-21503 → attrs → black → paths.
  `graph-readability.cjs` passed with all 48 nodes/75 edges, zero measured overlap,
  stable selection/path interactions and responsive widths.
- A fresh, unmocked browser submission of the same public URL completed as scan
  `24ade495-66ec-4766-9942-2daf7da3d328` at commit
  `f0e6e6b09c0090728fd1edbb09ee3f39fd1e3b9b`: GitHub ingestion, persisted queue,
  worker, both databases and live OSV ran; 47/47 packages were CHECKED and 53
  current alias-deduplicated findings were stored. Browser finding/package/path/
  impact inspection passed. `black@23.3.0` again returned 0 downstream packages,
  1 exposed app and 1 exact depth-1 path with HIGH evidence; dependency metadata
  remained explicitly incomplete. Live advisory totals can change over time.
- OpenAPI JSON parses. Full third-party OpenAPI conformance and large 10k/100k
  performance/load testing were not performed.

The synthetic provider/UI tests are deterministic. The database workflow uses
real databases but controlled GitHub/provider fixtures. Saved-scan browser/API
reads are real persisted data; creation POSTs are replayed to open that scan. The
fresh live scan used no mocked API requests.

## Approved Blast Radius model

Approved calculation for package `p` in one scan:

1. `A = exposed_application_roots(p) / total_application_roots`.
2. `D = downstream_packages(p) / (total_package_versions - 1)`.
3. Let `paths(p)` be the exact number of simple application-to-`p` paths. Let
   `M = max(paths(q))` over every package-version `q` in the same scan. Define
   `P = ln(1 + paths(p)) / ln(1 + M)` to keep a combinatorial path count from
   dominating. If `M=0`, `P=0`.
4. `score = round(100 × (0.50A + 0.30D + 0.20P))`, clamped to 0–100.

Zero denominators produce a zero component: no application roots means `A=0`;
fewer than two package nodes means `D=0`; no paths means `P=0`. A disconnected
package therefore scores 0/LOW. Reverse reach uses unique nodes, so cycles do not
duplicate reach; multiplicity uses simple paths. If any required selected or
workspace path count is capped/depth-limited, or dependency evidence is incomplete,
the implementation withholds the numeric score instead of substituting a lower bound.
The UI continues to show the independently valid snapshot metrics. Scores
are comparable only among package versions in the same scan because all three
denominators are scan-relative. They are structural impact classifications, never
exploit probabilities or CVSS replacements.

Retained blueprint bands (product-defined, not industry standards): 0–19 LOW,
20–39 MODERATE, 40–59 ELEVATED, 60–79 HIGH, 80–100 VERY HIGH. Retained weights:
50% application reach, 30% downstream-package reach, 20% path multiplicity.
Weights are named configuration constants and returned as explanation data.

Worked examples (all evidence complete and path counts exact):

| Shape / selected node | A | D | P | Proposed result |
|---|---:|---:|---:|---:|
| Chain `App→A→B→X`, select X; 1 app, 3 packages, workspace M=1 | 1/1 | 2/2 | ln2/ln2 | 100 VERY HIGH |
| Same chain, select B | 1/1 | 1/2 | ln2/ln2 | 85 VERY HIGH |
| Diamond `App→A→X`, `App→B→X`, select X; M=2 | 1/1 | 2/2 | ln3/ln3 | 100 VERY HIGH |
| Two apps `App1/App2→A→X`, select X; 2 apps, 2 packages, M=2 | 2/2 | 1/1 | ln3/ln3 | 100 VERY HIGH |
| An isolated X beside an otherwise connected graph | 0 | 0 | 0 | 0 LOW |

This choice of logarithmic path normalization and the “score unavailable when
capped/incomplete” policy were approved by the user. An alternative denominator
or an interval score would materially change product meaning and will not be
chosen silently.

## Remaining work and decisions

### Approved score implementation and verification

The user approved the formula and conservative availability policy. Named constants
retain weights 0.50/0.30/0.20. Components expose numerators, denominators,
normalized values and contributions in score points. The UI displays full-scan
application/package totals, scan-relative wording, prototype bands, and the
limits of interpreting HIGH/VERY HIGH. Null score/classification values always
include reason codes; zero is reserved for an available zero result.

The analysis counts all simple application-path prefixes in one full-scan pass,
recording each package's count and the normalization maximum. A 16-entry
process cache keys results by scan ID, topology fingerprint and completeness.
Concurrent calls share normalization; cache admission waits at most two seconds.
Eviction/process restart permits bounded recomputation. Time-limited failures
are not cached. Work is bounded at 200,000 node/edge visits and two seconds per
traversal, in addition to depth 100 and per-package path cap 1,000.
The service supplies a five-second context. Filtered paths and score normalization
are independent. If another package's count is capped, the score is null with
SCAN_PATH_COUNT_CAPPED and blocking_package. Incomplete scans skip normalization
and return INCOMPLETE_DEPENDENCY_EVIDENCE without modifying stored limitations.

Verification for this approved follow-up (2026-09-10):

- Go worked examples: chain X=100, B=85; diamond/shared X=100; isolated=0;
  single application/dependency=70 HIGH. Empty/no-application cases, cycles,
  distinct versions, cap/depth/work/cancellation, filter independence, cache
  invalidation/isolation, rounding and all band boundaries are tested.
- Full Go suite with PHASE1_INTEGRATION=1 passed in Docker on real PostgreSQL/
  Neo4j. CompleteSyntheticBlastRadius writes an independent complete synthetic
  graph and verifies 70/HIGH through HTTP. No live metadata limitations are removed.
- TypeScript and Docker production build passed. Focused race/vet checks passed.
- Chromium phase3-impact.cjs verifies numeric/component rendering using explicit
  complete-evidence API fixtures and unchanged scores after application filtering.
  It also verifies another package's capped normalization is explained without
  hiding valid paths. The deterministic rapid finding-selection regression passed.
  Final database regression confirms complete-evidence limitation fields are
  empty arrays suitable for the inspector, rather than null.
- Chromium finding-selection-live.cjs verifies the persisted Poetry scan
  579acc06-c7c0-45c1-8bba-a5eb0002d2db displays Score unavailable with
  INCOMPLETE_DEPENDENCY_EVIDENCE while preserving finding/package/path behavior.
  These are live database/API reads; only scan-opening POST responses are replayed.
- Full fresh live ingestion evidence earlier in this document remains historical;
  this follow-up does not claim another provider query.

Scores are computed on read, not stored in findings snapshots. Cache is local
to one API process; distributed caching, performance/load benchmarking and full
OpenAPI conformance remain deferred infrastructure work. Other final blueprint
features remain in FEATURE_MATRIX. No Phase 4 work is included.

The approved score and classification are implemented. Real ingestion still discovers
one root application, so multi-application correctness is synthetic only.
Broader load behavior, SCC visualization, path ordering policy, pagination and
observability remain incomplete. All later blueprint capabilities remain tracked
in FEATURE_MATRIX. Phase 4 is not authorized and will not start automatically.

## Where the Docker build is hosted

After running `docker compose up -d --build` from the repository root, Compose
publishes the local prototype on the host machine at:

| Service | Host address | Purpose |
|---|---|---|
| ChainReaction UI | `http://localhost:3000` | Main browser workspace |
| ChainReaction API | `http://localhost:8080` | REST API |
| API health check | `http://localhost:8080/healthz` | API availability check |
| Neo4j Browser | `http://localhost:7474` | Local graph database browser |
| Neo4j Bolt | `neo4j://localhost:7687` | Graph driver connection |
| PostgreSQL | `localhost:5432` | Local relational database connection |

These are Docker port mappings for local development, not a public deployment.
From another machine, `localhost` must be replaced with the Docker host address
and the ports must be intentionally exposed through that host's firewall/network.
