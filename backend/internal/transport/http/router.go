package httptransport

import (
	"encoding/json"
	"errors"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/service"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func Router(projects service.ProjectService, scans service.ScanService, graph service.GraphService, findings ...service.FindingService) http.Handler {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if q.Method == "OPTIONS" {
				w.WriteHeader(204)
				return
			}
			next.ServeHTTP(w, q)
		})
	})
	r.Get("/healthz", func(w http.ResponseWriter, q *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	r.Post("/api/v1/projects", func(w http.ResponseWriter, q *http.Request) {
		var v struct {
			RepositoryURL string `json:"repository_url"`
		}
		if json.NewDecoder(q.Body).Decode(&v) != nil || v.RepositoryURL == "" {
			fail(w, 400, "INVALID_REQUEST", "repository_url is required")
			return
		}
		p, e := projects.CreateProject(q.Context(), v.RepositoryURL)
		if e != nil {
			notFound(w, e)
			return
		}
		respond(w, 201, p)
	})
	r.Get("/api/v1/projects/{id}", func(w http.ResponseWriter, q *http.Request) {
		p, e := projects.GetProject(q.Context(), chi.URLParam(q, "id"))
		if e != nil {
			notFound(w, e)
			return
		}
		respond(w, 200, p)
	})
	r.Post("/api/v1/projects/{id}/scans", func(w http.ResponseWriter, q *http.Request) {
		x, e := scans.CreateScan(q.Context(), chi.URLParam(q, "id"))
		if e != nil {
			notFound(w, e)
			return
		}
		respond(w, 202, x)
	})
	r.Get("/api/v1/scans/{id}", func(w http.ResponseWriter, q *http.Request) {
		x, e := scans.GetScan(q.Context(), chi.URLParam(q, "id"))
		if e != nil {
			notFound(w, e)
			return
		}
		respond(w, 200, x)
	})
	r.Get("/api/v1/scans/{id}/status", func(w http.ResponseWriter, q *http.Request) {
		x, e := scans.GetScan(q.Context(), chi.URLParam(q, "id"))
		if e != nil {
			notFound(w, e)
			return
		}
		respond(w, 200, map[string]any{"id": x.ID, "commit_sha": x.CommitSHA, "status": x.Status, "stage": x.Stage, "progress": x.Progress, "error_code": x.ErrorCode, "error_message": x.ErrorMessage, "incomplete": x.Incomplete, "limitations": x.Limitations, "enrichment_status": x.EnrichmentStatus})
	})
	r.Get("/api/v1/scans/{id}/graph", func(w http.ResponseWriter, q *http.Request) {
		v, e := graph.GetGraph(q.Context(), chi.URLParam(q, "id"))
		if e != nil {
			notFound(w, e)
			return
		}
		respond(w, 200, v)
	})
	r.Get("/api/v1/scans/{scan}/packages/{package}", func(w http.ResponseWriter, q *http.Request) {
		packageID, err := url.PathUnescape(chi.URLParam(q, "package"))
		if err != nil {
			fail(w, 400, "INVALID_REQUEST", "invalid package identifier")
			return
		}
		v, e := graph.GetPackage(q.Context(), chi.URLParam(q, "scan"), packageID)
		if e != nil {
			notFound(w, e)
			return
		}
		respond(w, 200, v)
	})
	r.Get("/api/v1/scans/{scan}/packages/{package}/paths", func(w http.ResponseWriter, q *http.Request) {
		limit := 20
		if x := q.URL.Query().Get("limit"); x != "" {
			if v, e := strconv.Atoi(x); e == nil && v > 0 && v <= 100 {
				limit = v
			}
		}
		packageID, err := url.PathUnescape(chi.URLParam(q, "package"))
		if err != nil {
			fail(w, 400, "INVALID_REQUEST", "invalid package identifier")
			return
		}
		v, e := graph.GetPaths(q.Context(), chi.URLParam(q, "scan"), packageID, limit)
		if e != nil {
			notFound(w, e)
			return
		}
		respond(w, 200, v)
	})
	if len(findings) > 0 {
		handler.RegisterFindings(r, findings[0])
	}
	return r
}
func respond(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, code, message string) {
	respond(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
func notFound(w http.ResponseWriter, e error) {
	if errors.Is(e, domain.ErrNotFound) {
		fail(w, 404, "NOT_FOUND", "resource not found")
		return
	}
	for _, code := range []string{"INVALID_REQUEST", "INVALID_REPOSITORY_URL", "SCAN_NOT_READY"} {
		if strings.HasPrefix(e.Error(), code+":") {
			status := 400
			if code == "SCAN_NOT_READY" {
				status = 409
			}
			fail(w, status, code, e.Error())
			return
		}
	}
	fail(w, 500, "INTERNAL_ERROR", "request failed")
}
