package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luisferreira32/stickerio/internal"
)

func main() {
	// The context should be the one controlling the lifecycle of the program
	// ensure external SIGINT and SIGTERM are gracefully handled.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database := internal.NewStickerioRepository(os.Getenv("DB_HOST"))
	handlers := internal.NewServerHandler(database)

	router := chi.NewRouter()

	// middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(internal.WithAuthentication)

	// routes
	router.Route("/api", func(router chi.Router) {
		router.Get("/", handlers.GetWelcome)
	})
	router.Route("/api/cities", func(router chi.Router) {
		router.Get("/", handlers.ListCityInfo)

		router.Route(fmt.Sprintf("/{%s}", internal.CityIDKey), func(router chi.Router) {
			router.Use(internal.WithCityIDContext)
			router.Get("/info", handlers.GetCityInfo)
			router.Get("/", handlers.GetCity)

			router.Route("/unitq", func(router chi.Router) {
				router.Get("/", handlers.ListUnitQueueItem)
				router.Route(fmt.Sprintf("/{%s}", internal.UnitQueueItemIDKey), func(router chi.Router) {
					router.Use(internal.WithUnitQueueItemIDContext)
					router.Get("/", handlers.GetUnitQueueItem)
				})
			})
			router.Route("/buildingq", func(router chi.Router) {
				router.Get("/", handlers.ListBuildingQueueItems)
				router.Route(fmt.Sprintf("/{%s}", internal.BuildingQueueItemIDKey), func(router chi.Router) {
					router.Use(internal.WithBuildingQueueItemIDContext)
					router.Get("/", handlers.GetBuildingQueueItem)
				})
			})
		})
	})
	router.Route("/api/movements", func(router chi.Router) {
		router.Get("/", handlers.ListMovements)

		router.Route(fmt.Sprintf("/{%s}", internal.MovementIDKey), func(router chi.Router) {
			router.Use(internal.WithMovementIDContext)
			router.Get("/", handlers.GetMovement)
		})
	})

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           router,
	}

	go func() {
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			log.Printf("server closed")
		} else if err != nil {
			log.Printf("error starting server: %v", err)
		}
	}()
	log.Printf("started stickerio-api server from %s", os.Args[0])
	<-ctx.Done()
}
