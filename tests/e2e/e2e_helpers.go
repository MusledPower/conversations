package e2e

import (
	"backend-test/internal/config"
	"backend-test/internal/http-server/handlers/BookingCancelHandler"
	"backend-test/internal/http-server/handlers/BookingCreateHandler"
	"backend-test/internal/http-server/handlers/DummyLoginHandler"
	"backend-test/internal/http-server/handlers/FindBookingHandler"
	"backend-test/internal/http-server/handlers/GetRoomsListHandler"
	"backend-test/internal/http-server/handlers/GetSlotsHandler"
	"backend-test/internal/http-server/handlers/MyBookingHandler"
	"backend-test/internal/http-server/handlers/createRoomHandler"
	"backend-test/internal/http-server/handlers/createScheduleHandler"
	"backend-test/internal/http-server/middlewares"
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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// Request/Response типы

type CreateRoomRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Capacity    *int    `json:"capacity,omitempty"`
}

type CreateScheduleRequest struct {
	DaysOfWeek []int  `json:"days_of_week"`
	Start      string `json:"start"`
	End        string `json:"end"`
}

type CreateBookingRequest struct {
	UserID         uuid.UUID `json:"user_id"`
	SlotID         uuid.UUID `json:"slot_id"`
	ConferenceLink string    `json:"conference_link"`
}

type AuthResponse struct {
	Status int    `json:"status"`
	Token  string `json:"token"`
	Error  string `json:"error,omitempty"`
}

type RoomResponse struct {
	Status int    `json:"status"`
	Room   *Room  `json:"room,omitempty"`
	Error  string `json:"error,omitempty"`
}

type Room struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Capacity    *int      `json:"capacity,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoomListResponse struct {
	Status int    `json:"status"`
	Rooms  []Room `json:"rooms,omitempty"`
	Error  string `json:"error,omitempty"`
}

type ScheduleResponse struct {
	Status   int       `json:"status"`
	Schedule *Schedule `json:"schedule,omitempty"`
	Error    string    `json:"error,omitempty"`
}

type Schedule struct {
	ID         uuid.UUID `json:"id"`
	RoomID     uuid.UUID `json:"room_id"`
	DaysOfWeek []int     `json:"days_of_week"`
	Start      string    `json:"start"`
	End        string    `json:"end"`
	CreatedAt  time.Time `json:"created_at"`
}

type SlotsResponse struct {
	Status int    `json:"status"`
	Slots  []Slot `json:"slots,omitempty"`
	Error  string `json:"error,omitempty"`
}

type Slot struct {
	ID        uuid.UUID  `json:"id"`
	RoomID    uuid.UUID  `json:"room_id"`
	Start     time.Time  `json:"start"`
	End       time.Time  `json:"end"`
	IsBooked  bool       `json:"is_booked"`
	BookingID *uuid.UUID `json:"booking_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type BookingResponse struct {
	Status  int      `json:"status"`
	Booking *Booking `json:"room,omitempty"`
	Error   string   `json:"error,omitempty"`
}

