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
	eventSourcer := internal.NewEventSourcer(database)
	eventSourcer.StartEventsWorker(ctx, 10*time.Second)
	handlers := internal.NewServerHandler(database, eventSourcer)

	router := chi.NewRouter()

	// middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(internal.WithAuthentication)
	router.Use(internal.WithPagination)

	// routes
	router.Route(fmt.Sprintf("/%s", internal.APIVersion), func(router chi.Router) {
		router.Get("/", handlers.GetWelcome)
		router.Route("/cities", func(router chi.Router) {
			router.Get("/", handlers.ListCityInfo)
			router.Post("/", handlers.CreateCity)
			router.With(internal.WithCityIDContext).Route(fmt.Sprintf("/{%s}", internal.CityID), func(router chi.Router) {
				router.Get("/", handlers.GetCity)
				router.Get("/", handlers.DeleteCity)
				router.Get("/info", handlers.GetCityInfo)
				router.Route("/unitqitems", func(router chi.Router) {
					router.Get("/", handlers.ListUnitQueueItem)
					router.Post("/", handlers.QueueUnit)
					router.With(internal.WithUnitQueueItemIDContext).Route(fmt.Sprintf("/{%s}", internal.ItemID), func(router chi.Router) {
						router.Get("/", handlers.GetUnitQueueItem)
					})
				})
				router.Route("/buildingqitems", func(router chi.Router) {
					router.Get("/", handlers.ListBuildingQueueItems)
					router.Post("/", handlers.QueueBuilding)
					router.With(internal.WithBuildingQueueItemIDContext).Route(fmt.Sprintf("/{%s}", internal.ItemID), func(router chi.Router) {
						router.Get("/", handlers.GetBuildingQueueItem)
					})
				})
			})
		})
		router.Route("/movements", func(router chi.Router) {
			router.Use(internal.WithPagination)
			router.Get("/", handlers.ListMovements)
			router.Post("/start", handlers.StartMovement)

			router.Route(fmt.Sprintf("/{%s}", internal.MovementID), func(router chi.Router) {
				router.Use(internal.WithMovementIDContext)
				router.Get("/", handlers.GetMovement)
			})
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
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
	log.Printf("started stickerio-api server from %s on port %s", os.Args[0], port)
	<-ctx.Done()
}
