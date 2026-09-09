# ChainReaction — Architecture Blueprint
## Open-Source Supply Chain Risk Intelligence
**Status:** Architecture draft for prototype planning  
**Implementation status:** NOT STARTED  
**Architect:** Ayush   
**Builder:** Codex, only after explicit implementation approval

---

## 1. Executive Summary

ChainReaction is a graph-based software supply-chain risk analysis platform.

Its purpose is not merely to report vulnerable packages. It is designed to answer the harder questions security teams actually face:

- Which dependency matters most in *this* software ecosystem?
- Which applications are potentially exposed if a package is compromised?
- Through which dependency paths does that exposure exist?
- How large is the blast radius?
- Why is one package more important than another even if its CVSS score is lower?
- Which remediation would reduce the most downstream risk?
- How confident is the system in its conclusions?

The system is centered on an explicit dependency graph. The graph, not an LLM, is the source of truth for topology, path discovery, propagation analysis, blast-radius calculation, structural criticality, and remediation simulation.

AI/LLM functionality is not required for the core prototype and should not be used to invent risk conclusions. If an LLM is added later, it may only summarize already-computed evidence.

---

## 2. Problem-Statement Alignment

The hackathon problem asks for a system that:

1. Maps relationships in a software dependency ecosystem.
2. Identifies critical dependencies.
3. Identifies affected applications.
4. Shows propagation paths.
5. Prioritizes mitigation.
6. Makes reasoning visible.
7. Distinguishes broadly exposed components from isolated ones.
8. Helps teams focus mitigation where one intervention can reduce downstream exposure.

ChainReaction is designed directly around those requirements.

---

## 3. Product Vocabulary

### 3.1 Package
An installable software component from an ecosystem.

Examples:
- `log-parser@2.4.1` in Python (illustrative package)
- `requests@2.32.3` in Python (PyPI)

A package node must be version-specific for accurate vulnerability matching.

Canonical identity should be represented internally as a Package URL (purl) when possible.

Example:
`pkg:pypi/log-parser@2.4.1`

---

### 3.2 Application
A root software unit that depends on packages.

Examples:
- `payment-api`
- `frontend`
- `worker`
- `auth-service`

One repository may contain multiple applications.

---

### 3.3 Direct Dependency
A dependency explicitly declared by an application.

Example:

`payment-api -> requests`

Here, `requests` is a direct dependency of `payment-api`.

---

### 3.4 Transitive Dependency
A package reached through another dependency.

Example:

`payment-api -> package-a -> log-parser`

`log-parser` is transitive from the perspective of `payment-api`.

---

### 3.5 Dependency Edge
A directed relation:

`A DEPENDS_ON B`

The stored graph direction is from dependent to dependency.

Example:

`payment-api -> package-a -> log-parser`

---

### 3.6 Upstream Dependency
In ChainReaction terminology, an upstream dependency is something the current component depends on.

If:

`A -> B`

then B is upstream of A.

---

### 3.7 Downstream Dependent
Something that depends, directly or transitively, on the selected package.

If:

`App -> A -> X`

then `A` and `App` are downstream dependents of `X` from an impact perspective.

ChainReaction uses reverse traversal of `DEPENDS_ON` edges to calculate downstream exposure.

---

### 3.8 Downstream Packages
The number of unique package-version nodes that depend directly or transitively on the selected package.

For a selected package X:

`App -> A -> B -> X`
`App -> C -> X`

The downstream package set is `{A, B, C}`.

Applications are counted separately.

---

### 3.9 Applications Exposed
The number of unique application roots that can reach the selected dependency.

This means the selected package lies somewhere in their resolved dependency graph.

This indicates *potential dependency exposure*, not guaranteed exploitability.

---

### 3.10 Dependency Path
A complete route from an application root to the selected package.

Example:

`Payment API -> sdk-a -> parser-b -> log-parser`

That is one dependency path.

If the same application reaches `log-parser` through two branches, those are two independent paths.

Path counts are useful because repeated reachability can indicate a dependency is deeply embedded or shared through several branches.

---

### 3.11 Maximum Depth
The largest number of dependency edges between an affected application root and the selected package.

Example:

`App -> A -> B -> C -> D -> log-parser`

Depth = 5.

Maximum depth helps explain how deeply nested a dependency is. It is contextual evidence and should not by itself be treated as “more depth = more danger.”

---

### 3.12 Blast Radius
A project-relative measure of how much of the software ecosystem is structurally reachable from a selected package when impact is traced toward dependents/applications.

It summarizes:
- application reach,
- downstream package reach,
- path multiplicity,
- optionally critical-application reach when application criticality is available.

Blast Radius is not the same as CVSS.

CVSS describes vulnerability severity.
Blast Radius describes structural ecosystem impact.

---

### 3.13 Blast Radius Score
A normalized 0–100 project-relative score.

Initial transparent prototype formula:

- 50% Application Reach
- 30% Downstream Package Reach
- 20% Path Multiplicity

Where each component is normalized within the current analyzed project/workspace.

Example:

`88/100 — VERY HIGH`

This means the package is structurally capable of affecting a large portion of the analyzed dependency ecosystem relative to other packages in that project.

The precise weights must be configuration constants and shown in documentation. They must not be hidden.

---

### 3.14 Blast Radius Classification
Suggested prototype bands:

