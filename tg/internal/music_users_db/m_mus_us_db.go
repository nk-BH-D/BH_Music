package music_db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	pq "github.com/lib/pq"
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

type PostgresUs struct {
	DB_us *sql.DB
}

func NewUsPostgres(db_us_url string, smoc, smic int) (*PostgresUs, error) {
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

	p := &PostgresUs{DB_us: db}
	if err := p.ensureUsSchema(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *PostgresUs) CloseUs() error {
	return p.DB_us.Close()
}

func (p *PostgresUs) ensureUsSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	schrema := `
	CREATE TABLE IF NOT EXISTS users (
		user_id BIGINT PRIMARY KEY,
		chat_id BIGINT,
		accept BOOLEAN NOT NULL DEFAULT FALSE,
		status TEXT NOT NULL,
		file_id_list TEXT[] NOT NULL DEFAULT'{}', 
		created_ad TIMESTAMP WITH TIME ZONE DEFAULT now()
	);`
	_, err := p.DB_us.ExecContext(ctx, schrema)
	return err
}

func (p *PostgresUs) InsertUser(ctx context.Context, user_id, chat_id int64, status string) error {
	_, err := p.DB_us.ExecContext(
		ctx,
		"INSERT INTO users (user_id, chat_id, accept, status) VALUES ($1,$2,$3,$4)",
		user_id,
		chat_id,
		false,
		status,
	)
	return err
}

func (p *PostgresUs) GetUserStatus(ctx context.Context, user_id int64) (bool, string, error) {
	var (
		accept bool
		status string
	)
	row := p.DB_us.QueryRowContext(
		ctx,
		"SELECT accept, status FROM users WHERE user_id=$1",
		user_id,
	)
	if err := row.Scan(&accept, &status); err != nil {
		return false, "", err
	}
	return accept, status, nil
}

func (p *PostgresUs) UpUserAccept(ctx context.Context, user_id int64, accept bool) error {
	_, err := p.DB_us.ExecContext(
		ctx,
		"UPDATE users SET accept=$2 WHERE user_id=$1",
		user_id,
		accept,
	)
	return err
}

func (p *PostgresUs) UpdateFileIDList(ctx context.Context, user_id int64, file_id string) error {
	_, err := p.DB_us.ExecContext(
		ctx,
		"UPDATE users SET file_id_list = array_append(file_id_list, $1) WHERE user_id = $2",
		file_id, user_id,
	)
	return err
}

func (p *PostgresUs) GetFileIDList(ctx context.Context, user_id int64) ([]string, error) {
	var fileIDL []string
	row := p.DB_us.QueryRowContext(
		ctx,
		"SELECT file_id_list FROM users WHERE user_id = $1",
		user_id,
	)
	if err := row.Scan(pq.Array(&fileIDL)); err != nil {
		return nil, err
	}
	return fileIDL, nil
}
