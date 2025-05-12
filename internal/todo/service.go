package todo

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/todoflow-labs/shared-dtos/dto"
)

// Service provides business logic for listing todos.
type Service interface {
	ListPaginated(ctx context.Context, q string, limit int, offset int) ([]dto.SearchResult, error)
}

type service struct {
	repo   Repo
	logger *zerolog.Logger
}

// NewService creates a new Service instance.
func NewService(r Repo, logger *zerolog.Logger) Service {
	return &service{repo: r, logger: logger}
}

// ListPaginated performs a full-text search with pagination.
func (s *service) ListPaginated(ctx context.Context, q string, limit int, offset int) ([]dto.SearchResult, error) {
	s.logger.Debug().
		Str("query", q).
		Int("limit", limit).
		Int("offset", offset).
		Msg("starting service.ListPaginated")

	todos, err := s.repo.SearchTodos(ctx, q, limit, offset)
	if err != nil {
		s.logger.Error().Err(err).Msg("SearchTodos failed")
		return nil, err
	}

	s.logger.Debug().Int("results", len(todos)).Msg("service.ListPaginated completed")
	return todos, nil
}
