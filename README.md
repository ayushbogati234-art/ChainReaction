# ChainReaction

ChainReaction is a local graph-first analyzer for public Python Poetry repositories. Phase 1 provides commit-pinned dependency graphs and paths; Phase 2 adds OSV enrichment and evidence-backed findings; Phase 3 adds scan-scoped reverse-reachability and bounded path evidence. Phase 3 includes the approved scan-relative Blast Radius score with conservative availability reasons. The full product specification remains the [architecture blueprint](ChainReaction-Architecture_Blueprint.md); track progress in the [feature matrix](docs/FEATURE_MATRIX.md), [Phase 3](docs/phases/PHASE_3.md), and [handoff](docs/HANDOFF.md).

## Run locally

1. Copy `.env.example` to `.env`.
2. Run `docker compose up --build`.
3. Open `http://localhost:3000` and submit a public repository with a supported root Poetry pair.

The API is on `http://localhost:8080`; its Phase 1 contract is in [docs/API.md](docs/API.md).

## Supported Python subset

Phase 1 supports one application at repository root with a Poetry `[tool.poetry]` manifest and an adjacent `poetry.lock` containing `[[package]]` records and package dependency tables. It reads package names, exact lock versions, direct root dependencies, dependency edges, markers, extras, categories/groups, and source type where represented. An edge is always stored as `dependent -> dependency`.

`requirements.txt`, manifest-only resolution, nested applications, Git/path dependencies, dynamic metadata, non-unique lock matches, and environment-specific selection are intentionally not resolved. Such inputs fail clearly when they cannot form the supported pair; encountered ambiguous or conditional evidence makes a completed scan `incomplete` and includes limitations. No unconstrained declaration is converted into an exact version or relationship.

The analysis environment is deliberately non-selective in Phase 1: marker-bearing relations are preserved but not evaluated for a target Python/platform/extras environment. Therefore their presence does not prove an active installed dependency.

## Safety and scope

Repository ingestion uses GitHub metadata followed by a commit-SHA-pinned archive download. It applies HTTP timeouts, a 10 MiB archive response cap, and a 1 MiB selected-file cap. ChainReaction never clones, installs, builds, imports, or runs scanned repository code, tests, hooks, or scripts.

PostgreSQL stores projects, scans, failures, a persisted job queue, package/advisory snapshots and findings. Neo4j stores scan-scoped dependencies and vulnerability relationships. Scores, simulations, mitigation, systemic detection, additional formats and multi-application discovery remain roadmap work.

Phase 2 queries `https://api.osv.dev/v1/query` using validated default-PyPI package identities and exact numeric releases. Other version forms or custom sources are explicitly skipped. No OSV key is required. Requests have an 8-second timeout, a 4 MiB response cap, up to five pages, 200 ms spacing, and a 60-second per-scan enrichment budget. Rate-limit/unavailable responses stop further queries in that scan; coverage records explain what failed or was not queried. Create a new scan to retry; snapshots are not a cross-scan cache.

Provider failures preserve completed dependency processing with degraded/partial enrichment. Dependency incompleteness and provider coverage are separate. Findings describe potential dependency exposure, not confirmed exploitation; provider fixed versions are not verified compatible upgrades. See [Phase 2 evidence and limitations](docs/phases/PHASE_2.md).

## Test

Run `cd backend && go test ./...`. The suite covers Poetry parsing/direction/directness, unsupported input, incomplete ambiguous conditions, diamond/shared dependency behavior, and cycle-safe capped path enumeration. The controlled inputs and graph shapes live under `fixtures/python/`. Docker is not available in every WSL installation; enable Docker Desktop WSL integration to run the full Compose flow.

Phase 1 closeout checks, architecture boundaries, exact commands, live scan IDs,
and verification limits are recorded in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).
The database integration test under backend/tests is opt-in and uses real
PostgreSQL/Neo4j with controlled GitHub responses. The browser test under
frontend/tests drives the actual local application and public GitHub requests.
