package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Burgatski/pack-calculator/internal/handler"
	"github.com/Burgatski/pack-calculator/internal/model"
	"github.com/Burgatski/pack-calculator/internal/storage"
)

func newHandler() *handler.Handler {
	return handler.New(storage.NewMemory())
}

func marshalBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return bytes.NewReader(b)
}


func TestGetPackSizes_ReturnsDefaults(t *testing.T) {
	h := newHandler()
	rr := httptest.NewRecorder()
	h.GetPackSizes(rr, httptest.NewRequest(http.MethodGet, "/api/packs", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var resp model.PackSizesResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.PackSizes) == 0 {
		t.Error("expected non-empty default pack sizes")
	}
}

func TestPutPackSizes(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{
			name:       "valid_sizes_accepted",
			body:       model.PackSizesResponse{PackSizes: []int{100, 200, 500}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty_sizes_rejected",
			body:       model.PackSizesResponse{PackSizes: []int{}},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "negative_size_rejected",
			body:       model.PackSizesResponse{PackSizes: []int{-1, 500}},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "invalid_json_rejected",
			body:       "not-valid-json{{{",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler()
			req := httptest.NewRequest(http.MethodPut, "/api/packs", marshalBody(t, tc.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h.PutPackSizes(rr, req)
			if rr.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rr.Code, tc.wantStatus, rr.Body.String())
			}
		})
	}
}

func TestPutPackSizes_ResponseReflectsUpdate(t *testing.T) {
	h := newHandler()
	newSizes := []int{100, 300, 700}
	req := httptest.NewRequest(http.MethodPut, "/api/packs",
		marshalBody(t, model.PackSizesResponse{PackSizes: newSizes}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.PutPackSizes(rr, req)

	var resp model.PackSizesResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.PackSizes) != len(newSizes) {
		t.Errorf("response pack count = %d, want %d", len(resp.PackSizes), len(newSizes))
	}
}

func TestPostCalculate(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{
			name:       "valid_order_ok",
			body:       model.CalculateRequest{Order: 501},
			wantStatus: http.StatusOK,
		},
		{
			name:       "zero_order_rejected",
			body:       model.CalculateRequest{Order: 0},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "negative_order_rejected",
			body:       model.CalculateRequest{Order: -1},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid_json_rejected",
			body:       "bad",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler()
			req := httptest.NewRequest(http.MethodPost, "/api/calculate", marshalBody(t, tc.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h.PostCalculate(rr, req)
			if rr.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rr.Code, tc.wantStatus, rr.Body.String())
			}
		})
	}
}

func TestPostCalculate_ResponseFields(t *testing.T) {
	h := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/calculate",
		marshalBody(t, model.CalculateRequest{Order: 501}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.PostCalculate(rr, req)

	var resp model.CalculateResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Packs) == 0 {
		t.Error("expected non-empty Packs map")
	}
	if resp.TotalItems < 501 {
		t.Errorf("TotalItems %d is less than order 501", resp.TotalItems)
	}
	if resp.TotalPacks <= 0 {
		t.Errorf("TotalPacks = %d, want > 0", resp.TotalPacks)
	}
}

func TestGetHealth(t *testing.T) {
	h := newHandler()
	rr := httptest.NewRecorder()
	h.GetHealth(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}
