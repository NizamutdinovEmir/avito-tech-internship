package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"avito-tech-internship/internal/service"
)

// ErrorCode represents error codes from OpenAPI spec
type ErrorCode string

const (
	ErrorCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists    ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged    ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound    ErrorCode = "NOT_FOUND"
)

// ErrorResponse represents error response structure
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// writeError writes error response with appropriate status code
func writeError(w http.ResponseWriter, code ErrorCode, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ErrorResponse{}
	resp.Error.Code = string(code)
	resp.Error.Message = message

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Failed to encode error response", "error", err)
	}
}

// handleServiceError converts service errors to HTTP responses
func handleServiceError(w http.ResponseWriter, err error) {
	switch err {
	case service.ErrTeamExists:
		writeError(w, ErrorCodeTeamExists, "team_name already exists", http.StatusBadRequest)
	case service.ErrTeamNotFound:
		writeError(w, ErrorCodeNotFound, "team not found", http.StatusNotFound)
	case service.ErrUserNotFound:
		writeError(w, ErrorCodeNotFound, "user not found", http.StatusNotFound)
	case service.ErrPRNotFound:
		writeError(w, ErrorCodeNotFound, "PR not found", http.StatusNotFound)
	case service.ErrPRExists:
		writeError(w, ErrorCodePRExists, "PR id already exists", http.StatusConflict)
	case service.ErrPRMerged:
		writeError(w, ErrorCodePRMerged, "cannot reassign on merged PR", http.StatusConflict)
	case service.ErrNotAssigned:
		writeError(w, ErrorCodeNotAssigned, "reviewer is not assigned to this PR", http.StatusConflict)
	case service.ErrNoCandidate:
		writeError(w, ErrorCodeNoCandidate, "no active replacement candidate in team", http.StatusConflict)
	case service.ErrAuthorNotFound:
		writeError(w, ErrorCodeNotFound, "author/team not found", http.StatusNotFound)
	default:
		slog.Error("Unhandled service error", "error", err)
		writeError(w, ErrorCodeNotFound, "internal server error", http.StatusInternalServerError)
	}
}
