package slogger

import (
	"log/slog"
	"os"
)

// newFileHandler создает и настраивает обработчик логирования (slog.Handler) для записи в файл.
//
// Функция анализирует переданную конфигурацию cfg.Format и возвращает:
// - JSON-обработчик (slog.JSONHandler), если указан формат FormatJSON.
// - Текстовый обработчик (slog.TextHandler) в формате "ключ=значение", если указан FormatText
//   или если передан неизвестный/неподдерживаемый формат (поведение по умолчанию).
//
// Для обоих типов обработчиков устанавливается минимальный уровень логирования из cfg.Level.

func newFileHandler(cfg FileConfig, file *os.File) slog.Handler {
	switch cfg.Format {
	case FormatJSON:
		return slog.NewJSONHandler(file, &slog.HandlerOptions{
			Level: cfg.Level.SlogLevel(),
		})
	case FormatText:
		return slog.NewTextHandler(file, &slog.HandlerOptions{
			Level: cfg.Level.SlogLevel(),
		})
	}

	return slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: cfg.Level.SlogLevel(),
	})
}

// newStdoutHandler создает и настраивает обработчик логирования (slog.Handler) для вывода в стандартный поток (консоль).
//
// Функция гибко определяет целевой поток для записи:
// - Если в конфигурации явно задан cfg.Writer, вывод направляется в него.
// - Если cfg.Writer не указан (равен nil), в качестве потока по умолчанию используется os.Stdout.
//
// В зависимости от настройки cfg.Format возвращается либо JSONHandler, либо TextHandler
// с ограничением по уровню логирования из cfg.Level. Текстовый формат является поведением по умолчанию.

func newStdoutHandler(cfg StdoutConfig) slog.Handler {

	writer := cfg.Writer

	if writer == nil {
		writer = os.Stdout
	}

	switch cfg.Format {
	case FormatJSON:
		return slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: cfg.Level.SlogLevel(),
		})
	case FormatText:
		return slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level: cfg.Level.SlogLevel(),
		})
	}

	return slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: cfg.Level.SlogLevel(),
	})
}
