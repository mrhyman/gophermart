package middleware

import (
	"context"
	"net/http"

	"github.com/mrhyman/gophermart/internal/auth"
	"github.com/mrhyman/gophermart/internal/logger"
	"github.com/mrhyman/gophermart/internal/model"
)

func WithAuth(secret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(r.Context())

			ce, err := auth.NewCookieEncoder(secret)
			if err != nil {
				log.With("err", err.Error()).Warn()
				http.Error(w, model.ErrWentWrong.Error(), http.StatusInternalServerError)
				return
			}

			cookie, err := r.Cookie("X-Auth-Token")
			if err != nil {
				log.With("err", err.Error()).Warn()
				http.Error(w, model.ErrWentWrong.Error(), http.StatusUnauthorized)
				return
			}

			userID, err := ce.DecodeUserID(cookie)
			if err != nil || userID == "" {
				log.With("err", err.Error()).Warn()
				http.Error(w, model.ErrUnknownUser.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), model.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}
