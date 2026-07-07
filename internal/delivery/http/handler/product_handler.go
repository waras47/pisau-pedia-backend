package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type ProductHandler struct {
	productUsecase *usecase.ProductUsecase
}

func NewProductHandler(productUsecase *usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{productUsecase: productUsecase}
}

func (h *ProductHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	result, err := h.productUsecase.ListProducts(c.Request().Context(), usecase.ProductListInput{
		Page:         page,
		PerPage:      perPage,
		CategorySlug: c.QueryParam("category"),
		Search:       c.QueryParam("search"),
		Sort:         c.QueryParam("sort"),
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list products", nil)
	}

	return response.SuccessPaginated(c, http.StatusOK, dto.ToProductListItems(result.Products), response.Meta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *ProductHandler) GetBySlug(c echo.Context) error {
	product, err := h.productUsecase.GetProductBySlug(c.Request().Context(), c.Param("slug"))
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return response.Error(c, http.StatusNotFound, "product not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get product", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToProductDetailResponse(product))
}

func (h *ProductHandler) Create(c echo.Context) error {
	var req dto.CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	product, err := h.productUsecase.CreateProduct(c.Request().Context(), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create product", nil)
	}
	return response.Success(c, http.StatusCreated, "Product created", dto.ToProductDetailResponse(product))
}

func (h *ProductHandler) Update(c echo.Context) error {
	var req dto.UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	product, err := h.productUsecase.UpdateProduct(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return response.Error(c, http.StatusNotFound, "product not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update product", nil)
	}
	return response.Success(c, http.StatusOK, "Product updated", dto.ToProductDetailResponse(product))
}

func (h *ProductHandler) Delete(c echo.Context) error {
	err := h.productUsecase.DeleteProduct(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return response.Error(c, http.StatusNotFound, "product not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to delete product", nil)
	}
	return response.Success(c, http.StatusOK, "Product deleted", nil)
}
