# Phase 5 — explainable risk and systemic dependencies

Authorized 2026-09-25. The user approved systemic detection at criticality
≥80/100 with complete enrichment and no known advisory, and withholding overall
risk when exploitability evidence is missing. The blueprint remains authoritative.

## Model and evidence

`analysis/risk` calculates package-level prioritization using the blueprint's
35% severity, 30% Blast Radius, 20% structural criticality and 15% exploitability
evidence weights. The API returns the model version, normalized component values,
weights, contributions, source evidence, reasons and deterministic explanation.
The total is rounded once to an integer; unavailable values/contributions are
null and weights are never redistributed. This is not an exploit probability.

Severity is the maximum CVSS **base** score across matched advisories, multiplied
by ten. Every advisory must contain a supported valid CVSS v3.0/v3.1 vector and
have HIGH matching confidence before the total is available. Duplicate aliases
do not add points. Missing severity, provider labels, numeric strings and v2/v4
vectors are not guessed or converted. Other valid temporal/environmental metrics
are validated but do not alter base severity. Implementation follows the
[FIRST equations and rounding](https://www.first.org/cvss/v3.1/specification-document).

The bounded prototype exploitability source is explicit exploit-code maturity
in a valid saved OSV CVSS v3 vector: `E:P`, `E:F` or `E:H` supplies a positive
evidence signal (100 normalized, 15 weighted points). Missing, `E:X` and `E:U`
remain unknown. This binary positive-evidence rule is a prototype product policy,
not CVSS's temporal scoring formula. Generic URLs and OSV `EVIDENCE` references
alone are not scored: the [OSV schema](https://ossf.github.io/osv-schema/) describes
them more broadly as demonstrations of vulnerability validity. No reference is
fetched or executed. A positive signal does not establish exploitation in the
analyzed application. Severity and maturity can originate in different matched
advisories; their source IDs/vectors are returned explicitly.

No matched advisory produces unavailable advisory-based risk, not zero risk.
Available topology and severity components remain inspectable. Coverage must
match every package ID in the topology exactly once, with CHECKED status,
nonzero timestamps and COMPLETE OSV enrichment/graph status. Incomplete dependency
evidence and capped/work-limited normalization retain Phase 3/4 null-score policy.
Application/path display filters do not affect package risk or systemic results.

## Systemic findings

`analysis/systemic` returns MATCH only with available structural criticality
≥80, complete scan enrichment, and no matched advisory for that package. It also
checks persisted findings to avoid a false match if evidence is inconsistent.
Below-threshold/advisory-present packages return NOT_MATCHED; insufficient
evidence returns UNAVAILABLE. These are distinct from vulnerability findings.
The provider check timestamp and threshold accompany the result. Labels say
“no known advisory returned at scan time”; they do not imply safety or freshness
beyond that stored check. The 80 threshold is a product rule, not a standard.

## API and UI

- Existing package `/impact` adds `risk` and `systemic_dependency`.
- `GET /api/v1/scans/{id}/systemic-dependencies` lists matching packages with
  COMPLETE/PARTIAL assessment status and evaluated/total counts. It reads one
  scan snapshot and reuses bounded normalization caches; a five-second request
  budget stops further package evaluation and exposes partial results.
- Unknown scans/packages return 404; pending/failed scans return 409.
- Missing legacy enrichment stays unknown. Enrichment outages withhold scores
  while retaining graph evidence. `/paths` does not depend on enrichment storage.
- The UI lists systemic dependencies separately, selects their package on click,
  and displays risk components, unavailable reasons and check timestamps in the
  inspector. Requests are aborted on scan changes and graph positioning is retained.

Scores/findings are derived on read from immutable scan evidence, not persisted
as new advisory rows. No new external provider, simulation, mitigation ranking,
remediation or later-phase work is included.

## Verification

Verified for the documented prototype subset on 2026-09-25 and 2026-09-28:

- Full Go suite passed, including real PostgreSQL/Neo4j integration inside the
  Compose network (`PHASE1_INTEGRATION=1 go test ./... -count=1`). The synthetic
  complete scan transitioned from unavailable legacy evidence, to a systemic
  finding after persisted coverage, to risk 90/100 after explicit advisory evidence.
- Focused race checks passed for analysis, services and API tests.
- `go vet ./...` passed on resume.
- Frontend TypeScript and production build passed.
- Browser verification passed using isolated headless Edge through CDP:
  `phase5-risk.cjs`, `phase3-impact.cjs`, `finding-selection.cjs`, and
  `phase2-workflow.cjs`. Checked numeric/unknown components, systemic selection,
  incomplete coverage, stable graph positions, narrow UI and no page errors.
- OpenAPI JSON and all 71 internal schema references validated; diff check clean.
- On 2026-09-28, rebuilt and started the local Compose API, worker and frontend.
  Browser smoke check at localhost:3000 confirmed Phase 5 UI and healthy API.
  The saved 47-package scan `35c8adff-730c-43e4-b71d-2d9ba8c6d6f1` returned
  HTTP 200 and PARTIAL systemic assessment with incomplete dependency/enrichment
  reasons. The `astunparse@1.6.3` impact response also returned HTTP 200, all four
  risk components, null overall risk with reasons, UNAVAILABLE systemic status,
  and its valid dependency path. These were read-only checks of saved evidence.

Unit/HTTP fixtures cover the numeric formula, valid/malformed CVSS vectors,
missing evidence, coverage integrity, the 79/80 boundary, known advisories,
filter independence, scan isolation, cancellation and storage outages.

Verification limitations: browser data uses deterministic API fixtures; the
database workflow uses real databases with controlled input/provider fixtures.
No fresh unmocked repository scan or large-graph performance certification was
performed for Phase 5. Normal real scans may withhold risk because dependency
evidence is incomplete or OSV has no explicit exploit-code maturity. CVSS v2/v4,
additional exploitability sources and cross-scan comparisons remain unsupported.

The direct host database test initially hit a PostgreSQL authentication conflict;
running inside the project's Compose network resolved it without changing any
credentials. Chromium download timed out, so browser tests used the installed
Edge with an isolated temporary profile. Browser scripts support optional
`TEST_BROWSER_CDP`, `TEST_FRONTEND_URL` and `PLAYWRIGHT_MODULE` settings.
