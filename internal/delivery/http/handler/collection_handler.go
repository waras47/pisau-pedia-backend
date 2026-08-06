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

type CollectionHandler struct {
	collectionUsecase *usecase.CollectionUsecase
}

func NewCollectionHandler(collectionUsecase *usecase.CollectionUsecase) *CollectionHandler {
	return &CollectionHandler{collectionUsecase: collectionUsecase}
}

func (h *CollectionHandler) List(c echo.Context) error {
	collections, err := h.collectionUsecase.ListCollections(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list collections", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToCollectionResponses(collections))
}

func (h *CollectionHandler) GetBySlug(c echo.Context) error {
	collection, err := h.collectionUsecase.GetCollectionBySlug(c.Request().Context(), c.Param("slug"))
	if err != nil {
		if errors.Is(err, repository.ErrCollectionNotFound) {
			return response.Error(c, http.StatusNotFound, "collection not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get collection", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToCollectionResponse(collection))
}

func (h *CollectionHandler) Create(c echo.Context) error {
	var req dto.CreateCollectionRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	collection, err := h.collectionUsecase.CreateCollection(c.Request().Context(), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create collection", nil)
	}
	return response.Success(c, http.StatusCreated, "Collection created", dto.ToCollectionResponse(collection))
}

func (h *CollectionHandler) Update(c echo.Context) error {
	var req dto.UpdateCollectionRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	collection, err := h.collectionUsecase.UpdateCollection(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrCollectionNotFound) {
			return response.Error(c, http.StatusNotFound, "collection not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update collection", nil)
	}
	return response.Success(c, http.StatusOK, "Collection updated", dto.ToCollectionResponse(collection))
}

func (h *CollectionHandler) Delete(c echo.Context) error {
	err := h.collectionUsecase.DeleteCollection(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrCollectionNotFound) {
			return response.Error(c, http.StatusNotFound, "collection not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to delete collection", nil)
	}
	return response.Success(c, http.StatusOK, "Collection deleted", nil)
}
