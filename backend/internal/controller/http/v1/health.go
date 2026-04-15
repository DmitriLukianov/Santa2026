package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type healthHandler struct {
	db *pgxpool.Pool
}

func newHealthHandler(db *pgxpool.Pool) *healthHandler {
	return &healthHandler{db: db}
}

type healthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services,omitempty"`
}

func (h *healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	services := map[string]string{}
	overallOK := true

	// Проверяем БД
	if err := h.db.Ping(ctx); err != nil {
		services["database"] = "unavailable"
		overallOK = false
	} else {
		services["database"] = "ok"
	}

	resp := healthResponse{Services: services}
	if overallOK {
		resp.Status = "ok"
		w.WriteHeader(http.StatusOK)
	} else {
		resp.Status = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
