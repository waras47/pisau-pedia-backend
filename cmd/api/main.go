package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/handler"
	appmw "github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/middleware"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/router"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/infrastructure/mysql"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/infrastructure/payment"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/config"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/database"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/exchangerate"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/logger"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/storage"
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
	orderRepo := mysql.NewOrderRepository(db)
	serviceRequestRepo := mysql.NewServiceRequestRepository(db)
	reviewRepo := mysql.NewReviewRepository(db)
	couponRepo := mysql.NewCouponRepository(db)
	newsletterRepo := mysql.NewNewsletterRepository(db)
	notificationRepo := mysql.NewNotificationRepository(db)

	// Usecases
	notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, refreshTokenRepo, cfg.JWT, notificationUsecase)
	userUsecase := usecase.NewUserUsecase(userRepo, addressRepo, orderRepo)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo)
	productUsecase := usecase.NewProductUsecase(productRepo, categoryRepo, notificationUsecase)
	couponUsecase := usecase.NewCouponUsecase(couponRepo)

	paymentGateway := payment.NewDummyGateway(cfg.FrontendURL)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, productRepo, paymentGateway, couponUsecase, notificationUsecase)
	serviceRequestUsecase := usecase.NewServiceRequestUsecase(serviceRequestRepo, notificationUsecase)
	reviewUsecase := usecase.NewReviewUsecase(reviewRepo, productRepo)
	newsletterUsecase := usecase.NewNewsletterUsecase(newsletterRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	userHandler := handler.NewUserHandler(userUsecase)
	categoryHandler := handler.NewCategoryHandler(categoryUsecase)
	productHandler := handler.NewProductHandler(productUsecase)
	orderHandler := handler.NewOrderHandler(orderUsecase)
	serviceRequestHandler := handler.NewServiceRequestHandler(serviceRequestUsecase)
	reviewHandler := handler.NewReviewHandler(reviewUsecase)
	couponHandler := handler.NewCouponHandler(couponUsecase)
	newsletterHandler := handler.NewNewsletterHandler(newsletterUsecase)
	notificationHandler := handler.NewNotificationHandler(notificationUsecase)

	jwtAuth := appmw.JWTAuth(cfg.JWT.Secret)

	store, err := storage.New(cfg.Minio)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init storage")
	}
	uploadHandler := handler.NewUploadHandler(store)

	exchangeRateClient := exchangerate.New()
	exchangeRateHandler := handler.NewExchangeRateHandler(exchangeRateClient)

	router.Register(e, router.Dependencies{
		JWTAuth:               jwtAuth,
		AuthHandler:           authHandler,
		UserHandler:           userHandler,
		CategoryHandler:       categoryHandler,
		ProductHandler:        productHandler,
		UploadHandler:         uploadHandler,
		ExchangeRateHandler:   exchangeRateHandler,
		OrderHandler:          orderHandler,
		ServiceRequestHandler: serviceRequestHandler,
		ReviewHandler:         reviewHandler,
		CouponHandler:         couponHandler,
		NewsletterHandler:     newsletterHandler,
		NotificationHandler:   notificationHandler,
	})

	log.Info().Str("port", cfg.App.Port).Msg("starting server")
	if err := e.Start(":" + cfg.App.Port); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("server stopped")
	}
}
