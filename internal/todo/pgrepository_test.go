// internal/todo/pgrepository_test.go
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

	logger := zerolog.New(nil)
	repo := todo.NewPgRepository(mockDB, &logger)

	now := time.Now()
	mockDB.ExpectQuery("SELECT id, title, completed, created_at").
		WithArgs(3).
		WillReturnRows(pgxmock.NewRows([]string{"id", "title", "completed", "created_at"}).
			AddRow("1", "Task A", false, now).
			AddRow("2", "Task B", true, now),
		)

	results, err := repo.SearchTodos(context.Background(), "", 3)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "Task A", results[0].Title)

	assert.NoError(t, mockDB.ExpectationsWereMet())
}
