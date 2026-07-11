package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/config"
)

// bcryptCost is set above the library default (10) as part of the extra
// hardening requested for this project — higher cost means brute-forcing a
// leaked hash is meaningfully slower, at an acceptable login-latency cost.
const bcryptCost = 12

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type AuthUsecase struct {
	userRepo            repository.UserRepository
	refreshTokenRepo    repository.RefreshTokenRepository
	jwtCfg              config.JWTConfig
	notificationUsecase *NotificationUsecase
}

func NewAuthUsecase(userRepo repository.UserRepository, refreshTokenRepo repository.RefreshTokenRepository, jwtCfg config.JWTConfig, notificationUsecase *NotificationUsecase) *AuthUsecase {
	return &AuthUsecase{userRepo: userRepo, refreshTokenRepo: refreshTokenRepo, jwtCfg: jwtCfg, notificationUsecase: notificationUsecase}
}

func (u *AuthUsecase) Register(ctx context.Context, email, password, fullName, phone string) (*entity.User, *TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	existing, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, nil, err
	}
	if existing != nil {
		return nil, nil, ErrEmailAlreadyRegistered
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, nil, err
	}

	user := &entity.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hash),
		FullName:     fullName,
		Role:         entity.RoleCustomer,
		IsActive:     true,
	}
	if phone != "" {
		user.Phone = &phone
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	// Notification failures shouldn't block registration.
	_ = u.notificationUsecase.NotifyCustomerRegistered(ctx, user)

	tokens, err := u.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	return user, tokens, nil
}

func (u *AuthUsecase) Login(ctx context.Context, email, password string) (*entity.User, *TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, nil, ErrAccountDisabled
	}

	tokens, err := u.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	return user, tokens, nil
}

func (u *AuthUsecase) RefreshToken(ctx context.Context, refreshTokenRaw string) (*TokenPair, error) {
	tokenHash := hashToken(refreshTokenRaw)

	stored, err := u.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	if !stored.IsValid(time.Now()) {
		return nil, ErrInvalidRefreshToken
	}

	user, err := u.userRepo.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, ErrAccountDisabled
	}

	// Rotation: revoke the token being redeemed before issuing a new pair,
	// so a stolen-but-not-yet-used refresh token can only be replayed once
	// before it stops working.
	if err := u.refreshTokenRepo.Revoke(ctx, stored.ID); err != nil {
		return nil, err
	}

	return u.generateTokenPair(ctx, user)
}

func (u *AuthUsecase) Logout(ctx context.Context, refreshTokenRaw string) error {
	tokenHash := hashToken(refreshTokenRaw)

	stored, err := u.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) {
			return nil
		}
		return err
	}
	return u.refreshTokenRepo.Revoke(ctx, stored.ID)
}

func (u *AuthUsecase) generateTokenPair(ctx context.Context, user *entity.User) (*TokenPair, error) {
	accessExpiry := time.Duration(u.jwtCfg.AccessExpiryMinutes) * time.Minute

	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": string(user.Role),
		"exp":  time.Now().Add(accessExpiry).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedAccess, err := accessToken.SignedString([]byte(u.jwtCfg.Secret))
	if err != nil {
		return nil, err
	}

	rawRefresh, err := generateOpaqueToken()
	if err != nil {
		return nil, err
	}

	refreshExpiry := time.Duration(u.jwtCfg.RefreshExpiryHours) * time.Hour
	refreshRecord := &entity.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: hashToken(rawRefresh),
		ExpiresAt: time.Now().Add(refreshExpiry),
	}
	if err := u.refreshTokenRepo.Create(ctx, refreshRecord); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  signedAccess,
		RefreshToken: rawRefresh,
		ExpiresIn:    int(accessExpiry.Seconds()),
	}, nil
}

// generateOpaqueToken returns a cryptographically random, base64url-encoded
// refresh token. It is intentionally not a JWT: revocation only requires a
// database lookup, no blocklist needed.
func generateOpaqueToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buf), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
