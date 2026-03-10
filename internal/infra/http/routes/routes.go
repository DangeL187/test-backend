package routes

import (
	"github.com/labstack/echo/v5"

	"back/internal/app"
	"back/internal/features/user/middleware"
)

func RegisterRoutes(e *echo.Echo, app *app.App) {
	api := e.Group("/api")
	api.Use(middleware.Auth(app.Config))

	api.POST("/v1/withdrawals", app.TransactionHandler.CreateWithdrawal)
	api.GET("/v1/withdrawals/:id", app.TransactionHandler.GetWithdrawal)
	api.POST("/v1/withdrawals/:id/confirm", app.TransactionHandler.ConfirmWithdrawal)
}
