package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	r.Get("/todos", handler.List)

	metrics.Init(cfg.MetricsAddr)
	logger.Info().Msgf("metrics listening on %s", cfg.MetricsAddr)

	logger.Info().Msgf("HTTP listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, r); err != nil {
		logger.Fatal().Err(err).Msg("http.ListenAndServe failed")
	}
}
