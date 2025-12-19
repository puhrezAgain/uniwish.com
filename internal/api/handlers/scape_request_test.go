/*
uniwish.com/interal/api/handlers/scape_request_test

test scrape request handler
*/
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"uniwish.com/internal/api/errors"
)

type FakeItemService struct {
	id  string
	err error
}

func (s *FakeItemService) Create(ctx context.Context, _ string) (string, error) {
	return s.id, s.err
}

func TestCreateItemService(t *testing.T) {
	tests := []struct {
		name                 string
		service              FakeItemService
		expectedStatus       int
		expectedJSONResponse createScrapeRequestResponse
	}{
		{
			name: "invalid_input",
			service: FakeItemService{
				id:  "",
				err: errors.ErrInputInvalid,
			},
			expectedStatus:       http.StatusBadRequest,
			expectedJSONResponse: createScrapeRequestResponse{},
		},
		{
			name: "store_unavailable",
			service: FakeItemService{
				id:  "",
				err: errors.ErrStoreUnsupported,
			},
			expectedStatus:       http.StatusUnprocessableEntity,
			expectedJSONResponse: createScrapeRequestResponse{},
		},
		{
			name: "internal_error",
			service: FakeItemService{
				id:  "",
				err: errors.ErrUnavailable,
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedJSONResponse: createScrapeRequestResponse{},
		},
		{
			name: "healthy",
			service: FakeItemService{
				id:  "fakeid",
				err: nil,
			},
			expectedStatus: http.StatusAccepted,
			expectedJSONResponse: createScrapeRequestResponse{
				ID:     "fakeid",
				Status: "pending",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := &tt.service
			handler := NewCreateItemHandler(srv)

			req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(`{"url": "fake.com"}`))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Code == http.StatusAccepted {
				var actualResp createScrapeRequestResponse
				if err := json.NewDecoder(rr.Body).Decode(&actualResp); err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(tt.expectedJSONResponse, actualResp) {
					t.Fatalf("expected %+v, got %+v", tt.expectedJSONResponse, actualResp)
				}
			}

		})
	}
}
