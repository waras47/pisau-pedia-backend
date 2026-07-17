package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
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
	googleClient        *googleoauth.Client

	exchangeMu    sync.Mutex
	exchangeCodes map[string]exchangeEntry
}

// exchangeEntry briefly holds a completed Google login result behind a
// single-use opaque code — see IssueExchangeCode.
type exchangeEntry struct {
	user      *entity.User
	tokens    *TokenPair
	expiresAt time.Time
}

func NewAuthUsecase(userRepo repository.UserRepository, refreshTokenRepo repository.RefreshTokenRepository, jwtCfg config.JWTConfig, notificationUsecase *NotificationUsecase, googleClient *googleoauth.Client) *AuthUsecase {
	return &AuthUsecase{
		userRepo:            userRepo,
		refreshTokenRepo:    refreshTokenRepo,
		jwtCfg:              jwtCfg,
		notificationUsecase: notificationUsecase,
		googleClient:        googleClient,
		exchangeCodes:       make(map[string]exchangeEntry),
	}
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

// GoogleEnabled reports whether Google Sign-In has real credentials
// configured — the frontend uses this (via /auth/google/status) to decide
// whether to show the button at all.
func (u *AuthUsecase) GoogleEnabled() bool {
	return u.googleClient != nil && u.googleClient.Enabled()
}

// GoogleAuthURL returns the Google consent-screen URL to redirect the
// browser to. state must be a fresh random value the caller can verify on
// the way back (CSRF protection) — see AuthHandler.GoogleLogin.
func (u *AuthUsecase) GoogleAuthURL(state string) (string, error) {
	if !u.GoogleEnabled() {
		return "", ErrGoogleAuthUnavailable
	}
	return u.googleClient.AuthCodeURL(state), nil
}

// GoogleCallback exchanges the authorization code Google redirected back
// with for the user's verified profile, then finds or creates the matching
// local account:
//   - google_id already linked  -> that account (repeat login)
//   - email matches an existing account -> link google_id to it (a user who
//     registered with a password can add Google sign-in without creating a
//     second account, safe because Google already verified the email)
//   - neither -> brand new customer account
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
			if err := u.userRepo.Update(ctx, byEmail); err != nil {
				return nil, nil, err
			}
			user = byEmail
		}
	}

	if user == nil {
		// Unusable random password — this account can only ever sign in via
		// Google, but password_hash stays NOT NULL so no other code path
		// (login, change-password) needs a nil check.
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
		newUser := &entity.User{
			ID:           uuid.New().String(),
			Email:        email,
			PasswordHash: string(hash),
			GoogleID:     &info.Sub,
			FullName:     fullName,
			Role:         entity.RoleCustomer,
			IsActive:     true,
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

// exchangeCodeTTL is deliberately short — the code only needs to survive one
// browser redirect (Google callback -> frontend callback page -> immediate
// POST to consume it), typically well under a second.
const exchangeCodeTTL = 2 * time.Minute

// IssueExchangeCode hands a completed login result a single-use code instead
// of putting the real access/refresh tokens in a URL (browser history,
// server logs, Referer headers). The frontend's callback page immediately
// exchanges it server-side for the real tokens via ConsumeExchangeCode.
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

// ConsumeExchangeCode redeems a code issued by IssueExchangeCode. Codes are
// single-use: a second redemption attempt (replay) always fails.
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
