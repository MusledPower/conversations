package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookingCompleteFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	suite := setupSuite(t)
	defer t.Log("=== Начало E2E теста: Полный сценарий бронирования ===")

	adminID := uuid.New()
	userID := uuid.New()

	t.Logf("Admin ID: %s", adminID)
	t.Logf("User ID: %s", userID)

	t.Run(
		"Step 1: Получение JWT токенов", func(t *testing.T) {
			adminToken, err := suite.GetAuthToken(adminID, "admin")
			require.NoError(t, err, "Не удалось получить токен admin")
			require.NotEmpty(t, adminToken)
			t.Logf("✓ Admin токен получен")

			userToken, err := suite.GetAuthToken(userID, "user")
			require.NoError(t, err, "Не удалось получить токен user")
			require.NotEmpty(t, userToken)
			t.Logf("✓ User токен получен")
		},
	)

	adminToken, _ := suite.GetAuthToken(adminID, "admin")
	userToken, _ := suite.GetAuthToken(userID, "user")

	// Шаг 2: Создание комнаты
	var roomID uuid.UUID
	t.Run(
		"Step 2: Admin создаёт комнату", func(t *testing.T) {
			room, err := suite.CreateRoom(
				adminToken, CreateRoomRequest{
					Name:        "Переговорная 1",
					Description: stringPtr("Большая комната для встреч"),
					Capacity:    intPtr(10),
				},
			)
			require.NoError(t, err)
			require.NotNil(t, room)
			require.NotEqual(t, uuid.Nil, room.ID)

			roomID = room.ID
			t.Logf("✓ Комната создана: %s (ID: %s)", room.Name, room.ID)
		},
	)

	// Шаг 3: Создание расписания
	t.Run(
		"Step 3: Admin создаёт расписание для комнаты", func(t *testing.T) {
			schedule, err := suite.CreateSchedule(
				adminToken, roomID, CreateScheduleRequest{
					DaysOfWeek: []int{1, 2, 3, 4, 5}, // Пн-Пт
					Start:      "09:00",
					End:        "18:00",
				},
			)
			require.NoError(t, err)
			require.NotNil(t, schedule)
			t.Logf("✓ Расписание создано: %v, %s-%s", schedule.DaysOfWeek, schedule.Start, schedule.End)

			// Даём время на создание слотов
			time.Sleep(200 * time.Millisecond)
		},
	)

	// Шаг 4: Получение списка комнат
	t.Run(
		"Step 4: User получает список доступных комнат", func(t *testing.T) {
			rooms, err := suite.GetRoomList(userToken)
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(rooms), 1, "Должна быть хотя бы одна комната")

			foundRoom := false
			for _, room := range rooms {
				if room.ID == roomID {
					foundRoom = true
					t.Logf("✓ Найдена созданная комната в списке: %s", room.Name)
				}
			}
			require.True(t, foundRoom, "Созданная комната должна быть в списке")
		},
	)

	// Шаг 5: Получение свободных слотов
	var slotID uuid.UUID
	t.Run(
		"Step 5: User получает свободные слоты на завтра", func(t *testing.T) {
			tomorrow := time.Now().AddDate(0, 0, 1)

			slots, err := suite.GetSlots(userToken, roomID, tomorrow)
			require.NoError(t, err)
			require.Greater(t, len(slots), 0, "Должны быть доступные слоты")

			slotID = slots[0].ID
			t.Logf("✓ Найдено %d свободных слотов на %s", len(slots), tomorrow.Format("2006-01-02"))
			t.Logf(
				"  Первый слот: %s - %s",
				slots[0].Start.Format("15:04"),
				slots[0].End.Format("15:04"),
			)
		},
	)

	// Шаг 6: Создание бронирования
	var bookingID uuid.UUID
	t.Run(
		"Step 6: User бронирует слот", func(t *testing.T) {
			booking, err := suite.CreateBooking(
				userToken, CreateBookingRequest{
					UserID:         userID,
					SlotID:         slotID,
					ConferenceLink: "https://meet.google.com/abc-def-ghi",
				},
			)
			require.NoError(t, err)
			require.NotNil(t, booking)
			require.Equal(t, slotID, booking.SlotID)

			bookingID = booking.ID
			t.Logf("✓ Бронирование создано: ID=%s, Link=%s", booking.ID, booking.ConferenceLink)
		},
	)

	// Шаг 7: Проверка своих бронирований
	t.Run(
		"Step 7: User проверяет свои бронирования", func(t *testing.T) {
			bookings, err := suite.GetMyBookings(userToken, userID)
			require.NoError(t, err)
			require.Len(t, bookings, 1, "Должно быть ровно одно бронирование")
			require.Equal(t, bookingID, bookings[0].ID)
			t.Logf("✓ Найдено бронирование в личном списке")
		},
	)

	// Шаг 8: Попытка повторного бронирования (должна провалиться)
	t.Run(
		"Step 8: Попытка повторного бронирования того же слота", func(t *testing.T) {
			_, err := suite.CreateBooking(
				userToken, CreateBookingRequest{
					UserID:         userID,
					SlotID:         slotID,
					ConferenceLink: "https://meet.google.com/xyz-123-456",
				},
			)
			require.Error(t, err, "Повторное бронирование должно быть запрещено")
			t.Logf("✓ Повторное бронирование корректно заблокировано")
		},
	)

	// Шаг 9: Admin смотрит все бронирования
	t.Run(
		"Step 9: Admin получает список всех бронирований", func(t *testing.T) {
			bookings, err := suite.GetAllBookings(adminToken, 1, 20)
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(bookings), 1, "Должно быть хотя бы одно бронирование")

			found := false
			for _, b := range bookings {
				if b.ID == bookingID {
					found = true
				}
			}
			require.True(t, found, "Созданное бронирование должно быть в общем списке")
			t.Logf("✓ Бронирование видно в общем списке (%d всего)", len(bookings))
		},
	)

	// Шаг 10: Отмена бронирования
	t.Run(
		"Step 10: User отменяет бронирование", func(t *testing.T) {
			err := suite.CancelBooking(userToken, bookingID, userID)
			require.NoError(t, err)
			t.Logf("✓ Бронирование успешно отменено")
		},
	)

	// Шаг 11: Проверка, что бронирование отменено
	t.Run(
		"Step 11: Проверка отмены бронирования", func(t *testing.T) {
			bookings, err := suite.GetMyBookings(userToken, userID)
			require.NoError(t, err)

			// В зависимости от логики вашего приложения:
			// Либо бронирования нет в списке вообще
			// Либо оно есть, но с флагом is_cancelled=true

			activeBookings := 0
			for _, b := range bookings {
				if !b.IsCancelled {
					activeBookings++
				}
			}

			require.Equal(t, 0, activeBookings, "Не должно быть активных бронирований после отмены")
			t.Logf("✓ Активных бронирований: 0")
		},
	)

	t.Log("=== ✓ E2E тест успешно завершён ===")
}

