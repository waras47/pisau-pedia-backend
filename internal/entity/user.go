package entity

import "time"

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

type User struct {
	ID              string     `db:"id"`
	Email           string     `db:"email"`
	PasswordHash    string     `db:"password_hash" json:"-"`
	// GoogleID is the Google account's stable "sub" claim, set once a user
	// signs in via Google. A user who registered with email/password first
	// can still link Google later — same row, GoogleID just gets filled in.
	GoogleID        *string    `db:"google_id" json:"-"`
	FullName        string     `db:"full_name"`
	Phone           *string    `db:"phone"`
	Role            Role       `db:"role"`
	AvatarURL       *string    `db:"avatar_url"`
	IsActive        bool       `db:"is_active"`
	EmailVerifiedAt *time.Time `db:"email_verified_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
