# Phase 1 architecture and verification

The repository follows Sections 9 and 10 of
[the architecture blueprint](../ChainReaction-Architecture_Blueprint.md).
Python is the only analyzed ecosystem; the Go backend and Next.js frontend are
implementation technologies. Phase 2 has not started.

## Responsibilities

|Location|Implemented responsibility|
|-|-|
|backend/cmd/api, backend/cmd/worker|Configuration, process signals, invoking startup|
|backend/internal/app/api.go, worker.go|Database connections, dependency wiring, HTTP server and worker startup, shutdown|
|backend/internal/domain|Client-independent project, scan, graph and evidence types|
|backend/internal/transport/http/router.go|Routing, CORS, HTTP validation, JSON and status mapping; delegates to services|
|backend/internal/service/project\_service.go|Repository URL validation and project operations through persistence interfaces|
|backend/internal/service/scan\_service.go|Ingestion, resolution and graph/scan persistence orchestration|
|backend/internal/service/graph\_service.go|Scan-readiness checks and graph operations through graph.Repository|
|backend/internal/jobs/queue.go, scan\_job.go, worker.go|Queue interface, claiming, execution timeout, failure recording and polling|
|backend/internal/ingest/github|Validated GitHub URL, commit-pinned archive retrieval and size limits|
|backend/internal/resolver/resolver.go|Shared resolver interface and neutral output|
|backend/internal/resolver/python|pyproject.toml and poetry.lock parsing, conditions, provenance and incompleteness|
|backend/internal/graph/repository.go|Graph persistence/query interface|
|backend/internal/graph/neo4j/repository.go|Cypher, graph writes, scoped topology, package lookup and simple paths|
|backend/internal/persistence/postgres/db.go|SQL, migrations, transactional scan/job persistence and abandoned-job recovery|
|frontend/src/app|Route and layout|
|frontend/src/features/scan|Scan submission, polling, errors and selection state|
|frontend/src/components/graph|Cytoscape rendering and node selection|
|frontend/src/lib/api, frontend/src/types|API client and response types|
|backend/tests, frontend/tests|Closeout regression/integration tests and live browser script|

HTTP handlers and services do not import Postgres or Neo4j drivers. Domain types
do not import infrastructure clients. Queue SQL remains in the Postgres adapter
behind the jobs.Queue interface. No scoring or other deferred analysis modules
were created.

Some implemented modules are consolidated: domain/model.go, the HTTP router,
GitHub ingestor, Postgres adapter and Neo4j repository. The in-memory simple-path
helper remains in internal/graph as a topology operation. These consolidations
preserve layer responsibilities. No empty directories or deferred stubs were
added.

## Runtime behavior

Creating a scan inserts its scan record and queue entry in one transaction.
The worker claims with FOR UPDATE SKIP LOCKED, runs service orchestration with
a two-minute context timeout, and records terminal scan/job states together in
a transaction. Cancellation does not prevent a separate, bounded attempt to
record failure. Claim failures and failures to persist terminal state are logged.
Jobs left RUNNING by a crashed worker are marked FAILED/WORKER\_INTERRUPTED after
a three-minute lease when another worker polls. This is terminal recovery;
automatic retry/re-resolution is not implemented. A user can start a new scan.

Graphs are exposed only after a COMPLETED scan. Application and package nodes
are scoped by scan\_id. Package adjacency filters both endpoints; path queries
require every traversed node to belong to the selected scan. Missing scans and
packages produce 404, and pending/failed scan graph requests produce 409.
Browser-encoded package identifiers are decoded before lookup.

Poetry metadata is parsed without selecting an environment or executing project
code. Markers, extras, groups, source evidence, optional flags and declared
constraints are retained. Missing or ambiguous lock matches are omitted and
explained. A completed scan can be incomplete: completion means processing
finished, not that environment-specific installation or exploitability was proven.
HIGH confidence on a package describes lock metadata evidence.

## Closeout evidence — 2026-09-08

### Input provenance follow-up — 2026-09-09

The retrieval layer extracts only root `pyproject.toml` and `poetry.lock` from
the commit-pinned archive. The Python resolver requires both nonempty files.
`parsePyproject` supplies `[tool.poetry].name` and main, legacy development,
and Poetry group dependency declarations. `parsePoetryLock` supplies package
versions and package-to-package dependency declarations. Direct edges record
`pyproject.toml + poetry.lock`; package-version evidence records `poetry.lock`.
The package source label never implied that the manifest was ignored.

