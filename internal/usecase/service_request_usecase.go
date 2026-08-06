package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/mailer"
)

var ErrInvalidServiceRequestType = errors.New("invalid service request type")

type CreateServiceRequestInput struct {
	Type          string
	CustomerName  string
	CustomerEmail string
	CustomerPhone *string
	Message       string
}

type ServiceRequestListInput struct {
	Page    int
	PerPage int
	Type    string
	Status  string
}

type ServiceRequestListResult struct {
	Requests   []entity.ServiceRequest
	Page       int
	PerPage    int
	Total      int64
	TotalPages int64
}

type UpdateServiceRequestInput struct {
	Status      entity.ServiceRequestStatus
	QuotedPrice *int64
	AdminNotes  *string
}

type ServiceRequestUsecase struct {
	repo                repository.ServiceRequestRepository
	notificationUsecase *NotificationUsecase
	mailer              *mailer.Mailer
}

func NewServiceRequestUsecase(repo repository.ServiceRequestRepository, notificationUsecase *NotificationUsecase, mailer *mailer.Mailer) *ServiceRequestUsecase {
	return &ServiceRequestUsecase{repo: repo, notificationUsecase: notificationUsecase, mailer: mailer}
}

func (u *ServiceRequestUsecase) CreateRequest(ctx context.Context, input CreateServiceRequestInput) (*entity.ServiceRequest, error) {
	reqType := entity.ServiceRequestType(input.Type)
	if reqType != entity.ServiceRequestTypeSharpening && reqType != entity.ServiceRequestTypeEngraving {
		return nil, ErrInvalidServiceRequestType
	}

	now := time.Now()
	req := &entity.ServiceRequest{
		ID:            uuid.New().String(),
		Type:          reqType,
		Status:        entity.ServiceRequestStatusPending,
		CustomerName:  input.CustomerName,
		CustomerEmail: input.CustomerEmail,
		CustomerPhone: input.CustomerPhone,
		Message:       input.Message,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := u.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	// Notification failures shouldn't block request creation.
	_ = u.notificationUsecase.NotifyServiceRequestCreated(ctx, req)

	return req, nil
}

func (u *ServiceRequestUsecase) ListRequests(ctx context.Context, input ServiceRequestListInput) (*ServiceRequestListResult, error) {
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

	requests, total, err := u.repo.FindAll(ctx, repository.ServiceRequestFilter{
		Page:    page,
		PerPage: perPage,
		Type:    input.Type,
		Status:  input.Status,
	})
	if err != nil {
		return nil, err
	}

	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}

	return &ServiceRequestListResult{Requests: requests, Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}, nil
}

func (u *ServiceRequestUsecase) ListMyRequests(ctx context.Context, email string, page, perPage int) (*ServiceRequestListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	requests, total, err := u.repo.FindAll(ctx, repository.ServiceRequestFilter{
		Page:    page,
		PerPage: perPage,
		Email:   email,
	})
	if err != nil {
		return nil, err
	}

	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}

	return &ServiceRequestListResult{Requests: requests, Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}, nil
}

func (u *ServiceRequestUsecase) GetRequest(ctx context.Context, id string) (*entity.ServiceRequest, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *ServiceRequestUsecase) UpdateRequest(ctx context.Context, id string, input UpdateServiceRequestInput) (*entity.ServiceRequest, error) {
	req, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	oldStatus := req.Status

	if input.Status != "" {
		req.Status = input.Status
	}
	if input.QuotedPrice != nil {
		req.QuotedPrice = input.QuotedPrice
	}
	if input.AdminNotes != nil {
		req.AdminNotes = input.AdminNotes
	}

	if err := u.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	if req.Status != oldStatus {
		_ = u.notificationUsecase.NotifyServiceRequestStatusChanged(ctx, req, oldStatus, req.Status)

		if u.mailer != nil && u.mailer.Enabled() {
			notes := ""
			if req.AdminNotes != nil {
				notes = *req.AdminNotes
			}
			if err := u.mailer.SendServiceRequestStatusEmail(
				req.CustomerEmail, req.CustomerName,
				string(req.Type), string(oldStatus), string(req.Status), notes,
			); err != nil {
				slog.Error("failed to send service request status email", "err", err, "to", req.CustomerEmail)
			}
		}
	}

	return req, nil
}
