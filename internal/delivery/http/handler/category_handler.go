package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type CategoryHandler struct {
	categoryUsecase *usecase.CategoryUsecase
}

func NewCategoryHandler(categoryUsecase *usecase.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{categoryUsecase: categoryUsecase}
}

func (h *CategoryHandler) List(c echo.Context) error {
	categories, err := h.categoryUsecase.ListCategories(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list categories", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToCategoryResponses(categories))
}

func (h *CategoryHandler) Create(c echo.Context) error {
	var req dto.CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	category, err := h.categoryUsecase.CreateCategory(c.Request().Context(), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create category", nil)
	}
	return response.Success(c, http.StatusCreated, "Category created", dto.ToCategoryResponse(category))
}

func (h *CategoryHandler) Update(c echo.Context) error {
	var req dto.UpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	category, err := h.categoryUsecase.UpdateCategory(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return response.Error(c, http.StatusNotFound, "category not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update category", nil)
	}
	return response.Success(c, http.StatusOK, "Category updated", dto.ToCategoryResponse(category))
}

func (h *CategoryHandler) Delete(c echo.Context) error {
	err := h.categoryUsecase.DeleteCategory(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return response.Error(c, http.StatusNotFound, "category not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to delete category", nil)
	}
	return response.Success(c, http.StatusOK, "Category deleted", nil)
}
