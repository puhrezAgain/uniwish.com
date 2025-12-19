/*
uniwish.com/interal/api/services/scrape_request

scrape_request http service
*/
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"uniwish.com/internal/api/errors"
)

type ScrapeRequester interface {
	Request(ctx context.Context, rawUrl string) (uuid.UUID, error)
}

type CreateScrapeRequestHandler struct {
	service ScrapeRequester
}

func NewCreateItemHandler(srv ScrapeRequester) *CreateScrapeRequestHandler {
	return &CreateScrapeRequestHandler{service: srv}
}

type createScrapeRequestRequest struct {
	URL string `json:"url"`
}

type createScrapeRequestResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (h *CreateScrapeRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req createScrapeRequestRequest
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid_json"}`, http.StatusBadRequest)
		return
	}

	id, err := h.service.Request(r.Context(), req.URL)

	if err != nil {
		switch err {
		case errors.ErrInputInvalid:
			http.Error(w, `{"error": "missing_or_invalid_url"}`, http.StatusBadRequest)
		case errors.ErrStoreUnsupported:
			http.Error(w, `{"error": "unsupported_store"}`, http.StatusUnprocessableEntity)
		default:
			http.Error(w, `{"error": "internal_error"}`, http.StatusInternalServerError)
		}
		return
	}

	resp := createScrapeRequestResponse{
		ID:     id,
		Status: "pending",
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}
