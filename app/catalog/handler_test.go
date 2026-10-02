package catalog_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/mocks"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var assertAnError = errors.New("boom")

func decimalPtr(v string) *decimal.Decimal {
	d := decimal.RequireFromString(v)
	return &d
}

func sampleCategory() models.ProductCategory {
	return models.ProductCategory{ID: 1, Code: models.CategoryClothing, Name: "Clothing"}
}

type CatalogHandlerSuite struct {
	suite.Suite
	repo    *mocks.ProductsRepository
	handler *catalog.CatalogHandler
	mux     *http.ServeMux
}

func (s *CatalogHandlerSuite) SetupTest() {
	s.repo = mocks.NewProductsRepository(s.T())
	s.handler = catalog.NewCatalogHandler(s.repo)

	// A real ServeMux is needed so GET /catalog/{code} populates r.PathValue("code").
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("GET /catalog", s.handler.HandleGet)
	s.mux.HandleFunc("GET /catalog/{code}", s.handler.HandleGetByCode)
}

func (s *CatalogHandlerSuite) doGet(target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	return rec
}

func TestCatalogHandlerSuite(t *testing.T) {
	suite.Run(t, new(CatalogHandlerSuite))
}

func (s *CatalogHandlerSuite) TestHandleGet_DefaultPagination() {
	products := []models.Product{
		{Code: "PROD001", Price: decimal.RequireFromString("10.99"), Category: sampleCategory()},
	}
	s.repo.On("GetAllProducts", 0, 10, models.ProductFilters{}).Return(products, int64(1), nil)

	rec := s.doGet("/catalog")

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("application/json", rec.Header().Get("Content-Type"))

	var body catalog.Response
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	s.Equal(int64(1), body.TotalNumber)
	s.Require().Len(body.Products, 1)
	s.Equal("PROD001", body.Products[0].Code)
	s.Equal("CLOTHING", body.Products[0].Category)
	s.Equal(10.99, body.Products[0].Price)

	s.repo.AssertExpectations(s.T())
}

func (s *CatalogHandlerSuite) TestHandleGet_CustomPaginationAndFilters() {
	s.repo.On("GetAllProducts", 5, 20, models.ProductFilters{
		CategoryCode:  "SHOES",
		PriceLessThan: decimalPtr("15"),
	}).Return([]models.Product{}, int64(0), nil)

	rec := s.doGet("/catalog?offset=5&limit=20&category=SHOES&price_max=15")

	s.Equal(http.StatusOK, rec.Code)
	s.repo.AssertExpectations(s.T())
}

func (s *CatalogHandlerSuite) TestHandleGet_LimitIsClampedToMax() {
	s.repo.On("GetAllProducts", 0, 100, models.ProductFilters{}).Return([]models.Product{}, int64(0), nil)

	rec := s.doGet("/catalog?limit=500")

	s.Equal(http.StatusOK, rec.Code)
	s.repo.AssertExpectations(s.T())
}

func (s *CatalogHandlerSuite) TestHandleGet_InvalidOffset() {
	rec := s.doGet("/catalog?offset=-1")

	s.Equal(http.StatusBadRequest, rec.Code)
	s.repo.AssertNotCalled(s.T(), "GetAllProducts", mock.Anything, mock.Anything, mock.Anything)
}

func (s *CatalogHandlerSuite) TestHandleGet_InvalidLimit() {
	rec := s.doGet("/catalog?limit=abc")

	s.Equal(http.StatusBadRequest, rec.Code)
	s.repo.AssertNotCalled(s.T(), "GetAllProducts", mock.Anything, mock.Anything, mock.Anything)
}

func (s *CatalogHandlerSuite) TestHandleGet_InvalidPriceMax() {
	rec := s.doGet("/catalog?price_max=not-a-number")

	s.Equal(http.StatusBadRequest, rec.Code)
	s.repo.AssertNotCalled(s.T(), "GetAllProducts", mock.Anything, mock.Anything, mock.Anything)
}

func (s *CatalogHandlerSuite) TestHandleGet_RepositoryError() {
	s.repo.On("GetAllProducts", 0, 10, models.ProductFilters{}).
		Return(nil, int64(0), assertAnError)

	rec := s.doGet("/catalog")

	s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *CatalogHandlerSuite) TestHandleGetByCode_Success() {
	productPrice := decimal.RequireFromString("20.00")
	product := &models.Product{
		Code:     "PROD004",
		Price:    productPrice,
		Category: sampleCategory(),
		Variants: []models.Variant{
			{SKU: "SKU004A", Name: "Variant A", Price: decimalPtr("16.99")},
			{SKU: "SKU004C", Name: "Variant C", Price: nil}, // inherits the product price
		},
	}
	s.repo.On("GetProductByCode", "PROD004").Return(product, nil)

	rec := s.doGet("/catalog/PROD004")

	s.Equal(http.StatusOK, rec.Code)

	var body catalog.ProductDetails
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	s.Equal("PROD004", body.Code)
	s.Equal("CLOTHING", body.Category)
	s.Equal(20.00, body.Price)
	s.Require().Len(body.Variants, 2)
	s.Equal(16.99, body.Variants[0].Price)
	s.Equal(20.00, body.Variants[1].Price) // inherited from the product

	s.repo.AssertExpectations(s.T())
}

func (s *CatalogHandlerSuite) TestHandleGetByCode_NotFound() {
	s.repo.On("GetProductByCode", "UNKNOWN").Return(nil, models.ErrProductNotFound)

	rec := s.doGet("/catalog/UNKNOWN")

	s.Equal(http.StatusNotFound, rec.Code)
	s.repo.AssertExpectations(s.T())
}

func (s *CatalogHandlerSuite) TestHandleGetByCode_RepositoryError() {
	s.repo.On("GetProductByCode", "PROD001").Return(nil, assertAnError)

	rec := s.doGet("/catalog/PROD001")

	s.Equal(http.StatusInternalServerError, rec.Code)
	s.repo.AssertExpectations(s.T())
}

func (s *CatalogHandlerSuite) TestHandleGetByCode_MissingCode() {
	// Calling the handler directly (bypassing the mux) since the registered
	// pattern "/catalog/{code}" never dispatches here with an empty code.
	req := httptest.NewRequest(http.MethodGet, "/catalog/", nil)
	rec := httptest.NewRecorder()
	s.handler.HandleGetByCode(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.repo.AssertNotCalled(s.T(), "GetProductByCode", mock.Anything)
}
