package todo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/todoflow-labs/shared-dtos/dto"
)

// Repo abstracts DB access for todos.
type Repo interface {
	SearchTodos(ctx context.Context, q string, limit, offset int) ([]dto.SearchResult, error)
	FindByID(ctx context.Context, id string) (*dto.SearchResult, error)
}

// PGXQueryIface enables mocking or plugging in pgxmock.
type PGXQueryIface interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type pgRepository struct {
	db     PGXQueryIface
	logger *zerolog.Logger
}

// NewPgRepository returns a Repo backed by PostgreSQL.
func NewPgRepository(db PGXQueryIface, logger *zerolog.Logger) Repo {
	return &pgRepository{db: db, logger: logger}
}

func getUserID(ctx context.Context) string {
	uid, ok := ctx.Value("user_id").(string)

	if !ok || uid == "" {
		return ""
	}
	return uid
}

// SearchTodos runs a full-text search or fallback listing, scoped by user_id.
func (r *pgRepository) SearchTodos(ctx context.Context, q string, limit, offset int) ([]dto.SearchResult, error) {
	userID := getUserID(ctx)
	if userID == "" {
		r.logger.Warn().Msg("user_id missing in context")
		return nil, fmt.Errorf("unauthorized")
	}

	r.logger.Debug().
		Str("query", q).
		Int("limit", limit).
		Int("offset", offset).
		Str("user_id", userID).
		Msg("executing paginated search")

	var (
		rows pgx.Rows
		err  error
	)

	if q == "" {
		rows, err = r.db.Query(ctx, `
			SELECT id, user_id, title, description, completed,
			       created_at, updated_at, due_date, priority, tags
			FROM todo
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`, userID, limit, offset)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, user_id, title, description, completed,
			       created_at, updated_at, due_date, priority, tags
			FROM todo
			WHERE user_id = $1
			  AND search_vector @@ plainto_tsquery('simple', $2)
			ORDER BY ts_rank(search_vector, plainto_tsquery('simple', $2)) DESC
			LIMIT $3 OFFSET $4
		`, userID, q, limit, offset)
	}

	if err != nil {
		r.logger.Error().Err(err).Msg("search query failed")
		return nil, err
	}
	defer rows.Close()

	var results []dto.SearchResult

	for rows.Next() {
		var todo dto.SearchResult
		err := rows.Scan(
			&todo.ID,
			&todo.UserID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
			&todo.DueDate,
			&todo.Priority,
			&todo.Tags,
		)
		if err != nil {
			r.logger.Error().Err(err).Msg("row scan failed")
			continue
		}
		results = append(results, todo)
	}

	r.logger.Debug().Int("result_count", len(results)).Msg("search completed")
	return results, nil
}

// FindByID fetches a single todo by ID, scoped to user_id from context.
func (r *pgRepository) FindByID(ctx context.Context, id string) (*dto.SearchResult, error) {
	userID := getUserID(ctx)
	if userID == "" {
		r.logger.Warn().Msg("user_id missing in context")
		return nil, fmt.Errorf("unauthorized")
	}

	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, title, description, completed,
		       created_at, updated_at, due_date, priority, tags
		FROM todo
		WHERE id = $1 AND user_id = $2
	`, id, userID)

	var todo dto.SearchResult
	err := row.Scan(
		&todo.ID,
		&todo.UserID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.DueDate,
		&todo.Priority,
		&todo.Tags,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Info().Str("id", id).Str("user_id", userID).Msg("todo not found")
			return nil, fmt.Errorf("todo not found")
		}
		r.logger.Error().Err(err).Str("id", id).Msg("FindByID query failed")
		return nil, err
	}

	return &todo, nil
}
