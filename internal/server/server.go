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
			Addr:    cfg.ServerAddress,
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
	dmw := DefaultMiddleware(cfg)

	r.Post("/api/user/register", dmw(nil))
	r.Post("/api/user/login", dmw(nil))
	r.Post("/api/user/orders", dmw(nil))
	r.Get("/api/user/orders", dmw(nil))
	r.Get("/api/user/balance", dmw(nil))
	r.Post("/api/user/balance/withdraw", dmw(nil))
	r.Get("/api/user/withdrawals", middleware.WithLogging(nil))

	return r
}

func DefaultMiddleware(cfg config.AppConfig) func(http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.WithAuth(cfg.HashKey)(
			middleware.WithGzip(
				middleware.WithLogging(h),
			),
		)
	}
}
