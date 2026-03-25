package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s is required", key)
	}
	return val
}

func main() {
	migrationPath := mustEnv("MIGRATIONS_PATH")
	pgUser := mustEnv("PG_USER")
	pgPass := os.Getenv("PG_PASS") // пароль можно пустой
	pgHost := mustEnv("PG_HOST")
	pgPort := mustEnv("PG_PORT")
	pgDB := mustEnv("PG_DB")

	pgSSL := os.Getenv("PG_SSL")
	if pgSSL == "" {
		pgSSL = "disable"
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		pgUser,
		url.QueryEscape(pgPass),
		pgHost,
		pgPort,
		pgDB,
		pgSSL,
	)

	m, err := migrate.New(
		"file://"+migrationPath,
		dsn,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("all migrations are up to date")
			return
		}
		log.Fatal(err)
	}

	log.Println("migrations applied successfully")
}
