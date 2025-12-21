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

	"github.com/google/uuid"
	"uniwish.com/internal/api/errors"
)

var fakeId uuid.UUID = uuid.New()

type FakeScrapeRequester struct {
	id  uuid.UUID
	err error
}

func (s *FakeScrapeRequester) Request(ctx context.Context, _ string) (uuid.UUID, error) {
	return s.id, s.err
}

func FakeJSON() string {
	return `{"url": "fake.com"}`
}

func TestCreateScrapeRequester(t *testing.T) {
	tests := []struct {
		name                 string
		service              FakeScrapeRequester
		payload              string
		expectedStatus       int
		expectedJSONResponse createScrapeRequestResponse
	}{
		{
			name: "invalid_input",
			service: FakeScrapeRequester{
				id:  fakeId,
				err: errors.ErrInputInvalid,
			},
			payload:              FakeJSON(),
			expectedStatus:       http.StatusBadRequest,
			expectedJSONResponse: createScrapeRequestResponse{},
		},
		{
			name: "store_unavailable",
			service: FakeScrapeRequester{
				id:  fakeId,
				err: errors.ErrStoreUnsupported,
			},
			payload:              FakeJSON(),
			expectedStatus:       http.StatusUnprocessableEntity,
			expectedJSONResponse: createScrapeRequestResponse{},
		},
		{
			name: "internal_error",
			service: FakeScrapeRequester{
				id:  fakeId,
				err: errors.ErrUnavailable,
			},
			payload:              FakeJSON(),
			expectedStatus:       http.StatusInternalServerError,
			expectedJSONResponse: createScrapeRequestResponse{},
		},
		{
			name: "healthy",
			service: FakeScrapeRequester{
				id:  fakeId,
				err: nil,
			},
			payload:        FakeJSON(),
			expectedStatus: http.StatusAccepted,
			expectedJSONResponse: createScrapeRequestResponse{
				ID:     fakeId,
				Status: "pending",
			},
		},
		{
			name: "bad_json",
			service: FakeScrapeRequester{
				id:  fakeId,
				err: nil,
			},
			payload:              `{"url":}`,
			expectedStatus:       http.StatusBadRequest,
			expectedJSONResponse: createScrapeRequestResponse{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := &tt.service
			handler := NewCreateItemHandler(srv)

			req := httptest.NewRequest(http.MethodPost, "/scrape-requests", strings.NewReader(tt.payload))
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