Fixed reporting gaps: application source/confidence and the two input filenames
now survive Neo4j persistence; status polling includes the commit SHA. Mixed
`[project]` dependencies, optional dependencies or dynamic metadata, conflicting
application names, and non-Poetry dependency groups return `UNSUPPORTED_INPUT`.
Unknown dependency fields/types and git/path/URL declarations omit the affected
relationship and explain why. Invalid TOML also returns `UNSUPPORTED_INPUT`.
Simple exact numeric release mismatches omit the relationship as inconsistent.
Version ranges and other constraint forms remain name-matched evidence and now
explicitly mark the result incomplete because compatibility is not verified.
This does not introduce a Poetry solver or change metadata confidence scores.
Lock content-hash/freshness verification is still unimplemented and disclosed.

The UI's “Supported inputs and scan provenance” disclosure is available at all
screen widths. It lists recorded root files, commit and application identity,
explains file responsibilities and limitations, and distinguishes package-version
evidence from direct declarations (including group and constraint). Older scans
without recorded input files request a new scan instead of inventing provenance.

Controlled regression: `TestManifestChangesWithIdenticalLock` reads the original
`fixtures/python/chain/pyproject.toml` and the alternative
`fixtures/python/provenance/pyproject.toml` against the **same bytes** from
`fixtures/python/chain/poetry.lock`. Both retain `a@1.0.0`, `b@2.0.0` and `a -> b`.
The application changes from `chain` to `changed-application`, and its direct edge
changes from `app -> a` to `app -> b`. Direct flags change accordingly. A tox.ini
entry in the simulated archive has no effect. The live database test also verifies
the changed persisted path and absence of an application path to the removed
direct dependency; GitHub archive responses alone are simulated in these tests.

Passed backend `go test ./...`, frontend TypeScript checking, Docker production
build, and `PHASE1_INTEGRATION=1 go test ./tests -run TestDatabaseWorkflow -v
-count=1` in a temporary Go container on the Compose network. The initial localhost
database attempt failed authentication; the Compose-network run using `.env`
passed. Test cleanup removes only its generated schema and scan graph records.

Fresh real GitHub ingestion used
`https://github.com/Hasenpfote/python-poetry-example`, scan
`b1a39eb6-6948-4ef5-8a35-04476ef1a4a6`, commit
`f0e6e6b09c0090728fd1edbb09ee3f39fd1e3b9b`. It completed with 48 nodes, 75 edges,
12 direct declarations and 123 limitations (including unverified version
compatibility). Persisted application evidence records both root input files
and source `pyproject.toml`.

Chromium regression passed against that saved scan, replaying only creation
responses to open it and making real status/graph/package/path GET requests.
It checks recorded filenames and commit, the tox scope notice, separate inspector
evidence labels, and all graph controls, route highlighting, stable selection,
element preservation and responsive inspector widths (1000/760/390px). Opening
the provenance panel initially exposed stale canvas pointer coordinates; observing
page resizing fixed it and the rerun passed without browser errors.

`tox.ini` parsing remains outside Phase 1. Separate tox analysis would require
approved input retrieval/parsing, explicit tox version and environment/factor
selection, handling inheritance/substitutions and referenced requirements or
constraints files, an environment-scoped dependency model and UI, and a policy
for mapping environment declarations to exact versions when Poetry's lock does
not cover them. It must keep test-environment dependencies distinct from
application dependencies and preserve the no-repository-code-execution rule.
The presence of the `tox` package in poetry.lock does not describe tox environments.

The automated controlled fixtures and the live public GitHub run are different
verification layers. No controlled public fixture URL was supplied during this
session. The controlled cases use checked-in fixtures and simulated GitHub HTTP
responses; the browser run uses an inspected public Poetry example at the exact
commit recorded below. No repository was created or published for testing.

### Automated checks

Passed:

* GOCACHE=/tmp/chainreaction-go-cache go test ./... from backend.
* GOCACHE=/tmp/chainreaction-go-cache go test -race ./... from backend.
* GOCACHE=/tmp/chainreaction-go-cache go vet ./... from backend.
* npm run build from frontend; Next.js compilation and type checking passed.
* git diff --check.

The normal Go suite skips the database integration test unless explicitly
enabled. Existing tests cover ambiguity, diamond/shared paths, cycle avoidance
and path caps. New regression tests cover commit-pinned metadata ingestion,
root-only inputs, unsupported formats, file/archive limits, conditional evidence,
missing edges, worker errors and failure recording after cancellation.

The database integration suite also passed, with real PostgreSQL and Neo4j:

```sh
docker build --target build -t chainreaction-phase1-tests backend
docker run --rm --network chainreaction\_default --env-file .env \\
  -e PHASE1\_INTEGRATION=1 \\
  --mount type=bind,source=/home/anish/ChainReaction/fixtures,target=/fixtures,readonly \\
  chainreaction-phase1-tests go test ./tests -v -count=1
```

