/*
uniwish.com/interal/api/handlers/health

simple health endpoint
*/
package handlers

import (
	"context"
	"net/http"
)

type HealthChecker interface {
	Check(ctx context.Context) error
}
type HealthHandler struct {
	service HealthChecker
}

func NewHealthHandler(srv HealthChecker) *HealthHandler {
	return &HealthHandler{
		service: srv,
	}
}

func (s *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := s.service.Check(r.Context())
	if err != nil {
		http.Error(w, "health check error", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
