package handler

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/storage"
)

const maxUploadSize = 5 << 20 // 5MB

var allowedImageTypes = map[string]bool{
	"image/jpeg": true, "image/png": true, "image/webp": true,
}

type UploadHandler struct {
	storage      *storage.Storage
	orderUsecase *usecase.OrderUsecase
}

func NewUploadHandler(s *storage.Storage, orderUsecase *usecase.OrderUsecase) *UploadHandler {
	return &UploadHandler{storage: s, orderUsecase: orderUsecase}
}

func (h *UploadHandler) UploadImage(c echo.Context) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "missing image file", nil)
	}
	if fileHeader.Size > maxUploadSize {
		return response.Error(c, http.StatusBadRequest, "image too large (max 5MB)", nil)
	}
	contentType := fileHeader.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return response.Error(c, http.StatusUnprocessableEntity, "unsupported image type", nil)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to read file", nil)
	}
	defer src.Close()

	ext := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}[contentType]
	key := fmt.Sprintf("products/%s%s", uuid.New().String(), ext)

	url, err := h.storage.UploadImage(c.Request().Context(), key, src, fileHeader.Size, contentType)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to upload image", nil)
	}

	return response.Success(c, http.StatusOK, "Image uploaded", map[string]string{"url": url})
}

// UploadPaymentProof lets a customer attach a receipt/screenshot to their own
// order after checkout — part of the manual/static payment flow (no
// automated gateway confirms payment for us). Public like GetByID: the order
// UUID in the path is the only "credential" needed, matching how the rest of
// the guest-checkout flow already works.
func (h *UploadHandler) UploadPaymentProof(c echo.Context) error {
	orderID := c.Param("id")

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "missing image file", nil)
	}
	if fileHeader.Size > maxUploadSize {
		return response.Error(c, http.StatusBadRequest, "image too large (max 5MB)", nil)
	}
	contentType := fileHeader.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return response.Error(c, http.StatusUnprocessableEntity, "unsupported image type", nil)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to read file", nil)
	}
	defer src.Close()

	ext := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}[contentType]
	key := fmt.Sprintf("payment-proofs/%s%s", uuid.New().String(), ext)

	url, err := h.storage.UploadImage(c.Request().Context(), key, src, fileHeader.Size, contentType)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to upload image", nil)
	}

	order, err := h.orderUsecase.UploadPaymentProof(c.Request().Context(), orderID, url)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "order not found", nil)
	}

	return response.Success(c, http.StatusOK, "Payment proof uploaded", dto.ToOrderResponse(order))
}
