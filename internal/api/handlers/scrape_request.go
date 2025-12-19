/*
uniwish.com/interal/api/services/scrape_request

scrape_request http service
*/
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"uniwish.com/internal/api/errors"
)

type ScrapeRequestCreator interface {
	Create(ctx context.Context, rawUrl string) (string, error)
}

type CreateScrapeRequestHandler struct {
	service ScrapeRequestCreator
}

func NewCreateItemHandler(srv ScrapeRequestCreator) *CreateScrapeRequestHandler {
	return &CreateScrapeRequestHandler{service: srv}
}

type createScrapeRequestRequest struct {
	URL string `json:"url"`
}

type createScrapeRequestResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (h *CreateScrapeRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req createScrapeRequestRequest
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid_json"}`, http.StatusBadRequest)
		return
	}

	id, err := h.service.Create(r.Context(), req.URL)

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
