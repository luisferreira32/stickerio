package internal

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ContextKey string

func (s ContextKey) String() string { return string(s) }

const (
	PlayerIDKey            ContextKey = "playerID"
	CityIDKey              ContextKey = "cityID"
	MovementIDKey          ContextKey = "movementID"
	UnitQueueItemIDKey     ContextKey = "unitQueueItemID"
	BuildingQueueItemIDKey ContextKey = "buildingQueueItemID"
)

func WithAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: validate token and add info to context
		ctx := context.WithValue(r.Context(), PlayerIDKey, "foo")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithCityIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cityID := chi.URLParam(r, CityIDKey.String())
		ctx := context.WithValue(r.Context(), CityIDKey, cityID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithMovementIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movementID := chi.URLParam(r, MovementIDKey.String())
		ctx := context.WithValue(r.Context(), MovementIDKey, movementID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithUnitQueueItemIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		unitQueueItemID := chi.URLParam(r, UnitQueueItemIDKey.String())
		ctx := context.WithValue(r.Context(), UnitQueueItemIDKey, unitQueueItemID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithBuildingQueueItemIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buildingQueueItemID := chi.URLParam(r, BuildingQueueItemIDKey.String())
		ctx := context.WithValue(r.Context(), BuildingQueueItemIDKey, buildingQueueItemID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
