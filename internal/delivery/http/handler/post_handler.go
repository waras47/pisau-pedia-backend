package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

// ---------- PostCategoryHandler ----------

type PostCategoryHandler struct {
	usecase *usecase.PostCategoryUsecase
}

func NewPostCategoryHandler(u *usecase.PostCategoryUsecase) *PostCategoryHandler {
	return &PostCategoryHandler{usecase: u}
}

func (h *PostCategoryHandler) List(c echo.Context) error {
	cats, err := h.usecase.List(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list post categories", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToPostCategoryListResponse(cats))
}

func (h *PostCategoryHandler) GetBySlug(c echo.Context) error {
	cat, err := h.usecase.GetBySlug(c.Request().Context(), c.Param("slug"))
	if err != nil {
		return response.Error(c, http.StatusNotFound, "post category not found", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToPostCategoryResponse(cat))
}

func (h *PostCategoryHandler) Create(c echo.Context) error {
	var req dto.CreatePostCategoryRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	cat, err := h.usecase.Create(c.Request().Context(), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create post category", nil)
	}
	return response.Success(c, http.StatusCreated, "Post category created", dto.ToPostCategoryResponse(cat))
}

func (h *PostCategoryHandler) Update(c echo.Context) error {
	var req dto.UpdatePostCategoryRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	cat, err := h.usecase.Update(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to update post category", nil)
	}
	return response.Success(c, http.StatusOK, "Post category updated", dto.ToPostCategoryResponse(cat))
}

func (h *PostCategoryHandler) Delete(c echo.Context) error {
	if err := h.usecase.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to delete post category", nil)
	}
	return response.Success(c, http.StatusOK, "Post category deleted", nil)
}

// ---------- PostHandler ----------

type PostHandler struct {
	usecase *usecase.PostUsecase
}

func NewPostHandler(u *usecase.PostUsecase) *PostHandler {
	return &PostHandler{usecase: u}
}

// postPaginationMeta mirrors the page/per_page defaulting done inside
// postRepository.FindAll, so the meta returned to the client matches what
// was actually queried.
func postPaginationMeta(page, perPage int, total int64) response.Meta {
	if perPage <= 0 {
		perPage = 10
	}
	if page <= 0 {
		page = 1
	}
	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}
	return response.Meta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}
}

func (h *PostHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	filter := repository.PostFilter{
		Page:         page,
		PerPage:      perPage,
		Status:       c.QueryParam("status"),
		CategorySlug: c.QueryParam("category"),
		Sort:         c.QueryParam("sort"),
	}

	posts, total, err := h.usecase.List(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list posts", nil)
	}
	return response.SuccessPaginated(c, http.StatusOK, dto.ToPostListResponse(posts), postPaginationMeta(filter.Page, filter.PerPage, total))
}

func (h *PostHandler) PublicList(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	filter := repository.PostFilter{
		Page:         page,
		PerPage:      perPage,
		Status:       "published",
		CategorySlug: c.QueryParam("category"),
		Sort:         c.QueryParam("sort"),
	}

	posts, total, err := h.usecase.List(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list posts", nil)
	}
	return response.SuccessPaginated(c, http.StatusOK, dto.ToPostListResponse(posts), postPaginationMeta(filter.Page, filter.PerPage, total))
}

func (h *PostHandler) GetBySlug(c echo.Context) error {
	post, err := h.usecase.GetBySlug(c.Request().Context(), c.Param("slug"))
	if err != nil {
		return response.Error(c, http.StatusNotFound, "post not found", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToPostResponse(post))
}

func (h *PostHandler) Create(c echo.Context) error {
	var req dto.CreatePostRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	post, err := h.usecase.Create(c.Request().Context(), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create post", nil)
	}
	return response.Success(c, http.StatusCreated, "Post created", dto.ToPostResponse(post))
}

func (h *PostHandler) Update(c echo.Context) error {
	var req dto.UpdatePostRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}
	post, err := h.usecase.Update(c.Request().Context(), c.Param("id"), req.ToInput())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to update post", nil)
	}
	return response.Success(c, http.StatusOK, "Post updated", dto.ToPostResponse(post))
}

func (h *PostHandler) Delete(c echo.Context) error {
	if err := h.usecase.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to delete post", nil)
	}
	return response.Success(c, http.StatusOK, "Post deleted", nil)
}