- 0–19: LOW
- 20–39: MODERATE
- 40–59: ELEVATED
- 60–79: HIGH
- 80–100: VERY HIGH

These are product classifications, not industry standards. The UI must state that.

---

### 3.15 Structural Criticality
A graph-topology measure independent of a specific CVE.

A package can be structurally critical even when it has no known vulnerability.

Candidate inputs:
- reverse fan-in,
- betweenness centrality,
- number of unique applications depending on it,
- shared dependency concentration,
- number of independent paths through it.

---

### 3.16 Potential Exposure
A conservative term used instead of saying an application “is exploited.”

If an application has a dependency path to a compromised/vulnerable package, ChainReaction can say:

`Potentially exposed through dependency path`

It must not claim exploitability without evidence.

---

### 3.17 Systemic Dependency
A package with high structural importance but no currently known vulnerability.

Example:

- Known CVEs: 0
- Used by 23 applications
- Structural criticality: 97/100

Label:
`SYSTEMIC DEPENDENCY — no known current vulnerability`

---

### 3.18 Mitigation Priority
A ranking of remediation actions based on expected reduction of project-level risk, not only vulnerability severity.

---

### 3.19 Risk Reduction Potential (RRP)
A counterfactual metric:

`RRP = (baseline aggregate risk - simulated aggregate risk) / baseline aggregate risk * 100`

Example:

- Baseline aggregate risk: 91
- Simulated risk after remediation: 24
- RRP: 73.6%

The prototype must clearly state when this is a modelled counterfactual rather than a fully resolved new dependency tree.

---

### 3.20 Confidence
A provenance-based indicator explaining how trustworthy a graph/finding assertion is.

Suggested levels:

**HIGH**
- exact dependency version from lockfile
- exact package-version vulnerability match

**MEDIUM**
- version resolved from manifest/tooling
- dependency relation derived without a lockfile

**LOW**
- inferred or incomplete dependency information

Confidence should be attached to evidence, not generated as vague AI confidence.

---

## 4. What the Graph Is Doing

### 4.1 Stored Direction

The graph stores:

`dependent -> dependency`

Example:

```text
Payment API
    |
    v
payment-sdk
    |
    v
json-parser
    |
    v
log-parser@2.4.1
```

This answers:

> What does Payment API depend on?

---

### 4.2 Impact Traversal

When the user selects `log-parser@2.4.1`, ChainReaction reverses the traversal direction conceptually:

```text
log-parser@2.4.1
    ^
    |
json-parser
    ^
    |
payment-sdk
    ^
    |
Payment API
```

This answers:

> What depends on log-parser?

That reverse traversal is the foundation for:
- downstream package count,
- exposed application count,
- path discovery,
- blast radius,
- compromise simulation,
- mitigation impact.

---

### 4.3 Multiple Paths

Example:

```text
                Payment API
                 /        \
                v          v
            package-a   package-b
                \          /
                 v        v
                log-parser
```

There are two dependency paths from `Payment API` to `log-parser`.

This matters because the dependency is embedded in two different branches.

---

### 4.4 Shared Dependency Across Applications

```text
Payment API ----\
Inventory API ---+--> package-a --> log-parser
Customer UI ----/
```

`log-parser` reaches three application roots.

Even if its vulnerability severity is lower than another isolated package, its ecosystem impact may be larger.

---

### 4.5 Cycles

Real dependency data can contain cycles.

ChainReaction must:
- detect cycles,
- never recurse forever,
- use visited-node protection,
- count simple paths only,
- cap path enumeration,
- optionally collapse strongly connected components for complex graphs.

Prototype path display should stop at a safe maximum such as 100 displayed paths and report a capped aggregate such as `1000+` where exact enumeration is impractical.

---

## 5. Selected Prototype Direction

The following features have been selected as prototype priorities/candidates. The implementation order is **not yet frozen**.

### Repository ingestion
- Method A: public GitHub repository URL

### Selected product capabilities
- Conservative exposure language and vulnerability ≠ guaranteed exploitability
- Explainable risk score
- Propagation-path explorer
- Simulate Compromise
- Simulate Fix / Remediation
- Mitigation Priority Engine
- Risk Reduction Potential
- Finding Detail Page
- Search, filters, and graph modes
- Systemic dependency detection even without a known CVE
- Confidence / uncertainty reporting

### Architecture/product constraints retained
- Explicit intelligence engines, not one vague “AI engine”
- No chatbot-first interface
- MVP discipline
- Controlled demo scenario
- “From vulnerability severity to ecosystem impact” positioning
- Clear Butterfly Effect narrative
- Real-world/business viability in presentation
- Seven product pillars:
  Discover, Enrich, Understand, Trace, Prioritize, Simulate, Explain

---

## 6. Prototype Functional Scope

### 6.1 Input
Primary input:

`Public GitHub repository URL`

Example:
`https://github.com/org/repo`

Prototype V1 analyzed ecosystem:
- Python projects only.

This scope concerns repositories analyzed by ChainReaction. The proposed Go backend and React/Next.js frontend remain unchanged. JavaScript/TypeScript/npm and Go/Go Modules analysis are outside Prototype V1.

All previously selected product capabilities remain in scope. This ecosystem correction does not approve or freeze the first implementation phase.

