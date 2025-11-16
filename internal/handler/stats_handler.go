package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"avito-tech-internship/internal/service"
)

type StatsHandler struct {
	prService *service.PullRequestService
}

func NewStatsHandler(prService *service.PullRequestService) *StatsHandler {
	return &StatsHandler{prService: prService}
}

// GetStats handles GET /stats
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, ErrorCodeNotFound, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.prService.GetStats()
	if err != nil {
		slog.Error("Failed to get stats", "error", err)
		writeError(w, ErrorCodeNotFound, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		slog.Error("Failed to encode stats response", "error", err)
	}
}
