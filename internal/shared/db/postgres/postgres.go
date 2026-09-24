package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/CimaCha/gophermart-loyal-service/migrations"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

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

	// Здесь подтягиваем миграции при старте БД.

	log.Info(
		"running db migrations",
	)

	db, err := sql.Open("pgx", cfg.URI)
	if err != nil {
		log.Error(
			"faild to open sql db",
			"err", err,
		)
		pool.Close()
		return nil, fmt.Errorf("open db for migrations: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Error(
			"failed to set dialect",
			"err", err,
		)
		return nil, fmt.Errorf("goose set dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		pool.Close()
		return nil, fmt.Errorf("goose up: %w", err)
	}

	log.Info(
		"database initialized successfully",
	)

	return pool, nil
}
