package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type postgres struct {
	db DB
}

func NewPG(ctx context.Context, connString string) (*postgres, error) {
	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Error().Err(err).Msg("Unable to connect to database")
	}

	return &postgres{db}, nil
}

func (pg *postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *postgres) Close() {
	pg.db.Close()
}

func GetConnectionUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
}
