package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"back/internal/shared/config"
)

func Auth(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if cfg.Token != tokenStr {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			c.Set("user_id", 1)

			return next(c)
		}
	}
}
