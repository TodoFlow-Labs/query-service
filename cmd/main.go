package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/blevesearch/bleve/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/todoflow-labs/query-service/internal/config"
	"github.com/todoflow-labs/shared-dtos/dto"
	"github.com/todoflow-labs/shared-dtos/logging"
	"github.com/todoflow-labs/shared-dtos/metrics"
)

func main() {
	// 1) Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("config.Load failed")
	}
	logger := logging.New(cfg.LogLevel)
	logger.Info().Msg("query-service starting")

	// 2) Initialize metrics
	metrics.Init(cfg.MetricsAddr)
	logger.Info().Msgf("metrics server listening on %s", cfg.MetricsAddr)

	// 2) Connect to CockroachDB
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("db connect failed")
	}
	defer db.Close()
	logger.Info().Msg("connected to database")

	// 3) Build HTTP router
	r := chi.NewRouter()
	r.Get("/todos", func(w http.ResponseWriter, r *http.Request) {
		// 3a) Run Bleve search to get IDs
		index, err := bleve.Open(cfg.BleveIndexPath)
		if err != nil {
			logger.Error().Err(err).Msg("bleve.Open failed")
			http.Error(w, "search unavailable", http.StatusInternalServerError)
			return
		}

		qs := r.URL.Query().Get("q")
		var req *bleve.SearchRequest
		if qs == "" {
			q := bleve.NewMatchAllQuery()
			req = bleve.NewSearchRequest(q)
		} else {
			q := bleve.NewMatchQuery(qs)
			req = bleve.NewSearchRequest(q)
		}

		req.Size = 100
		res, err := index.Search(req)
		if cerr := index.Close(); cerr != nil {
			logger.Error().Err(cerr).Msg("bleve.Close failed")
		}
		if err != nil {
			logger.Error().Err(err).Msg("bleve.Search failed")
			http.Error(w, "search error", http.StatusInternalServerError)
			return
		}

		// 3b) Collect the IDs
		ids := make([]string, 0, len(res.Hits))
		for _, hit := range res.Hits {
			ids = append(ids, hit.ID)
		}
		if len(ids) == 0 {
			// no matches
			json.NewEncoder(w).Encode(dto.ListTodosResponse{Todos: nil})
			return
		}

		// 3c) Fetch full records from DB
		// use PostgreSQL array operator ANY($1)
		rows, err := db.Query(context.Background(), `
			SELECT id, title, completed, created_at
			FROM todos.todo
			WHERE id = ANY($1)
			ORDER BY created_at DESC
		`, ids)
		if err != nil {
			logger.Error().Err(err).Msg("db query failed")
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// 3d) Scan into DTO
		var resp dto.ListTodosResponse
		for rows.Next() {
			var sr dto.SearchResult
			if err := rows.Scan(&sr.ID, &sr.Title, &sr.Completed, &sr.CreatedAt); err != nil {
				logger.Error().Err(err).Msg("row scan failed")
				continue
			}
			resp.Todos = append(resp.Todos, sr)
		}

		// 3e) Sort results in the same order as Bleve hits
		idToResult := make(map[string]dto.SearchResult, len(resp.Todos))
		for _, t := range resp.Todos {
			idToResult[t.ID] = t
		}
		ordered := make([]dto.SearchResult, 0, len(ids))
		for _, id := range ids {
			if t, ok := idToResult[id]; ok {
				ordered = append(ordered, t)
			}
		}
		resp.Todos = ordered

		// 3f) Return JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// 4) Start HTTP server
	logger.Info().Msgf("about to ListenAndServe on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, r); err != nil {
		logger.Fatal().Err(err).Msg("http.ListenAndServe failed")
	}
}
