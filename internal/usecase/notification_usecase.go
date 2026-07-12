package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type NotificationListInput struct {
	Page       int
	PerPage    int
	UnreadOnly bool
}

type NotificationListResult struct {
	Notifications []entity.Notification
	Page          int
	PerPage       int
	Total         int64
	TotalPages    int64
}

type NotificationUsecase struct {
	repo repository.NotificationRepository
}

func NewNotificationUsecase(repo repository.NotificationRepository) *NotificationUsecase {
	return &NotificationUsecase{repo: repo}
}

func (u *NotificationUsecase) create(ctx context.Context, module entity.NotificationModule, notifType, title, message string, referenceID, link *string) error {
	return u.repo.Create(ctx, &entity.Notification{
		ID:          uuid.New().String(),
		Module:      module,
		Type:        notifType,
		Title:       title,
		Message:     message,
		ReferenceID: referenceID,
		Link:        link,
	})
}

func (u *NotificationUsecase) ListNotifications(ctx context.Context, input NotificationListInput) (*NotificationListResult, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	perPage := input.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	notifications, total, err := u.repo.FindAll(ctx, repository.NotificationFilter{
		Page:       page,
		PerPage:    perPage,
		UnreadOnly: input.UnreadOnly,
	})
	if err != nil {
		return nil, err
	}

	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}

	return &NotificationListResult{Notifications: notifications, Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}, nil
}

func (u *NotificationUsecase) CountUnread(ctx context.Context) (int64, error) {
	return u.repo.CountUnread(ctx)
}

func (u *NotificationUsecase) MarkAsRead(ctx context.Context, id string) error {
	return u.repo.MarkAsRead(ctx, id)
}

func (u *NotificationUsecase) MarkAllAsRead(ctx context.Context) error {
	return u.repo.MarkAllAsRead(ctx)
}

// --- Per-event helpers, called from other usecases after a successful mutation ---

func ptr(s string) *string { return &s }

func (u *NotificationUsecase) NotifyOrderCreated(ctx context.Context, order *entity.Order) error {
	return u.create(ctx, entity.NotificationModuleOrder, "order_created",
		"Pesanan Baru",
		fmt.Sprintf("Pesanan baru dari %s senilai Rp%d", order.CustomerName, order.Total),
		ptr(order.ID), ptr("/admin/orders?status=pending"))
}

func (u *NotificationUsecase) NotifyOrderStatusChanged(ctx context.Context, order *entity.Order, oldStatus, newStatus entity.OrderStatus) error {
	return u.create(ctx, entity.NotificationModuleOrder, "order_status_changed",
		"Status Pesanan Diperbarui",
		fmt.Sprintf("Pesanan #%s dari %s berubah dari %s menjadi %s", order.ID[:8], order.CustomerName, oldStatus, newStatus),
		ptr(order.ID), ptr("/admin/orders"))
}

func (u *NotificationUsecase) NotifyOrderPaid(ctx context.Context, order *entity.Order) error {
	return u.create(ctx, entity.NotificationModuleOrder, "order_paid",
		"Pembayaran Diterima",
		fmt.Sprintf("Pesanan #%s dari %s telah dibayar (Rp%d)", order.ID[:8], order.CustomerName, order.Total),
		ptr(order.ID), ptr("/admin/orders"))
}

func (u *NotificationUsecase) NotifyServiceRequestCreated(ctx context.Context, req *entity.ServiceRequest) error {
	return u.create(ctx, entity.NotificationModuleService, "service_created",
		"Permintaan Servis Baru",
		fmt.Sprintf("Permintaan %s baru dari %s", req.Type, req.CustomerName),
		ptr(req.ID), ptr(serviceRequestLink(req.Type)))
}

func (u *NotificationUsecase) NotifyServiceRequestStatusChanged(ctx context.Context, req *entity.ServiceRequest, oldStatus, newStatus entity.ServiceRequestStatus) error {
	return u.create(ctx, entity.NotificationModuleService, "service_status_changed",
		"Status Servis Diperbarui",
		fmt.Sprintf("Permintaan %s dari %s berubah dari %s menjadi %s", req.Type, req.CustomerName, oldStatus, newStatus),
		ptr(req.ID), ptr(serviceRequestLink(req.Type)))
}

func serviceRequestLink(t entity.ServiceRequestType) string {
	if t == entity.ServiceRequestTypeEngraving {
		return "/admin/engravings"
	}
	return "/admin/sharpening"
}

func (u *NotificationUsecase) NotifyCustomerRegistered(ctx context.Context, user *entity.User) error {
	return u.create(ctx, entity.NotificationModuleCustomer, "customer_registered",
		"Customer Baru",
		fmt.Sprintf("%s (%s) baru saja mendaftar", user.FullName, user.Email),
		ptr(user.ID), ptr("/admin/customers"))
}

func (u *NotificationUsecase) NotifyProductCreated(ctx context.Context, product *entity.Product) error {
	return u.create(ctx, entity.NotificationModuleProduct, "product_created",
		"Produk Baru Ditambahkan",
		fmt.Sprintf("%s telah ditambahkan ke katalog", product.Name),
		ptr(product.ID), ptr("/admin/products"))
}

func (u *NotificationUsecase) NotifyProductDeleted(ctx context.Context, product *entity.Product) error {
	return u.create(ctx, entity.NotificationModuleProduct, "product_deleted",
		"Produk Dihapus",
		fmt.Sprintf("%s telah dihapus dari katalog", product.Name),
		nil, ptr("/admin/products"))
}

func (u *NotificationUsecase) NotifyProductLowStock(ctx context.Context, product *entity.Product) error {
	return u.create(ctx, entity.NotificationModuleProduct, "product_low_stock",
		"Stok Menipis",
		fmt.Sprintf("Stok %s tinggal %d unit", product.Name, product.Stock),
		ptr(product.ID), ptr("/admin/products"))
}
