package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type ReviewHandler struct {
	usecase *usecase.ReviewUsecase
}

func NewReviewHandler(usecase *usecase.ReviewUsecase) *ReviewHandler {
	return &ReviewHandler{usecase: usecase}
}

func (h *ReviewHandler) Create(c echo.Context) error {
	var req dto.CreateReviewRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	review, err := h.usecase.CreateReview(c.Request().Context(), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return response.Error(c, http.StatusNotFound, "product not found", nil)
		}
		if errors.Is(err, usecase.ErrInvalidRating) {
			return response.Error(c, http.StatusUnprocessableEntity, "rating must be between 1 and 5", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to create review", nil)
	}
	return response.Success(c, http.StatusCreated, "Review submitted", dto.ToReviewResponse(review))
}

// ListPublic only ever returns approved reviews — the status is hardcoded
// here rather than trusted from a query param, so this handler is safe to
// mount on a public route without leaking pending/rejected reviews.
func (h *ReviewHandler) ListPublic(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	result, err := h.usecase.ListReviews(c.Request().Context(), usecase.ReviewListInput{
		Page:        page,
		PerPage:     perPage,
		ProductSlug: c.QueryParam("product_slug"),
		Status:      string(entity.ReviewStatusApproved),
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list reviews", nil)
	}

	return response.SuccessPaginated(c, http.StatusOK, dto.ToReviewResponses(result.Reviews), response.Meta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

// CreateAdmin lets an admin insert a review directly (e.g. one collected
// offline), optionally setting the initial status instead of always
// starting "pending" like the public Create endpoint.
func (h *ReviewHandler) CreateAdmin(c echo.Context) error {
	var req dto.AdminCreateReviewRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	review, err := h.usecase.CreateReview(c.Request().Context(), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return response.Error(c, http.StatusNotFound, "product not found", nil)
		}
		if errors.Is(err, usecase.ErrInvalidRating) {
			return response.Error(c, http.StatusUnprocessableEntity, "rating must be between 1 and 5", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to create review", nil)
	}
	return response.Success(c, http.StatusCreated, "Review created", dto.ToReviewResponse(review))
}

// List is the admin moderation view — any status, filterable.
func (h *ReviewHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	result, err := h.usecase.ListReviews(c.Request().Context(), usecase.ReviewListInput{
		Page:        page,
		PerPage:     perPage,
		ProductSlug: c.QueryParam("product_slug"),
		Status:      c.QueryParam("status"),
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list reviews", nil)
	}

	return response.SuccessPaginated(c, http.StatusOK, dto.ToReviewResponses(result.Reviews), response.Meta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *ReviewHandler) Update(c echo.Context) error {
	var req dto.UpdateReviewRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	review, err := h.usecase.UpdateReview(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		if errors.Is(err, repository.ErrReviewNotFound) {
			return response.Error(c, http.StatusNotFound, "review not found", nil)
		}
		if errors.Is(err, usecase.ErrInvalidRating) {
			return response.Error(c, http.StatusUnprocessableEntity, "rating must be between 1 and 5", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update review", nil)
	}
	return response.Success(c, http.StatusOK, "Review updated", dto.ToReviewResponse(review))
}

func (h *ReviewHandler) Delete(c echo.Context) error {
	if err := h.usecase.DeleteReview(c.Request().Context(), c.Param("id")); err != nil {
		if errors.Is(err, repository.ErrReviewNotFound) {
			return response.Error(c, http.StatusNotFound, "review not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to delete review", nil)
	}
	return response.Success(c, http.StatusOK, "Review deleted", nil)
}
