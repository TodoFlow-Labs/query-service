package todo_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/todoflow-labs/query-service/internal/todo"
	"github.com/todoflow-labs/shared-dtos/dto"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) SearchTodos(ctx context.Context, q string, limit int, offset int) ([]dto.SearchResult, error) {
	args := m.Called(ctx, q, limit, offset)
	return args.Get(0).([]dto.SearchResult), args.Error(1)
}

func (m *mockRepo) FindByID(ctx context.Context, id string) (*dto.SearchResult, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*dto.SearchResult), args.Error(1)
}

func TestService_List(t *testing.T) {
	logger := zerolog.New(nil)
	repo := new(mockRepo)
	svc := todo.NewService(repo, &logger)

	expected := []dto.SearchResult{
		{ID: "1", Title: "Test", Completed: false, CreatedAt: time.Now()},
	}
	repo.On("SearchTodos", mock.Anything, "test", 5, 0).Return(expected, nil)

	results, err := svc.ListPaginated(context.Background(), "test", 5, 0)
	assert.NoError(t, err)
	assert.Equal(t, expected, results)
	repo.AssertExpectations(t)
}

func TestService_GetByID(t *testing.T) {
	logger := zerolog.New(nil)
	repo := new(mockRepo)
	svc := todo.NewService(repo, &logger)

	expected := &dto.SearchResult{
		ID:        "123",
		Title:     "Read docs",
		Completed: false,
		CreatedAt: time.Now(),
	}

	repo.On("FindByID", mock.Anything, "123").Return(expected, nil)

	result, err := svc.GetByID(context.Background(), "123")
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}
