package catalog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

type Response struct {
	Products    []Product `json:"products"`
	TotalNumber int64     `json:"total_number"`
}

type Product struct {
	Code     string  `json:"code"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
}

type ProductsRepository interface {
	GetAllProducts(offset, limit int, filters models.ProductFilters) ([]models.Product, int64, error)
}

type CatalogHandler struct {
	repo ProductsRepository
}

func NewCatalogHandler(r ProductsRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	offset, limit, err := parsePagination(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filters, err := parseFiltering(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, total, err := h.repo.GetAllProducts(offset, limit, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Map response
	products := make([]Product, len(res))
	for i, p := range res {
		products[i] = Product{
			Code:     p.Code,
			Category: string(p.Category.Code),
			Price:    p.Price.InexactFloat64(),
		}
	}

	// Return the products as a JSON response
	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Products:    products,
		TotalNumber: total,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func parsePagination(q url.Values) (offset, limit int, err error) {
	offset = 0
	limit = 10

	if v := q.Get("offset"); v != "" {
		offset, err = strconv.Atoi(v)
		if err != nil || offset < 0 {
			return 0, 0, fmt.Errorf("invalid offset")
		}
	}
	if v := q.Get("limit"); v != "" {
		limit, err = strconv.Atoi(v)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid limit")
		}
	}
	limit = max(1, min(limit, 100))

	return offset, limit, nil
}

func parseFiltering(q url.Values) (models.ProductFilters, error) {
	var filters models.ProductFilters

	if v := q.Get("category"); v != "" {
		filters.CategoryCode = v
	}

	if v := q.Get("price_max"); v != "" {
		price, err := decimal.NewFromString(v)
		if err != nil {
			return models.ProductFilters{}, fmt.Errorf("invalid price_max")
		}
		filters.PriceLessThan = &price
	}

	return filters, nil
}
