package categories_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexrodfe/golang-challenge-1.2.0/app/categories"
	"github.com/alexrodfe/golang-challenge-1.2.0/mocks"
	"github.com/alexrodfe/golang-challenge-1.2.0/models"
	"github.com/stretchr/testify/suite"
)

var assertAnError = errors.New("boom")

type CategoriesHandlerSuite struct {
	suite.Suite
	repo    *mocks.CategoriesRepository
	handler *categories.CategoriesHandler
	mux     *http.ServeMux
}

func (s *CategoriesHandlerSuite) SetupTest() {
	s.repo = mocks.NewCategoriesRepository(s.T())
	s.handler = categories.NewCategoriesHandler(s.repo)

	s.mux = http.NewServeMux()
	s.mux.HandleFunc("GET /categories", s.handler.HandleGet)
	s.mux.HandleFunc("POST /categories", s.handler.HandleCreate)
}

func (s *CategoriesHandlerSuite) doGet(target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	return rec
}

func (s *CategoriesHandlerSuite) doPost(target string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, target, bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	return rec
}

func TestCategoriesHandlerSuite(t *testing.T) {
	suite.Run(t, new(CategoriesHandlerSuite))
}

func (s *CategoriesHandlerSuite) TestHandleGet_Success() {
	s.repo.On("GetAllCategories").Return([]models.ProductCategory{
		{ID: 1, Code: models.CategoryClothing, Name: "Clothing"},
		{ID: 2, Code: models.CategoryShoes, Name: "Shoes"},
	}, nil)

	rec := s.doGet("/categories")

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("application/json", rec.Header().Get("Content-Type"))

	var body categories.Response
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	s.Require().Len(body.Categories, 2)
	s.Equal("CLOTHING", body.Categories[0].Code)
	s.Equal("Clothing", body.Categories[0].Name)

	s.repo.AssertExpectations(s.T())
}

func (s *CategoriesHandlerSuite) TestHandleGet_RepositoryError() {
	s.repo.On("GetAllCategories").Return(nil, assertAnError)

	rec := s.doGet("/categories")

	s.Equal(http.StatusInternalServerError, rec.Code)
	s.repo.AssertExpectations(s.T())
}

func (s *CategoriesHandlerSuite) TestHandleCreate_Success() {
	s.repo.On("CreateCategory", models.ProductCategory{Code: "BAGS", Name: "Bags"}).
		Return(&models.ProductCategory{ID: 3, Code: "BAGS", Name: "Bags"}, nil)

	rec := s.doPost("/categories", []byte(`{"code":"BAGS","name":"Bags"}`))

	s.Equal(http.StatusCreated, rec.Code)

	var body categories.Category
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body))
	s.Equal("BAGS", body.Code)
	s.Equal("Bags", body.Name)

	s.repo.AssertExpectations(s.T())
}

func (s *CategoriesHandlerSuite) TestHandleCreate_InvalidBody() {
	rec := s.doPost("/categories", []byte(`not-json`))

	s.Equal(http.StatusBadRequest, rec.Code)
	s.repo.AssertNotCalled(s.T(), "CreateCategory")
}

func (s *CategoriesHandlerSuite) TestHandleCreate_MissingFields() {
	rec := s.doPost("/categories", []byte(`{"code":""}`))

	s.Equal(http.StatusBadRequest, rec.Code)
	s.repo.AssertNotCalled(s.T(), "CreateCategory")
}

func (s *CategoriesHandlerSuite) TestHandleCreate_AlreadyExists() {
	s.repo.On("CreateCategory", models.ProductCategory{Code: "CLOTHING", Name: "Clothing"}).
		Return(nil, models.ErrCategoryAlreadyExists)

	rec := s.doPost("/categories", []byte(`{"code":"CLOTHING","name":"Clothing"}`))

	s.Equal(http.StatusConflict, rec.Code)
	s.repo.AssertExpectations(s.T())
}

func (s *CategoriesHandlerSuite) TestHandleCreate_RepositoryError() {
	s.repo.On("CreateCategory", models.ProductCategory{Code: "BAGS", Name: "Bags"}).
		Return(nil, assertAnError)

	rec := s.doPost("/categories", []byte(`{"code":"BAGS","name":"Bags"}`))

	s.Equal(http.StatusInternalServerError, rec.Code)
	s.repo.AssertExpectations(s.T())
}
