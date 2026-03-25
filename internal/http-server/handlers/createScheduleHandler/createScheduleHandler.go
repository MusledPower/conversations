package createScheduleHandler

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/lib/sl"
	"backend-test/internal/services/scheduleService"
	"backend-test/internal/services/slotsService"
	"backend-test/internal/storage"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const InvalidRequest = "Неверный запрос"
const InternalServerError = "Внутренняя ошибка сервера"

type Request struct {
	DaysOfWeek []int  `json:"days_of_week" validate:"required,min=1,dive,min=1,max=7"`
	Start      string `json:"start" validate:"required,datetime=15:04"`
	End        string `json:"end" validate:"required,datetime=15:04"`
}

type Response struct {
	Status   int              `json:"status"`
	Error    string           `json:"error,omitempty" validate:"error,omitempty"`
	Schedule *models.Schedule `json:"schedule,omitempty"`
}

type TimeSlot struct {
	Start time.Time
	End   time.Time
}

// CreateSchedule godoc
// @Summary Создать расписание для переговорки
// @Description Доступно только admin. После создания изменить нельзя.
// @Tags Schedules
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param roomId path string true "UUID переговорки"
// @Param schedule body Request true "Данные расписания"
// @Success 201 {object} Response "Расписание создано"
// @Failure 400 {object} Response "Неверный запрос (например daysOfWeek вне 1-7)"
// @Failure 401 {object} Response "Не авторизован"
// @Failure 403 {object} Response "Доступ запрещён"
// @Failure 404 {object} Response "Переговорка не найдена"
// @Failure 409 {object} Response "Расписание уже существует"
// @Failure 500 {object} Response "Внутренняя ошибка сервера"
// @Router /api/rooms/{roomId}/schedule/create [post]
func New(s *scheduleService.ScheduleService, ss *slotsService.SlotsService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.BookingCreateHandler"

		log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		roomID, err := uuid.Parse(chi.URLParam(r, "roomId"))
		if err != nil {
			log.Error(InternalServerError, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})

			return
		}

		var req Request

		err = render.Decode(r, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, Response{Status: http.StatusBadRequest, Error: InvalidRequest})

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error(InvalidRequest, sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, Response{Status: http.StatusBadRequest, Error: InvalidRequest})

			return
		}

		start, err := time.Parse("15:04", req.Start)
		if err != nil {
			log.Error(InternalServerError, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})

			return
		}
		end, err := time.Parse("15:04", req.End)
		if err != nil {
			log.Error(InternalServerError, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})

			return
		}

		schedule, err := s.CreateSchedule(roomID, req.DaysOfWeek, start, end)
		if err != nil {
			if errors.Is(err, storage.ErrScheduleExists) {
				log.Error(storage.ErrScheduleExists.Error(), sl.Err(err))
				w.WriteHeader(http.StatusConflict)

				render.JSON(w, r, Response{Status: http.StatusConflict, Error: storage.ErrScheduleExists.Error()})

				return
			}

			if errors.Is(err, storage.ErrSlotNotFound) {
				log.Error(storage.ErrSlotNotFound.Error(), sl.Err(err))
				w.WriteHeader(http.StatusNotFound)

				render.JSON(w, r, Response{Status: http.StatusNotFound, Error: storage.ErrSlotNotFound.Error()})

				return
			}

			if errors.Is(err, storage.ErrRoomNotFound) {
				log.Error(storage.ErrScheduleExists.Error(), sl.Err(err))
				w.WriteHeader(http.StatusNotFound)

				render.JSON(w, r, Response{Status: http.StatusNotFound, Error: storage.ErrRoomNotFound.Error()})

				return
			}

			log.Error(InternalServerError, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})

			return
		}

		slots := generateSlots(req.DaysOfWeek, start, end, time.Now(), time.Now().AddDate(0, 0, 7))

		for _, slot := range slots {
			err := ss.CreateSlotsService(roomID, slot.Start, slot.End)
			if err != nil {
				log.Error(InternalServerError, sl.Err(err))
				w.WriteHeader(http.StatusInternalServerError)

				render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})

				return
			}
		}
		w.WriteHeader(http.StatusCreated)

		render.JSON(w, r, Response{Status: http.StatusCreated, Schedule: schedule})

		return
	}
}

func generateSlots(
	daysOfWeek []int,
	startTime, endTime time.Time,
	rangeStart, rangeEnd time.Time,
) []TimeSlot {
	var slots []TimeSlot

	daySet := make(map[time.Weekday]bool)

	for _, d := range daysOfWeek {
		daySet[time.Weekday(d%7)] = true
	}

	for d := rangeStart; !d.After(rangeEnd); d = d.AddDate(0, 0, 1) {

		if !daySet[d.Weekday()] {
			continue
		}

		current := time.Date(d.Year(), d.Month(), d.Day(), startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
		endOfDay := time.Date(d.Year(), d.Month(), d.Day(), endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

		for current.Before(endOfDay) {
			slotEnd := current.Add(30 * time.Minute)
			if slotEnd.After(endOfDay) {
				slotEnd = endOfDay
			}

			slots = append(
				slots, TimeSlot{
					Start: current,
					End:   slotEnd,
				},
			)

			current = slotEnd
		}
	}

	return slots
}
