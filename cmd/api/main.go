package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/handler"
	appmw "github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/middleware"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/infrastructure/mysql"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/config"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/database"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/logger"
	pkgvalidator "github.com/pisaupediaprojek/pisau-pedia-backend/pkg/validator"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.App.Debug)

	db, err := database.Connect(cfg.DB.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	e := echo.New()
	e.HideBanner = true
	e.Validator = pkgvalidator.New()
	e.HTTPErrorHandler = httpErrorHandler(cfg.App.Debug, log)

	e.Use(echomw.Recover())
	e.Use(appmw.RequestLogger(log))
	e.Use(appmw.Secure(cfg.App.Env == "production"))
	e.Use(appmw.BodyLimit())
	e.Use(appmw.CORS(cfg.FrontendURL))

	// Repositories
	userRepo := mysql.NewUserRepository(db)
	addressRepo := mysql.NewAddressRepository(db)
	refreshTokenRepo := mysql.NewRefreshTokenRepository(db)
	categoryRepo := mysql.NewCategoryRepository(db)
	productRepo := mysql.NewProductRepository(db)

	// Usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, refreshTokenRepo, cfg.JWT)
	userUsecase := usecase.NewUserUsecase(userRepo, addressRepo)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo)
	productUsecase := usecase.NewProductUsecase(productRepo, categoryRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	userHandler := handler.NewUserHandler(userUsecase)
	categoryHandler := handler.NewCategoryHandler(categoryUsecase)
	productHandler := handler.NewProductHandler(productUsecase)

	jwtAuth := appmw.JWTAuth(cfg.JWT.Secret)

	v1 := e.Group("/api/v1")

	auth := v1.Group("/auth", appmw.AuthRateLimiter())
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)
	auth.POST("/logout", authHandler.Logout, jwtAuth)

	users := v1.Group("/users/me", jwtAuth)
	users.GET("", userHandler.GetProfile)
	users.PATCH("", userHandler.UpdateProfile)
	users.POST("/change-password", userHandler.ChangePassword)
	users.GET("/addresses", userHandler.ListAddresses)
	users.POST("/addresses", userHandler.CreateAddress)
	users.PATCH("/addresses/:id", userHandler.UpdateAddress)
	users.DELETE("/addresses/:id", userHandler.DeleteAddress)

	v1.GET("/categories", categoryHandler.List)
	v1.GET("/products", productHandler.List)
	v1.GET("/products/:slug", productHandler.GetBySlug)

	admin := v1.Group("/admin", jwtAuth, appmw.RequireRole("admin"))
	admin.POST("/products", productHandler.Create)
	admin.PATCH("/products/:id", productHandler.Update)
	admin.DELETE("/products/:id", productHandler.Delete)
	admin.POST("/categories", categoryHandler.Create)
	admin.PATCH("/categories/:id", categoryHandler.Update)
	admin.DELETE("/categories/:id", categoryHandler.Delete)

	log.Info().Str("port", cfg.App.Port).Msg("starting server")
	if err := e.Start(":" + cfg.App.Port); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("server stopped")
	}
}