---

### 6.2 Python Dependency Inputs
Current candidate inputs:
- `pyproject.toml`
- `requirements.txt`
- `poetry.lock`

The architects have not yet frozen the exact support and implementation order.

Use exact versions and dependency relationships only when supported by the available evidence. A manifest declaration or version constraint must not be treated as an exact resolved transitive dependency graph. Lockfile entries must be interpreted with their applicable environment conditions, extras, and dependency groups; their presence alone does not prove installation or runtime use.

Preserve provenance and report incomplete or ambiguous dependency information through confidence and uncertainty reporting. Do not invent missing transitive versions or relationships.

---

### 6.3 Python Resolution Boundary
The Python Dependency Resolver replaces the Node.js and Go resolvers and produces the shared dependency model through metadata analysis only.

It must not:
- install the scanned project or its dependencies,
- execute `setup.py` or build hooks,
- build the repository,
- run tests or project binaries,
- run arbitrary repository scripts.

If available metadata cannot establish a complete graph, report the limitation. Any metadata resolution tooling and its isolation policy must be approved as part of the first build phase; it must not execute repository code.

---

## 7. Recommended High-Level Architecture

```text
                          Browser
                             |
                             v
                    Next.js / React UI
                             |
                          HTTPS/JSON
                             |
                             v
                       Go API Service
                             |
               +-------------+-------------+
               |                           |
               v                           v
        PostgreSQL                    Job Queue
     metadata/source of truth             |
                                           v
                                      Scan Worker
                                           |
                      +--------------------+--------------------+
                      |                    |                    |
                      v                    v                    v
               GitHub Ingestor       Dependency Resolver   Vuln Enricher
                      |                    |                    |
                      +--------------------+--------------------+
                                           |
                                           v
                                    Graph Builder
                                           |
                                           v
                                         Neo4j
                                           |
                   +-----------------------+----------------------+
                   |                       |                      |
                   v                       v                      v
             Blast Radius             Risk Engine         Path/Graph Engine
                   |                       |                      |
                   +-----------------------+----------------------+
                                           |
                                           v
                                  Simulation / Mitigation
                                           |
                                           v
                                       Findings
                                           |
                                           v
                                      API -> UI
```

---

## 8. Why PostgreSQL + Neo4j

### PostgreSQL
Use for transactional/domain data:
- projects
- repositories
- scans
- scan status
- findings
- vulnerability snapshots
- simulation records
- configuration
- audit metadata

### Neo4j
Use for graph topology:
- application nodes
- package-version nodes
- dependency edges
- graph traversal
- reverse reachability
- path discovery
- centrality/topology calculations

This separation avoids forcing transactional workflow data into a graph database while still using a graph database for the problem it is strongest at.

For hackathon deployment, both run in Docker Compose.

---

## 9. Monorepo Structure

