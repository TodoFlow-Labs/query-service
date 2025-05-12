package todo

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/todoflow-labs/shared-dtos/dto"
)

// Handler deals with HTTP.
type Handler struct {
	svc    Service
	logger *zerolog.Logger
}

// NewHandler wires the todo endpoints.
func NewHandler(svc Service, logger *zerolog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// List responds to GET /todos?q=...
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug().Msg("handling /todos request")

	query := r.URL.Query().Get("q")

	// Parse size (default: 100)
	sizeStr := r.URL.Query().Get("s")
	size := 100
	if sizeStr != "" {
		var err error
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size <= 0 {
			h.logger.Error().Err(err).Msg("invalid size parameter")
			http.Error(w, "invalid size parameter", http.StatusBadRequest)
			return
		}
	}

	// Parse page (default: 0)
	pageStr := r.URL.Query().Get("page")
	page := 0
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 0 {
			h.logger.Error().Err(err).Msg("invalid page parameter")
			http.Error(w, "invalid page parameter", http.StatusBadRequest)
			return
		}
	}

	offset := page * size

	h.logger.Debug().
		Str("query", query).
		Int("size", size).
		Int("page", page).
		Int("offset", offset).
		Msg("received /todos paginated request")

	todos, err := h.svc.ListPaginated(r.Context(), query, size, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("ListPaginated failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().Int("count", len(todos)).Msg("returning paginated results")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.ListTodosResponse{Todos: todos})
}
