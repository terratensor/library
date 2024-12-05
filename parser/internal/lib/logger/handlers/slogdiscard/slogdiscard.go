// Package slogdiscard Помимо стандартных реализаций, нам все же придется написать одну свою —
// DiscardHandler. В таком виде логгер будет игнорировать все сообщения, которые мы в него
// отправляем, — это понадобится в тестах. Создадим пакет slogdiscard
// и имплементируем в нем интерфейс slog.Handler
package slogdiscard

import (
	"context"
	"log/slog"
)

// NewDiscardLogger creates a new logger that discards all log messages.
func NewDiscardLogger() *slog.Logger {
	// Create a new discard handler
	handler := NewDiscardHandler()

	// Create a new logger with the discard handler
	logger := slog.New(handler)

	// Return the new logger
	return logger
}

type DiscardHandler struct {
}

func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

func (h *DiscardHandler) Handle(_ context.Context, _ slog.Record) error {
	// Просто игнорируем запись журнала
	return nil
}

func (h *DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// Возвращает тот же обработчик, так как нет атрибутов для сохранения
	return h
}

func (h *DiscardHandler) WithGroup(_ string) slog.Handler {
	// Возвращает тот же обработчик, так как нет группы для сохранения
	return h
}

func (h *DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	// Всегда возвращает false, так как запись журнала игнорируется
	return false
}
