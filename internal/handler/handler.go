package handler

import (
	"net/http"
	"sync/atomic"

	"github.com/domifgit/chirpy/internal/database"
)

type Handler struct {
	dbQueries      *database.Queries
	platform       string
	fileserverHits atomic.Int32
}

func New(db *database.Queries, platform string) *Handler {
	return &Handler{
		dbQueries: db,
		platform:  platform,
	}
}

func (h *Handler) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