type Booking struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	SlotID         uuid.UUID `json:"slot_id"`
	ConferenceLink string    `json:"conference_link"`
	IsCancelled    bool      `json:"is_cancelled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type BookingListResponse struct {
	Status     int        `json:"status"`
	Booking    []Booking  `json:"room,omitempty"` // В вашем API это "room"
	Pagination Pagination `json:"pagination,omitempty"`
	Error      string     `json:"error,omitempty"`
}

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

type E2ETestSuite struct {
	server  *httptest.Server
	client  *http.Client
	storage *db.Storage
}

func setupSuite(t *testing.T) *E2ETestSuite {
	testDBURL := os.Getenv("TEST_DB_URL")
	if testDBURL == "" {
		testDBURL = "postgres://postgres:12345678@localhost:5433/testdb?sslmode=disable"
	}

	t.Logf("🔌 Подключаемся к тестовой БД: %s", testDBURL)

	cfg := &config.Config{
		Env:            "test",
		StorageURL:     testDBURL,
		JWTSecret:      "test-secret-key-for-e2e-tests",
		ExpirationTime: 24 * time.Hour,
		HTTPServer: config.HTTPServer{
			URL: ":8080",
		},
	}

	// Инициализируем хранилище
	str, err := db.New(cfg.StorageURL)
	if err != nil {
		t.Fatalf("❌ Не удалось подключиться к БД: %v", err)
	}

	t.Log("✅ Подключение к БД успешно")

	// Создаём logger
	log := slog.New(
		slog.NewTextHandler(
			os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelError, // Минимальные логи для чистого вывода тестов
			},
		),
	)

	// Инициализируем репозитории
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

	// Создаём роутер (копия из app.go)
	router := initTestRouter(log, RoomServ, SlotServ, bookingServ, scheduleServ, auth, userServ, cfg)

	// Создаём тестовый сервер
	server := httptest.NewServer(router)

	t.Logf("🚀 Тестовый сервер запущен: %s", server.URL)

	return &E2ETestSuite{
		server:  server,
		client:  &http.Client{Timeout: 10 * time.Second},
		storage: str,
	}
}

// initTestRouter - копия функции initRouter из app.go
func initTestRouter(
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

	return r
}

func (s *E2ETestSuite) Cleanup() {
	if s.server != nil {
		s.server.Close()
	}
	if s.storage != nil {
		s.storage.DB.Close()
	}
}

func (s *E2ETestSuite) DoRequest(method, path, token string, body []byte) (*http.Response, error) {
	url := s.server.URL + path

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return s.client.Do(req)
}

func (s *E2ETestSuite) GetAuthToken(userID uuid.UUID, role string) (string, error) {
	body := map[string]interface{}{
		"user_id": userID,
		"role":    role,
	}
	bodyBytes, _ := json.Marshal(body)

	resp, err := s.DoRequest("POST", "/api/dummyLogin", "", bodyBytes)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyContent, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get token: status %d, body: %s", resp.StatusCode, string(bodyContent))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}

	if authResp.Token == "" {
		return "", fmt.Errorf("empty token in response")
	}

	return authResp.Token, nil
}

func (s *E2ETestSuite) CreateRoom(token string, req CreateRoomRequest) (*Room, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.DoRequest("POST", "/api/rooms/create", token, bodyBytes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var roomResp RoomResponse
	if err := json.Unmarshal(bodyContent, &roomResp); err != nil {
		return nil, fmt.Errorf(
			"failed to decode room response: %w, body: %s",
			err,
			string(bodyContent),
		)
	}

	if resp.StatusCode != http.StatusCreated &&
		roomResp.Status != http.StatusCreated {

		return nil, fmt.Errorf(
			"failed to create room: http_status=%d api_status=%d body=%s",
			resp.StatusCode,
			roomResp.Status,
			string(bodyContent),
		)
	}

	return roomResp.Room, nil
}

func (s *E2ETestSuite) GetRoomList(token string) ([]Room, error) {
	resp, err := s.DoRequest("GET", "/api/rooms/list", token, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyContent, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to get room list: status %d, body: %s", resp.StatusCode, string(bodyContent))
	}

	var listResp RoomListResponse
	if err := json.Unmarshal(bodyContent, &listResp); err != nil {
		return nil, fmt.Errorf("failed to decode room list: %w, body: %s", err, string(bodyContent))
	}

	return listResp.Rooms, nil
}

func (s *E2ETestSuite) CreateSchedule(token string, roomID uuid.UUID, req CreateScheduleRequest) (*Schedule, error) {
	bodyBytes, _ := json.Marshal(req)

	path := fmt.Sprintf("/api/rooms/%s/schedule/create", roomID)
	resp, err := s.DoRequest("POST", path, token, bodyBytes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyContent, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create schedule: status %d, body: %s", resp.StatusCode, string(bodyContent))
	}

	var schedResp ScheduleResponse
	if err := json.Unmarshal(bodyContent, &schedResp); err != nil {
		return nil, fmt.Errorf("failed to decode schedule: %w, body: %s", err, string(bodyContent))
	}

	return schedResp.Schedule, nil
}

func (s *E2ETestSuite) GetSlots(token string, roomID uuid.UUID, date time.Time) ([]Slot, error) {
	path := fmt.Sprintf("/api/rooms/%s/slots/list?date=%s", roomID, date.Format("2006-01-02"))

	resp, err := s.DoRequest("GET", path, token, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyContent, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to get slots: status %d, body: %s", resp.StatusCode, string(bodyContent))
	}

	var slotsResp SlotsResponse
	if err := json.Unmarshal(bodyContent, &slotsResp); err != nil {
		return nil, fmt.Errorf("failed to decode slots: %w, body: %s", err, string(bodyContent))
	}

	return slotsResp.Slots, nil
}

func (s *E2ETestSuite) CreateBooking(token string, req CreateBookingRequest) (*Booking, error) {
	bodyBytes, _ := json.Marshal(req)

	resp, err := s.DoRequest("POST", "/api/booking/create", token, bodyBytes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyContent, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create booking: status %d, body: %s", resp.StatusCode, string(bodyContent))
	}

	var bookingResp BookingResponse
	if err := json.Unmarshal(bodyContent, &bookingResp); err != nil {
		return nil, fmt.Errorf("failed to decode booking: %w, body: %s", err, string(bodyContent))
	}

	return bookingResp.Booking, nil
}

func (s *E2ETestSuite) GetMyBookings(token string, userID uuid.UUID) ([]Booking, error) {
	path := fmt.Sprintf("/api/bookings/my?userId=%s", userID)

	resp, err := s.DoRequest("GET", path, token, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyContent, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get my bookings: status %d, body: %s", resp.StatusCode, string(bodyContent))
	}

	var listResp BookingListResponse
	if err := json.Unmarshal(bodyContent, &listResp); err != nil {
		return nil, fmt.Errorf("failed to decode bookings: %w, body: %s", err, string(bodyContent))
	}

	return listResp.Booking, nil
}

func (s *E2ETestSuite) GetAllBookings(token string, page, pageSize int) ([]Booking, error) {
	path := fmt.Sprintf("/api/bookings/list?page=%d&pageSize=%d", page, pageSize)

	resp, err := s.DoRequest("GET", path, token, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyContent, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get all bookings: status %d, body: %s", resp.StatusCode, string(bodyContent))
	}

	var listResp BookingListResponse
	if err := json.Unmarshal(bodyContent, &listResp); err != nil {
		return nil, fmt.Errorf("failed to decode bookings: %w, body: %s", err, string(bodyContent))
	}

	return listResp.Booking, nil
}

func (s *E2ETestSuite) CancelBooking(token string, bookingID, userID uuid.UUID) error {
	body := map[string]interface{}{
		"booking_id": bookingID,
		"user_id":    userID,
	}
	bodyBytes, _ := json.Marshal(body)

	path := fmt.Sprintf("/api/bookings/%s/cancel", bookingID)
	resp, err := s.DoRequest("PUT", path, token, bodyBytes)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyContent, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to cancel booking: status %d, body: %s", resp.StatusCode, string(bodyContent))
	}

	return nil
}
