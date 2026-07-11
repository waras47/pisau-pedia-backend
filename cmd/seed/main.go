package main

import (
	"context"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/infrastructure/mysql"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/config"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/database"

	"errors"

	"github.com/google/uuid"
)

const bcryptCost = 12

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	email := os.Getenv("SEED_ADMIN_EMAIL")
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	name := os.Getenv("SEED_ADMIN_NAME")
	if email == "" || password == "" || name == "" {
		log.Fatal("SEED_ADMIN_EMAIL, SEED_ADMIN_PASSWORD, and SEED_ADMIN_NAME must be set")
	}

	db, err := database.Connect(cfg.DB.DSN())
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	userRepo := mysql.NewUserRepository(db)
	ctx := context.Background()

	existing, err := userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		log.Fatalf("find user: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	if existing != nil {
		existing.Role = entity.RoleAdmin
		existing.PasswordHash = string(hash)
		existing.IsActive = true
		if err := userRepo.Update(ctx, existing); err != nil {
			log.Fatalf("promote existing user: %v", err)
		}
		log.Printf("promoted existing user %s to admin", email)
		return
	}

	admin := &entity.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hash),
		FullName:     name,
		Role:         entity.RoleAdmin,
		IsActive:     true,
	}
	if err := userRepo.Create(ctx, admin); err != nil {
		log.Fatalf("create admin user: %v", err)
	}
	log.Printf("created admin user %s", email)
}