```text
ChainReaction/
├── README.md
├── docker-compose.yml
├── Makefile
├── .env.example
│
├── docs/
│   ├── ARCHITECTURE.md
│   ├── RISK_MODEL.md
│   ├── API.md
│   ├── GRAPH_MODEL.md
│   └── adr/
│       ├── 001-postgres-plus-neo4j.md
│       ├── 002-job-worker.md
│       └── 003-no-code-execution-during-scan.md
│
├── backend/
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   ├── api/
│   │   │   └── main.go
│   │   └── worker/
│   │       └── main.go
│   │
│   ├── internal/
│   │   ├── app/
│   │   │   ├── api.go
│   │   │   └── worker.go
│   │   │
│   │   ├── config/
│   │   │   └── config.go
│   │   │
│   │   ├── domain/
│   │   │   ├── project.go
│   │   │   ├── repository.go
│   │   │   ├── scan.go
│   │   │   ├── application.go
│   │   │   ├── package.go
│   │   │   ├── vulnerability.go
│   │   │   ├── finding.go
│   │   │   ├── exposure.go
│   │   │   ├── simulation.go
│   │   │   └── mitigation.go
│   │   │
│   │   ├── transport/
│   │   │   └── http/
│   │   │       ├── router.go
│   │   │       ├── middleware/
│   │   │       │   ├── request_id.go
│   │   │       │   ├── recovery.go
│   │   │       │   └── logging.go
│   │   │       └── handler/
│   │   │           ├── project_handler.go
│   │   │           ├── scan_handler.go
│   │   │           ├── graph_handler.go
│   │   │           ├── finding_handler.go
│   │   │           ├── simulation_handler.go
│   │   │           └── search_handler.go
│   │   │
│   │   ├── service/
│   │   │   ├── project_service.go
│   │   │   ├── scan_service.go
│   │   │   ├── graph_service.go
│   │   │   ├── finding_service.go
│   │   │   ├── simulation_service.go
│   │   │   └── mitigation_service.go
│   │   │
│   │   ├── ingest/
│   │   │   └── github/
│   │   │       ├── client.go
│   │   │       ├── url.go
│   │   │       ├── tree.go
│   │   │       └── manifests.go
│   │   │
│   │   ├── resolver/
│   │   │   ├── resolver.go
│   │   │   └── python/
│   │   │       ├── resolver.go
│   │   │       ├── pyproject.go
│   │   │       ├── requirements.go
│   │   │       └── poetry_lock.go
│   │   │
│   │   ├── vulnerability/
│   │   │   ├── provider.go
│   │   │   ├── osv/
│   │   │   │   ├── client.go
│   │   │   │   └── mapper.go
│   │   │   └── match.go
│   │   │
│   │   ├── graph/
│   │   │   ├── repository.go
│   │   │   ├── model.go
│   │   │   └── neo4j/
│   │   │       ├── repository.go
│   │   │       ├── writer.go
│   │   │       ├── traversal.go
│   │   │       └── paths.go
│   │   │
│   │   ├── analysis/
│   │   │   ├── blast/
│   │   │   │   ├── calculator.go
│   │   │   │   └── normalize.go
│   │   │   ├── centrality/
│   │   │   │   └── calculator.go
│   │   │   ├── risk/
│   │   │   │   ├── calculator.go
│   │   │   │   ├── weights.go
│   │   │   │   └── explanation.go
│   │   │   ├── confidence/
│   │   │   │   └── calculator.go
│   │   │   └── systemic/
│   │   │       └── detector.go
│   │   │
│   │   ├── simulation/
│   │   │   ├── compromise.go
│   │   │   ├── remediation.go
│   │   │   └── result.go
│   │   │
│   │   ├── mitigation/
│   │   │   ├── ranker.go
│   │   │   └── rrp.go
│   │   │
│   │   ├── jobs/
│   │   │   ├── queue.go
│   │   │   ├── worker.go
│   │   │   └── scan_job.go
│   │   │
│   │   ├── persistence/
│   │   │   └── postgres/
│   │   │       ├── db.go
│   │   │       ├── project_repo.go
│   │   │       ├── scan_repo.go
│   │   │       ├── finding_repo.go
│   │   │       └── simulation_repo.go
│   │   │
│   │   └── observability/
│   │       ├── logger.go
│   │       └── metrics.go
│   │
│   ├── migrations/
│   └── tests/
│
├── frontend/
│   ├── package.json
│   ├── src/
│   │   ├── app/
│   │   ├── components/
│   │   │   ├── graph/
│   │   │   ├── findings/
│   │   │   ├── filters/
│   │   │   ├── simulations/
│   │   │   └── layout/
│   │   ├── features/
│   │   │   ├── project/
│   │   │   ├── scan/
│   │   │   ├── graph/
│   │   │   ├── findings/
│   │   │   └── simulation/
│   │   ├── lib/
│   │   │   ├── api/
│   │   │   └── graph/
│   │   └── types/
│   └── tests/
│
├── fixtures/
│   ├── python/
│   └── demo/
│
└── scripts/
    ├── dev.sh
    └── seed-demo.sh
```

Follow this monorepo organization and the layer responsibilities in Section 10. The Python parser files show the planned locations for candidate inputs, not a frozen implementation order; create implementations only for the approved phase. The resolver is implemented in Go under backend/internal/resolver/python/. Backend go.mod/go.sum and frontend package.json describe ChainReaction itself, not analyzed ecosystems. Any later filename changes must preserve these module responsibilities.

---

## 10. Backend Layer Responsibilities

### 10.1 `cmd/api`
Only bootstraps the HTTP API.

Responsibilities:
- load configuration
- initialize Postgres
- initialize Neo4j
- initialize services
- register HTTP routes
- start server
- graceful shutdown

No business logic.

---

### 10.2 `cmd/worker`
Bootstraps asynchronous scan processing.

Responsibilities:
- connect to queue
- receive scan jobs
- run ingestion/resolution/enrichment/analysis pipeline
- persist state
- handle retry/failure

No HTTP logic.

---

### 10.3 `domain`
Pure domain objects and enums.

Examples:
- `ScanStatus`
- `PackageVersion`
- `DependencyEdge`
- `Finding`
- `ExposurePath`
- `BlastRadius`
- `RiskScore`
- `ConfidenceLevel`

Domain package should not depend on HTTP, Neo4j, Postgres, or OSV clients.

---

### 10.4 `transport/http`
Converts HTTP requests to service calls and service results to JSON.

Handlers must not contain risk formulas, graph algorithms, SQL, Cypher, or GitHub logic.

---

### 10.5 `service`
Application orchestration.

Example:
`SimulationService.RunCompromise(packageID)`

It coordinates graph traversal + analysis + persistence but delegates algorithms to focused packages.

---

### 10.6 `ingest/github`
Responsible only for safe repository metadata/file retrieval.

Must:
- validate GitHub URL
- normalize owner/repo/ref
- fetch repository tree
- find supported manifest/lock files
- enforce size/time limits

Must not:
- execute repo code
- run package scripts
- build project

---

### 10.7 `resolver`
Converts ecosystem-specific files into a neutral dependency model.

Shared output:

```go
type ResolvedApplication struct {
    ID           string
    Name         string
    Ecosystem    string
    Packages     []PackageVersion
    Dependencies []DependencyEdge
    Evidence     []Evidence
}
```

The Python Dependency Resolver implements the shared resolver interface. Its input-specific parsers live in backend/internal/resolver/python/ and handle the approved subset of pyproject.toml, requirements.txt, and poetry.lock. Preserve dependency conditions and provenance, and explicitly report unresolved versions or relationships. Input support order and exact confidence rules remain to be frozen.

---

### 10.8 `vulnerability`
Provider abstraction.