This suite uses a unique Postgres schema and unique graph scan IDs, removing only
its own test data afterward. GitHub HTTP responses alone are simulated.
Observed checks:

* Queue entry visible from a separate database connection before processing.
* Controlled chain completes with 3 nodes, 2 edges, a recorded commit SHA and
app:chain -> a@1.0.0 -> b@2.0.0.
* Raw and percent-encoded package IDs both work.
* Two scans with different locked versions cannot retrieve each other's packages.
* A deliberately injected cross-scan edge is excluded from graph/adjacency reads.
* Unsupported inputs persist FAILED and UNSUPPORTED\_INPUT.
* Conditional input completes with incomplete=true and persisted marker evidence.
* Abandoned worker lease produces FAILED and WORKER\_INTERRUPTED, including when
no new job is available.

### Live integration and browser verification

The actual Compose API, worker, PostgreSQL, Neo4j and frontend were rebuilt/run.
Chromium executed frontend/tests/live-workflow.cjs with no mocked network routes.
It submitted a GitHub URL through the form, observed completion and the graph
canvas, selected a package, read dependency paths, clicked a different Cytoscape
node, checked the incomplete-result warning and verified a visible unsupported
input failure. No browser page errors were observed.

Successful repository:
[Hasenpfote/python-poetry-example](https://github.com/Hasenpfote/python-poetry-example)

* Commit: f0e6e6b09c0090728fd1edbb09ee3f39fd1e3b9b.
* Final successful scan: 5284dce3-a714-4f11-91e6-bb8ad93112a5.
* 48 nodes (one application and 47 packages), 75 dependency edges.
* COMPLETED / job COMPLETED, one processing attempt, incomplete=true.
* 122 evidence limitations, including unselected environments/extras/groups.
* Package button selected pytest@7.2.2; graph node click selected isort@5.11.5.
* pytest paths: app:python-poetry-example -> pytest@7.2.2 and
app:python-poetry-example -> pytest-mock@3.10.0 -> pytest@7.2.2.
* A separate live API request with limit=1 returned one path and truncated=true.
* Live scans 323bdcb1-bd66-4b7c-8e76-1d32d5eeef9b and
5284dce3-a714-4f11-91e6-bb8ad93112a5 each returned only their own scan\_id nodes.

The nonempty octocat/Hello-World repository failed with UNSUPPORTED\_INPUT in both
the UI and persisted scan/job records:
db6f4334-f838-48e4-8abd-de75fdffa86a.
An earlier attempt using the empty remote ChainReaction repository returned
GitHub 409; it persisted GITHUB\_METADATA\_ERROR and rendered the error in the UI
(scan d9c18160-cf18-4ede-8d64-d8ab9025d36b).

The first browser attempt exposed a percent-encoded package lookup bug, which was
fixed and regression-tested before the successful rerun. The earlier raw-URL
tests alone had not detected it.

Browser command used in this environment:

```sh
LD\_LIBRARY\_PATH=/tmp/chainreaction-browser-RR6AjF/libs/usr/lib/x86\_64-linux-gnu \\
PLAYWRIGHT\_MODULE=/tmp/chainreaction-browser-RR6AjF/node\_modules/playwright \\
PLAYWRIGHT\_BROWSERS\_PATH=/tmp/chainreaction-browser-RR6AjF/browsers \\
TEST\_SCREENSHOT=/tmp/chainreaction-phase1-live.png \\
node frontend/tests/live-workflow.cjs
```

Playwright, Chromium and missing NSS/NSPR libraries were downloaded under /tmp;
no project dependency or system package was installed for this test. The browser
script can use another Playwright installation through PLAYWRIGHT\_MODULE.

### Limits and environment notes

Host port 5432 did not reach the intended Compose database with its configured
credentials, and direct Docker bridge access from the host timed out. Running
the integration suite inside the Compose network resolved that test blocker.
Chromium initially lacked shared libraries; temporary library extraction resolved
that blocker. All five application services were running at handoff.

No blocker remains for the tested local workflow. A live run against a
user-controlled public fixture repository is still unverified because no such URL
was provided; the exact public-example snapshot above was used instead. Large
graph load testing, a forced live database outage, and production deployment are
not claimed as verified. Synthetic tests cover injected worker/persistence errors.

Five live scan records (three completed and two failed) were retained as evidence.
Existing user files and the user's blueprint edits were preserved. Phase 2,
vulnerability enrichment, scoring and simulation remain unimplemented.
