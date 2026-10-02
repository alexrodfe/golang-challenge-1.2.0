package categories

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type Response struct {
	Categories []Category `json:"categories"`
}

type Category struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CategoriesRepository interface {
	GetAllCategories() ([]models.ProductCategory, error)
	CreateCategory(category models.ProductCategory) (*models.ProductCategory, error)
}

type CategoriesHandler struct {
	repo CategoriesRepository
}

func NewCategoriesHandler(r CategoriesRepository) *CategoriesHandler {
	return &CategoriesHandler{
		repo: r,
	}
}

func (h *CategoriesHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	res, err := h.repo.GetAllCategories()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	categories := make([]Category, len(res))
	for i, c := range res {
		categories[i] = Category{
			Code: string(c.Code),
			Name: c.Name,
		}
	}

	api.OKResponse(w, Response{Categories: categories})
}

func (h *CategoriesHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" || req.Name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "code and name are required")
		return
	}

	created, err := h.repo.CreateCategory(models.ProductCategory{
		Code: models.CategoryCode(req.Code),
		Name: req.Name,
	})
	if err != nil {
		if errors.Is(err, models.ErrCategoryAlreadyExists) {
			api.ErrorResponse(w, http.StatusConflict, err.Error())
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Category{Code: string(created.Code), Name: created.Name})
}
