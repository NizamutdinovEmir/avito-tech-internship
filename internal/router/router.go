package router

import (
	"database/sql"
	"net/http"
	"time"

	"avito-tech-internship/internal/handler"
	"avito-tech-internship/internal/repository/postgres"
	"avito-tech-internship/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SetupRouter creates and configures the HTTP router with all routes
func SetupRouter(db *sql.DB) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Swagger UI
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	r.Get("/swagger/", handler.ServeSwaggerUI)
	r.HandleFunc("/swagger/*", handler.ServeSwaggerUI)

	// Initialize repositories
	teamRepo := postgres.NewTeamRepository(db)
	userRepo := postgres.NewUserRepository(db)
	prRepo := postgres.NewPullRequestRepository(db)

	// Initialize services
	teamService := service.NewTeamService(teamRepo)
	userService := service.NewUserService(userRepo)
	prService := service.NewPullRequestService(prRepo, userRepo, teamRepo)
	bulkDeactivateService := service.NewBulkDeactivateService(userRepo, prRepo, teamRepo)

	// Initialize handlers
	teamHandler := handler.NewTeamHandler(teamService)
	userHandler := handler.NewUserHandler(userService, prService)
	prHandler := handler.NewPullRequestHandler(prService)
	statsHandler := handler.NewStatsHandler(prService)
	bulkDeactivateHandler := handler.NewBulkDeactivateHandler(bulkDeactivateService)

	// API routes
	r.Route("/team", func(r chi.Router) {
		r.Post("/add", teamHandler.CreateTeam)
		r.Get("/get", teamHandler.GetTeam)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", userHandler.SetIsActive)
		r.Get("/getReview", userHandler.GetReview)
		r.Post("/bulkDeactivate", bulkDeactivateHandler.BulkDeactivate)
	})

	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", prHandler.CreatePR)
		r.Post("/merge", prHandler.MergePR)
		r.Post("/reassign", prHandler.ReassignReviewer)
	})

	// Statistics endpoint
	r.Get("/stats", statsHandler.GetStats)

	return r
}
