package psq

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Riter/E-Shop/internal/config"
	_ "github.com/lib/pq"
)

type Postgres struct {
    DB *sql.DB
}

func NewPostgres(ctx context.Context, cfg config.PostgresConfig) (*Postgres, error) {
    db, err := sql.Open("postgres", cfg.DSN())
    if err != nil {
        return nil, fmt.Errorf("failed to open postgres: %w", err)
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(time.Hour)

    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping postgres: %w", err)
    }

    return &Postgres{DB: db}, nil
}


func (p *Postgres) GetStorage() *sql.DB{
	return p.DB
}
