package main

import (
	"errors"
	"flag"
	"fmt"

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

	dbURL := fmt.Sprintf(
		"postgres://admin:123@localhost:5432/mydb?sslmode=disable&x-migrations-table=%s",
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
