/*
uniwish.com/interal/api/handlers/product

product endpoint for getting list of products and getting a particular product
*/
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"uniwish.com/internal/api/services"
)

type DefaultProductHandler struct {
	service services.ProductReaderService
}

func NewDefaultProductHandler(service services.ProductReaderService) *DefaultProductHandler {
	return &DefaultProductHandler{service: service}
}
func (h *DefaultProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rawId := r.PathValue("id")

	if rawId == "" {
		http.Error(w, `{"error": "invalid_path"}`, http.StatusBadRequest)
		return
	}

	productId, err := uuid.Parse(rawId)
	if err != nil {
		http.Error(w, `{"error": "invalid_id"}`, http.StatusBadRequest)
		return
	}

	product, err := h.service.Get(r.Context(), productId)

	if err != nil {
		switch err {
		default:
			http.Error(w, `{"error": "internal_error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)

}
func (h *DefaultProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	list, err := h.service.List(r.Context())

	if err != nil {
		// TODO better error handling
		switch err {
		default:
			http.Error(w, `{"error": "internal_error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(list)
}
