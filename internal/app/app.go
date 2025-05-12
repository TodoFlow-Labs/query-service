package app

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/todoflow-labs/query-service/internal/config"
	"github.com/todoflow-labs/query-service/internal/todo"
	"github.com/todoflow-labs/shared-dtos/logging"
	"github.com/todoflow-labs/shared-dtos/metrics"
)

// Run loads config, wires everything, and blocks in ListenAndServe (fatal on error).
func Run() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("config.Load failed")
	}

	logger := logging.New(cfg.LogLevel).With().Str("service", "query-service").Logger()
	logger.Info().Msg("starting query-service")

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("db connect failed")
	}
	defer db.Close()

	repo := todo.NewPgRepository(db, &logger)
	svc := todo.NewService(repo, &logger)
	handler := todo.NewHandler(svc, &logger)

	r := chi.NewRouter()
	// Middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(jsonContentType)
	r.Use(authMiddleware)

	r.Get("/todos", handler.List)

	metrics.Init(cfg.MetricsAddr)
	logger.Info().Msgf("metrics listening on %s", cfg.MetricsAddr)

	// Error handlers
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "route not found")
		logger.Warn().Str("path", r.URL.Path).Msg("404 not found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		logger.Warn().Str("path", r.URL.Path).Msg("405 method not allowed")
	})

	logger.Info().Msgf("query-service listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, r); err != nil {
		logger.Fatal().Err(err).Msg("HTTP server failed")
	}
}

// Forces JSON Content-Type for all responses
func jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Writes a structured JSON error
func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			if os.Getenv("ENV") == "development" {
				userID = "test-user"
			} else {
				http.Error(w, "X-User-ID header required", http.StatusUnauthorized)
				return
			}
		}
		user_key := "user_id"
		ctx := context.WithValue(r.Context(), user_key, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
