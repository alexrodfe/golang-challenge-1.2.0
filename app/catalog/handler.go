package catalog

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/alexrodfe/golang-challenge-1.2.0/app/api"
	"github.com/alexrodfe/golang-challenge-1.2.0/models"
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

type ProductDetails struct {
	Code     string    `json:"code"`
	Category string    `json:"category"`
	Price    float64   `json:"price"`
	Variants []Variant `json:"variants"`
}

type Variant struct {
	SKU   string  `json:"sku"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type ProductsRepository interface {
	GetAllProducts(offset, limit int, filters models.ProductFilters) ([]models.Product, int64, error)
	GetProductByCode(code string) (*models.Product, error)
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
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	filters, err := parseFiltering(r.URL.Query())
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	res, total, err := h.repo.GetAllProducts(offset, limit, filters)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
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

	api.OKResponse(w, Response{
		Products:    products,
		TotalNumber: total,
	})
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

func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "missing product code")
		return
	}

	product, err := h.repo.GetProductByCode(code)
	if err != nil {
		if errors.Is(err, models.ErrProductNotFound) {
			api.ErrorResponse(w, http.StatusNotFound, "product not found")
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	variants := make([]Variant, len(product.Variants))
	for i, v := range product.Variants {
		variants[i] = Variant{
			SKU:   v.SKU,
			Name:  v.Name,
			Price: v.EffectivePrice(*product).InexactFloat64(),
		}
	}

	api.OKResponse(w, ProductDetails{
		Code:     product.Code,
		Category: string(product.Category.Code),
		Price:    product.Price.InexactFloat64(),
		Variants: variants,
	})
}
