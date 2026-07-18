package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type ProfilePatch struct {
	FullName  *string
	Email     *string
	Phone     *string
	AvatarURL *string
}

type AddressInput struct {
	Label       *string
	FullName    string
	Phone       *string
	AddressLine string
	City        string
	Province    *string
	PostalCode  string
	IsDefault   bool
}

type UserUsecase struct {
	userRepo    repository.UserRepository
	addressRepo repository.AddressRepository
	orderRepo   repository.OrderRepository
}

func NewUserUsecase(userRepo repository.UserRepository, addressRepo repository.AddressRepository, orderRepo repository.OrderRepository) *UserUsecase {
	return &UserUsecase{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		orderRepo:   orderRepo,
	}
}

type CustomerListInput struct {
	Page    int
	PerPage int
	Search  string
}

type CustomerWithStats struct {
	entity.User
	OrderCount int64
	TotalSpent int64
}

type CustomerListResult struct {
	Customers  []CustomerWithStats
	Page       int
	PerPage    int
	Total      int64
	TotalPages int64
}

func (u *UserUsecase) ListCustomers(ctx context.Context, input CustomerListInput) (*CustomerListResult, error) {
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

	users, total, err := u.userRepo.FindAll(ctx, repository.UserFilter{
		Page:    page,
		PerPage: perPage,
		Role:    string(entity.RoleCustomer),
		Search:  input.Search,
	})
	if err != nil {
		return nil, err
	}

	emails := make([]string, len(users))
	for i, user := range users {
		emails[i] = user.Email
	}
	statsByEmail, err := u.orderRepo.GetCustomerStatsBulk(ctx, emails)
	if err != nil {
		return nil, err
	}

	customers := make([]CustomerWithStats, 0, len(users))
	for _, user := range users {
		stats := statsByEmail[user.Email]
		customers = append(customers, CustomerWithStats{User: user, OrderCount: stats.OrderCount, TotalSpent: stats.TotalSpent})
	}

	totalPages := total / int64(perPage)
	if total%int64(perPage) != 0 {
		totalPages++
	}

	return &CustomerListResult{Customers: customers, Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}, nil
}

func (u *UserUsecase) GetCustomer(ctx context.Context, id string) (*CustomerWithStats, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	orderCount, totalSpent, err := u.orderRepo.GetCustomerStats(ctx, user.Email)
	if err != nil {
		return nil, err
	}
	return &CustomerWithStats{User: *user, OrderCount: orderCount, TotalSpent: totalSpent}, nil
}

func (u *UserUsecase) ListCustomerAddresses(ctx context.Context, userID string) ([]entity.Address, error) {
	return u.addressRepo.FindByUserID(ctx, userID)
}

func (u *UserUsecase) UpdateCustomerStatus(ctx context.Context, id string, isActive bool) (*entity.User, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user.IsActive = isActive
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserUsecase) GetProfile(ctx context.Context, userID string) (*entity.User, error) {
	return u.userRepo.FindByID(ctx, userID)
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, userID string, patch ProfilePatch) (*entity.User, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if patch.Email != nil && *patch.Email != user.Email {
		existing, err := u.userRepo.FindByEmail(ctx, *patch.Email)
		if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != user.ID {
			return nil, ErrEmailAlreadyRegistered
		}
		user.Email = *patch.Email
	}
	if patch.FullName != nil {
		user.FullName = *patch.FullName
	}
	if patch.Phone != nil {
		user.Phone = patch.Phone
	}
	if patch.AvatarURL != nil {
		user.AvatarURL = patch.AvatarURL
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserUsecase) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCurrentPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)

	return u.userRepo.Update(ctx, user)
}

func (u *UserUsecase) ListAddresses(ctx context.Context, userID string) ([]entity.Address, error) {
	return u.addressRepo.FindByUserID(ctx, userID)
}

func (u *UserUsecase) CreateAddress(ctx context.Context, userID string, input AddressInput) (*entity.Address, error) {
	if input.IsDefault {
		if err := u.addressRepo.UnsetDefaultForUser(ctx, userID); err != nil {
			return nil, err
		}
	}

	address := &entity.Address{
		ID:          uuid.New().String(),
		UserID:      userID,
		Label:       input.Label,
		FullName:    input.FullName,
		Phone:       input.Phone,
		AddressLine: input.AddressLine,
		City:        input.City,
		Province:    input.Province,
		PostalCode:  input.PostalCode,
		IsDefault:   input.IsDefault,
	}
	if err := u.addressRepo.Create(ctx, address); err != nil {
		return nil, err
	}
	return address, nil
}

func (u *UserUsecase) UpdateAddress(ctx context.Context, userID, addressID string, input AddressInput) (*entity.Address, error) {
	address, err := u.mustOwnAddress(ctx, userID, addressID)
	if err != nil {
		return nil, err
	}

	if input.IsDefault && !address.IsDefault {
		if err := u.addressRepo.UnsetDefaultForUser(ctx, userID); err != nil {
			return nil, err
		}
	}

	address.Label = input.Label
	address.FullName = input.FullName
	address.Phone = input.Phone
	address.AddressLine = input.AddressLine
	address.City = input.City
	address.Province = input.Province
	address.PostalCode = input.PostalCode
	address.IsDefault = input.IsDefault

	if err := u.addressRepo.Update(ctx, address); err != nil {
		return nil, err
	}
	return address, nil
}

func (u *UserUsecase) DeleteAddress(ctx context.Context, userID, addressID string) error {
	if _, err := u.mustOwnAddress(ctx, userID, addressID); err != nil {
		return err
	}
	return u.addressRepo.Delete(ctx, addressID)
}

// mustOwnAddress loads the address and explicitly checks UserID matches the
// caller — deliberately not relying on a WHERE user_id = ? clause alone, so
// the ownership intent stays readable here for anyone auditing this later.
func (u *UserUsecase) mustOwnAddress(ctx context.Context, userID, addressID string) (*entity.Address, error) {
	address, err := u.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return nil, err
	}
	if address.UserID != userID {
		return nil, ErrForbidden
	}
	return address, nil
}
