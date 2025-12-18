/*
uniwish.com/interal/api/handlers/health_test

test health handler
*/
package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"uniwish.com/internal/api/errors"
)

type FakeHealthService struct {
	err error
}

func (s *FakeHealthService) Check(ctx context.Context) error {
	return s.err
}

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name           string
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "healthy",
			serviceError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unhealthy",
			serviceError:   errors.ErrUnavailable,
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := &FakeHealthService{err: tt.serviceError}
			handler := NewHealthHandler(srv)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
