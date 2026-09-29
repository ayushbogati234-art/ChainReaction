# ChainReaction project handoff

Updated 2026-09-21 — Phases 1–3 are closed out. Phase 4 structural criticality
is implemented as a separate, topology-only score with conservative availability.
The project is not complete.

## Authority and exact approved scope

The [architecture blueprint](../ChainReaction-Architecture_Blueprint.md) remains
the complete product specification. Coverage is tracked in
[FEATURE_MATRIX.md](FEATURE_MATRIX.md); phase evidence is in
[Phase 1](ARCHITECTURE.md), [Phase 2](phases/PHASE_2.md), and
[Phase 3](phases/PHASE_3.md). Do not replace or narrow the blueprint.

The current authorization is Phase 4 structural criticality: transparent,
scan-relative fan-in, application-reach, concentration and distinct-simple-path
components; conservative score availability; impact API/UI evidence; tests; and
the documentation phase split. Later phases remain unapproved. Do not stage or
commit without explicit permission.

## Completed and unfinished

Phase 2 is committed by the user at `a3c6798` and includes the inspector fix:
package changes clear findings owned by another package, same-package findings are
preserved, and stale detail responses are aborted and identity-checked. Both
deterministic rapid/delayed-response and saved-scan regressions pass.

Uncommitted Phase 3 work adds a typed Neo4j topology snapshot, pure
`analysis/blast` reachability/path analysis, service completeness context, paths
and impact APIs, and inspector/graph integration. It distinguishes exact snapshot
reach counts from incomplete discovery, exact/capped path counts, displayed
representatives and exact/observed-only maximum depth. The UI filters by synthetic
or persisted application roots and keeps Cytoscape positions/viewport stable.

The approved score is implemented in analysis/blast/normalize.go. It returns
component evidence, full-scan totals, rounded classifications, and null scores
with reason codes for incomplete evidence or inexact normalization. A bounded
cache reuses scan-wide counts across clicks; path display/application filters do
not affect scores. Execution and work limits protect both path traversals. Overall Risk Score, structural criticality, simulations,
mitigation/RRP, Systemic Dependency, search/filters/modes, broader inputs/caches/
observability and the full demo remain in FEATURE_MATRIX.

## Architecture decisions and user corrections

Python-only analysis; Go/Next.js are implementation technology. Supported input
remains paired root `pyproject.toml` + `poetry.lock`: manifest identity/direct
declarations versus lock version/transitive evidence. `tox.ini` and broader inputs
remain unsupported. Scans are metadata-only and never execute repository code.

Use the existing monorepo and prescribed responsibility boundaries. Domain is
client-independent; Neo4j owns scan-scoped persistence/query adaptation;
`analysis/blast` owns algorithms; services orchestrate; HTTP parses/delegates;
startup stays in app; jobs retain worker flow. Consolidated files are acceptable.
Cytoscape, dependent → dependency direction, dark styling and inspector behavior
are preserved. Reachability is exact only within the persisted snapshot; scan
incompleteness is a separate evidence limitation. Never describe impact as
confirmed exploitation or a numeric score as exploit probability.

## Files and Git state

Branch `main`; HEAD `a3c6798 (phase 2 completed)`, tracking `origin/main`. Current
resume check found 34 staged files, with no new commit since the prior session.
The existing staging was preserved. This handoff refresh is an unstaged change
on top of the staged handoff; no staging or commit was performed during resume.
Modified/new paths are listed in
PHASE_3.md under “Files changed”; they are concentrated in:

```text
backend/internal/{domain,graph,analysis/blast,service,transport/http}
backend/tests/integration_test.go
frontend/src/{components,features,lib/graph,types}
frontend/tests
docs/{API.md,openapi.json,FEATURE_MATRIX.md,HANDOFF.md,phases}
```

No existing commit was rewritten, no Phase 2 file was discarded, the blueprint
has no intended modification, and nothing should be staged until user review.
Always recheck `git status --short` before handoff.

