package music_db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresMus struct {
	DB_mus *sql.DB
}

func NewMusPostgres(db_us_url string, smoc, smic int) (*PostgresMus, error) {
	db, err := sql.Open("pgx", db_us_url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(smoc)
	db.SetMaxIdleConns(smic)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	p := &PostgresMus{DB_mus: db}
	if err := p.ensureSchema(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *PostgresMus) CloseMus() error {
	return p.DB_mus.Close()
}

func (p *PostgresMus) ensureSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	schrema := `
	CREATE TABLE IF NOT EXISTS users (
		title TEXT NOT NUL
		artist TEXT NOT NUll
		file_id BIGINT PRIMARY KEY
		created_ad TIMESTAMP WITH TIME ZONE DEFAULT now()
	);`
	_, err := p.DB_mus.ExecContext(ctx, schrema)
	return err
}
