# Phase 1–3 API

Machine-readable contract: [OpenAPI 3.1](openapi.json). Phase evidence:
[Phase 1](ARCHITECTURE.md), [Phase 2](phases/PHASE_2.md), and
[Phase 3](phases/PHASE_3.md). Full product coverage
and deferred API endpoints remain in [FEATURE_MATRIX.md](FEATURE_MATRIX.md).

All endpoints use JSON and require no authentication in the local prototype.

| Method | Endpoint | Purpose |
| --- | --- | --- |
| `POST` | `/api/v1/projects` | Create or retrieve a project by `repository_url`. |
| `GET` | `/api/v1/projects/{id}` | Retrieve a project. |
| `POST` | `/api/v1/projects/{id}/scans` | Enqueue a scan and return `202`. |
| `GET` | `/api/v1/scans/{id}` | Retrieve persisted scan record. |
| `GET` | `/api/v1/scans/{id}/status` | Retrieve scan stage, progress, error, and limitations. |
| `GET` | `/api/v1/scans/{id}/graph` | Return scan-scoped nodes and `dependent -> dependency` edges. |
| `GET` | `/api/v1/scans/{id}/packages/{package_id}` | Inspect package evidence and adjacent packages. |
| `GET` | `/api/v1/scans/{id}/packages/{package_id}/paths?max_paths=20&max_depth=100&application_id=...` | Return bounded simple application-to-package paths, lengths, aggregate-count status, maximum-depth status and evidence confidence. |
| `GET` | `/api/v1/scans/{id}/packages/{package_id}/impact` | Return unique reverse-reachable packages/applications and the same bounded path evidence. |
| `GET` | `/api/v1/scans/{id}/findings` | Stored enrichment status, per-package coverage and deduplicated findings. |
| `GET` | `/api/v1/scans/{id}/findings/{finding_id}` | Scan-scoped advisory, matching evidence, references, fixes and confidence. |

Findings reads require a completed scan (409 otherwise); unknown scans or finding
IDs return 404, including IDs from another scan. Older scans return NOT_RUN with
empty evidence lists, not a clean vulnerability result. The local prototype
returns the complete persisted findings report without server-side pagination.

Scan/status responses expose `enrichment_status` separately from `status` and
dependency `incomplete`/`limitations`. COMPLETE enrichment means every retained
package was fully checked by this provider under the supported matching rules;
it does not imply complete dependency discovery, safety or exploitability.
PARTIAL indicates incomplete coverage or vulnerability graph persistence;
DEGRADED means no package coverage succeeded; NOT_RUN covers older/unprocessed
scans. Per-package states are CHECKED, PARTIAL, FAILED, SKIPPED and UNQUERIED,
with reasons and successful-response timestamps where available. Empty CHECKED
results mean no matching advisories returned at scan time only.

Findings expose separate `dependency_confidence`, `vulnerability_match_confidence`
and conservative minimum `confidence`. Each advisory retains source ID, aliases,
raw metadata, affected entries, supplied severity/CVSS, references and applicable
fixed-version evidence. Missing data stays absent/empty. CVSS vectors and labels
are provider values, not calculated risk. Phase 3 impact counts are structural
evidence. The approved scan-relative Blast Radius is returned as `blast_radius`
on impact responses. No overall risk or structural-criticality score,
simulations or ranking endpoints are available.

Errors use `{ "error": { "code": "...", "message": "..." } }`. Relevant codes include `UNSUPPORTED_INPUT`, `INVALID_REPOSITORY_URL`, `GITHUB_DOWNLOAD_ERROR`, and `REPOSITORY_FILE_TOO_LARGE`.

The local browser origin is http://localhost:3000; preflight allows GET, POST and
Content-Type. Repository URLs must be HTTPS github.com owner/repository URLs,
without credentials, query parameters, fragments or branch paths.

Unknown scans/packages return 404 NOT_FOUND. Graph requests before a scan
completes return 409 SCAN_NOT_READY. Package and application identifiers can be
percent-encoded. `max_paths` defaults to 20 and allows 1–100 displayed paths;
the legacy `limit` alias remains accepted. `max_depth` defaults to 100 and allows
1–100 edges. Invalid bounds and application filters return 400.

`total_paths` is exact only when `count_exact=true`. When the aggregate safety
cap of 1,000 is reached, `count_capped=true` and the value is a lower bound
(displayed as `1000+`). If the depth bound is reached, `depth_limited=true`.
`maximum_depth` is exact only when `maximum_depth_exact=true`; otherwise it is the
maximum observed within bounded results. Paths are cycle-safe simple paths.
Reverse reachability counts are unique and exact within the persisted isolated
scan snapshot, excluding the selected package and counting applications separately.
`evidence_complete=false` means discovery limitations can leave real relationships
outside that snapshot. None of these fields prove installation or exploitability.

A COMPLETED scan may have incomplete=true and a nonempty limitations array.
The impact response includes `blast_radius.score` (integer 0–100 or null),
`classification` (string or null), full-scan `total_applications` and
`total_packages`, and component numerators, denominators, normalized values,
weights and point contributions. Path component numerator/denominator are
ln(1 + count). Scores use all applications/packages, independently of path filters
or display limits. The approved weights are 0.50/0.30/0.20; zero denominators give
zero components. Rounded bands are LOW 0–19, MODERATE 20–39, ELEVATED 40–59,
HIGH 60–79 and VERY HIGH 80–100.

Unavailable scores and classifications are null, with `reason_codes`:
`INCOMPLETE_DEPENDENCY_EVIDENCE`, `SCAN_PATH_COUNT_CAPPED`,
`SCAN_PATH_DEPTH_LIMIT`, `WORK_LIMIT`, or `EXECUTION_LIMIT`.
`blocking_package` identifies a package whose capped count/depth prevented exact
scan-wide normalization, even if the selected package itself has few paths.
Components are null when unavailable; valid reachability/path metrics remain.
Analysis allows 200,000 work units and two seconds per traversal, plus a
five-second service context. Scan normalization is reused in a bounded process
cache keyed by scan ID, full topology fingerprint and dependency completeness.
HIGH/VERY HIGH are scan-relative structural classifications, not vulnerability
severity, exploitability or absolute business impact.

Worker failures persist error_code/error_message on the scan and terminal state
on the job. WORKER_INTERRUPTED indicates an expired worker lease; a new scan can
be submitted. Internal infrastructure errors return a generic 500 response.

See [architecture and closeout evidence](ARCHITECTURE.md) for verified behavior
and the distinction between fixture tests and live browser integration.
