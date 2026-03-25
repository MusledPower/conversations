package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	DB *sql.DB
}

func New(storageURL string) (*Storage, error) {
	const op = "storage.pg.New"

	db, err := sql.Open("pgx", storageURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{DB: db}, nil
}
