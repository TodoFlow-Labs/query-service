package todo_test

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/todoflow-labs/query-service/internal/todo"
)

func TestPgRepository_SearchTodos_NoQuery(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	repo := todo.NewPgRepository(mockDB, &logger)

	now := time.Now()
	dueDate := now.Add(24 * time.Hour)
	priority := 1
	tags := []string{"home", "urgent"}

	// Set expectation matching full query
	mockDB.ExpectQuery("SELECT id, user_id, title, description, completed,").
		WithArgs("test-user-1", 3, 0).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "title", "description", "completed",
			"created_at", "updated_at", "due_date", "priority", "tags",
		}).
			AddRow("1", "test-user-1", "Task A", "desc A", false, now, now, &dueDate, &priority, tags).
			AddRow("2", "test-user-1", "Task B", "desc B", true, now, now, nil, nil, []string{}),
		)

	// Inject user_id into context
	ctx := context.WithValue(context.Background(), "user_id", "test-user-1")
	results, err := repo.SearchTodos(ctx, "", 3, 0)

	assert.NoError(t, err)
	assert.Len(t, results, 2)

	assert.Equal(t, "1", results[0].ID)
	assert.Equal(t, "Task A", results[0].Title)
	assert.Equal(t, "test-user-1", results[0].UserID)
	assert.Equal(t, tags, results[0].Tags)

	assert.NoError(t, mockDB.ExpectationsWereMet())
}

func TestPgRepository_FindByID(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockDB.Close()

	logger := zerolog.New(zerolog.NewTestWriter(t))
	repo := todo.NewPgRepository(mockDB, &logger)

	now := time.Now()
	dueDate := now.Add(48 * time.Hour)
	priority := 2
	tags := []string{"work", "go"}

	// Expect the query with correct args
	mockDB.ExpectQuery("SELECT id, user_id, title, description, completed,").
		WithArgs("abc123", "user-42").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "title", "description", "completed",
			"created_at", "updated_at", "due_date", "priority", "tags",
		}).AddRow(
			"abc123", "user-42", "My Todo", "Detail", false,
			now, now, &dueDate, &priority, tags,
		))

	ctx := context.WithValue(context.Background(), "user_id", "user-42")
	result, err := repo.FindByID(ctx, "abc123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.ID)
	assert.Equal(t, "My Todo", result.Title)
	assert.Equal(t, "user-42", result.UserID)
	assert.Equal(t, tags, result.Tags)

	assert.NoError(t, mockDB.ExpectationsWereMet())
}
