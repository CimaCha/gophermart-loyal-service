// Package migrations содержит SQL-скрипты миграций базы данных
// и предоставляет встроенный доступ к ним во время компиляции.
package migrations

import "embed"

// EmbedMigrations содержит встроенную (embedded) файловую систему.
// Используется для автоматического применения миграций при старте сервиса accrual.
//go:embed *.sql
var EmbedMigrations embed.FS
