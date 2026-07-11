package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type SubscribeInput struct {
	Email  string
	Name   *string
	Source string
}

type SubscriberListInput struct {
	Page    int
	PerPage int
	Status  string
	Search  string
}

type SubscriberListResult struct {
	Subscribers []entity.NewsletterSubscriber
	Page        int
	PerPage     int
	Total       int64
	TotalPages  int64
}

type NewsletterUsecase struct {
	repo repository.NewsletterRepository
}

func NewNewsletterUsecase(repo repository.NewsletterRepository) *NewsletterUsecase {
	return &NewsletterUsecase{repo: repo}
}

// Subscribe is idempotent — resubscribing an unsubscribed email flips it
// back to active instead of erroring, since that's what a customer filling
// the form again almost certainly wants.
func (u *NewsletterUsecase) Subscribe(ctx context.Context, input SubscribeInput) (*entity.NewsletterSubscriber, error) {
	existing, err := u.repo.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, repository.ErrSubscriberNotFound) {
		return nil, err
	}

	if existing != nil {
		if existing.Status == entity.SubscriberStatusUnsubscribed {
			if err := u.repo.UpdateStatus(ctx, existing.ID, entity.SubscriberStatusActive); err != nil {
				return nil, err
			}
			existing.Status = entity.SubscriberStatusActive
		}
		return existing, nil
	}

	source := entity.SubscriberSource(input.Source)
	if source == "" {
		source = entity.SubscriberSourceFooter
	}

	sub := &entity.NewsletterSubscriber{
		ID:     uuid.New().String(),
		Email:  input.Email,
		Name:   input.Name,
		Source: source,
		Status: entity.SubscriberStatusActive,
	}
	if err := u.repo.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (u *NewsletterUsecase) ListSubscribers(ctx context.Context, input SubscriberListInput) (*SubscriberListResult, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	perPage := input.PerPage
	if perPage < 1 {
		perPage = 50
	}
	if perPage > 200 {
		perPage = 200
	}

	subs, total, err := u.repo.FindAll(ctx, repository.SubscriberFilter{
		Page:    page,
		PerPage: perPage,
		Status:  input.Status,
		Search:  input.Search,
	})
	if err != nil {
		return nil, err
	}

	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}

	return &SubscriberListResult{Subscribers: subs, Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}, nil
}

func (u *NewsletterUsecase) Unsubscribe(ctx context.Context, id string) error {
	if _, err := u.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return u.repo.UpdateStatus(ctx, id, entity.SubscriberStatusUnsubscribed)
}

func (u *NewsletterUsecase) DeleteSubscriber(ctx context.Context, id string) error {
	if _, err := u.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return u.repo.Delete(ctx, id)
}
