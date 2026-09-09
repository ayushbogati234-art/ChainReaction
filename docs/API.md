# Phase 1 API

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
