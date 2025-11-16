package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"avito-tech-internship/internal/domain"
	"avito-tech-internship/internal/service"
)

type PullRequestHandler struct {
	prService *service.PullRequestService
}

func NewPullRequestHandler(prService *service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{prService: prService}
}

// CreatePR handles POST /pullRequest/create
func (h *PullRequestHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, ErrorCodeNotFound, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, ErrorCodeNotFound, "invalid request body", http.StatusBadRequest)
		return
	}

	pr := &domain.PullRequest{
		PullRequestID:   req.PullRequestID,
		PullRequestName: req.PullRequestName,
		AuthorID:        req.AuthorID,
	}

	if err := h.prService.CreatePR(pr); err != nil {
		handleServiceError(w, err)
		return
	}

	// Get created PR to return full data with assigned reviewers
	createdPR, err := h.prService.GetPR(pr.PullRequestID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]*domain.PullRequest{
		"pr": createdPR,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}

// MergePR handles POST /pullRequest/merge
func (h *PullRequestHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, ErrorCodeNotFound, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, ErrorCodeNotFound, "invalid request body", http.StatusBadRequest)
		return
	}

	pr, err := h.prService.MergePR(req.PullRequestID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]*domain.PullRequest{
		"pr": pr,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}

// ReassignReviewer handles POST /pullRequest/reassign
func (h *PullRequestHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, ErrorCodeNotFound, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, ErrorCodeNotFound, "invalid request body", http.StatusBadRequest)
		return
	}

	pr, newUserID, err := h.prService.ReassignReviewer(req.PullRequestID, req.OldUserID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"pr":          pr,
		"replaced_by": newUserID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}