// TestUnauthorizedAccess проверяет, что endpoints защищены авторизацией
func TestUnauthorizedAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	suite := setupSuite(t)
	defer suite.Cleanup()

	tests := []struct {
		name       string
		method     string
		url        string
		body       interface{}
		wantStatus int
		wantError  string
	}{
		{
			name:       "Создание комнаты без токена",
			method:     "POST",
			url:        "/api/rooms/create",
			body:       map[string]string{"name": "Test"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Список комнат без токена",
			method:     "GET",
			url:        "/api/rooms/list",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Создание бронирования без токена",
			method:     "POST",
			url:        "/api/booking/create",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				var body []byte
				if tt.body != nil {
					body, _ = json.Marshal(tt.body)
				}

				resp, err := suite.DoRequest(tt.method, tt.url, "", body)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(
					t, tt.wantStatus, resp.StatusCode,
					"Endpoint должен требовать авторизацию",
				)
			},
		)
	}
}

func TestRoleBasedAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	suite := setupSuite(t)
	defer suite.Cleanup()

	userID := uuid.New()
	userToken, _ := suite.GetAuthToken(userID, "user")

	t.Run(
		"User не может создавать комнаты", func(t *testing.T) {
			_, err := suite.CreateRoom(
				userToken, CreateRoomRequest{
					Name: "Unauthorized Room",
				},
			)
			require.Error(t, err, "User не должен иметь права создавать комнаты")
		},
	)

	t.Run(
		"User не может создавать расписания", func(t *testing.T) {
			roomID := uuid.New()
			_, err := suite.CreateSchedule(
				userToken, roomID, CreateScheduleRequest{
					DaysOfWeek: []int{1, 2, 3},
					Start:      "09:00",
					End:        "18:00",
				},
			)
			require.Error(t, err, "User не должен иметь права создавать расписания")
		},
	)

	t.Run(
		"User не может получать список всех бронирований", func(t *testing.T) {
			_, err := suite.GetAllBookings(userToken, 1, 20)
			require.Error(t, err, "User не должен видеть все бронирования")
		},
	)
}

