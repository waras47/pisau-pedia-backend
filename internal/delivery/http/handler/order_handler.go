package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/export"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

const dateLayout = "2006-01-02"

// parseReportRange reads ?from=&to= (YYYY-MM-DD), defaulting to the last 30
// days when omitted or unparseable — shared by every report endpoint so the
// JSON preview and the export always agree on what range they're showing.
func parseReportRange(c echo.Context) (time.Time, time.Time) {
	now := time.Now()
	from := now.AddDate(0, 0, -30)
	to := now

	if v := c.QueryParam("from"); v != "" {
		if parsed, err := time.Parse(dateLayout, v); err == nil {
			from = parsed
		}
	}
	if v := c.QueryParam("to"); v != "" {
		if parsed, err := time.Parse(dateLayout, v); err == nil {
			to = parsed.Add(24*time.Hour - time.Second)
		}
	}
	return from, to
}

type OrderHandler struct {
	orderUsecase *usecase.OrderUsecase
}

func NewOrderHandler(orderUsecase *usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{orderUsecase: orderUsecase}
}

func (h *OrderHandler) Create(c echo.Context) error {
	var req dto.CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	input := req.ToInput()
	// Set only when appmw.OptionalJWTAuth verified a real access token —
	// never trust a client-sent user id. Empty means guest checkout, which
	// stays fully supported.
	input.UserID = currentUserID(c)

	order, err := h.orderUsecase.CreateOrder(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, usecase.ErrEmptyOrder) {
			return response.Error(c, http.StatusUnprocessableEntity, "order must have at least one valid item", nil)
		}
		if errors.Is(err, usecase.ErrInsufficientStock) {
			return response.Error(c, http.StatusConflict, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to create order", nil)
	}
	return response.Success(c, http.StatusCreated, "Order created", dto.ToOrderResponse(order))
}

// CreateManual lets an admin record an order taken outside the storefront
// (e.g. WhatsApp) so it counts in reports — created already paid, no
// payment gateway or shipping-cost lookup involved.
func (h *OrderHandler) CreateManual(c echo.Context) error {
	var req dto.CreateManualOrderRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	order, err := h.orderUsecase.CreateOrder(c.Request().Context(), req.ToInput())
	if err != nil {
		if errors.Is(err, usecase.ErrEmptyOrder) {
			return response.Error(c, http.StatusUnprocessableEntity, "order must have at least one valid item", nil)
		}
		if errors.Is(err, usecase.ErrInsufficientStock) {
			return response.Error(c, http.StatusConflict, err.Error(), nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to create order", nil)
	}
	return response.Success(c, http.StatusCreated, "Order created", dto.ToOrderResponse(order))
}

func (h *OrderHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	result, err := h.orderUsecase.ListOrders(c.Request().Context(), usecase.OrderListInput{
		Page:          page,
		PerPage:       perPage,
		Status:        c.QueryParam("status"),
		CustomerEmail: c.QueryParam("customer_email"),
		Search:        c.QueryParam("search"),
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list orders", nil)
	}

	return response.SuccessPaginated(c, http.StatusOK, dto.ToOrderResponses(result.Orders), response.Meta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *OrderHandler) GetByID(c echo.Context) error {
	order, err := h.orderUsecase.GetOrder(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return response.Error(c, http.StatusNotFound, "order not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get order", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToOrderResponse(order))
}

// ListMine returns the authenticated customer's own orders — never guest
// orders, even ones placed with the same email (see
// docs/16-plan-konfirmasi-pesanan-diterima-review.md).
func (h *OrderHandler) ListMine(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	result, err := h.orderUsecase.ListMyOrders(c.Request().Context(), currentUserID(c), page, perPage)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list orders", nil)
	}

	return response.SuccessPaginated(c, http.StatusOK, dto.ToOrderResponses(result.Orders), response.Meta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

// GetMine returns one order only if it belongs to the authenticated
// customer — a 404 covers both "doesn't exist" and "not yours".
func (h *OrderHandler) GetMine(c echo.Context) error {
	order, err := h.orderUsecase.GetMyOrder(c.Request().Context(), currentUserID(c), c.Param("id"))
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return response.Error(c, http.StatusNotFound, "order not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get order", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToOrderResponse(order))
}

// ConfirmReceived lets the authenticated customer self-report their package
// arrived. It never changes Status — see
// docs/16-plan-konfirmasi-pesanan-diterima-review.md for why.
func (h *OrderHandler) ConfirmReceived(c echo.Context) error {
	order, err := h.orderUsecase.ConfirmReceived(c.Request().Context(), currentUserID(c), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrOrderNotFound):
			return response.Error(c, http.StatusNotFound, "order not found", nil)
		case errors.Is(err, usecase.ErrOrderNotEligibleForConfirmation):
			return response.Error(c, http.StatusConflict, "order is not eligible for receipt confirmation yet", nil)
		default:
			return response.Error(c, http.StatusInternalServerError, "failed to confirm receipt", nil)
		}
	}
	return response.Success(c, http.StatusOK, "Order marked as received", dto.ToOrderResponse(order))
}

func (h *OrderHandler) UpdateStatus(c echo.Context) error {
	var req dto.UpdateOrderStatusRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	err := h.orderUsecase.UpdateOrderStatus(c.Request().Context(), c.Param("id"), entity.OrderStatus(req.Status), req.TrackingNumber, req.ShippingEvidenceURL)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return response.Error(c, http.StatusNotFound, "order not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update order status", nil)
	}
	return response.Success(c, http.StatusOK, "Order status updated", nil)
}

func (h *OrderHandler) UpdatePaymentStatus(c echo.Context) error {
	var req dto.UpdatePaymentStatusRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "validation failed", err.Error())
	}

	err := h.orderUsecase.UpdatePaymentStatus(c.Request().Context(), c.Param("id"), entity.PaymentStatus(req.PaymentStatus))
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return response.Error(c, http.StatusNotFound, "order not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to update payment status", nil)
	}
	return response.Success(c, http.StatusOK, "Payment status updated", nil)
}

func (h *OrderHandler) GetSalesReport(c echo.Context) error {
	from, to := parseReportRange(c)

	report, err := h.orderUsecase.GetSalesReport(c.Request().Context(), from, to)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to build sales report", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToSalesReportResponse(report))
}

func salesReportTable(report *usecase.SalesReportResult) export.Table {
	table := export.Table{
		Title:   fmt.Sprintf("Sales Report (%s - %s)", report.From.Format(dateLayout), report.To.Format(dateLayout)),
		Headers: []string{"Order ID", "Tanggal", "Customer", "Status", "Pembayaran", "Total (Rp)"},
	}
	for _, o := range report.Orders {
		table.Rows = append(table.Rows, []string{
			o.ID,
			o.CreatedAt.Format(dateLayout),
			o.CustomerName,
			string(o.Status),
			string(o.PaymentStatus),
			fmt.Sprintf("%d", o.Total),
		})
	}
	return table
}

func (h *OrderHandler) ExportSalesReport(c echo.Context) error {
	from, to := parseReportRange(c)

	report, err := h.orderUsecase.GetSalesReport(c.Request().Context(), from, to)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to build sales report", nil)
	}

	table := salesReportTable(report)
	filename := fmt.Sprintf("sales-report-%s_%s", from.Format(dateLayout), to.Format(dateLayout))

	switch c.QueryParam("format") {
	case "pdf":
		bytes, err := export.ToPDF(table)
		if err != nil {
			return response.Error(c, http.StatusInternalServerError, "failed to generate PDF", nil)
		}
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, filename))
		return c.Blob(http.StatusOK, "application/pdf", bytes)
	default:
		bytes, err := export.ToExcel(table)
		if err != nil {
			return response.Error(c, http.StatusInternalServerError, "failed to generate Excel file", nil)
		}
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.xlsx"`, filename))
		return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", bytes)
	}
}
