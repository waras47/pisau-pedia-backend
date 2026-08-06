package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/config"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/googleoauth"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/mailer"
)

const bcryptCost = 12

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type AuthUsecase struct {
	userRepo            repository.UserRepository
	refreshTokenRepo    repository.RefreshTokenRepository
	verificationRepo    repository.EmailVerificationRepository
	jwtCfg              config.JWTConfig
	notificationUsecase *NotificationUsecase
	googleClient        *googleoauth.Client
	mailer              *mailer.Mailer
	frontendURL         string

	exchangeMu    sync.Mutex
	exchangeCodes map[string]exchangeEntry
}

type exchangeEntry struct {
	user      *entity.User
	tokens    *TokenPair
	expiresAt time.Time
}

func NewAuthUsecase(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	verificationRepo repository.EmailVerificationRepository,
	jwtCfg config.JWTConfig,
	notificationUsecase *NotificationUsecase,
	googleClient *googleoauth.Client,
	m *mailer.Mailer,
	frontendURL string,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:            userRepo,
		refreshTokenRepo:    refreshTokenRepo,
		verificationRepo:    verificationRepo,
		jwtCfg:              jwtCfg,
		notificationUsecase: notificationUsecase,
		googleClient:        googleClient,
		mailer:              m,
		frontendURL:         frontendURL,
		exchangeCodes:       make(map[string]exchangeEntry),
	}
}

func (u *AuthUsecase) Register(ctx context.Context, email, password, fullName, phone string) (*entity.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	existing, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyRegistered
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	_ = u.notificationUsecase.NotifyCustomerRegistered(ctx, user)

	go u.sendVerificationEmail(user)

	return user, nil
}

func (u *AuthUsecase) sendVerificationEmail(user *entity.User) {
	if u.mailer == nil || !u.mailer.Enabled() {
		return
	}

	rawToken, err := generateOpaqueToken()
	if err != nil {
		return
	}

	token := &entity.EmailVerificationToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: hashToken(rawToken),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := u.verificationRepo.Create(context.Background(), token); err != nil {
		return
	}

	verifyURL := fmt.Sprintf("%s/account/verify-email?token=%s", u.frontendURL, rawToken)
	_ = u.mailer.SendVerificationEmail(user.Email, user.FullName, verifyURL)
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

	if user.EmailVerifiedAt == nil {
		return nil, nil, ErrEmailNotVerified
	}

	tokens, err := u.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	return user, tokens, nil
}

func (u *AuthUsecase) VerifyEmail(ctx context.Context, rawToken string) error {
	tokenHash := hashToken(rawToken)

	stored, err := u.verificationRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrVerificationTokenNotFound) {
			return ErrVerificationTokenInvalid
		}
		return err
	}

	if stored.IsExpired() {
		return ErrVerificationTokenInvalid
	}

	user, err := u.userRepo.FindByID(ctx, stored.UserID)
	if err != nil {
		return err
	}

	now := time.Now()
	user.EmailVerifiedAt = &now
	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return u.verificationRepo.DeleteByUserID(ctx, user.ID)
}

func (u *AuthUsecase) ResendVerification(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrInvalidCredentials
		}
		return err
	}

	if user.EmailVerifiedAt != nil {
		return ErrEmailAlreadyVerified
	}

	_ = u.verificationRepo.DeleteByUserID(ctx, user.ID)

	go u.sendVerificationEmail(user)

	return nil
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

func (u *AuthUsecase) GoogleEnabled() bool {
	return u.googleClient != nil && u.googleClient.Enabled()
}

func (u *AuthUsecase) GoogleAuthURL(state string) (string, error) {
	if !u.GoogleEnabled() {
		return "", ErrGoogleAuthUnavailable
	}
	return u.googleClient.AuthCodeURL(state), nil
}

func (u *AuthUsecase) GoogleCallback(ctx context.Context, code string) (*entity.User, *TokenPair, error) {
	if !u.GoogleEnabled() {
		return nil, nil, ErrGoogleAuthUnavailable
	}

	info, err := u.googleClient.Exchange(ctx, code)
	if err != nil {
		return nil, nil, err
	}
	email := strings.ToLower(strings.TrimSpace(info.Email))

	user, err := u.userRepo.FindByGoogleID(ctx, info.Sub)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, nil, err
	}

	if user == nil {
		byEmail, err := u.userRepo.FindByEmail(ctx, email)
		if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, err
		}
		if byEmail != nil {
			byEmail.GoogleID = &info.Sub
			if byEmail.EmailVerifiedAt == nil {
				now := time.Now()
				byEmail.EmailVerifiedAt = &now
			}
			if err := u.userRepo.Update(ctx, byEmail); err != nil {
				return nil, nil, err
			}
			user = byEmail
		}
	}

	if user == nil {
		randomPassword, err := generateOpaqueToken()
		if err != nil {
			return nil, nil, err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcryptCost)
		if err != nil {
			return nil, nil, err
		}

		fullName := info.Name
		if fullName == "" {
			fullName = email
		}
		now := time.Now()
		newUser := &entity.User{
			ID:              uuid.New().String(),
			Email:           email,
			PasswordHash:    string(hash),
			GoogleID:        &info.Sub,
			FullName:        fullName,
			Role:            entity.RoleCustomer,
			IsActive:        true,
			EmailVerifiedAt: &now,
		}
		if info.Picture != "" {
			newUser.AvatarURL = &info.Picture
		}
		if err := u.userRepo.Create(ctx, newUser); err != nil {
			return nil, nil, err
		}
		_ = u.notificationUsecase.NotifyCustomerRegistered(ctx, newUser)
		user = newUser
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

const exchangeCodeTTL = 2 * time.Minute

func (u *AuthUsecase) IssueExchangeCode(user *entity.User, tokens *TokenPair) (string, error) {
	code, err := generateOpaqueToken()
	if err != nil {
		return "", err
	}

	u.exchangeMu.Lock()
	defer u.exchangeMu.Unlock()
	u.pruneExpiredExchangeCodesLocked()
	u.exchangeCodes[code] = exchangeEntry{user: user, tokens: tokens, expiresAt: time.Now().Add(exchangeCodeTTL)}
	return code, nil
}

func (u *AuthUsecase) ConsumeExchangeCode(code string) (*entity.User, *TokenPair, error) {
	u.exchangeMu.Lock()
	defer u.exchangeMu.Unlock()

	entry, ok := u.exchangeCodes[code]
	delete(u.exchangeCodes, code)
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, nil, ErrInvalidExchangeCode
	}
	return entry.user, entry.tokens, nil
}

func (u *AuthUsecase) pruneExpiredExchangeCodesLocked() {
	now := time.Now()
	for code, entry := range u.exchangeCodes {
		if now.After(entry.expiresAt) {
			delete(u.exchangeCodes, code)
		}
	}
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
