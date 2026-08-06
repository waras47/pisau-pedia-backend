package repository

import (
	"context"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

type ServiceRequestFilter struct {
	Page    int
	PerPage int
	Type    string
	Status  string
	Email   string
}

type ServiceRequestRepository interface {
	FindAll(ctx context.Context, filter ServiceRequestFilter) ([]entity.ServiceRequest, int64, error)
	FindByID(ctx context.Context, id string) (*entity.ServiceRequest, error)
	Create(ctx context.Context, req *entity.ServiceRequest) error
	Update(ctx context.Context, req *entity.ServiceRequest) error
}
