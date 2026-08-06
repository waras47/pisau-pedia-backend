package repository

import (
	"context"
	"errors"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
)

var (
	ErrConfiguratorShapeNotFound     = errors.New("configurator shape not found")
	ErrConfiguratorBladeNotFound     = errors.New("configurator blade not found")
	ErrConfiguratorHandleNotFound    = errors.New("configurator handle not found")
	ErrConfiguratorAccessoryNotFound = errors.New("configurator accessory not found")
)

type ConfiguratorShapeRepository interface {
	FindAll(ctx context.Context) ([]entity.ConfiguratorShape, error)
	FindByID(ctx context.Context, id string) (*entity.ConfiguratorShape, error)
	Create(ctx context.Context, shape *entity.ConfiguratorShape) error
	Update(ctx context.Context, shape *entity.ConfiguratorShape) error
	Delete(ctx context.Context, id string) error
}

type ConfiguratorBladeRepository interface {
	FindAll(ctx context.Context) ([]entity.ConfiguratorBlade, error)
	FindByShapeID(ctx context.Context, shapeID string) ([]entity.ConfiguratorBlade, error)
	FindByID(ctx context.Context, id string) (*entity.ConfiguratorBlade, error)
	Create(ctx context.Context, blade *entity.ConfiguratorBlade) error
	Update(ctx context.Context, blade *entity.ConfiguratorBlade) error
	Delete(ctx context.Context, id string) error
}

type ConfiguratorHandleRepository interface {
	FindAll(ctx context.Context) ([]entity.ConfiguratorHandle, error)
	FindByID(ctx context.Context, id string) (*entity.ConfiguratorHandle, error)
	Create(ctx context.Context, handle *entity.ConfiguratorHandle) error
	Update(ctx context.Context, handle *entity.ConfiguratorHandle) error
	Delete(ctx context.Context, id string) error
}

type ConfiguratorAccessoryRepository interface {
	FindAll(ctx context.Context) ([]entity.ConfiguratorAccessory, error)
	FindByID(ctx context.Context, id string) (*entity.ConfiguratorAccessory, error)
	Create(ctx context.Context, accessory *entity.ConfiguratorAccessory) error
	Update(ctx context.Context, accessory *entity.ConfiguratorAccessory) error
	Delete(ctx context.Context, id string) error
}
