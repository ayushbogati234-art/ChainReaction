# Phase 1–2 API

Machine-readable contract: [OpenAPI 3.1](openapi.json). Phase evidence:
[Phase 1](ARCHITECTURE.md), [Phase 2](phases/PHASE_2.md). Full product coverage
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
| `GET` | `/api/v1/scans/{id}/packages/{package_id}/paths?limit=20` | Return representative simple application-to-package paths. |
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
are provider values, not calculated risk. No overall risk/Blast Radius/structural
score, simulations or ranking endpoints are available.

Errors use `{ "error": { "code": "...", "message": "..." } }`. Relevant codes include `UNSUPPORTED_INPUT`, `INVALID_REPOSITORY_URL`, `GITHUB_DOWNLOAD_ERROR`, and `REPOSITORY_FILE_TOO_LARGE`.

The local browser origin is http://localhost:3000; preflight allows GET, POST and
Content-Type. Repository URLs must be HTTPS github.com owner/repository URLs,
without credentials, query parameters, fragments or branch paths.

Unknown scans/packages return 404 NOT_FOUND. Graph requests before a scan
completes return 409 SCAN_NOT_READY. Package identifiers can be percent-encoded
(for example pytest%407.2.2). Path limits default to 20 and allow 1–100 results.
Paths are simple, potentially conditional metadata relationships, not proof of
installation or exploitability. Empty path results use an empty JSON array.

A COMPLETED scan may have incomplete=true and a nonempty limitations array.
Worker failures persist error_code/error_message on the scan and terminal state
on the job. WORKER_INTERRUPTED indicates an expired worker lease; a new scan can
be submitted. Internal infrastructure errors return a generic 500 response.

See [architecture and closeout evidence](ARCHITECTURE.md) for verified behavior
and the distinction between fixture tests and live browser integration.