```go
type Provider interface {
    Query(ctx context.Context, packages []PackageVersion) ([]VulnerabilityMatch, error)
}
```

Default prototype provider:
- OSV-compatible provider

Provider-specific data must be mapped into ChainReaction's internal vulnerability representation.

---

### 10.9 `graph`
Owns graph persistence and topology queries.

Interface examples:
- write application graph
- get dependency subgraph
- reverse reachable nodes
- get paths application -> package
- calculate/return graph statistics
- search package nodes

Business services should not embed raw Cypher.

---

### 10.10 `analysis/blast`
Calculates:
- downstream package reach
- exposed applications
- path multiplicity
- maximum depth
- normalized blast score
- classification
- explanation components

---

### 10.11 `analysis/risk`
Combines:
- vulnerability severity
- blast radius
- structural criticality
- exploitability evidence

Initial model:

- Vulnerability Severity: 35%
- Blast Radius: 30%
- Structural Criticality: 20%
- Exploitability Evidence: 15%

All components must be returned to the UI.

No black-box single number.

---

### 10.12 `analysis/confidence`
Computes evidence confidence from data provenance.

Example:
- lockfile exact version -> high
- manifest/tool resolution -> medium
- inferred relation -> low

---

### 10.13 `analysis/systemic`
Finds high-criticality packages even when known CVEs = 0.

This produces a different finding category from vulnerability findings.

---

### 10.14 `simulation`
Pure counterfactual analysis.

**Compromise simulation**
- mark selected package as assumed compromised
- traverse reverse dependency graph
- calculate affected packages/apps/paths
- produce exposure graph overlay

**Remediation simulation**
Prototype mode:
- assume selected vulnerability risk is removed/fixed
- preserve current topology unless a re-resolved graph is available
- recompute risk and exposure metrics

The UI must label this accurately.

Future precise mode:
- resolve candidate package version
- rebuild affected dependency subgraph
- recompute with actual topology

---

### 10.15 `mitigation`
Ranks candidate actions.

Inputs:
- baseline risk state
- candidate remediation simulations

Outputs:
- ordered actions
- RRP
- affected apps
- risk before/after
- explanation

---

## 11. Scan Pipeline

### 11.1 API request

`POST /api/v1/projects`

Payload:

```json
{
  "repository_url": "https://github.com/example/project"
}
```

Creates project.

---

### 11.2 Start scan

`POST /api/v1/projects/{project_id}/scans`

API:
1. creates scan row
2. status = `QUEUED`
3. enqueues scan ID
4. returns immediately

---

### 11.3 Worker state machine

```text
QUEUED
  |
  v
FETCHING_REPOSITORY
  |
  v
DISCOVERING_APPLICATIONS
  |
  v
RESOLVING_DEPENDENCIES
  |
  v
BUILDING_GRAPH
  |
  v
ENRICHING_VULNERABILITIES
  |
  v
ANALYZING_TOPOLOGY
  |
  v
CALCULATING_RISK
  |
  v
GENERATING_FINDINGS
  |
  v
COMPLETED
```

Failure:
`FAILED`

Each stage records:
- start/end timestamp
- error code
- human-readable message
- progress percentage

---

## 12. Graph Data Model

### Node labels

#### `Application`
Properties:
- id
- project_id
- scan_id
- name
- path
- ecosystem

#### `PackageVersion`
Properties:
- id
- purl
- ecosystem
- name
- version

#### `Vulnerability`
Properties:
- id
- source
- advisory_id
- cvss
- severity

---

### Relationships

```text
(Application)-[:DEPENDS_ON]->(PackageVersion)

(PackageVersion)-[:DEPENDS_ON]->(PackageVersion)

(PackageVersion)-[:AFFECTED_BY]->(Vulnerability)
```

Dependency relationship properties:
- direct
- evidence_source
- confidence
- manifest_path

---

## 13. Canonical Identity

Package nodes must be deduplicated by canonical version identity.

Preferred key:
`purl`

Fallback:
`ecosystem + normalized_name + version`

Never merge:
- different versions
- different ecosystems
- unresolved version with resolved version without explicit evidence

---

## 14. Blast Radius Algorithm

Given selected package `P`:

1. Reverse-traverse `DEPENDS_ON`.
2. Collect unique reachable `PackageVersion` nodes.
3. Collect unique reachable `Application` nodes.
4. Determine application-to-P dependency paths.
5. Calculate maximum dependency depth.
6. Normalize reach metrics against project graph totals.
7. Calculate score.
8. Return evidence.

Example result:

```json
{
  "package": "log-parser@2.4.1",
  "downstream_packages": 17,
  "applications_exposed": 4,
  "dependency_paths": 23,
  "maximum_depth": 5,
  "score": 88,
  "classification": "VERY_HIGH"
}
```

`88` is a structural project-relative score, not an exploit probability.

---

## 15. Path Counting Safety

Dependency path counts can explode combinatorially.

Rules:
- count only simple paths
- use cycle guards
- hard-cap full path enumeration
- return top N representative paths
- keep aggregate count separately when feasible
- if capped, return `is_capped=true`

UI example:
`1,000+ paths; showing 25 representative paths`

---

## 16. Explainable Risk Model

For every risk score return:

