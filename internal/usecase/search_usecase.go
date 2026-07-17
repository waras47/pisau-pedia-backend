package usecase

import (
	"context"
	"strings"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

// searchResultLimit caps each category to a handful of matches — this
// powers a quick-jump dropdown, not a paginated search results page.
const searchResultLimit = 5

type GlobalSearchResult struct {
	Products  []entity.Product
	Customers []entity.User
	Orders    []entity.Order
}

type SearchUsecase struct {
	productRepo repository.ProductRepository
	userRepo    repository.UserRepository
	orderRepo   repository.OrderRepository
}

func NewSearchUsecase(productRepo repository.ProductRepository, userRepo repository.UserRepository, orderRepo repository.OrderRepository) *SearchUsecase {
	return &SearchUsecase{productRepo: productRepo, userRepo: userRepo, orderRepo: orderRepo}
}

func (u *SearchUsecase) GlobalSearch(ctx context.Context, query string) (*GlobalSearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return &GlobalSearchResult{}, nil
	}

	products, _, err := u.productRepo.FindAll(ctx, repository.ProductFilter{
		Search: query, Page: 1, PerPage: searchResultLimit,
	})
	if err != nil {
		return nil, err
	}

	customers, _, err := u.userRepo.FindAll(ctx, repository.UserFilter{
		Search: query, Role: string(entity.RoleCustomer), Page: 1, PerPage: searchResultLimit,
	})
	if err != nil {
		return nil, err
	}

	orders, _, err := u.orderRepo.FindAll(ctx, repository.OrderFilter{
		Search: query, Page: 1, PerPage: searchResultLimit,
	})
	if err != nil {
		return nil, err
	}

	return &GlobalSearchResult{Products: products, Customers: customers, Orders: orders}, nil
}
