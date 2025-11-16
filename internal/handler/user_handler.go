package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"avito-tech-internship/internal/domain"
	"avito-tech-internship/internal/service"
)

type UserHandler struct {
	userService        *service.UserService
	pullRequestService *service.PullRequestService
}

func NewUserHandler(userService *service.UserService, prService *service.PullRequestService) *UserHandler {
	return &UserHandler{
		userService:        userService,
		pullRequestService: prService,
	}
}

// SetIsActive handles POST /users/setIsActive
func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, ErrorCodeNotFound, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, ErrorCodeNotFound, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.SetIsActive(req.UserID, req.IsActive)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]*domain.User{
		"user": user,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}

// GetReview handles GET /users/getReview?user_id=...
func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, ErrorCodeNotFound, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, ErrorCodeNotFound, "user_id parameter is required", http.StatusBadRequest)
		return
	}

	prs, err := h.pullRequestService.GetPRsByReviewer(userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}
