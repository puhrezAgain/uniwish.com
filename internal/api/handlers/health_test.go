/*
uniwish.com/interal/api/services/health_test

test health service
*/
package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
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
			serviceError:   errors.ErrUnsupported,
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
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
