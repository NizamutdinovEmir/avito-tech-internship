package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"avito-tech-internship/internal/service"
)

type BulkDeactivateHandler struct {
	bulkDeactivateService *service.BulkDeactivateService
}

func NewBulkDeactivateHandler(bulkDeactivateService *service.BulkDeactivateService) *BulkDeactivateHandler {
	return &BulkDeactivateHandler{bulkDeactivateService: bulkDeactivateService}
}

// BulkDeactivate handles POST /users/bulkDeactivate
func (h *BulkDeactivateHandler) BulkDeactivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, ErrorCodeNotFound, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TeamName string   `json:"team_name"`
		UserIDs  []string `json:"user_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, ErrorCodeNotFound, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.TeamName == "" {
		writeError(w, ErrorCodeNotFound, "team_name is required", http.StatusBadRequest)
		return
	}

	if len(req.UserIDs) == 0 {
		writeError(w, ErrorCodeNotFound, "user_ids is required", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	if err := h.bulkDeactivateService.BulkDeactivate(req.TeamName, req.UserIDs); err != nil {
		slog.Error("Failed to bulk deactivate users", "error", err, "team", req.TeamName, "users", req.UserIDs)
		handleServiceError(w, err)
		return
	}

	duration := time.Since(startTime)
	slog.Info("Bulk deactivation completed",
		"team", req.TeamName,
		"users_count", len(req.UserIDs),
		"duration_ms", duration.Milliseconds())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"deactivated_users": req.UserIDs,
		"team_name":         req.TeamName,
		"duration_ms":       duration.Milliseconds(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}
