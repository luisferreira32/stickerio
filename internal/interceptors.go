package internal

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ContextKey string

const (
	PlayerIDKey ContextKey = "playerID"
	CityIDKey   ContextKey = "cityID"
)

func WithCityIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cityID := chi.URLParam(r, "cityID")
		ctx := context.WithValue(r.Context(), CityIDKey, cityID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: validate token and add info to context
		ctx := context.WithValue(r.Context(), PlayerIDKey, "foo")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
