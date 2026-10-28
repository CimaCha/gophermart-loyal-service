package slogger

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Closer отвечает за безопасное закрытие всех файлов логов, открытых при инициализации.
// Использует sync.Once, чтобы гарантировать, что закрытие произойдет ровно один раз,
// и собирает все возникшие ошибки с помощью errors.Join.
type Closer struct {
	files []*os.File // Список указателей на открытые файлы логов
	once  sync.Once  // Гарантирует потокобезопасный и однократный вызов Close
	errs  error      // Переменная для аккумулирования ошибок при закрытии файлов
}

var (
	ErrNoLevelOrDirectory = errors.New("missing directory or level")
)

// New инициализирует новый экземпляр *slog.Logger на основе переданной конфигурации Config.
//
// Функция возвращает:
// - *slog.Logger: готовый объект логера (который может писать одновременно в консоль и несколько файлов).
// - *Closer: структуру для освобождения ресурсов (закрытия файлов), которую нужно вызвать при завершении работы приложения.
// - error: ошибку инициализации, если не удалось создать директорию или открыть файлы.
func New(cfg *Config) (*slog.Logger, *Closer, error) {

	handler, closer, err := buildHandler(cfg)

	if err != nil {
		return nil, nil, err
	}

	return slog.New(handler), closer, err
}

// MustNew работает аналогично функции New, но вызывает panic в случае возникновения ошибки.
//
// Рекомендуется использовать на самом старте приложения (например, в функции main),
// когда невозможность настроить логирование является критической ошибкой инициализации.
func MustNew(cfg *Config) (*slog.Logger, *Closer) {
	slogger, closer, err := New(cfg)
	if err != nil {
		panic(err)
	}
	return slogger, closer
}

// buildHandler — это внутренняя функция сборки, которая объединяет все активные потоки вывода.
// Она обрабатывает вывод в Stdout и перебирает список конфигураций для файлов.
//
// В случае ошибки на этапе создания файлов, функция гарантирует закрытие уже открытых в этом цикле дескрипторов (через defer).
func buildHandler(cfg *Config) (handler slog.Handler, closer *Closer, err error) {

	handlers := make([]slog.Handler, 0)

	files := make([]*os.File, 0)

	defer func() {
		if err != nil {
			for _, file := range files {
				file.Close()
			}
		}
	}()

	if cfg.Stdout.Enabled {
		handlers = append(handlers, newStdoutHandler(cfg.Stdout))
	}

	for _, fileConfig := range cfg.Files {
		if fileConfig.Enabled {

			file, err := setupLogFile(cfg.Directory, fileConfig.Name, fileConfig.Level)
			if err != nil {
				return nil, nil, err
			}
			files = append(files, file)
			handlers = append(handlers, newFileHandler(fileConfig, file))
		}
	}

	handler = slog.NewMultiHandler(handlers...)

	return handler, &Closer{
		files: files,
	}, nil

}

// setupLogFile подготавливает файл для записи логов на диске.
//
// Функция:
// 1. Создает всю цепочку директорий (os.MkdirAll), если она еще не существует.
// 2. Генерирует уникальное имя файла по шаблону: YYYY-MM-DDTHH-MM-SS.MMMMMM-LEVEL-NAME.log в формате UTC.
// 3. Открывает файл в режиме "только запись", "создать, если нет" и "дописывать в конец" (append).
func setupLogFile(dir, name string, level Level) (*os.File, error) {

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to make log dir: %w", err)
	}
	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000000")

	logFilePath := filepath.Join(dir, fmt.Sprintf("%s-%s-%s.log", timestamp, level, name))

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open app log file: %w", err)
	}

	return logFile, nil
}

// Close последовательно закрывает все файлы логов, удерживаемые структурой Closer.
//
// Метод безопасен для многократного вызова, благодаря sync.Once.
// Если один или несколько файлов не удалось закрыть, метод не прерывает цикл,
// а собирает все возникшие ошибки вместе и возвращает их общим списком в конце.
func (c *Closer) Close() error {

	c.once.Do(func() {
		for _, file := range c.files {
			if err := file.Close(); err != nil {
				c.errs = errors.Join(c.errs, fmt.Errorf("failed to close file with name %s: %w", file.Name(), err))
			}
		}
	})

	return c.errs
}
