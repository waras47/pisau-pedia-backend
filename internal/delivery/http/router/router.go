package router

import (
	"github.com/labstack/echo/v4"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/handler"
	appmw "github.com/pisaupediaprojek/pisau-pedia-backend/internal/delivery/http/middleware"
)

type Dependencies struct {
	JWTAuth         echo.MiddlewareFunc
	AuthHandler     *handler.AuthHandler
	UserHandler     *handler.UserHandler
	CategoryHandler *handler.CategoryHandler
	ProductHandler  *handler.ProductHandler
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
	v1.GET("/products", deps.ProductHandler.List)
	v1.GET("/products/:slug", deps.ProductHandler.GetBySlug)

	admin := v1.Group("/admin", deps.JWTAuth, appmw.RequireRole("admin"))
	admin.POST("/products", deps.ProductHandler.Create)
	admin.PATCH("/products/:id", deps.ProductHandler.Update)
	admin.DELETE("/products/:id", deps.ProductHandler.Delete)
	admin.POST("/categories", deps.CategoryHandler.Create)
	admin.PATCH("/categories/:id", deps.CategoryHandler.Update)
	admin.DELETE("/categories/:id", deps.CategoryHandler.Delete)
}
