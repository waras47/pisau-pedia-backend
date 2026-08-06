package usecase

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type ConfiguratorShapeInput struct {
	Name        string
	Category    string
	Description *string
	ImageURL    *string
	SortOrder   int
}

type ConfiguratorBladeInput struct {
	ShapeID        string
	Name           string
	Steel          string
	LengthMm       int
	Price          float64
	CompareAtPrice *float64
	Description    *string
	Specifications map[string]string
	ImageURL       *string
	SortOrder      int
}

type ConfiguratorHandleInput struct {
	Name       string
	Material   string
	PriceDelta float64
	ImageURL   *string
	SortOrder  int
}

type ConfiguratorAccessoryInput struct {
	Name      string
	Price     float64
	ImageURL  *string
	SortOrder int
}

type ConfiguratorUsecase struct {
	shapeRepo     repository.ConfiguratorShapeRepository
	bladeRepo     repository.ConfiguratorBladeRepository
	handleRepo    repository.ConfiguratorHandleRepository
	accessoryRepo repository.ConfiguratorAccessoryRepository
}

func NewConfiguratorUsecase(
	shapeRepo repository.ConfiguratorShapeRepository,
	bladeRepo repository.ConfiguratorBladeRepository,
	handleRepo repository.ConfiguratorHandleRepository,
	accessoryRepo repository.ConfiguratorAccessoryRepository,
) *ConfiguratorUsecase {
	return &ConfiguratorUsecase{shapeRepo: shapeRepo, bladeRepo: bladeRepo, handleRepo: handleRepo, accessoryRepo: accessoryRepo}
}

// --- Shape ---

func (u *ConfiguratorUsecase) ListShapes(ctx context.Context) ([]entity.ConfiguratorShape, error) {
	return u.shapeRepo.FindAll(ctx)
}

func (u *ConfiguratorUsecase) GetShape(ctx context.Context, id string) (*entity.ConfiguratorShape, error) {
	return u.shapeRepo.FindByID(ctx, id)
}

func (u *ConfiguratorUsecase) CreateShape(ctx context.Context, in ConfiguratorShapeInput) (*entity.ConfiguratorShape, error) {
	s := &entity.ConfiguratorShape{
		ID:          uuid.New().String(),
		Name:        in.Name,
		Category:    in.Category,
		Description: in.Description,
		ImageURL:    in.ImageURL,
		SortOrder:   in.SortOrder,
	}
	if err := u.shapeRepo.Create(ctx, s); err != nil {
		return nil, err
	}
	return u.shapeRepo.FindByID(ctx, s.ID)
}

func (u *ConfiguratorUsecase) UpdateShape(ctx context.Context, id string, in ConfiguratorShapeInput) (*entity.ConfiguratorShape, error) {
	s, err := u.shapeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != "" {
		s.Name = in.Name
	}
	if in.Category != "" {
		s.Category = in.Category
	}
	s.Description = in.Description
	s.ImageURL = in.ImageURL
	s.SortOrder = in.SortOrder
	if err := u.shapeRepo.Update(ctx, s); err != nil {
		return nil, err
	}
	return u.shapeRepo.FindByID(ctx, s.ID)
}

func (u *ConfiguratorUsecase) DeleteShape(ctx context.Context, id string) error {
	if _, err := u.shapeRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return u.shapeRepo.Delete(ctx, id)
}

// --- Blade ---

func (u *ConfiguratorUsecase) ListBlades(ctx context.Context) ([]entity.ConfiguratorBlade, error) {
	return u.bladeRepo.FindAll(ctx)
}

func (u *ConfiguratorUsecase) GetBlade(ctx context.Context, id string) (*entity.ConfiguratorBlade, error) {
	return u.bladeRepo.FindByID(ctx, id)
}

func (u *ConfiguratorUsecase) CreateBlade(ctx context.Context, in ConfiguratorBladeInput) (*entity.ConfiguratorBlade, error) {
	if _, err := u.shapeRepo.FindByID(ctx, in.ShapeID); err != nil {
		return nil, err
	}

	specs, _ := json.Marshal(in.Specifications)

	b := &entity.ConfiguratorBlade{
		ID:             uuid.New().String(),
		ShapeID:        in.ShapeID,
		Name:           in.Name,
		Steel:          in.Steel,
		LengthMm:       in.LengthMm,
		Price:          in.Price,
		CompareAtPrice: in.CompareAtPrice,
		Description:    in.Description,
		Specifications: specs,
		ImageURL:       in.ImageURL,
		SortOrder:      in.SortOrder,
	}
	if err := u.bladeRepo.Create(ctx, b); err != nil {
		return nil, err
	}
	return u.bladeRepo.FindByID(ctx, b.ID)
}

