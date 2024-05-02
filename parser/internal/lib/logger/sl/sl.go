// Package sl (сокращенно от slog), в который добавим некоторые функции для работы с логгером.
// Они пригодятся в будущем.
package sl

import "log/slog"

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
