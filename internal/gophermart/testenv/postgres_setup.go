package testenv

import (
	"context"
	"database/sql"
	"embed"
	"log"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	pg "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type Environment struct {
	Pool        *pgxpool.Pool
	db          *sql.DB
	pgContainer *pg.PostgresContainer
}

func Setup(ctx context.Context, migrationsFS embed.FS) *Environment {

	pgContainer, err := pg.Run(
		ctx,
		"postgres:18.6-alpine",
		pg.WithDatabase("db"),
		pg.WithUsername("postgres"),
		pg.WithPassword("postgres"),
		pg.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("container run: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("pg container connection string: %v", err)
	}

	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("parse pool config: %v", err)
	}

	poolCfg.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		pgxdecimal.Register(c.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		log.Fatalf("pgxpool new with config: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("pool ping: %v", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("sql open: %v", err)
	}

	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose set dialect: %v", err)
	}

	if err := goose.Up(db, "."); err != nil {
		pool.Close()
		log.Fatalf("goose up: %v", err)
	}

	return &Environment{
		Pool:        pool,
		db:          db,
		pgContainer: pgContainer,
	}
}

func (env *Environment) Close(ctx context.Context) {
	if env.Pool != nil {
		env.Pool.Close()
	}
	if env.db != nil {
		_ = env.db.Close()
	}
	if env.pgContainer != nil {
		_ = env.pgContainer.Terminate(ctx)
	}
}
