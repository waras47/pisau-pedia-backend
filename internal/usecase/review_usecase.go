package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

var ErrInvalidRating = errors.New("rating must be between 1 and 5")

type CreateReviewInput struct {
	ProductSlug   string
	CustomerName  string
	CustomerEmail *string
	Rating        uint
	Content       string
	Photos        []string
	Status        entity.ReviewStatus
}

// UpdateReviewInput fields are pointers so admin edits can be partial —
// nil means "leave as-is".
type UpdateReviewInput struct {
	CustomerName  *string
	CustomerEmail *string
	Rating        *uint
	Content       *string
	Status        *entity.ReviewStatus
}

type ReviewListInput struct {
	Page        int
	PerPage     int
	ProductSlug string
	Status      string
	Scope       string
}

type ReviewListResult struct {
	Reviews    []entity.Review
	Page       int
	PerPage    int
	Total      int64
	TotalPages int64
}

type ReviewUsecase struct {
	reviewRepo          repository.ReviewRepository
	productRepo         repository.ProductRepository
	notificationUsecase *NotificationUsecase
}

func NewReviewUsecase(reviewRepo repository.ReviewRepository, productRepo repository.ProductRepository, notificationUsecase *NotificationUsecase) *ReviewUsecase {
	return &ReviewUsecase{reviewRepo: reviewRepo, productRepo: productRepo, notificationUsecase: notificationUsecase}
}

func (u *ReviewUsecase) CreateReview(ctx context.Context, input CreateReviewInput) (*entity.Review, error) {
	if input.Rating < 1 || input.Rating > 5 {
		return nil, ErrInvalidRating
	}

	// Empty ProductSlug means this is a "shop review" — feedback about the
	// store itself, not tied to any product.
	var productID *string
	if input.ProductSlug != "" {
		product, err := u.productRepo.FindBySlug(ctx, input.ProductSlug)
		if err != nil {
			return nil, err
		}
		productID = &product.ID
	}

	status := input.Status
	if status == "" {
		status = entity.ReviewStatusPending
	}

	review := &entity.Review{
		ID:            uuid.New().String(),
		ProductID:     productID,
		CustomerName:  input.CustomerName,
		CustomerEmail: input.CustomerEmail,
		Rating:        input.Rating,
		Content:       input.Content,
		Photos:        entity.StringSlice(input.Photos),
		Status:        status,
	}
	if err := u.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	if u.notificationUsecase != nil {
		_ = u.notificationUsecase.NotifyReviewCreated(ctx, review)
	}

	if review.ProductID != nil {
		stats, err := u.reviewRepo.GetApprovedStatsByProduct(ctx, *review.ProductID)
		if err != nil {
			return nil, err
		}
		if err := u.productRepo.UpdateRatingStats(ctx, *review.ProductID, stats.RatingAvg, stats.ReviewCount); err != nil {
			return nil, err
		}
	}

	return review, nil
}

func (u *ReviewUsecase) ListReviews(ctx context.Context, input ReviewListInput) (*ReviewListResult, error) {
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

	reviews, total, err := u.reviewRepo.FindAll(ctx, repository.ReviewFilter{
		Page:        page,
		PerPage:     perPage,
		ProductSlug: input.ProductSlug,
		Status:      input.Status,
		Scope:       input.Scope,
	})
	if err != nil {
		return nil, err
	}

	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}

	return &ReviewListResult{Reviews: reviews, Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}, nil
}

// UpdateReview applies a partial edit (any subset of fields) to a review,
// then recomputes the parent product's rating_avg/review_count from
// approved reviews — those denormalized columns only ever reflect reality
// through this path, so this runs regardless of which fields changed.
func (u *ReviewUsecase) UpdateReview(ctx context.Context, id string, input UpdateReviewInput) (*entity.Review, error) {
	review, err := u.reviewRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.CustomerName != nil {
		review.CustomerName = *input.CustomerName
	}
	if input.CustomerEmail != nil {
		review.CustomerEmail = input.CustomerEmail
	}
	if input.Rating != nil {
		if *input.Rating < 1 || *input.Rating > 5 {
			return nil, ErrInvalidRating
		}
		review.Rating = *input.Rating
	}
	if input.Content != nil {
		review.Content = *input.Content
	}
	if input.Status != nil {
		review.Status = *input.Status
	}

	if err := u.reviewRepo.Update(ctx, review); err != nil {
		return nil, err
	}

	if review.ProductID != nil {
		stats, err := u.reviewRepo.GetApprovedStatsByProduct(ctx, *review.ProductID)
		if err != nil {
			return nil, err
		}
		if err := u.productRepo.UpdateRatingStats(ctx, *review.ProductID, stats.RatingAvg, stats.ReviewCount); err != nil {
			return nil, err
		}
	}

	return review, nil
}

func (u *ReviewUsecase) DeleteReview(ctx context.Context, id string) error {
	review, err := u.reviewRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.reviewRepo.Delete(ctx, id); err != nil {
		return err
	}

	if review.ProductID == nil {
		return nil
	}
	stats, err := u.reviewRepo.GetApprovedStatsByProduct(ctx, *review.ProductID)
	if err != nil {
		return err
	}
	return u.productRepo.UpdateRatingStats(ctx, *review.ProductID, stats.RatingAvg, stats.ReviewCount)
}
