package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/webpush"
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
	repo    repository.NotificationRepository
	pushRepo repository.PushSubscriptionRepository
	pusher   *webpush.Pusher
	frontURL string
	log      zerolog.Logger
}

func NewNotificationUsecase(repo repository.NotificationRepository, pushRepo repository.PushSubscriptionRepository, pusher *webpush.Pusher, frontURL string, log zerolog.Logger) *NotificationUsecase {
	return &NotificationUsecase{repo: repo, pushRepo: pushRepo, pusher: pusher, frontURL: frontURL, log: log}
}

func (u *NotificationUsecase) create(ctx context.Context, module entity.NotificationModule, notifType, title, message string, referenceID, link *string) error {
	err := u.repo.Create(ctx, &entity.Notification{
		ID:          uuid.New().String(),
		Module:      module,
		Type:        notifType,
		Title:       title,
		Message:     message,
		ReferenceID: referenceID,
		Link:        link,
	})
	if err != nil {
		return err
	}

	go u.sendPushToAdmins(title, message, link)
	return nil
}

func (u *NotificationUsecase) sendPushToAdmins(title, message string, link *string) {
	if u.pusher == nil || u.pushRepo == nil {
		return
	}
	ctx := context.Background()
	subs, err := u.pushRepo.FindAllAdmin(ctx)
	if err != nil {
		u.log.Error().Err(err).Msg("push: failed to fetch admin subscriptions")
		return
	}

	pushURL := ""
	if link != nil {
		pushURL = u.frontURL + *link
	}

	payload := webpush.Payload{
		Title: title,
		Body:  message,
		Icon:  u.frontURL + "/logo-pisaupedia.png",
		URL:   pushURL,
		Tag:   "pisaupedia-admin",
	}

	for _, sub := range subs {
		if err := u.pusher.Send(webpush.Subscription{
			Endpoint: sub.Endpoint,
			P256dh:   sub.P256dh,
			Auth:     sub.Auth,
		}, payload); err != nil {
			if webpush.IsGone(err) {
				_ = u.pushRepo.DeleteByEndpoint(ctx, sub.Endpoint)
			}
			u.log.Warn().Err(err).Str("endpoint", sub.Endpoint[:40]).Msg("push: send failed")
		}
	}
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

func (u *NotificationUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

// --- Per-event helpers, called from other usecases after a successful mutation ---

func ptr(s string) *string { return &s }

func (u *NotificationUsecase) NotifyOrderCreated(ctx context.Context, order *entity.Order) error {
	return u.create(ctx, entity.NotificationModuleOrder, "order_created",
		"Pesanan Baru",
		fmt.Sprintf("Pesanan baru dari %s senilai Rp%d", order.CustomerName, order.Total),
		ptr(order.ID), ptr("/pisaupedia/admin/orders?status=pending"))
}

func (u *NotificationUsecase) NotifyOrderStatusChanged(ctx context.Context, order *entity.Order, oldStatus, newStatus entity.OrderStatus) error {
	return u.create(ctx, entity.NotificationModuleOrder, "order_status_changed",
		"Status Pesanan Diperbarui",
		fmt.Sprintf("Pesanan #%s dari %s berubah dari %s menjadi %s", order.ID[:8], order.CustomerName, oldStatus, newStatus),
		ptr(order.ID), ptr("/pisaupedia/admin/orders"))
}

func (u *NotificationUsecase) NotifyOrderPaid(ctx context.Context, order *entity.Order) error {
	return u.create(ctx, entity.NotificationModuleOrder, "order_paid",
		"Pembayaran Diterima",
		fmt.Sprintf("Pesanan #%s dari %s telah dibayar (Rp%d)", order.ID[:8], order.CustomerName, order.Total),
		ptr(order.ID), ptr("/pisaupedia/admin/orders"))
}

// NotifyOrderReceived fires when a customer self-confirms their package
// arrived — informational for admin, doesn't imply Status changed (see
// docs/16-plan-konfirmasi-pesanan-diterima-review.md).
func (u *NotificationUsecase) NotifyOrderReceived(ctx context.Context, order *entity.Order) error {
	return u.create(ctx, entity.NotificationModuleOrder, "order_received",
		"Pesanan Dikonfirmasi Diterima",
		fmt.Sprintf("Pesanan #%s dikonfirmasi diterima oleh %s", order.ID[:8], order.CustomerName),
		ptr(order.ID), ptr("/pisaupedia/admin/orders"))
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
		return "/pisaupedia/admin/engravings"
	}
	return "/pisaupedia/admin/sharpening"
}

func (u *NotificationUsecase) NotifyCustomerRegistered(ctx context.Context, user *entity.User) error {
	return u.create(ctx, entity.NotificationModuleCustomer, "customer_registered",
		"Customer Baru",
		fmt.Sprintf("%s (%s) baru saja mendaftar", user.FullName, user.Email),
		ptr(user.ID), ptr("/pisaupedia/admin/customers"))
}

func (u *NotificationUsecase) NotifyProductCreated(ctx context.Context, product *entity.Product) error {
	return u.create(ctx, entity.NotificationModuleProduct, "product_created",
		"Produk Baru Ditambahkan",
		fmt.Sprintf("%s telah ditambahkan ke katalog", product.Name),
		ptr(product.ID), ptr("/pisaupedia/admin/products"))
}

func (u *NotificationUsecase) NotifyProductDeleted(ctx context.Context, product *entity.Product) error {
	return u.create(ctx, entity.NotificationModuleProduct, "product_deleted",
		"Produk Dihapus",
		fmt.Sprintf("%s telah dihapus dari katalog", product.Name),
		nil, ptr("/pisaupedia/admin/products"))
}

func (u *NotificationUsecase) NotifyProductLowStock(ctx context.Context, product *entity.Product) error {
	return u.create(ctx, entity.NotificationModuleProduct, "product_low_stock",
		"Stok Menipis",
		fmt.Sprintf("Stok %s tinggal %d unit", product.Name, product.Stock),
		ptr(product.ID), ptr("/pisaupedia/admin/products"))
}

func (u *NotificationUsecase) NotifyReviewCreated(ctx context.Context, review *entity.Review) error {
	return u.create(ctx, entity.NotificationModuleReview, "review_created",
		"Review Baru",
		fmt.Sprintf("Review baru dari %s (⭐%d)", review.CustomerName, review.Rating),
		ptr(review.ID), ptr("/pisaupedia/admin/reviews"))
}