func (u *ConfiguratorUsecase) UpdateBlade(ctx context.Context, id string, in ConfiguratorBladeInput) (*entity.ConfiguratorBlade, error) {
	b, err := u.bladeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.ShapeID != "" {
		if _, err := u.shapeRepo.FindByID(ctx, in.ShapeID); err != nil {
			return nil, err
		}
		b.ShapeID = in.ShapeID
	}
	if in.Name != "" {
		b.Name = in.Name
	}
	if in.Steel != "" {
		b.Steel = in.Steel
	}
	if in.LengthMm > 0 {
		b.LengthMm = in.LengthMm
	}
	b.Price = in.Price
	b.CompareAtPrice = in.CompareAtPrice
	b.Description = in.Description
	b.ImageURL = in.ImageURL
	b.SortOrder = in.SortOrder
	if in.Specifications != nil {
		specs, _ := json.Marshal(in.Specifications)
		b.Specifications = specs
	}
	if err := u.bladeRepo.Update(ctx, b); err != nil {
		return nil, err
	}
	return u.bladeRepo.FindByID(ctx, b.ID)
}

func (u *ConfiguratorUsecase) DeleteBlade(ctx context.Context, id string) error {
	if _, err := u.bladeRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return u.bladeRepo.Delete(ctx, id)
}

// --- Handle ---

func (u *ConfiguratorUsecase) ListHandles(ctx context.Context) ([]entity.ConfiguratorHandle, error) {
	return u.handleRepo.FindAll(ctx)
}

func (u *ConfiguratorUsecase) GetHandle(ctx context.Context, id string) (*entity.ConfiguratorHandle, error) {
	return u.handleRepo.FindByID(ctx, id)
}

func (u *ConfiguratorUsecase) CreateHandle(ctx context.Context, in ConfiguratorHandleInput) (*entity.ConfiguratorHandle, error) {
	h := &entity.ConfiguratorHandle{
		ID:         uuid.New().String(),
		Name:       in.Name,
		Material:   in.Material,
		PriceDelta: in.PriceDelta,
		ImageURL:   in.ImageURL,
		SortOrder:  in.SortOrder,
	}
	if err := u.handleRepo.Create(ctx, h); err != nil {
		return nil, err
	}
	return u.handleRepo.FindByID(ctx, h.ID)
}

func (u *ConfiguratorUsecase) UpdateHandle(ctx context.Context, id string, in ConfiguratorHandleInput) (*entity.ConfiguratorHandle, error) {
	h, err := u.handleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != "" {
		h.Name = in.Name
	}
	if in.Material != "" {
		h.Material = in.Material
	}
	h.PriceDelta = in.PriceDelta
	h.ImageURL = in.ImageURL
	h.SortOrder = in.SortOrder
	if err := u.handleRepo.Update(ctx, h); err != nil {
		return nil, err
	}
	return u.handleRepo.FindByID(ctx, h.ID)
}

func (u *ConfiguratorUsecase) DeleteHandle(ctx context.Context, id string) error {
	if _, err := u.handleRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return u.handleRepo.Delete(ctx, id)
}

// --- Accessory ---

func (u *ConfiguratorUsecase) ListAccessories(ctx context.Context) ([]entity.ConfiguratorAccessory, error) {
	return u.accessoryRepo.FindAll(ctx)
}

func (u *ConfiguratorUsecase) GetAccessory(ctx context.Context, id string) (*entity.ConfiguratorAccessory, error) {
	return u.accessoryRepo.FindByID(ctx, id)
}

func (u *ConfiguratorUsecase) CreateAccessory(ctx context.Context, in ConfiguratorAccessoryInput) (*entity.ConfiguratorAccessory, error) {
	a := &entity.ConfiguratorAccessory{
		ID:        uuid.New().String(),
		Name:      in.Name,
		Price:     in.Price,
		ImageURL:  in.ImageURL,
		SortOrder: in.SortOrder,
	}
	if err := u.accessoryRepo.Create(ctx, a); err != nil {
		return nil, err
	}
	return u.accessoryRepo.FindByID(ctx, a.ID)
}

func (u *ConfiguratorUsecase) UpdateAccessory(ctx context.Context, id string, in ConfiguratorAccessoryInput) (*entity.ConfiguratorAccessory, error) {
	a, err := u.accessoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != "" {
		a.Name = in.Name
	}
	a.Price = in.Price
	a.ImageURL = in.ImageURL
	a.SortOrder = in.SortOrder
	if err := u.accessoryRepo.Update(ctx, a); err != nil {
		return nil, err
	}
	return u.accessoryRepo.FindByID(ctx, a.ID)
}

func (u *ConfiguratorUsecase) DeleteAccessory(ctx context.Context, id string) error {
	if _, err := u.accessoryRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return u.accessoryRepo.Delete(ctx, id)
}
