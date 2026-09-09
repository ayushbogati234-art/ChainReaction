# Phase 2 — vulnerability enrichment and evidence-backed findings

Status: authorized, implementation in progress. Full scope authority is the
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

Data/API/UI changes, actual validation, files changed and limitations will be
recorded here as implemented. No Phase 2 capability is verified yet.