func TestConcurrentBookings(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	suite := setupSuite(t)
	defer suite.Cleanup()

	// Подготовка
	adminID := uuid.New()
	adminToken, _ := suite.GetAuthToken(adminID, "admin")

	room, err := suite.CreateRoom(
		adminToken,
		CreateRoomRequest{
			Name:     "Concurrent Test Room",
			Capacity: intPtr(1),
		},
	)
	require.NoError(t, err)
	require.NotNil(t, room)

	suite.CreateSchedule(
		adminToken,
		room.ID,
		CreateScheduleRequest{
			DaysOfWeek: []int{1, 2, 3, 4, 5},
			Start:      "09:00",
			End:        "18:00",
		},
	)

	time.Sleep(200 * time.Millisecond)

	// Получаем слот
	tomorrow := time.Now().AddDate(0, 0, 1)
	for tomorrow.Weekday() == time.Saturday || tomorrow.Weekday() == time.Sunday {
		tomorrow = tomorrow.AddDate(0, 0, 1)
	}

	slots, _ := suite.GetSlots(adminToken, room.ID, tomorrow)
	require.Greater(t, len(slots), 0)

	slotID := slots[0].ID

	// 5 пользователей пытаются одновременно забронировать
	const numUsers = 5
	results := make(chan error, numUsers)

	for i := 0; i < numUsers; i++ {
		go func(userNum int) {
			userID := uuid.New()
			token, _ := suite.GetAuthToken(userID, "user")

			_, err := suite.CreateBooking(
				token, CreateBookingRequest{
					UserID:         userID,
					SlotID:         slotID,
					ConferenceLink: fmt.Sprintf("https://meet.google.com/test-%d", userNum),
				},
			)

			results <- err
		}(i)
	}

	// Собираем результаты
	successCount := 0
	for i := 0; i < numUsers; i++ {
		err := <-results
		if err == nil {
			successCount++
		}
	}

	// Только одно бронирование должно быть успешным
	assert.Equal(
		t, 1, successCount,
		"При параллельных запросах только один должен забронировать слот",
	)
}

// TestInputValidation проверяет валидацию входных данных
func TestInputValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	suite := setupSuite(t)
	defer suite.Cleanup()

	adminID := uuid.New()
	adminToken, _ := suite.GetAuthToken(adminID, "admin")

	tests := []struct {
		name        string
		requestFunc func() error
		wantError   bool
	}{
		{
			name: "Создание комнаты без имени",
			requestFunc: func() error {
				_, err := suite.CreateRoom(
					adminToken, CreateRoomRequest{
						Description: stringPtr("No name"),
					},
				)
				return err
			},
			wantError: true,
		},
		{
			name: "Создание расписания с невалидными днями",
			requestFunc: func() error {
				roomID := uuid.New()
				_, err := suite.CreateSchedule(
					adminToken, roomID, CreateScheduleRequest{
						DaysOfWeek: []int{8, 9}, // Дни недели 1-7
						Start:      "09:00",
						End:        "18:00",
					},
				)
				return err
			},
			wantError: true,
		},
		{
			name: "Создание расписания с невалидным временем",
			requestFunc: func() error {
				roomID := uuid.New()
				_, err := suite.CreateSchedule(
					adminToken, roomID, CreateScheduleRequest{
						DaysOfWeek: []int{1, 2, 3},
						Start:      "25:00",
						End:        "18:00",
					},
				)
				return err
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := tt.requestFunc()
				if tt.wantError {
					require.Error(t, err, "Должна быть ошибка валидации")
				} else {
					require.NoError(t, err)
				}
			},
		)
	}
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