```json
{
  "score": 87,
  "components": {
    "vulnerability_severity": {
      "score": 31,
      "max": 35
    },
    "blast_radius": {
      "score": 28,
      "max": 30
    },
    "structural_criticality": {
      "score": 17,
      "max": 20
    },
    "exploitability_evidence": {
      "score": 11,
      "max": 15
    }
  }
}
```

Also return a deterministic explanation:
- high severity advisory
- reachable from 4 applications
- appears in 23 dependency paths
- high structural centrality
- exploitability evidence present/absent

Do not use an LLM to determine these facts.

---

## 17. Propagation Path Explorer

API:

`GET /api/v1/scans/{scan_id}/packages/{package_id}/paths`

Optional filters:
- application_id
- max_paths
- max_depth

Response:
- representative paths
- total/capped count
- path lengths
- evidence confidence

Frontend:
- select package
- choose application
- highlight one or more paths
- dim unrelated nodes

---

## 18. Simulate Compromise

API:

`POST /api/v1/scans/{scan_id}/simulations/compromise`

Payload:

```json
{
  "package_id": "..."
}
```

Backend:
1. load graph snapshot
2. treat package as assumed compromised
3. reverse traverse dependents
4. classify reached nodes as potentially exposed
5. calculate applications reached
6. collect paths
7. calculate blast radius
8. persist simulation
9. return overlay

The word “compromised” is an assumption for the simulation, not a claim about real exploitation.

---

## 19. Simulate Remediation

API:

`POST /api/v1/scans/{scan_id}/simulations/remediation`

Prototype payload:

```json
{
  "finding_id": "...",
  "mode": "remove_vulnerability_risk"
}
```

Backend:
1. load baseline metrics
2. counterfactually remove selected finding's vulnerability contribution
3. recompute finding/project risk
4. compute RRP
5. persist result

Important:
The prototype must not pretend it knows the exact future dependency topology unless it has actually re-resolved that target version.

---

## 20. Mitigation Priority Engine

For each remediable high-risk finding:

1. run/estimate remediation counterfactual
2. calculate RRP
3. score practicality
4. rank candidates

Prototype ranking can prioritize:
- RRP
- current risk score
- number of applications affected
- availability of known fixed version

Output:

```text
1. Upgrade package-x
   RRP: 72%
   Applications affected: 5

2. Upgrade package-y
   RRP: 34%
   Applications affected: 2
```

---

## 21. Finding Detail Contract

Each finding must expose:

### Identity
- package
- version
- ecosystem

### Risk
- overall score
- severity
- component breakdown

### Topology
- direct/transitive
- depth
- downstream package count
- exposed app count
- path count

### Why prioritized
Deterministic evidence bullets.

### Recommendation
- fixed version if provider supplies one
- otherwise general remediation action

### Confidence
- level
- evidence provenance

### Actions
- show paths
- simulate compromise
- simulate remediation

---

## 22. Search and Graph Modes

### Search
Support:
- package name
- application name
- advisory/CVE ID

### Filters
- severity
- ecosystem
- direct/transitive
- vulnerable only
- application
- risk classification
- confidence
- depth

### Graph modes
- Dependency View
- Vulnerability View
- Blast Radius View
- Application View
- Critical Path View

Backend returns graph data; frontend applies visual mode semantics.

---

## 23. Systemic Dependency Detection

A package can produce a systemic finding without a CVE.

Candidate rule:

If:
- structural criticality >= threshold
- application reach >= threshold

then:
`SYSTEMIC_DEPENDENCY`

Example:

```text
Known vulnerabilities: 0
Applications: 23
Structural criticality: 97/100
```

Message:
`No known vulnerability. High ecosystem importance. Monitor closely.`

This must be visually distinct from an active vulnerability finding.

---

## 24. Confidence Model

Confidence is evidence-derived. For Python inputs, distinguish declared constraints, exact locked versions, and evidenced dependency relationships. Incomplete transitive metadata or unresolved environment conditions must be reported; an exact package version alone does not establish a complete graph. The precise Python confidence rules remain to be frozen.

### Dependency relation confidence
HIGH:
- exact lockfile edge

MEDIUM:
- resolved from manifest/tool graph

LOW:
- incomplete/inferred edge

### Vulnerability match confidence
HIGH:
- exact ecosystem + name + version / purl match

MEDIUM:
- normalized version/range match with incomplete metadata

LOW:
- ambiguous identity

Overall finding confidence should be the minimum or weighted combination of the evidence confidence supporting it.

---

## 25. PostgreSQL Domain Model

Suggested tables:

### projects
- id
- name
- repository_url
- created_at

### scans
- id
- project_id
- commit_sha
- status
- progress
- started_at
- completed_at
- error_code
- error_message

### applications
- id
- scan_id
- name
- path
- ecosystem

### package_snapshots
- id
- scan_id
- purl
- name
- version
- ecosystem
- provenance
- confidence

### vulnerability_snapshots
- id
- scan_id
- advisory_id
- source
- severity
- cvss
- raw_metadata

### findings
- id
- scan_id
- package_snapshot_id
- vulnerability_snapshot_id nullable
- type
- risk_score
- blast_score
- structural_score
- confidence
- explanation_json

### simulations
- id
- scan_id
- type
- target_id
- baseline_json
- result_json
- created_at

