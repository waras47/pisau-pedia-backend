package usecase

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type SitePromoInput struct {
	Title           string
	Description     *string
	DiscountPercent int
	PopupImage      *string
	ApplyToAll      *bool
	ProductIDs      []string
	StartDate       string
	EndDate         string
	IsActive        *bool
}

type SitePromoUsecase struct {
	repo repository.SitePromoRepository
}

func NewSitePromoUsecase(repo repository.SitePromoRepository) *SitePromoUsecase {
	return &SitePromoUsecase{repo: repo}
}

func (u *SitePromoUsecase) List(ctx context.Context) ([]entity.SitePromo, error) {
	promos, err := u.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	for i := range promos {
		ids, _ := u.repo.GetProductIDs(ctx, promos[i].ID)
		promos[i].ProductIDs = ids
	}
	return promos, nil
}

func (u *SitePromoUsecase) GetByID(ctx context.Context, id string) (*entity.SitePromo, error) {
	promo, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ids, _ := u.repo.GetProductIDs(ctx, promo.ID)
	promo.ProductIDs = ids
	return promo, nil
}

func (u *SitePromoUsecase) GetActive(ctx context.Context) (*entity.SitePromo, error) {
	promo, err := u.repo.FindActive(ctx)
	if err != nil || promo == nil {
		return promo, err
	}
	ids, _ := u.repo.GetProductIDs(ctx, promo.ID)
	promo.ProductIDs = ids
	return promo, nil
}

func (u *SitePromoUsecase) Create(ctx context.Context, input SitePromoInput) (*entity.SitePromo, error) {
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return nil, err
	}

	applyToAll := true
	if input.ApplyToAll != nil {
		applyToAll = *input.ApplyToAll
	}

	promo := &entity.SitePromo{
		ID:              uuid.NewString(),
		Title:           input.Title,
		Description:     input.Description,
		DiscountPercent: input.DiscountPercent,
		PopupImage:      input.PopupImage,
		ApplyToAll:      applyToAll,
		StartDate:       startDate,
		EndDate:         endDate,
		IsActive:        true,
	}
	if input.IsActive != nil {
		promo.IsActive = *input.IsActive
	}

	if err := u.repo.Create(ctx, promo); err != nil {
		return nil, err
	}

	if !applyToAll && len(input.ProductIDs) > 0 {
		if err := u.repo.SetProductIDs(ctx, promo.ID, input.ProductIDs); err != nil {
			return nil, err
		}
		promo.ProductIDs = input.ProductIDs
	}

	return promo, nil
}

func (u *SitePromoUsecase) Update(ctx context.Context, id string, input SitePromoInput) (*entity.SitePromo, error) {
	promo, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	promo.Title = input.Title
	promo.Description = input.Description
	promo.DiscountPercent = input.DiscountPercent
	promo.PopupImage = input.PopupImage

	if input.ApplyToAll != nil {
		promo.ApplyToAll = *input.ApplyToAll
	}

	if input.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			return nil, err
		}
		promo.StartDate = startDate
	}
	if input.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return nil, err
		}
		promo.EndDate = endDate
	}
	if input.IsActive != nil {
		promo.IsActive = *input.IsActive
	}

	if err := u.repo.Update(ctx, promo); err != nil {
		return nil, err
	}

	if promo.ApplyToAll {
		_ = u.repo.SetProductIDs(ctx, promo.ID, nil)
		promo.ProductIDs = nil
	} else {
		if err := u.repo.SetProductIDs(ctx, promo.ID, input.ProductIDs); err != nil {
			return nil, err
		}
		promo.ProductIDs = input.ProductIDs
	}

	return promo, nil
}

func (u *SitePromoUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func ApplyPromoDiscount(price int64, discountPercent int) int64 {
	discounted := float64(price) * (1 - float64(discountPercent)/100)
	return int64(math.Round(discounted))
}
