package storage

import "errors"

var (
	ErrRoomNotFound      = errors.New("переговорка не найдена")
	ErrSlotNotFound      = errors.New("слот не найден")
	ErrSlotAlreadyBooked = errors.New("слот уже занят")
	ErrBookingNotFound   = errors.New("бронь не найдена")
	ErrScheduleExists    = errors.New("расписание уже существует")
	ErrBookingForbidden  = errors.New("cannot cancel another user's booking")
	ErrEmailAlreadyExist = errors.New("почта уже занята")
	ErrUserNotFound      = errors.New("пользователь не найден")
)