### mitigation_candidates
- id
- scan_id
- finding_id
- action_type
- target_version nullable
- rrp
- rank

Graph topology itself stays in Neo4j.

---

## 26. API Surface — Prototype

```text
POST   /api/v1/projects
GET    /api/v1/projects/:id

POST   /api/v1/projects/:id/scans
GET    /api/v1/scans/:id
GET    /api/v1/scans/:id/status

GET    /api/v1/scans/:id/graph
GET    /api/v1/scans/:id/packages
GET    /api/v1/scans/:id/packages/:package_id
GET    /api/v1/scans/:id/packages/:package_id/paths

GET    /api/v1/scans/:id/findings
GET    /api/v1/scans/:id/findings/:finding_id

POST   /api/v1/scans/:id/simulations/compromise
POST   /api/v1/scans/:id/simulations/remediation
GET    /api/v1/scans/:id/simulations/:simulation_id

GET    /api/v1/scans/:id/mitigations

GET    /api/v1/scans/:id/search?q=
```

API should be documented through OpenAPI before or alongside implementation.

---

## 27. Frontend Product Architecture

The frontend is a security analysis workspace, not a marketing page pretending to be a product.

Recommended primary layout:

```text
+----------------+--------------------------------+----------------------+
| PROJECT /      |                                | SELECTED PACKAGE /   |
| FILTERS        |       DEPENDENCY GRAPH         | FINDING DETAILS      |
|                |                                |                      |
| applications   |                                | risk                 |
| severity       |                                | blast radius         |
| ecosystem      |                                | why prioritized      |
| direct/trans.  |                                | paths                |
| confidence     |                                | simulation actions   |
+----------------+--------------------------------+----------------------+
```

### Frontend design rules
- no chatbot-first interface
- no giant gradient hero
- no random glassmorphism
- no excessive rounded cards
- no decorative “AI sparkles”
- no meaningless KPI card wall
- use tables, lists, graph, inspector panels
- use risk colors only for semantic state
- neutral base palette
- dense but readable security-tool layout
- purposeful animation only for graph traversal/simulation
- typography should prioritize information hierarchy
- show evidence close to the score it explains
- no generated marketing prose inside the analysis workspace

### Preferred graph library
Cytoscape.js is a strong candidate because this is a real graph exploration problem, not a flowchart editor.

---

## 28. Frontend Core Screens

### 28.1 Add Repository
Single serious input workflow:
- GitHub URL
- validate
- analyze

### 28.2 Scan Progress
Show actual pipeline stages, not a fake spinner.

### 28.3 Graph Workspace
Primary product screen.

### 28.4 Finding Detail
Right-side inspector or dedicated route.

### 28.5 Simulation Result
Overlay on graph plus before/after metrics.

### 28.6 Mitigation Priorities
Ranked list/table with RRP and explanation.

---

## 29. Asynchronous Processing

Repository scans should not block an HTTP request.

Use a worker queue.

Prototype queue options:
- PostgreSQL-backed job table
- Redis queue

Recommended for prototype simplicity:
PostgreSQL-backed queue or a small explicit worker table unless throughput requires Redis.

Do not add infrastructure merely for appearance.

---

## 30. Repository Safety

Scanning arbitrary public repositories is a security boundary.

Rules:
- never execute repository scripts
- never run `pip install` or `poetry install` on scanned projects or their dependencies
- never execute `setup.py` or repository build hooks
- never run scanned project tests (for example, `pytest`)
- never build repository binaries
- do not honor repo-provided shell hooks
- enforce download size limits
- enforce manifest size limits
- enforce resolver timeout
- sanitize archive paths
- prevent zip-slip/path traversal
- restrict resolver filesystem
- use isolated temporary directory
- delete temporary data after scan
- avoid secrets in logs

Python ingestion and resolution remain metadata-only. Any approved metadata tooling must run in a restricted environment and must not execute repository code. These scan-input restrictions do not prohibit normal building and testing of ChainReaction itself.

---

## 31. External Data Failure Strategy

If vulnerability provider is unavailable:
- dependency graph scan may still complete
- mark vulnerability enrichment as degraded
- do not fail all topology analysis
- show provider status

If GitHub rate-limits:
- return explicit retryable error
- preserve scan status

If graph DB is unavailable:
- fail graph-dependent analysis clearly

---

## 32. Caching

Cache:
- GitHub manifest fetches by commit SHA
- vulnerability queries by purl/version
- resolved package metadata where safe

Never silently reuse stale graph data for a new commit.

Every scan must be tied to a commit SHA.

---

## 33. Observability

Structured logs should include:
- request_id
- project_id
- scan_id
- stage
- duration
- provider
- error_code

Metrics:
- scan duration
- resolver duration
- graph node/edge count
- OSV request count
- scan failure rate
- path query latency
- simulation latency

---

## 34. Error Model

Stable API error shape:

```json
{
  "error": {
    "code": "UNSUPPORTED_ECOSYSTEM",
    "message": "No supported dependency manifest was found.",
    "request_id": "..."
  }
}
```

Suggested codes:
- INVALID_REPOSITORY_URL
- REPOSITORY_NOT_FOUND
- REPOSITORY_TOO_LARGE
- UNSUPPORTED_ECOSYSTEM
- DEPENDENCY_RESOLUTION_FAILED
- VULNERABILITY_PROVIDER_UNAVAILABLE
- GRAPH_WRITE_FAILED
- ANALYSIS_FAILED
- SIMULATION_FAILED

