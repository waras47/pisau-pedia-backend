package main

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/handler"
	appmw "github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/middleware"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/router"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/infrastructure/mysql"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/infrastructure/payment"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/config"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/database"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/exchangerate"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/googleoauth"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/mailer"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/komercepay"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/logger"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/rajaongkir"
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
	e.Use(appmw.CORS(cfg.FrontendURL, cfg.App.Env != "production"))

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

	sitePromoRepo := mysql.NewSitePromoRepository(db)
	siteContentRepo := mysql.NewSiteContentRepository(db)
	collectionRepo := mysql.NewCollectionRepository(db)
	cfgShapeRepo := mysql.NewConfiguratorShapeRepository(db)
	cfgBladeRepo := mysql.NewConfiguratorBladeRepository(db)
	cfgHandleRepo := mysql.NewConfiguratorHandleRepository(db)
	cfgAccessoryRepo := mysql.NewConfiguratorAccessoryRepository(db)
	postCategoryRepo := mysql.NewPostCategoryRepository(db)
	postRepo := mysql.NewPostRepository(db)
	emailVerificationRepo := mysql.NewEmailVerificationRepository(db)

	// Usecases
	notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
	googleOAuthClient := googleoauth.New(cfg.GoogleOAuth.ClientID, cfg.GoogleOAuth.ClientSecret, cfg.GoogleOAuth.RedirectURL)
	mailService := mailer.New(cfg.SMTP)
	authUsecase := usecase.NewAuthUsecase(userRepo, refreshTokenRepo, emailVerificationRepo, cfg.JWT, notificationUsecase, googleOAuthClient, mailService, cfg.FrontendURL)
	userUsecase := usecase.NewUserUsecase(userRepo, addressRepo, orderRepo)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo)
	collectionUsecase := usecase.NewCollectionUsecase(collectionRepo)
	sitePromoUsecase := usecase.NewSitePromoUsecase(sitePromoRepo)
	siteContentUsecase := usecase.NewSiteContentUsecase(siteContentRepo)
	productUsecase := usecase.NewProductUsecase(productRepo, categoryRepo, notificationUsecase)
	couponUsecase := usecase.NewCouponUsecase(couponRepo)

	shippingClient := rajaongkir.New(cfg.RajaOngkir.BaseURL, cfg.RajaOngkir.APIKey, cfg.RajaOngkir.OriginID)
	shippingUsecase := usecase.NewShippingUsecase(shippingClient, productRepo)

	komercePayClient := komercepay.New(cfg.KomercePayment.BaseURL, cfg.KomercePayment.APIKey)
	paymentUsecase := usecase.NewPaymentUsecase(komercePayClient, orderRepo, notificationUsecase, cfg.KomercePayment.CallbackKey)

	// The store takes payment manually (bank transfer / ShopeePay / DANA /
	// QRIS shown statically at checkout-success, confirmed by an admin) —
	// not through an automated gateway. DummyGateway is always used here;
	// komercePayClient/paymentUsecase above stay wired (webhook route still
	// registered) so a real gateway can be switched back in later without
	// re-plumbing everything, but nothing calls it while this is the case.
	dummyGateway := payment.NewDummyGateway(cfg.FrontendURL)
	var paymentGateway usecase.PaymentGateway = dummyGateway
	orderUsecase := usecase.NewOrderUsecase(orderRepo, productRepo, paymentGateway, couponUsecase, notificationUsecase, shippingClient, komercePayClient, mailService, log)
	serviceRequestUsecase := usecase.NewServiceRequestUsecase(serviceRequestRepo, notificationUsecase, mailService)
	reviewUsecase := usecase.NewReviewUsecase(reviewRepo, productRepo, notificationUsecase)
	newsletterUsecase := usecase.NewNewsletterUsecase(newsletterRepo)
	configuratorUsecase := usecase.NewConfiguratorUsecase(cfgShapeRepo, cfgBladeRepo, cfgHandleRepo, cfgAccessoryRepo)
	postCategoryUsecase := usecase.NewPostCategoryUsecase(postCategoryRepo)
	postUsecase := usecase.NewPostUsecase(postRepo)
	searchUsecase := usecase.NewSearchUsecase(productRepo, userRepo, orderRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authUsecase, cfg.FrontendURL)
	userHandler := handler.NewUserHandler(userUsecase)
	categoryHandler := handler.NewCategoryHandler(categoryUsecase)
	collectionHandler := handler.NewCollectionHandler(collectionUsecase)
	productHandler := handler.NewProductHandler(productUsecase, sitePromoUsecase)
	sitePromoHandler := handler.NewSitePromoHandler(sitePromoUsecase)
	siteContentHandler := handler.NewSiteContentHandler(siteContentUsecase)
	orderHandler := handler.NewOrderHandler(orderUsecase)
	serviceRequestHandler := handler.NewServiceRequestHandler(serviceRequestUsecase, userRepo)
	reviewHandler := handler.NewReviewHandler(reviewUsecase)
	couponHandler := handler.NewCouponHandler(couponUsecase)
	newsletterHandler := handler.NewNewsletterHandler(newsletterUsecase)
	notificationHandler := handler.NewNotificationHandler(notificationUsecase)
	shippingHandler := handler.NewShippingHandler(shippingUsecase)
	paymentHandler := handler.NewPaymentHandler(paymentUsecase)
	configuratorHandler := handler.NewConfiguratorHandler(configuratorUsecase)
	postCategoryHandler := handler.NewPostCategoryHandler(postCategoryUsecase)
	postHandler := handler.NewPostHandler(postUsecase)
	searchHandler := handler.NewSearchHandler(searchUsecase)

	jwtAuth := appmw.JWTAuth(cfg.JWT.Secret)
	optionalJWTAuth := appmw.OptionalJWTAuth(cfg.JWT.Secret)

	store, err := storage.New(cfg.Minio)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init storage")
	}
	uploadHandler := handler.NewUploadHandler(store, orderUsecase)

	exchangeRateClient := exchangerate.New()
	exchangeRateHandler := handler.NewExchangeRateHandler(exchangeRateClient)

	router.Register(e, router.Dependencies{
		JWTAuth:               jwtAuth,
		OptionalJWTAuth:       optionalJWTAuth,
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
		ShippingHandler:       shippingHandler,
		PaymentHandler:        paymentHandler,
		SearchHandler:         searchHandler,
		CollectionHandler:     collectionHandler,
		ConfiguratorHandler:   configuratorHandler,
		SitePromoHandler:      sitePromoHandler,
		SiteContentHandler:    siteContentHandler,
		PostCategoryHandler:   postCategoryHandler,
		PostHandler:           postHandler,
	})

	go runExpiredOrderSweep(orderUsecase, log)

	log.Info().Str("port", cfg.App.Port).Msg("starting server")
	if err := e.Start(":" + cfg.App.Port); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("server stopped")
	}
}

// runExpiredOrderSweep periodically cancels orders whose payment deadline
// passed without ever being paid and releases the stock they reserved —
// the automated counterpart to the admin's manual "Cek Status Pembayaran"
// button, for customers who simply never complete payment.
func runExpiredOrderSweep(orderUsecase *usecase.OrderUsecase, log zerolog.Logger) {
	const interval = 5 * time.Minute
	sweep := func() {
		released, err := orderUsecase.ReleaseExpiredOrders(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("expired order sweep failed")
			return
		}
		if released > 0 {
			log.Info().Int("released", released).Msg("expired order sweep: stock released")
		}
	}

	sweep() // run once at startup instead of waiting a full interval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		sweep()
	}
}
