package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"

	"identity-forecaster/internal/pkg/logger"
)

type postgres struct {
	*pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *postgres {
	return &postgres{pool}
}

func applyMigrations(DSN string) error {
	pg, err := sql.Open("pgx", DSN)
	if err != nil {
		return err
	}
	err = goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	err = pg.PingContext(context.Background())
	if err != nil {
		return err
	}

	err = goose.Run("up", pg, "./internal/app/forecaster/repository/migrations")
	if err != nil {
		return err
	}
	err = pg.Close()
	if err != nil {
		return err
	}

	return nil
}

func GetPgxPool(DSN string) (*pgxpool.Pool, error) {
	err := applyMigrations(DSN)
	if err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		logger.Logger().Infoln(err)
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.Logger().Infoln(err)
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (p *postgres) WithTransaction(ctx context.Context, txFunc func(context.Context, pgx.Tx) error) error {
	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback(ctx)
		if !errors.Is(err, pgx.ErrTxClosed) && err != nil {
			logger.Logger().Errorln(err)
		}
	}()

	err = txFunc(ctx, tx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (p *postgres) WithConnection(ctx context.Context, connFunc func(context.Context, *pgxpool.Conn) error) error {
	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return connFunc(ctx, conn)
}
