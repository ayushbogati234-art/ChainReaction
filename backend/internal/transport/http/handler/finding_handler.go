package handler

import (
	"encoding/json"
	"errors"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
)

func RegisterFindings(r chi.Router, s service.FindingService) {
	reply := func(w http.ResponseWriter, data any, err error) {
		if err != nil {
			status, code, message := 500, "INTERNAL_ERROR", "finding request failed"
			if errors.Is(err, domain.ErrNotFound) {
				status, code, message = 404, "NOT_FOUND", "resource not found"
			} else if strings.HasPrefix(err.Error(), "SCAN_NOT_READY:") {
				status, code, message = 409, "SCAN_NOT_READY", err.Error()
			}
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": code, "message": message}})
			return
		}
		json.NewEncoder(w).Encode(data)
	}
	r.Get("/api/v1/scans/{scan}/findings", func(w http.ResponseWriter, q *http.Request) {
		v, e := s.List(q.Context(), chi.URLParam(q, "scan"))
		reply(w, v, e)
	})
	r.Get("/api/v1/scans/{scan}/findings/{finding}", func(w http.ResponseWriter, q *http.Request) {
		v, e := s.Detail(q.Context(), chi.URLParam(q, "scan"), chi.URLParam(q, "finding"))
		reply(w, v, e)
	})
}
