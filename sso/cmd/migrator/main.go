package main

import (
	"errors"
	"flag"
	"fmt"
	"sso/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // важно!
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	//var storagePath, migrationPath, migratorTable string
	var migrationPath, migratorTable string

	flag.StringVar(&migrationPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migratorTable, "migrations-table", "migrations", "name of migrations table")
	flag.Parse()

	pgconf := config.LoadPostgresConfig()

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&x-migrations-table=%s",
		pgconf.User,
		pgconf.Password,
		pgconf.Host,
		pgconf.Port,
		pgconf.DBName,
		pgconf.SSLMode,
		migratorTable,
	)

	/*if storagePath == "" {
		panic("storage-path is required")
	}*/
	if migrationPath == "" {
		panic("migrations-path is required")
	}

	m, err := migrate.New(
		"file://"+migrationPath,
		dbURL,
	)

	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")

			return
		}

		panic(err)
	}

	fmt.Println("migrations applied successfully")

}
