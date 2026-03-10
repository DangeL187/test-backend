package server

import (
	"errors"
	"net/http"

	"github.com/DangeL187/erax"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"

	"back/internal/app"
	"back/internal/infra/http/routes"
)

type Server struct {
	engine *echo.Echo
}

func (s *Server) Run(addr string) error {
	zap.S().Infof("HTTP server launched on http://%s", addr)

	if err := s.engine.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return erax.Wrap(err, "failed to start HTTP server")
	}

	return nil
}

func NewServer(app *app.App) *Server {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	routes.RegisterRoutes(e, app)

	return &Server{engine: e}
}
