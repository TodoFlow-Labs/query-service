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
	query := r.URL.Query().Get("q")
	sizeStr := r.URL.Query().Get("s")

	size := 100
	if sizeStr != "" {
		var err error
		size, err = strconv.Atoi(sizeStr)
		if err != nil {
			h.logger.Error().Err(err).Msg("invalid size parameter")
			http.Error(w, "invalid size parameter", http.StatusBadRequest)
			return
		}
	}

	h.logger.Debug().Str("query", query).Int("size", size).Msg("received /todos request")

	todos, err := h.svc.List(r.Context(), query, size)
	if err != nil {
		h.logger.Error().Err(err).Msg("List failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().Int("count", len(todos)).Msg("returning search results")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.ListTodosResponse{Todos: todos})
}
