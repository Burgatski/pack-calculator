package model

type CalculateRequest struct {
	Order int `json:"order"`
}

type CalculateResponse struct {
	Packs      map[int]int `json:"packs"`
	TotalItems int         `json:"total_items"`
	TotalPacks int         `json:"total_packs"`
}

type PackSizesResponse struct {
	PackSizes []int `json:"pack_sizes"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
