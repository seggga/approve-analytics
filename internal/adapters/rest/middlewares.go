package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/seggga/approve-analytics/internal/domain/models"
)

type ctxKeyUser struct{}

// CheckAuth ...
func (s *Server) CheckAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken, err1 := r.Cookie("access")
		refreshToken, err2 := r.Cookie("refresh")

		if err1 != nil && err2 != nil {
			http.Error(w, "no token passed", http.StatusUnauthorized)
			return
		}

		tokens := &models.TokenPair{}
		if accessToken != nil {
			tokens.Access = accessToken.Value
		}
		if refreshToken != nil {
			tokens.Refresh = refreshToken.Value
		}

		tokenPair, err := s.auth.Authenticate(r.Context(), tokens)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// we have to update cookies with new tokens
		if tokenPair.Refreshed {
			// set cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "access",
				Value:    tokenPair.Access,
				Expires:  time.Now().Add(time.Minute * 1),
				HttpOnly: true,
			})

			// set refresh cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "refresh",
				Value:    tokenPair.Refresh,
				Expires:  time.Now().Add(time.Minute * 60),
				HttpOnly: true,
			})

		}
		ctx := r.Context()

		ctx = context.WithValue(ctx, ctxKeyUser{}, tokenPair.Login)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
