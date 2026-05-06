// Package model holds the request/response types for the API.
package model

// CalculateRequest is the body for POST /api/calculate.
type CalculateRequest struct {
	Order int `json:"order"`
}

// CalculateResponse is what POST /api/calculate returns.
type CalculateResponse struct {
	Packs      map[int]int `json:"packs"`
	TotalItems int         `json:"total_items"`
	TotalPacks int         `json:"total_packs"`
}

// PackSizesResponse is shared by GET and PUT /api/packs.
type PackSizesResponse struct {
	PackSizes []int `json:"pack_sizes"`
}

// ErrorResponse is returned on any non-2xx response.
type ErrorResponse struct {
	Error string `json:"error"`
}
