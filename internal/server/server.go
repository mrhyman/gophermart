package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mrhyman/gophermart/internal/config"
	"github.com/mrhyman/gophermart/internal/handler"
	"github.com/mrhyman/gophermart/internal/logger"
	"github.com/mrhyman/gophermart/internal/middleware"
)

type Server struct {
	Instance *http.Server
}

func New(cfg config.AppConfig, h handler.HTTPHandler) *Server {
	return &Server{
		Instance: &http.Server{
			Addr:    cfg.RunAddress,
			Handler: SetupMux(&h, cfg),
		},
	}
}

func (s *Server) Start(ctx context.Context) {
	logger.FromContext(ctx).Infof("listening on %s", s.Instance.Addr)
	if err := s.Instance.ListenAndServe(); err != nil {
		logger.FromContext(ctx).With("err", err.Error()).Fatal()
	}
}

func SetupMux(h *handler.HTTPHandler, cfg config.AppConfig) http.Handler {
	r := chi.NewRouter()
	publicMW := PublicMiddleware()
	authMW := AuthMiddleware(cfg)

	// Публичные роуты (без авторизации)
	r.Post("/api/user/register", publicMW(h.User.Register))
	r.Post("/api/user/login", publicMW(h.User.Login))

	// Защищенные роуты (с авторизацией)
	r.Post("/api/user/orders", authMW(h.Order.UploadOrder))
	r.Get("/api/user/orders", authMW(nil))
	r.Get("/api/user/balance", authMW(nil))
	r.Post("/api/user/balance/withdraw", authMW(nil))
	r.Get("/api/user/withdrawals", authMW(nil))

	return r
}

func PublicMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.WithGzip(
			middleware.WithLogging(h),
		)
	}
}

func AuthMiddleware(cfg config.AppConfig) func(http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.WithAuth(cfg.HashKey)(
			middleware.WithGzip(
				middleware.WithLogging(h),
			),
		)
	}
}
