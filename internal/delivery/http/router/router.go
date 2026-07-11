package router

import (
	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/handler"
	appmw "github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/middleware"
)

type Dependencies struct {
	JWTAuth               echo.MiddlewareFunc
	AuthHandler           *handler.AuthHandler
	UserHandler           *handler.UserHandler
	CategoryHandler       *handler.CategoryHandler
	ProductHandler        *handler.ProductHandler
	UploadHandler         *handler.UploadHandler
	ExchangeRateHandler   *handler.ExchangeRateHandler
	OrderHandler          *handler.OrderHandler
	ServiceRequestHandler *handler.ServiceRequestHandler
	ReviewHandler         *handler.ReviewHandler
	CouponHandler         *handler.CouponHandler
	NewsletterHandler     *handler.NewsletterHandler
	NotificationHandler   *handler.NotificationHandler
}

func Register(e *echo.Echo, deps Dependencies) {
	v1 := e.Group("/api/v1")

	auth := v1.Group("/auth", appmw.AuthRateLimiter())
	auth.POST("/register", deps.AuthHandler.Register)
	auth.POST("/login", deps.AuthHandler.Login)
	auth.POST("/refresh", deps.AuthHandler.Refresh)
	auth.POST("/logout", deps.AuthHandler.Logout, deps.JWTAuth)

	users := v1.Group("/users/me", deps.JWTAuth)
	users.GET("", deps.UserHandler.GetProfile)
	users.PATCH("", deps.UserHandler.UpdateProfile)
	users.POST("/change-password", deps.UserHandler.ChangePassword)
	users.GET("/addresses", deps.UserHandler.ListAddresses)
	users.POST("/addresses", deps.UserHandler.CreateAddress)
	users.PATCH("/addresses/:id", deps.UserHandler.UpdateAddress)
	users.DELETE("/addresses/:id", deps.UserHandler.DeleteAddress)

	v1.GET("/categories", deps.CategoryHandler.List)
	v1.GET("/categories/:slug", deps.CategoryHandler.GetBySlug)
	v1.GET("/products", deps.ProductHandler.List)
	v1.GET("/products/:slug", deps.ProductHandler.GetBySlug)
	v1.GET("/exchange-rate", deps.ExchangeRateHandler.Get)
	v1.POST("/orders", deps.OrderHandler.Create)
	v1.POST("/service-requests", deps.ServiceRequestHandler.Create)
	v1.GET("/reviews", deps.ReviewHandler.ListPublic)
	v1.POST("/reviews", deps.ReviewHandler.Create)
	v1.POST("/coupons/validate", deps.CouponHandler.Validate)
	v1.POST("/newsletter/subscribe", deps.NewsletterHandler.Subscribe)

	admin := v1.Group("/admin", deps.JWTAuth, appmw.RequireRole("admin"))
	admin.POST("/products", deps.ProductHandler.Create)
	admin.PATCH("/products/:id", deps.ProductHandler.Update)
	admin.DELETE("/products/:id", deps.ProductHandler.Delete)
	admin.POST("/categories", deps.CategoryHandler.Create)
	admin.PATCH("/categories/:id", deps.CategoryHandler.Update)
	admin.DELETE("/categories/:id", deps.CategoryHandler.Delete)

	admin.POST("/uploads/image", deps.UploadHandler.UploadImage)

	admin.GET("/orders", deps.OrderHandler.List)
	admin.GET("/orders/:id", deps.OrderHandler.GetByID)
	admin.PATCH("/orders/:id/status", deps.OrderHandler.UpdateStatus)
	admin.PATCH("/orders/:id/payment-status", deps.OrderHandler.UpdatePaymentStatus)

	admin.GET("/reports/sales", deps.OrderHandler.GetSalesReport)
	admin.GET("/reports/sales/export", deps.OrderHandler.ExportSalesReport)

	admin.GET("/reports/inventory", deps.ProductHandler.GetInventoryReport)
	admin.GET("/reports/inventory/export", deps.ProductHandler.ExportInventoryReport)

	admin.GET("/service-requests", deps.ServiceRequestHandler.List)
	admin.GET("/service-requests/:id", deps.ServiceRequestHandler.GetByID)
	admin.PATCH("/service-requests/:id", deps.ServiceRequestHandler.UpdateStatus)

	admin.GET("/customers", deps.UserHandler.ListCustomers)
	admin.GET("/customers/:id", deps.UserHandler.GetCustomer)
	admin.GET("/customers/:id/addresses", deps.UserHandler.GetCustomerAddresses)
	admin.PATCH("/customers/:id/status", deps.UserHandler.UpdateCustomerStatus)

	admin.GET("/reviews", deps.ReviewHandler.List)
	admin.PATCH("/reviews/:id", deps.ReviewHandler.UpdateStatus)
	admin.DELETE("/reviews/:id", deps.ReviewHandler.Delete)

	admin.GET("/coupons", deps.CouponHandler.List)
	admin.POST("/coupons", deps.CouponHandler.Create)
	admin.PATCH("/coupons/:id", deps.CouponHandler.Update)
	admin.DELETE("/coupons/:id", deps.CouponHandler.Delete)

	admin.GET("/newsletter/subscribers", deps.NewsletterHandler.ListSubscribers)
	admin.PATCH("/newsletter/subscribers/:id/unsubscribe", deps.NewsletterHandler.Unsubscribe)
	admin.DELETE("/newsletter/subscribers/:id", deps.NewsletterHandler.Delete)

	admin.GET("/notifications", deps.NotificationHandler.List)
	admin.GET("/notifications/unread-count", deps.NotificationHandler.UnreadCount)
	admin.PATCH("/notifications/:id/read", deps.NotificationHandler.MarkAsRead)
	admin.PATCH("/notifications/read-all", deps.NotificationHandler.MarkAllAsRead)
}
