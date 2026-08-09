package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/dto"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/export"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/response"
)

type ProductHandler struct {
	productUsecase   *usecase.ProductUsecase
	sitePromoUsecase *usecase.SitePromoUsecase
}

func NewProductHandler(productUsecase *usecase.ProductUsecase, sitePromoUsecase *usecase.SitePromoUsecase) *ProductHandler {
	return &ProductHandler{productUsecase: productUsecase, sitePromoUsecase: sitePromoUsecase}
}

func (h *ProductHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	ctx := c.Request().Context()
	result, err := h.productUsecase.ListProducts(ctx, usecase.ProductListInput{
		Page:         page,
		PerPage:      perPage,
		CategorySlug: c.QueryParam("category"),
		Search:       c.QueryParam("search"),
		Sort:         c.QueryParam("sort"),
		Badge:        c.QueryParam("badge"),
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list products", nil)
	}

	items := dto.ToProductListItems(result.Products)
	if promo, _ := h.sitePromoUsecase.GetActive(ctx); promo != nil {
		promoSet := make(map[string]struct{}, len(promo.ProductIDs))
		for _, pid := range promo.ProductIDs {
			promoSet[pid] = struct{}{}
		}
		for i := range items {
			if !promo.ApplyToAll {
				if _, ok := promoSet[items[i].ID]; !ok {
					continue
				}
			}
			if items[i].CompareAtPrice == nil {
				original := items[i].Price
				items[i].CompareAtPrice = &original
			}
			items[i].Price = usecase.ApplyPromoDiscount(items[i].Price, promo.DiscountPercent)
		}
	}

	return response.SuccessPaginated(c, http.StatusOK, items, response.Meta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

func (h *ProductHandler) GetBySlug(c echo.Context) error {
	ctx := c.Request().Context()
	product, err := h.productUsecase.GetProductBySlug(ctx, c.Param("slug"))
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return response.Error(c, http.StatusNotFound, "product not found", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "failed to get product", nil)
	}

	resp := dto.ToProductDetailResponse(product)
	if promo, _ := h.sitePromoUsecase.GetActive(ctx); promo != nil {
		applies := promo.ApplyToAll
		if !applies {
			for _, pid := range promo.ProductIDs {
				if pid == resp.ID {
					applies = true
					break
				}
			}
		}
		if applies {
			if resp.CompareAtPrice == nil {
				original := resp.Price
				resp.CompareAtPrice = &original
			}
			resp.Price = usecase.ApplyPromoDiscount(resp.Price, promo.DiscountPercent)
		}
	}

	return response.Success(c, http.StatusOK, "OK", resp)
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

func (h *ProductHandler) GetInventoryReport(c echo.Context) error {
	report, err := h.productUsecase.GetInventoryReport(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to build inventory report", nil)
	}
	return response.Success(c, http.StatusOK, "OK", dto.ToInventoryReportResponse(report))
}

func inventoryReportTable(report *usecase.InventoryReportResult) export.Table {
	table := export.Table{
		Title:   "Inventory Report",
		Headers: []string{"Nama Produk", "Kategori", "Stok", "Harga (Rp)", "Nilai Stok (Rp)"},
	}
	for _, p := range report.Products {
		category := "Tanpa Kategori"
		if p.CategoryName != nil {
			category = *p.CategoryName
		}
		table.Rows = append(table.Rows, []string{
			p.Name,
			category,
			fmt.Sprintf("%d", p.Stock),
			fmt.Sprintf("%d", p.Price),
			fmt.Sprintf("%d", p.Price*int64(p.Stock)),
		})
	}
	return table
}

func (h *ProductHandler) ExportInventoryReport(c echo.Context) error {
	report, err := h.productUsecase.GetInventoryReport(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to build inventory report", nil)
	}

	table := inventoryReportTable(report)
	filename := fmt.Sprintf("inventory-report-%s", time.Now().Format("2006-01-02"))

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
