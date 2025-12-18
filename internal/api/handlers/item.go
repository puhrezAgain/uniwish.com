/*
uniwish.com/interal/api/services/item

items http service
*/
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"uniwish.com/internal/api/errors"
)

type ItemCreator interface {
	Create(ctx context.Context, rawUrl string) (string, error)
}

type CreateItemHandler struct {
	service ItemCreator
}

func NewCreateItemHandler(srv ItemCreator) *CreateItemHandler {
	return &CreateItemHandler{service: srv}
}

type createItemRequest struct {
	URL string `json:"url"`
}

type createItemResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (h *CreateItemHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req createItemRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid_json"}`, http.StatusBadRequest)
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

	resp := createItemResponse{
		ID:     id,
		Status: "pending",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}
