# Phase 4 — structural criticality

Status: implemented; verification evidence is maintained with this phase as the
work closes. The [architecture blueprint](../../ChainReaction-Architecture_Blueprint.md)
remains authoritative.

## Approved scope

Phase 4 implements blueprint §§3.15, 10.9 and the topology-only evidence needed
to return structural criticality with a package impact response. It does not
implement the overall Risk Score, exploitability evidence, Systemic Dependency
findings, simulations, mitigation/RRP, search, filters, additional graph modes,
new inputs, persistence of scores, or broader caching/observability.

The approved project-relative model is:

1. 30% reverse fan-in: distinct immediate dependent application/package nodes,
   normalized to the maximum package fan-in in the scan.
2. 30% application reach: unique reverse-reachable application roots divided by
   all application roots in the scan.
3. 20% shared-dependency concentration: distinct immediate dependents divided
   by all other nodes in the scan.
4. 20% independent-path concentration: `ln(1 + distinct simple application
   paths)` normalized to the scan maximum.

The last factor counts distinct cycle-safe simple paths; it does not claim
edge-disjoint paths, independent failures, exploitability, or business impact.
All weights are named constants and returned with the component evidence.

## Evidence and availability

`analysis/criticality` is a pure domain analysis package. It receives only the
scan-filtered topology snapshot; Neo4j and HTTP types remain outside it. The
existing bounded simple-path counter is reused, with the same 10,000-node,
100,000-edge, depth, path-count, work and execution limits. A score is withheld
instead of using a lower bound when dependency evidence is incomplete or
scan-wide path normalization is capped, depth-limited, work-limited or timed
out. Valid reachability/path evidence remains visible independently.

The score is computed on read and cached only in the local API process, keyed by
scan ID and immutable topology fingerprint. It is not stored in findings or
treated as a cross-commit cache.

## Verification

The full Go suite passes, including deterministic formula, incomplete-evidence,
cache-isolation and HTTP integration assertions. TypeScript checking and OpenAPI
JSON parsing pass. Browser execution remains pending in this environment because
Docker is unavailable and the local `next` executable is invalid after a
lockfile-pinned install attempt; the deterministic browser fixture has been
extended to assert the rendered component evidence.

## Phase plan

The remaining blueprint is split into dependency-ordered phases:

| Phase | Scope | Prerequisite decision |
| --- | --- | --- |
| 4 | Structural criticality | Completed: transparent topology formula |
| 5 | Explainable overall risk and Systemic Dependency detection | Exploitability evidence source and systemic thresholds |
| 6 | Compromise/remediation simulations | Risk-only versus dependency re-resolution model |
| 7 | Mitigation priority and RRP | Practicality, fix-availability and zero-baseline policies |
| 8 | Search, filters and graph views | Query/filter semantics and mode interaction rules |
| 9 | Commit/provider caching, observability and resilience | Freshness, invalidation, retention and metric policies |
| 10 | Controlled Butterfly Effect demo and final blueprint cross-check | Approved demo data/scenario and completion criteria |

Later phases remain unapproved until explicitly authorized. This partition does
not remove any requirement from the architecture blueprint.
