package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"avito-tech-internship/internal/domain"
	"avito-tech-internship/internal/service"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

// CreateTeam handles POST /team/add
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, ErrorCodeNotFound, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var team domain.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		writeError(w, ErrorCodeNotFound, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.teamService.CreateTeam(&team); err != nil {
		handleServiceError(w, err)
		return
	}

	createdTeam, err := h.teamService.GetTeam(team.TeamName)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]*domain.Team{
		"team": createdTeam,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}

// GetTeam handles GET /team/get?team_name=...
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, ErrorCodeNotFound, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, ErrorCodeNotFound, "team_name parameter is required", http.StatusBadRequest)
		return
	}

	team, err := h.teamService.GetTeam(teamName)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(team); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}
