package todo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/todoflow-labs/shared-dtos/dto"
)

// Repo abstracts DB access.
type Repo interface {
	SearchTodos(ctx context.Context, q string, limit int) ([]dto.SearchResult, error)
}

// Service provides business logic.
type Service interface {
	List(ctx context.Context, q string, limit int) ([]dto.SearchResult, error)
}

// ---- service impl ----

type service struct {
	repo   Repo
	logger *zerolog.Logger
}

func NewService(r Repo, logger *zerolog.Logger) Service {
	return &service{repo: r, logger: logger}
}

func (s *service) List(ctx context.Context, q string, limit int) ([]dto.SearchResult, error) {
	s.logger.Debug().Str("query", q).Int("limit", limit).Msg("starting service.List")

	todos, err := s.repo.SearchTodos(ctx, q, limit)
	if err != nil {
		s.logger.Error().Err(err).Msg("SearchTodos failed")
		return nil, err
	}

	s.logger.Debug().Int("results", len(todos)).Msg("service.List completed")
	return todos, nil
}

// ---- PostgreSQL Repo ----

type pgRepository struct {
	db     *pgxpool.Pool
	logger *zerolog.Logger
}

func NewPgRepository(db *pgxpool.Pool, logger *zerolog.Logger) Repo {
	return &pgRepository{db: db, logger: logger}
}

func (r *pgRepository) SearchTodos(ctx context.Context, q string, limit int) ([]dto.SearchResult, error) {
	r.logger.Debug().Str("query", q).Int("limit", limit).Msg("executing full-text search")

	var rows pgx.Rows
	var err error

	if q == "" {
		rows, err = r.db.Query(ctx, `
			SELECT id, title, completed, created_at
			FROM todo
			ORDER BY created_at DESC
			LIMIT $1
		`, limit)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, title, completed, created_at
			FROM todo
			WHERE search_vector @@ plainto_tsquery('simple', $1)
			ORDER BY ts_rank(search_vector, plainto_tsquery('simple', $1)) DESC
			LIMIT $2
		`, q, limit)
	}

	if err != nil {
		r.logger.Error().Err(err).Msg("search query failed")
		return nil, err
	}
	defer rows.Close()

	var out []dto.SearchResult
	for rows.Next() {
		var t dto.SearchResult
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt); err != nil {
			r.logger.Error().Err(err).Msg("row scan failed")
			continue
		}
		out = append(out, t)
	}

	r.logger.Debug().Int("scanned_rows", len(out)).Msg("query completed")
	return out, nil
}