Resume verification (2026-09-11): Git status/history and Docker Compose status
were inspected. API/frontend/worker are running; PostgreSQL and Neo4j are healthy.
The test evidence below is from 2026-09-10; tests were not rerun during resume.

## Tests actually run and results

On 2026-09-10:

- `go test ./... -count=1` passed in Docker against real Compose PostgreSQL and
  Neo4j with controlled GitHub/OSV fixtures and integration enabled. It includes
  scan isolation, repeat-safe Phase 2 evidence and new impact API assertions.
- `go test -race ./... -count=1` and `go vet ./...` passed locally with required
  socket permission. Focused blast/graph/service/transport tests passed.
- `npm exec tsc -- --noEmit --incremental false` passed from `frontend/`.
- `docker compose up -d --build` passed, including Next.js production compilation.
- Deterministic Chromium Phase 3 multi-app/diamond UI passed: exact counts and
  lengths, depth, app filter, affected/path highlighting, stable graph,
  finding linkage, 390px width and no browser errors.
- Phase 2 deterministic browser selection/findings regressions passed.
- Saved scan `579acc06-c7c0-45c1-8bba-a5eb0002d2db` for
  `https://github.com/Hasenpfote/python-poetry-example` passed the real black
  CVE → attrs → black → paths browser sequence. Its real graph regression retained
  48 nodes/75 edges with zero measured overlap and stable responsive interactions.
  Live impact for `black@23.3.0` returned 0 downstream packages, 1 exposed app and
  1 exact depth-1 path; dependency evidence remains incomplete and is labelled so.
- A fresh unmocked browser run of the same public repository completed scan
  `24ade495-66ec-4766-9942-2daf7da3d328` at commit `f0e6e6b...`: 47/47 packages
  checked, 53 current findings, real queue/worker/GitHub/OSV/databases and impact
  inspection. Advisory counts are time-dependent.
- OpenAPI JSON parsing passed. Third-party conformance and large-scale performance
  testing remain gaps.

The approved formula follow-up passed the full Go suite with real database
integration, including an independent complete-evidence 70/HIGH HTTP fixture.
Unit tests cover worked examples, empty/disconnected graphs, cycles/versions,
incomplete/capped/depth/work/cancellation availability, cache invalidation,
filter independence and rounding/bands. Focused race/vet, TypeScript and Docker
build passed. Browser tests verified numeric components and filter independence;
the saved live Poetry scan correctly showed Score unavailable with
INCOMPLETE_DEPENDENCY_EVIDENCE. Exact evidence is in PHASE_3.md.

One combined verification command accidentally invoked `npm exec tsc` from
`backend/`, which fetched the unrelated deprecated npm `tsc` runner into npm's
cache and failed. It created no repository file. The correct frontend command was
then run and passed. This does not invalidate the independently passing Go race/vet.

See PHASE_3.md for the precise distinction between deterministic fixtures, real
databases, saved real-scan reads and the fresh unmocked workflow.

## Runtime, limitations and next step

The rebuilt stack is running at UI `http://localhost:3000`, API
`http://localhost:8080`, PostgreSQL `:5432`, Neo4j HTTP `:7474` and Bolt `:7687`.
Start/rebuild with `docker compose up -d --build`; inspect with
`docker compose ps`. Use local `.env` without documenting secrets. External
Playwright tooling currently exists under
`/tmp/chainreaction-graph-browser-0TMQ3w` and is not a project dependency.

Real repository ingestion still produces one application root; multi-application
impact correctness is synthetic. Dependency environment/markers/extras/groups
remain incomplete evidence. Path ordering is deterministic but not priority-
ranked. Aggregate path cap is 1,000; display cap 100; depth cap 100; snapshot cap
10k nodes/100k edges. Large graphs, SCC display, pagination, caching, full
observability and external OpenAPI conformance are unverified/unimplemented.

Phase 4 is authorized and implemented. The next concrete step is user review of
its documented evidence; Phase 5 and later phases require explicit authorization.
Do not stage or commit without authorization.
