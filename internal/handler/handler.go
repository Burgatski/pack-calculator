package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Burgatski/pack-calculator/internal/calculator"
	"github.com/Burgatski/pack-calculator/internal/model"
	"github.com/Burgatski/pack-calculator/internal/storage"
)

// Handler wires together the storage and HTTP layer.
type Handler struct {
	store storage.Storage
}

// New returns a Handler backed by the given storage.
func New(store storage.Storage) *Handler {
	return &Handler{store: store}
}

// RegisterRoutes registers all API routes on mux.
// The static UI route (GET /) is registered separately in main.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/packs", h.GetPackSizes)
	mux.HandleFunc("PUT /api/packs", h.PutPackSizes)
	mux.HandleFunc("POST /api/calculate", h.PostCalculate)
	mux.HandleFunc("GET /health", h.GetHealth)
}

// GetPackSizes returns the current pack sizes.
func (h *Handler) GetPackSizes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, model.PackSizesResponse{
		PackSizes: h.store.GetPackSizes(),
	})
}

// PutPackSizes replaces all pack sizes with the values from the request body.
func (h *Handler) PutPackSizes(w http.ResponseWriter, r *http.Request) {
	var req model.PackSizesResponse
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if err := h.store.SetPackSizes(req.PackSizes); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.PackSizesResponse{
		PackSizes: h.store.GetPackSizes(),
	})
}

// PostCalculate returns the optimal pack breakdown for an order.
func (h *Handler) PostCalculate(w http.ResponseWriter, r *http.Request) {
	var req model.CalculateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.Order <= 0 {
		writeError(w, http.StatusBadRequest, "order must be a positive integer")
		return
	}

	result, err := calculator.Calculate(req.Order, h.store.GetPackSizes())
	if err != nil {
		slog.Error("calculation failed", "order", req.Order, "error", err)
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.CalculateResponse{
		Packs:      result.Packs,
		TotalItems: result.TotalItems,
		TotalPacks: result.TotalPacks,
	})
}

// GetHealth is a simple liveness probe.
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, model.ErrorResponse{Error: msg})
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
