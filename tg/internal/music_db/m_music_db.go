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
	if err := p.ensureMusSchema(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *PostgresMus) CloseMus() error {
	return p.DB_mus.Close()
}

func (p *PostgresMus) ensureMusSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	schrema := `
	CREATE TABLE IF NOT EXISTS tracks (
    	file_unique_id TEXT PRIMARY KEY,
    	file_id TEXT NOT NULL,
    	title TEXT NOT NULL,
    	artist TEXT NOT NULL,
    	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    	UNIQUE (title, artist)
	);

	CREATE UNIQUE INDEX IF NOT EXISTS uniq_track
	ON tracks (LOWER(title), LOWER(artist));
	`

	_, err := p.DB_mus.ExecContext(ctx, schrema)
	return err
}

func (p *PostgresMus) InsertMusic(ctx context.Context, title, arctist, file_id, file_un_id string) error {
	_, err := p.DB_mus.ExecContext(
		ctx,
		"INSERT INTO tracks (file_unique_id, file_id, title, artist) VALUES ($1,$2,$3,$4)",
		file_un_id,
		file_id,
		title,
		arctist,
	)
	return err
}

func (p *PostgresMus) GetMusicByIndex(ctx context.Context, title, artist string) (string, string, error) {
	var (
		file_un_id string
		file_id    string
		row        *sql.Row = nil
	)
	if artist == "" {
		row = p.DB_mus.QueryRowContext(
			ctx,
			"SELECT file_unique_id, file_id FROM tracks WHERE LOWER(title) = LOWER($1)",
			title,
		)
	} else {
		row = p.DB_mus.QueryRowContext(
			ctx,
			"SELECT file_unique_id, file_id FROM tracks WHERE LOWER(title) = LOWER($1) AND LOWER(artist) ILIKE '%' || $2 || '%';",
			title,
			artist,
		)
	}
	if err := row.Scan(&file_un_id, &file_id); err != nil {
		return "", "", err
	}
	return file_un_id, file_id, nil
}
