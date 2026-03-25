package e2e

import (
	"backend-test/internal/config"
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
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestMain запускается перед всеми тестами
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

// setupTestServer создаёт тестовый сервер с реальной БД (testcontainers)
func setupTestServerWithDB(t *testing.T) (*httptest.Server, func()) {
	ctx := context.Background()

	// Запускаем PostgreSQL в Docker контейнере
	postgresContainer, err := testcontainers.GenericContainer(
		ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "postgres:18.1-alpine3.23",
				ExposedPorts: []string{"5432/tcp"},
				Env: map[string]string{
					"POSTGRES_USER":     "postgres",
					"POSTGRES_PASSWORD": "12345678",
					"POSTGRES_DB":       "testdb",
				},
				WaitingFor: wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(60 * time.Second),
			},
			Started: true,
		},
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %s", err)
	}

	// Получаем хост и порт контейнера
	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %s", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get container port: %s", err)
	}

	// Формируем строку подключения
	dbURL := fmt.Sprintf("postgres://postgres:12345678@%s:%s/testdb?sslmode=disable", host, port.Port())

	// Применяем миграции
	m, err := migrate.New(
		"file://../../migrations",
		dbURL,
	)
	if err != nil {
		t.Fatalf("Failed to create migrate instance: %s", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("Failed to run migrations: %s", err)
	}

	// Создаём конфиг для тестов
	cfg := &config.Config{
		Env:            "test",
		StorageURL:     dbURL,
		JWTSecret:      "test-secret-key-for-e2e-tests",
		ExpirationTime: 24 * time.Hour,
		HTTPServer: config.HTTPServer{
			URL: "localhost:8080",
		},
	}

	// Инициализируем хранилище
	str, err := db.New(cfg.StorageURL)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %s", err)
	}

	log := slog.Default()

	RoomsRepo := roomsRepository.NewRoomsRepository(str)
	SlotsRepo := slotsRepository.NewSlotRepository(str)
	BookingRepo := bookingRepository.New(str)
	ScheduleRepo := scheduleRepository.New(str)
	UserRepo := usersRepository.New(str)

	// Инициализируем сервисы
	RoomServ := roomService.NewRoomService(RoomsRepo, log)
	SlotServ := slotsService.New(SlotsRepo, log)
	bookingServ := bookingService.New(BookingRepo, log)
	scheduleServ := scheduleService.New(ScheduleRepo, log)
	auth := authService.New(cfg.JWTSecret, cfg.ExpirationTime)
	userServ := userService.New(UserRepo, log)

	router := initTestRouter(log, RoomServ, SlotServ, bookingServ, scheduleServ, auth, userServ, cfg)

	srv := httptest.NewServer(router)

	cleanup := func() {
		srv.Close()
		str.DB.Close()
		postgresContainer.Terminate(ctx)
	}

	return srv, cleanup
}
