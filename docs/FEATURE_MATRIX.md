# Blueprint feature matrix

Authority: [full blueprint](../ChainReaction-Architecture_Blueprint.md), preserved
intact. Evidence: [P1](ARCHITECTURE.md), [P2](phases/PHASE_2.md),
[P3](phases/PHASE_3.md), [P4](phases/PHASE_4.md). Current work:
[handoff](HANDOFF.md). Status describes the stated row, not its whole section.
`verified` means the cited bounded behavior was tested; `implemented` awaits
verification; `partial` has explicit gaps; `not started` has no implementation.
TBD means no later phase has been authorized. P2 rows are updated only with actual evidence; completed work is bounded by
its documented supported input/matching subset. Paths below are repository-relative;
`internal/` means `backend/internal/`, `src/` means `frontend/src/`.

| Blueprint | Requirement / capability | Status | Implementation locations | API / UI coverage | Evidence | Limitations / decisions | Phase |
|---|---|---|---|---|---|---|---|
|1,2,5,41|Graph-first intelligence; seven pillars and ecosystem-impact positioning|partial|internal/graph; src/components/graph|Dependency workspace|P1|Enrichment/finding evidence verified in P2; topology analytics/prioritization/simulations and business demo unfinished|P1/P2/TBD|
|1,5,38|Deterministic facts, optional AI summaries only; no chatbot|verified|resolver, graph; src/features/scan|No AI-generated scores|P1 code/browser|Future summarization not approved|P1|
|3.1,13|Version-specific Python normalized identity/purl; no version/ecosystem conflation|verified|internal/resolver/python|Package inspection|P1 resolver/isolation tests|Unresolved/custom sources limited|P1|
|3.2,4.4|Multiple application roots and shared applications|partial|internal/domain/exposure.go; analysis/blast|Impact/path APIs and UI accept multiple roots|P3 synthetic two-app unit/browser tests|Repository ingestion still discovers one root|P3/TBD|
|3.3–3.7,4.1|Direct/transitive, dependent -> dependency; upstream/downstream terminology|verified|resolver/python; graph/neo4j|Graph, inspector|P1 provenance fixture/browser|Conditional edges not active-installation proof|P1|
|3.8–3.9,4.2|Unique downstream packages and exposed application aggregates|verified|domain/exposure.go; analysis/blast/calculator.go|Impact API, package/finding inspector and graph region|P3 synthetic/real DB/saved scan/browser|Exact within isolated snapshot; discovery may be incomplete|P3|
|3.10,4.3|Complete representative paths; multiple branches|verified|analysis/blast/calculator.go; service/graph_service.go|Filtered path API and route highlighting|P3 diamond/shared-app/DB/browser|Displayed paths bounded independently from aggregate|P1/P3|
|3.11|Maximum dependency depth as context, not danger score|verified|analysis/blast/calculator.go|Exact/bounded labels in API/UI|P3 unit/browser|Bounded maximum is explicitly observed-only|P3|
|3.12–3.14,10.10,14|Blast reach, normalization, score, configurable weights and bands|verified|analysis/blast/{calculator,normalize}.go; phases/PHASE_3.md|Impact API and inspector score/components/totals/availability|P3 model/unit/real DB/browser verification|Approved formula implemented; null scores for incomplete/capped/work-limited evidence|P3|
|3.15,10.9|Structural criticality: fan-in, app reach, concentration, distinct simple paths|implemented|internal/analysis/criticality; service/graph_service.go|Impact API and inspector|P4 unit/API/UI verification|Approved transparent weights; betweenness intentionally not in this prototype formula|P4|
|3.16,38|Potential dependency exposure, no exploitability claims|verified|src/components/findings; features/scan|Evidence language/caveats|P1/P2 browser|No installation/exploitability claim|P1/P2|
|3.17,10.13,23|Systemic Dependency findings without known CVEs|not started|—|—|—|Criticality/reach thresholds and complete coverage needed|TBD|
|3.18,10.15,20|Mitigation Priority: RRP/risk/apps/practicality/fix availability ranking|not started|—|—|—|Ranking and practicality model pending|TBD|
|3.19,19|Risk Reduction Potential counterfactual and zero-baseline policy|not started|—|—|—|Aggregate risk and remediation model pending|TBD|
|3.20,10.12,24|Separate dependency and vulnerability confidence, overall evidence combination|partial|domain/model.go; resolver/python|Dependency confidence|P1|P2 separate exact/range evidence and conservative minimum verified; broader inference rules pending|P1/P2/TBD|
|4.5,15|Cycle guards, simple paths, bounded enumeration|verified|analysis/blast/calculator.go|max_paths/max_depth, 1000+ count semantics|P3 cap/cycle unit + browser|Graph size 10k nodes/100k edges; performance load gap|P1/P3|
|4.5,15|Cycle detection display, SCC collapse, aggregate capped counts|partial|analysis/blast; components/graph/ImpactEvidence.tsx|Cycle/cap notices|P3 unit/browser|SCC collapse remains optional/unimplemented|P3/TBD|
|5,6.1,28.1|Public GitHub URL input; Python-only correction|verified|ingest/github; service/project_service.go|Submit/error UI|P1 live/fixture|No JS/Go analyzed ecosystems|P1|
|6.2,10.7,39|Paired root pyproject.toml + poetry.lock; identity/declarations vs lock evidence|verified|resolver/python/{resolver,pyproject,poetry_lock}.go|Input evidence|P1 unchanged-lock fixture|Poetry subset, no PEP 621 dependencies|P1|
|6.2|requirements.txt and additional metadata inputs|not started|—|Explicit unsupported errors|P1 negative tests|No expansion incl. tox.ini authorized|TBD|
|6.2,24|Markers/extras/groups/optional/source provenance and uncertainty|verified|resolver/python; src/features/scan/InputEvidence|Limitations and group/constraint evidence|P1|No environment selection, full ranges or freshness validation|P1|
|6.3,30|No install/build/hooks/scripts/tests of scanned repositories|verified|ingest/github; resolver/python|Metadata-only disclosure|P1 fixtures/code trace|Tests/builds of ChainReaction itself allowed|P1/P2|
|7,8,29|API/worker/Postgres queue/Neo4j separation and Docker Compose|verified|internal/app; jobs; persistence; graph|Async scans, sync reads|P1 real DB workflow|No production auth/deployment|P1|
|9,10.1–10.7,10.9–10.10|Prescribed modules, startup-only wiring, pure domain, delegating handlers, interfaces|verified|cmd; app; service; transport; domain; graph; analysis/blast|Existing endpoints|P1 boundary audit + P3 mapping|Accepted consolidated files, no deferred stubs|P1/P2/P3|
|10.8|Vulnerability provider abstraction and OSV adapter|verified|internal/vulnerability/provider.go; osv/{client,mapper}.go|Worker + findings|P2 provider fixtures/live OSV|Exact numeric Python releases/default PyPI only|P2|
|10.11,16|Explainable Risk Score, severity/blast/criticality/exploitability components|not started|—|—|—|Phase 5: formula, weights and exploitability source pending|P5|
|10.14,18|Simulate Compromise assumption, reverse exposure, overlay, persistence|not started|—|—|—|No simulation authorization|TBD|
|10.14,19|Simulate Remediation baseline/counterfactual/re-resolution distinction|not started|—|—|—|Risk-only vs real re-resolution decision pending|TBD|
|11.1–11.2|Project creation, persisted queued scan, immediate response|verified|service; persistence/postgres; jobs|POST projects/scans|P1 separate DB observer|Queue retries not automatic|P1|
|11.3,28.2|Actual stages/progress and terminal failure|partial|jobs; service/scan_service; domain/scan|Polling UI|P1|No per-stage timestamp history; unapproved analysis stages absent|P1/P2/TBD|
|12|Application/package nodes and provenance-bearing DEPENDS_ON|verified|graph/neo4j|Topology and inspector|P1|Application project_id and package ecosystem properties absent; manifest path implicit|P1|
|12|Vulnerability nodes and scan-scoped AFFECTED_BY relationships|verified|internal/graph/neo4j/writer.go|Finding-linked AFFECTED_BY; dependency view preserved|P2 real DB/graph regression|One scan-scoped vulnerability snapshot per finding|P2|
|17|Propagation Path Explorer package select, filters, bounded routes, lengths, counts, confidence and dimming|verified|analysis/blast; service/graph_service; components/graph; lib/graph|Paths/impact APIs and inspector|P3 synthetic/DB/live browser|Real ingestion has one app; representative ordering is deterministic ID order|P1/P3|
|21 Identity|Finding package/version/ecosystem/advisory identity|verified|internal/domain/finding.go; service/finding_service.go|Findings list/detail|P2 DB/browser|Exact validated supported identities|P2|
|21 Risk|Provider severity/CVSS vs overall risk/components|partial|internal/vulnerability/osv/mapper.go; components/findings|Supplied severity/CVSS only|P2 missing-severity fixture/live|Overall risk/components deferred|P2/TBD|
|21 Topology|Finding direct/transitive, paths, depth/downstream/exposed/path counts|verified|analysis/blast; components/findings/FindingDetail.tsx; components/graph/ImpactEvidence.tsx|Finding/package impact evidence|P3 synthetic/live browser|Snapshot limitations remain visible; no prioritization score|P2/P3|
|21 Why prioritized|Deterministic matching evidence; prioritization explanation|partial|vulnerability/match.go; components/findings|Deterministic match explanations|P2 fixture/browser|Findings are not priority-ranked|P2/TBD|
|21 Recommendation|Provider fixed versions or unavailable, never guessed latest fix|verified|vulnerability/osv/mapper.go; components/findings|Applicable supplied fixed versions or explicit unavailable|P2 fixtures/live browser|Numeric affected interval fixes only; no compatibility guarantee|P2|
|21 Confidence/actions|Evidence provenance, show paths, simulations|partial|analysis/confidence/calculator.go; components/findings|Separate confidence and paths|P2 fixture/DB/browser|Simulation actions absent|P2/TBD|
|22 Search|Package/application/advisory/CVE search endpoint and UI|not started|—|—|—|Not implicitly removed|TBD|
|22 Filters|Severity/ecosystem/directness/vulnerable/app/risk/confidence/depth|not started|—|—|—|No full filtering contract yet|TBD|
|22 Graph modes|Dependency View|partial|src/components/graph|Current graph|P1|No mode switcher|P1|
|22 Graph modes|Vulnerability/Blast Radius/Application/Critical Path Views|not started|—|—|—|Findings-to-path connection is not all graph modes|TBD|
|25 projects/scans|Transactional projects, scan lifecycle and evidence limitations|verified|persistence/postgres/db.go|Projects/scans/status|P1 real DB|No distinct repository table|P1|
|25 applications|Transactional application snapshots|not started|—|Graph stores application|P1 graph only|Suggested table not yet implemented|TBD|
|25 package_snapshots|Scan package identity/provenance snapshots for findings|verified|persistence/postgres/finding_repo.go; migrations/002_findings.sql|Snapshot-backed findings|P2 real DB|Topology remains Neo4j|P2|
|25 vulnerability_snapshots/findings|Raw advisory snapshots, evidence, alias dedup, repeat safety|verified|persistence/postgres/finding_repo.go; migrations/002_findings.sql|Scan-scoped findings with raw source records|P2 aliases/repeat/isolation tests|Risk/blast/structural scores absent|P2|
|25 simulations/mitigation_candidates|Counterfactuals and ranked actions persistence|not started|—|—|—|Unapproved future phase|TBD|
|26|POST/GET projects, POST scans, GET scan/status|verified|transport/http/router.go|Existing API/UI|P1|Local unauthenticated prototype|P1|
|26|GET graph, package detail, package paths and package impact|verified|service/graph_service; graph/neo4j; transport/http/router.go|Existing API/UI|P3 real DB/API/browser|Impact is calculated on read from isolated snapshot|P1/P3|
|26|GET package list endpoint|not started|—|UI currently uses graph node list|—|Separate API contract pending|TBD|
|26|GET findings list and detail endpoints|verified|transport/http/handler/finding_handler.go; service/finding_service.go|List/detail + coverage|P2 DB/browser|Complete report; pagination not implemented|P2|
|26|POST compromise/remediation; GET simulation/mitigations/search|not started|—|—|—|Unapproved later phases|TBD|
|26|OpenAPI alongside implementation|implemented|docs/openapi.json; docs/API.md|Implemented P1–P3 routes and evidence schemas|P3 JSON/reference validation|Full contract conformance tooling not installed|P2/P3|
|27|Dense neutral security workspace, semantic colors, graph/inspector, no decorative AI UI|verified|src/app; components/graph|Responsive workspace|P3 390px Chromium|Future panels must retain usability|P1/P2/P3|
|28.3|Pannable graph, readable labels, root distinction, stable selection and controls|verified|components/graph; lib/graph|Fit/zoom/reset/path focus|P1 48-node browser|Large-graph load testing not done|P1|
|28.4|Finding detail inspector/page|partial|features/findings/FindingsPanel.tsx; components/findings/FindingDetail.tsx|Finding detail inspector|P2 live/fixture browser|Full §21 risk/topology/simulation contract deferred|P2/TBD|
|28.5–28.6|Simulation result and mitigation priorities screens|not started|—|—|—|No mock working actions|TBD|
|29|Persisted worker claiming, crash lease and failures|verified|jobs; postgres/db.go|Status/errors|P1|Automatic retry policy remains pending|P1|
|30|URL validation, pinned archive limits, root-only safe paths, bounded execution|verified|ingest/github; jobs/scan_job|Clear failures|P1 fixtures/live|Memory-only extraction: no temp repository filesystem|P1|
|30|Avoid secrets in logs; isolated metadata tooling|partial|ingest/github; config; app|No repository execution|P1 code audit|No extra tooling; comprehensive log audit pending|P1/P2|
|31|Provider outages preserve topology; per-package queried/unqueried/partial status|verified|vulnerability/osv; service/finding_service; postgres/finding_repo|COMPLETE/PARTIAL/DEGRADED/NOT_RUN + per-package states|P2 DB failure fixtures/browser|Retry requires new scan; live outage not forced|P2|
|31|GitHub failures explicit; rate-limit retryability|partial|ingest/github|Failed scan error|P1|Specific retry-after semantics not implemented|P1/TBD|
|31|Graph DB failure clearly fails graph analysis|partial|service; jobs|Failed scan|P1 injected errors|Forced live outage not verified|P1|
|32|Commit-keyed GitHub manifest cache|not started|—|—|—|Never reuse graph across commits|TBD|
|32|purl/version vulnerability cache with freshness|not started|—|—|—|Snapshots are evidence, not a reusable cache|TBD|
|32|Safe resolved metadata cache|not started|—|—|—|Cache invalidation policy pending|TBD|
|33|Structured request/project/scan/stage/duration/provider/error logs|partial|jobs/worker.go|Worker error logging|P1|Most structured fields missing|TBD|
|33|Scan/resolver/graph/OSV/failure/path/simulation metrics|not started|—|—|—|Instrumentation not implemented|TBD|
|34|Stable error code/message envelope, request_id|partial|transport/http/router.go; jobs|Errors shown|P1|request_id missing; future codes deferred|P1/P2/TBD|
|35 Unit|URL/identity/Poetry/conditions/incomplete/path/cycle tests|verified|resolver tests; backend/tests; graph tests|—|P1|Requirements tests await support|P1|
|35 Unit|Provider matching, alias dedup, missing evidence, timeout/rate-limit/partial failures|verified|vulnerability/osv/client_test.go; vulnerability/match_test.go; service/finding_service_test.go|—|P2 deterministic provider suite|Live results separately recorded in P2|P2|
|35 Unit|Blast/risk/confidence/RRP formulas|partial|analysis/blast/calculator_test.go; analysis/confidence|Impact components/confidence only|P3 graph-shape/cap tests|Blast formula verified; overall risk/RRP formulas deferred|P2/P3/TBD|
|35 Integration|Postgres, Neo4j, scan state machine|verified|backend/tests/integration_test.go|Real DB API workflow|P1|GitHub mocked in controlled tests|P1|
|35 Fixtures|Chain, diamond, shared-app, cycle, disconnected, versions, path-cap/incomplete shapes|verified|analysis/blast/calculator_test.go; backend/tests; frontend/tests/phase3-impact.cjs|Impact API/UI|P3 deterministic + real DB|Shared apps synthetic only|P1/P3|
|35 Fixtures|CVE low blast, lower CVSS high blast, systemic no CVE|not started|—|—|—|Scoring/application discovery not approved|TBD|
|35 E2E|URL -> graph -> finding -> compromise -> remediation -> ranking|partial|frontend/tests|P1 through graph/paths|P1|P2 live URL through finding/path workflow verified; simulations/ranking deferred|P1/P2/TBD|
|36|Controlled X/Y Butterfly Effect demo; sorting, explain risk, simulations/RRP/ranking|not started|—|—|—|Do not mark whole demo complete after P2|TBD|
|37|Explicit non-goals: broad ecosystems, images/K8s/runtime/IDE/bots/chat/PRs/billing/SSO/cloud/exploits|verified|Scope boundaries|Absent|P1 code/scope review|Retained as non-goals, not missing implemented features|P1/P2|
|38|Graph facts, explained scores, safe scans, async reads, pure domain/interfaces|partial|Existing layers|P1 workflow|P1|Deferred engines not implemented|P1/P2/TBD|
|39|Approved root Python subset, Postgres queue, Neo4j/Cytoscape, Docker no-auth prototype|verified|Existing project|P1 workflow|P1|All later product features remain in specification|P1|
|39|Blast normalization, criticality formula, exploitability source, remediation/deployment decisions|partial|analysis/criticality|Impact API/UI component evidence|P4|Blast and criticality formulas approved; remaining decisions still pending|P3/P4|
|40|Explicit phase authorization; no silent architecture/scope substitutions|implemented|HANDOFF; phases docs; this matrix|—|User authorized P2/P3/P4|Blast and criticality formulas approved; later phases require authorization|P2/P3/P4|
|41|Final project completion via full blueprint and demo cross-check|not started|This tracking document|—|—|Whole product is not complete|TBD|
