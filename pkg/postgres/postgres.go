// Package postgres предоставляет инструменты для настройки и работы
// с реляционной базой данных PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// TxBeginner это обёртка над pgxpool.Pool, которая предоставляет метод BeginFunc для начала транзакций.
type TxBeginner struct {
	pool *pgxpool.Pool
}

// NewTxBeginner создает новый экземпляр TxBeginner с указанным пулом соединений.
func NewTxBeginner(pool *pgxpool.Pool) *TxBeginner {
	return &TxBeginner{pool: pool}
}

// BeginFunc выполняет функцию fn в контексте транзакции. Если fn возвращает ошибку, транзакция откатывается, иначе коммитится.
func (t *TxBeginner) BeginFunc(ctx context.Context, fn func(pgx.Tx) error) error {
	return pgx.BeginFunc(ctx, t.pool, fn)
}

// New создает новый пул соединений с базой данных PostgreSQL и выполняет миграции.
func New(cfg Config, log *slog.Logger) (*pgxpool.Pool, error) {

	log.Info(
		"connecting to database",
	)

	poolCfg, err := pgxpool.ParseConfig(cfg.URI)
	if err != nil {
		log.Error(
			"failed to parse conn string",
			"err", err,
		)

		return nil, fmt.Errorf("parse config: %w", err)
	}

	poolCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		log.Error(
			"failed to initialize conntection pool",
			"err", err,
		)

		return nil, fmt.Errorf("new with config: %w", err)
	}

	return pool, nil
}

// SetupMigrations выполняет миграции базы данных с использованием встроенной файловой системы.
//
// gooseTableName позволяет использовать отдельную таблицу версий миграций
// для каждого сервиса при работе с одной базой данных.
func SetupMigrations(
	cfg Config,
	log *slog.Logger,
	migrationsFS embed.FS,
	gooseTableName string,
) error {

	log.Info(
		"running db migrations",
	)

	db, err := sql.Open("pgx", cfg.URI)
	if err != nil {
		log.Error(
			"failed to open sql db",
			"err", err,
		)

		return fmt.Errorf("open db for migrations: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrationsFS)
	goose.SetTableName(gooseTableName)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Error(
			"failed to set dialect",
			"err", err,
		)

		return fmt.Errorf("goose set dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	log.Info(
		"database migrations completed successfully",
	)

	return nil
}