---

## 35. Testing Strategy

### Unit tests
- URL normalization
- package identity
- pyproject.toml parsing
- requirements.txt parsing (when supported)
- poetry.lock parsing (when supported)
- Python dependency conditions, unresolved versions, and incomplete graph reporting
- blast score normalization
- risk formula
- confidence rules
- RRP
- path/cycle guards

### Integration tests
- Postgres repositories
- Neo4j graph writes/traversals
- OSV provider adapter with fixtures
- scan state machine

### Fixture graphs
Must include:
1. simple chain
2. diamond graph
3. shared dependency across apps
4. cycle
5. package with CVE but low blast radius
6. lower-CVSS package with high blast radius
7. systemic dependency with no CVE
8. path-count cap case

### End-to-end demo test
Repository URL -> completed scan -> graph -> finding -> compromise simulation -> remediation simulation -> mitigation ranking.

---

## 36. Demo Scenario

Purposefully construct or select a demo graph where:

### Package Y
- CVSS 9.8
- affects 1 application
- 1 path
- low blast radius

### Package X
- CVSS 8.2
- affects 3 applications
- 17 paths
- very high blast radius

Expected ChainReaction ranking:
Package X before Package Y due to ecosystem impact.

Demo sequence:
1. scan repository
2. show dependency graph
3. sort findings
4. select Package X
5. explain risk components
6. show propagation paths
7. Simulate Compromise
8. show affected apps
9. Simulate Remediation
10. show risk before/after
11. show RRP
12. show mitigation ranking

This is the project's “Butterfly Effect” moment.

---

## 37. Explicit Prototype Non-Goals

Do not implement initially:
- 20 ecosystems
- Docker image scanning
- Kubernetes scanning
- runtime exploit detection
- IDE extension
- Slack bot
- chatbot
- auto-generated PRs
- billing
- enterprise SSO
- cloud asset inventory
- arbitrary code execution
- vulnerability exploit generation

These may appear only as roadmap ideas.

---

## 38. Architecture Principles

1. **Graph facts before prose.**
2. **Deterministic evidence before AI interpretation.**
3. **Explain every score.**
4. **Never equate dependency reachability with guaranteed exploitability.**
5. **Version-specific package identity.**
6. **Repository scanning must not execute repo code.**
7. **Async scans, synchronous reads.**
8. **Domain logic independent of HTTP/database clients.**
9. **Graph queries behind interfaces.**
10. **Prototype scope stays narrow and demonstrable.**
11. **Frontend behaves like a security tool, not an AI landing page.**
12. **No implementation begins until architects freeze the first build phase.**

---

## 39. Phase 1 Decisions and Deferred Decisions

The following decisions are approved for Phase 1:

- ChainReaction analyzes Python projects only in Prototype V1. JavaScript/TypeScript/npm and Go/Go Modules are not analyzed ecosystems in this prototype. The Go backend and React/Next.js frontend remain ChainReaction implementation choices, not analyzed ecosystem choices.
- Start with one application at repository root.
- Support a paired root `pyproject.toml` and `poetry.lock` first. Standalone `requirements.txt`, manifest-only resolution, multi-application discovery, and other Python inputs return clear unsupported-input errors.
- Use PostgreSQL for scan/domain records and a PostgreSQL-backed persisted job queue.
- Use Neo4j for scan-scoped dependency topology and Cytoscape.js for the graph workspace.
- Use Docker Compose for local development and provide no authentication in this local prototype phase.
- The scan is metadata-only, pinned to a fetched GitHub commit SHA, and must never execute scanned repository code.
- Edges are stored as `dependent -> dependency`. Directness, markers, extras, groups, sources, provenance, and evidence limitations are preserved where supplied.
- Phase 1 does not select a Python/platform/extras environment. Marker-bearing relationships are retained but reported as incomplete rather than claimed active.

The following roadmap decisions remain intentionally deferred:

- exact blast-radius normalization
- exact structural-criticality formula
- exploitability source for prototype
- whether remediation simulation is risk-only first or includes actual version re-resolution
- deployment target

Phase 1 explicitly defers vulnerability enrichment, Blast Radius and structural-criticality formulas, overall Risk Score, compromise/remediation simulation, mitigation ranking, Risk Reduction Potential, systemic dependency detection, extra formats, and multi-application discovery. The product roadmap retains them, but the UI must not display invented scores or simulated functionality.

---

## 40. Builder Boundary

Codex is the implementation builder.

Codex may:
- inspect an existing repo read-only when asked
- explain code
- propose implementation steps
- implement only the explicitly approved phase

Codex must not:
- independently change architecture
- add unapproved dependencies
- expand scope
- create a chatbot
- redesign the product into a generic AI dashboard
- start coding before explicit architect approval
- silently substitute risk formulas
- run arbitrary repository code during scan implementation

If an architectural issue is discovered, Codex should stop and ask the architects for a decision.

---

## 41. Product North Star

**From vulnerability severity to ecosystem impact.**

ChainReaction should make this question easy to answer:

> “If I can fix only one thing first, which dependency should I fix, why, and how much downstream risk will that remove?”
