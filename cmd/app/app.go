package main

import (
	_ "backend-test/docs"
	"backend-test/internal/config"
	"backend-test/internal/http-server/handlers/BookingCancelHandler"
	"backend-test/internal/http-server/handlers/BookingCreateHandler"
	"backend-test/internal/http-server/handlers/DummyLoginHandler"
	"backend-test/internal/http-server/handlers/FindBookingHandler"
	"backend-test/internal/http-server/handlers/GetRoomsListHandler"
	"backend-test/internal/http-server/handlers/GetSlotsHandler"
	"backend-test/internal/http-server/handlers/InfoHandler"
	"backend-test/internal/http-server/handlers/LoginHandler"
	"backend-test/internal/http-server/handlers/MyBookingHandler"
	"backend-test/internal/http-server/handlers/RegisterHandler"
	"backend-test/internal/http-server/handlers/createRoomHandler"
	"backend-test/internal/http-server/handlers/createScheduleHandler"
	"backend-test/internal/http-server/middlewares"
	"backend-test/internal/lib/sl"
	"backend-test/internal/repository/pgxRepository/bookingRepository"
	"backend-test/internal/repository/pgxRepository/roomsRepository"
	"backend-test/internal/repository/pgxRepository/scheduleRepository"
	"backend-test/internal/repository/pgxRepository/slotsRepository"
	"backend-test/internal/repository/pgxRepository/usersRepository"
	authService "backend-test/internal/services/auth"
	"backend-test/internal/services/bookingService"
	"backend-test/internal/services/roomService"
	"backend-test/internal/services/scheduleService"
	"backend-test/internal/services/slotsService"
	"backend-test/internal/services/userService"
	"backend-test/internal/storage/db"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	envLocal = "local"
)

// @title Booking API
// @version 1.0
// @description API для бронирования переговорных комнат
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.MustLoad()

	log := initLogger(cfg.Env)
	log.Info("Starting backend application")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	str, err := db.New(cfg.StorageURL)
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
		os.Exit(1)
	}

	RoomsRepo := roomsRepository.NewRoomsRepository(str)
	SlotsRepo := slotsRepository.NewSlotRepository(str)
	BookingRepo := bookingRepository.New(str)
	ScheduleRepo := scheduleRepository.New(str)
	UserRepo := usersRepository.New(str)

	roomServ := roomService.NewRoomService(RoomsRepo, log)
	slotServ := slotsService.New(SlotsRepo, log)
	bookingServ := bookingService.New(BookingRepo, log)
	scheduleServ := scheduleService.New(ScheduleRepo, log)
	auth := authService.New(cfg.JWTSecret, cfg.ExpirationTime)
	userServ := userService.New(UserRepo, log)

	router := initRouter(
		log,
		roomServ,
		slotServ,
		bookingServ,
		scheduleServ,
		auth,
		userServ,
		cfg,
	)

	srv := &http.Server{
		Addr:    cfg.HTTPServer.URL,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start backend server", "error", err)
		}
	}()

	log.Info("backend server started")

	<-done
	log.Info("Shutting down backend server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("backend server shutdown failed", sl.Err(err))
	}
}

func initLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	return log
}

func initRouter(
	log *slog.Logger,
	rooms *roomService.RoomService,
	slots *slotsService.SlotsService,
	booking *bookingService.BookingService,
	schedule *scheduleService.ScheduleService,
	auth *authService.AuthService,
	user *userService.UserService,
	cfg *config.Config,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Route(
		"/api/rooms/create", func(r chi.Router) {
			r.With(
				middlewares.AuthMiddleware(cfg.JWTSecret),
				middlewares.RoleMiddleware("admin"),
			).Post("/", createRoomHandler.New(rooms, log))
		},
	)

	r.Route(
		"/api/rooms/list", func(r chi.Router) {
			r.With(
				middlewares.AuthMiddleware(cfg.JWTSecret),
				middlewares.RoleMiddleware("user", "admin"),
			).Get("/", GetRoomsListHandler.New(rooms, log))
		},
	)

	r.Route(
		"/api/rooms/{roomId}/slots/list", func(r chi.Router) {
			r.With(
				middlewares.AuthMiddleware(cfg.JWTSecret),
				middlewares.RoleMiddleware("user", "admin"),
			).Get("/", GetSlotsHandler.New(slots, log))
		},
	)

	r.Route(
		"/api/booking/create", func(r chi.Router) {
			r.With(
				middlewares.AuthMiddleware(cfg.JWTSecret),
				middlewares.RoleMiddleware("user"),
			).Post("/", BookingCreateHandler.New(booking, log))
		},
	)

	r.Route(
		"/api/bookings", func(r chi.Router) {
			r.Put("/", BookingCancelHandler.New(booking, log))
		},
	)

	r.Route(
		"/api/bookings/my", func(r chi.Router) {
			r.With(
				middlewares.AuthMiddleware(cfg.JWTSecret),
				middlewares.RoleMiddleware("user"),
			).Get("/", MyBookingHandler.New(booking, log))
		},
	)

	r.Route(
		"/api/bookings/{bookingId}/cancel", func(r chi.Router) {
			r.With(
				middlewares.AuthMiddleware(cfg.JWTSecret),
				middlewares.RoleMiddleware("user"),
			).Put("/", BookingCancelHandler.New(booking, log))
		},
	)

	r.Route(
		"/api/bookings/list", func(r chi.Router) {
			r.With(
				middlewares.AuthMiddleware(cfg.JWTSecret),
				middlewares.RoleMiddleware("admin"),
			).Get("/", FindBookingHandler.New(booking, log))
		},
	)

	r.Route(
		"/api/rooms/{roomId}/schedule/create", func(r chi.Router) {
			r.With(
				middlewares.AuthMiddleware(cfg.JWTSecret),
				middlewares.RoleMiddleware("admin"),
			).Post("/", createScheduleHandler.New(schedule, slots, log))
		},
	)

	r.Route(
		"/api/dummyLogin", func(r chi.Router) {
			r.Post("/", DummyLoginHandler.New(auth, user, log))
		},
	)

	r.Route(
		"/api/_info", func(r chi.Router) {
			r.Get("/", InfoHandler.New())
		},
	)

	r.Route(
		"/api/register", func(r chi.Router) {
			r.Post("/", RegisterHandler.New(user, log))
		},
	)

	r.Route(
		"/api/login", func(r chi.Router) {
			r.Post("/", LoginHandler.New(user, auth, log))
		},
	)

	r.Get("/docs/*", httpSwagger.WrapHandler)

	return r
}
