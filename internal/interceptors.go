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

	LastIDKey   ContextKey = "lastID"
	PageSizeKey ContextKey = "pageSize"
)

type PathParameterKey string
type QueryParameterKey string

func (s PathParameterKey) String() string  { return string(s) }
func (s QueryParameterKey) String() string { return string(s) }

const (
	APIVersion = "v1"

	CityID     PathParameterKey = "cityid"
	ItemID     PathParameterKey = "itemid"
	MovementID PathParameterKey = "movementid"

	LastID         QueryParameterKey = "lastid"
	PageSize       QueryParameterKey = "pagesize"
	PlayerID       QueryParameterKey = "playerid"
	LocationBounds QueryParameterKey = "locationbounds"
	OriginID       QueryParameterKey = "originid"
	DestinationID  QueryParameterKey = "destinationid"
)

func WithAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		playerID := "foo" // TODO: validate token and add info to context
		if playerID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorized"))
			return
		}

		ctx := context.WithValue(r.Context(), PlayerIDKey, playerID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithCityIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cityID := chi.URLParam(r, CityID.String())
		if cityID == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing /cityid/ path parameter"))
			return
		}

		ctx := context.WithValue(r.Context(), CityIDKey, cityID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithMovementIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		movementID := chi.URLParam(r, MovementID.String())
		ctx := context.WithValue(r.Context(), MovementIDKey, movementID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithUnitQueueItemIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		unitQueueItemID := chi.URLParam(r, ItemID.String())
		ctx := context.WithValue(r.Context(), UnitQueueItemIDKey, unitQueueItemID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithBuildingQueueItemIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buildingQueueItemID := chi.URLParam(r, ItemID.String())
		ctx := context.WithValue(r.Context(), BuildingQueueItemIDKey, buildingQueueItemID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithPagination(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastID := r.URL.Query().Get(LastID.String())
		pageSize := r.URL.Query().Get(PageSize.String())
		if pageSize == "" {
			pageSize = "10"
		}

		ctx := context.WithValue(r.Context(), LastIDKey, lastID)
		ctx = context.WithValue(ctx, PageSizeKey, pageSize)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
